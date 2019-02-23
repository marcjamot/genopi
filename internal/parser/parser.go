package parser

import (
	"errors"
	"fmt"
	"genopi/internal/common"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func FromPath(dir string) ([]common.Endpoint, []common.Struct, error) {
	paths, err := getPaths(dir)
	if err != nil {
		return nil, nil, err
	}

	methods := make([]common.Method, 0)
	structs := make([]common.Struct, 0)
	for _, path := range paths {
		m, s, err := readFileContent(path)
		if err != nil {
			return nil, nil, err
		}
		methods = append(methods, m...)
		structs = append(structs, s...)
	}

	endpoints := make([]common.Endpoint, 0)
	for _, method := range methods {
		e, err := parseEndpoint(method)
		if err != nil {
			// log.Printf("Skipping %s.%s: %v", method.Package, method.Name, err)
			continue
		}
		endpoints = append(endpoints, e)
	}

	return endpoints, structs, nil
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

func readFileContent(path string) ([]common.Method, []common.Struct, error) {
	methods := make([]common.Method, 0)
	structs := make([]common.Struct, 0)

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, nil, err
	}
	for _, decl := range f.Decls {
		switch x1 := decl.(type) {
		case *ast.GenDecl:
			if s, ok := readStruct(f, x1); ok {
				structs = append(structs, s)
			}
		case *ast.FuncDecl:
			if m, ok := readMethod(f, x1); ok {
				methods = append(methods, m)
			}
		}
	}

	return methods, structs, nil
}

func parseEndpoint(method common.Method) (common.Endpoint, error) {
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
			endpoint.Body = b
		} else if code, b, ok := tryResponse(c); ok {
			endpoint.Responses = append(endpoint.Responses, common.Response{
				Code: code,
				Type: b,
			})
		} else if i == 0 && endpoint.Name == "" {
			// Name is checked last to not accidentally match something else
			endpoint.Name = strings.TrimSpace(c)
		}
	}

	if err := verify(endpoint); err != nil {
		return common.Endpoint{}, err
	}

	return endpoint, nil
}

// verify that we have at least the minimal required info
func verify(endpoint common.Endpoint) error {
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

	return nil
}
