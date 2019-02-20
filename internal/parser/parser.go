package parser

import (
	"errors"
	"fmt"
	"genopi/internal/common"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var functionRegex = regexp.MustCompile(`func (.+)\(.+http.ResponseWriter,.+\*http.Request\)`)

type Method struct {
	Package  string
	Name     string
	Comments []string
}

type Struct struct {
	Package string
	Name    string
	Fields  []Field
}

type Field struct {
	Name     string
	Type     string
	Optional bool
	Array    bool
}

func FromPath(dir string) ([]common.Endpoint, error) {
	paths, err := getPaths(dir)
	if err != nil {
		return nil, err
	}

	methods := make([]Method, 0)
	structs := make(map[string]Struct, 0)
	for _, path := range paths {
		m, s, err := readFileContent(path)
		if err != nil {
			return nil, err
		}
		methods = append(methods, m...)
		for k, v := range s {
			structs[k] = v
		}
	}

	endpoints := make([]common.Endpoint, 0)
	for _, method := range methods {
		e, err := parseEndpoint(method, structs)
		if err != nil {
			log.Printf("Skipping %s.%s: %v", method.Package, method.Name, err)
			continue
		}
		endpoints = append(endpoints, e)
	}

	return endpoints, nil
}

func getPaths(dir string) ([]string, error) {
	paths := make([]string, 0)
	err := filepath.Walk(dir, func(p string, i os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !i.IsDir() && strings.HasSuffix(i.Name(), ".go") {
			paths = append(paths, p)
		}
		return nil
	})
	return paths, err
}

func readFileContent(path string) ([]Method, map[string]Struct, error) {
	methods := make([]Method, 0)
	structs := make(map[string]Struct, 0)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	for _, decl := range f.Decls {
		switch x1 := decl.(type) {
		case *ast.GenDecl:
			if s, ok := readStruct(f, x1); ok {
				structs[fmt.Sprintf("%s.%s", s.Package, s.Name)] = s
			}
		case *ast.FuncDecl:
			if m, ok := readMethod(f, x1); ok {
				methods = append(methods, m)
			}
		}
	}

	return methods, structs, nil
}

func parseEndpoint(method Method, structs map[string]Struct) (common.Endpoint, error) {
	endpoint := common.Endpoint{
		Name:        "",
		Method:      "",
		Path:        "",
		PathParams:  make(map[string]common.Param),
		QueryParams: make(map[string]common.Param),
		Headers:     make(map[string]common.Param),
		Body:        nil,
		Responses:   nil,
	}

	for i, c := range method.Comments {
		if m, p, ok := tryMethod(c); ok {
			endpoint.Method = m
			endpoint.Path = p
			continue
		} else if k, v, ok := tryParam(c, "{", "}"); ok {
			endpoint.PathParams[k] = v
		} else if k, v, ok := tryParam(c, "(", ")"); ok {
			endpoint.QueryParams[k] = v
		} else if k, v, ok := tryParam(c, "[", "]"); ok {
			endpoint.Headers[k] = v
		} else if b, ok := tryBody(c); ok {
			endpoint.Body = &b
		} else if i == 0 && endpoint.Name == "" {
			// Name is checked last to not accidentally match something else
			endpoint.Name = strings.TrimSpace(c)
		}
	}

	if err := verify(endpoint, structs); err != nil {
		return common.Endpoint{}, err
	}

	log.Printf("Endpoint: %v", endpoint)
	return endpoint, nil
}

// verify that we have at least the minimal required info
func verify(endpoint common.Endpoint, structs map[string]Struct) error {
	if endpoint.Name == "" {
		return errors.New("missing name")
	}
	if endpoint.Method == "" {
		return errors.New("missing method")
	}
	if endpoint.Path == "" {
		return errors.New("missing path")
	}

	for k := range endpoint.PathParams {
		if !strings.Contains(endpoint.Path, k) {
			return fmt.Errorf("path param %s missing in path: %s", k, endpoint.Path)
		}
	}

	if endpoint.Body != nil {
		if _, ok := structs[*endpoint.Body]; !ok {
			return fmt.Errorf("body struct not found: %s", *endpoint.Body)
		}
	}

	return nil
}
