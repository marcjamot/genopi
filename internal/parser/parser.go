package parser

import (
	"bufio"
	"fmt"
	"genopi/internal/common"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var functionRegex = regexp.MustCompile(`func (.+)\(.+http.ResponseWriter,.+\*http.Request\)`)

type Method struct {
	Comments []string
	Text     string
}

func FromPath(dir string) ([]common.Endpoint, error) {
	paths, err := getPaths(dir)
	if err != nil {
		return nil, err
	}

	methods := make([]Method, 0)
	for _, path := range paths {
		m, err := readMethods(path)
		if err != nil {
			return nil, err
		}
		methods = append(methods, m...)
	}

	endpoints := make([]common.Endpoint, 0)
	for _, method := range methods {
		e, err := parseEndpoint(method)
		if err != nil {
			return nil, err
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

func readMethods(path string) ([]Method, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	methods := make([]Method, 0)
	comments := make([]string, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		text := scanner.Text()

		if strings.HasPrefix(text, "// ") {
			comments = append(comments, text[3:])
		} else if functionRegex.MatchString(text) {
			methods = append(methods, Method{
				Comments: comments,
				Text:     text,
			})
			comments = make([]string, 0)
		} else if len(comments) > 0 {
			comments = make([]string, 0)
		}
	}

	return methods, nil
}

func parseEndpoint(method Method) (common.Endpoint, error) {
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
		} else if i == 0 && endpoint.Name == "" {
			// Name is checked last to not accidentally match something else
			endpoint.Name = strings.TrimSpace(c)
		}
	}

	if err := verify(endpoint); err != nil {
		return common.Endpoint{}, err
	}

	log.Printf("Endpoint: %v", endpoint)
	return endpoint, nil
}

// verify that we have at least the minimal required info
func verify(endpoint common.Endpoint) error {
	if endpoint.Name == "" {
		return fmt.Errorf("%s missing name", endpoint.Name)
	}
	if endpoint.Method == "" {
		return fmt.Errorf("%s missing method", endpoint.Name)
	}
	if endpoint.Path == "" {
		return fmt.Errorf("%s missing path", endpoint.Name)
	}
	return nil
}
