package generator

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/plankt/genopi/internal/common"
)

type generator struct {
	b *bytes.Buffer
}

func (g *generator) WriteString(indentation int, ss ...string) {
	for i := 0; i < indentation; i = i + 1 {
		g.b.WriteString("  ")
	}

	for _, s := range ss {
		g.b.WriteString(s)
	}
	g.b.WriteRune('\n')
}

func (g *generator) String() string {
	return g.b.String()
}

func OpenAPI3(api common.Api) error {
	g := &generator{b: &bytes.Buffer{}}

	g.WriteString(0, `openapi: "3.0.0"`)
	g.WriteString(0, "info:")
	g.WriteString(1, "title: ", api.Status.Title)
	g.WriteString(1, "version: \"", api.Status.Version, "\"")
	g.WriteString(0, "servers:")
	g.WriteString(1, "- url: ", api.Status.URL)

	paths(api, g)
	components(api, g)

	return ioutil.WriteFile(api.Status.Output, []byte(g.String()), 0644)
}

func paths(api common.Api, g *generator) {
	endpoints := make(map[string][]common.Endpoint)
	for _, e := range api.Endpoints {
		es, ok := endpoints[e.Path]
		if !ok {
			es = make([]common.Endpoint, 0)
		}
		es = append(es, e)
		endpoints[e.Path] = es
	}

	paths := make([]string, 0)
	for k := range endpoints {
		paths = append(paths, k)
	}
	sort.Slice(paths, func(i, j int) bool {
		return strings.Compare(paths[i], paths[j]) <= 0
	})

	g.WriteString(0, "paths:")
	for _, k := range paths {
		v := endpoints[k]
		sort.Slice(v, func(i, j int) bool {
			return methodOrder(v[i].Method) <= methodOrder(v[j].Method)
		})

		g.WriteString(1, k, ":")

		for _, e := range v {
			g.WriteString(2, fmt.Sprintf("%s:", strings.ToLower(e.Method)))
			g.WriteString(3, "summary: ", e.Name)
			if len(e.PathParams) > 0 || len(e.Headers) > 0 || len(e.QueryParams) > 0 {
				g.WriteString(3, "parameters:")
				for k, v := range e.PathParams {
					parameter(g, "path", k, v)
				}
				for k, v := range e.Headers {
					parameter(g, "header", k, v)
				}
				for k, v := range e.QueryParams {
					parameter(g, "query", k, v)
				}
			}

			if e.Body != nil {
				if _, ok := api.Structs[*e.Body]; ok {
					g.WriteString(3, "requestBody:")
					g.WriteString(4, "required: true")
					g.WriteString(4, "content:")
					g.WriteString(5, "application/json:")
					g.WriteString(6, "schema:")
					g.WriteString(7, fmt.Sprintf("$ref: '#/components/schemas/%s'", *e.Body))
				} else {
					log.Printf("Body: could not find the struct: %s", *e.Body)
				}
			}

			g.WriteString(3, "responses:")
			for _, r := range e.Responses {
				g.WriteString(4, fmt.Sprintf("'%d':", r.Code))
				g.WriteString(5, "description: ", http.StatusText(r.Code))
				if r.Type != nil {
					if _, ok := api.Structs[*r.Type]; ok {
						g.WriteString(5, "content:")
						g.WriteString(6, "application/json:")
						g.WriteString(7, "schema:")
						g.WriteString(8, fmt.Sprintf("$ref: '#/components/schemas/%s'", *r.Type))
					} else {
						log.Printf("%d: could not find the struct: %s", r.Code, *r.Type)
					}
				}
			}
		}
	}
}

func methodOrder(method string) int {
	switch method {
	case "get":
		return 0
	case "post":
		return 1
	case "put":
		return 2
	case "delete":
		return 3
	default:
		return 4
	}
}

func parameter(g *generator, typ string, name string, param common.Param) {
	g.WriteString(4, "- name: ", name)
	g.WriteString(4, "  in: ", typ)
	g.WriteString(4, "  description: ", param.Desc)
	g.WriteString(4, "  required: ", strconv.FormatBool(param.Required))
	g.WriteString(4, "  schema:")
	g.WriteString(5, "  type: ", param.Type)
}

func components(api common.Api, g *generator) {
	g.WriteString(0, "components:")
	g.WriteString(1, "schemas:")

	sorted := make([]common.Struct, 0)
	for _, s := range api.Structs {
		sorted = append(sorted, s)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return strings.Compare(sorted[i].Name, sorted[j].Name) <= 0
	})

	for _, s := range sorted {
		g.WriteString(2, s.FullName(), ":")
		g.WriteString(3, "type: object")
		g.WriteString(3, "properties:")
		for _, f := range s.Fields {
			g.WriteString(4, f.Name, ":")
			if f.Array {
				g.WriteString(5, "type: array")
				g.WriteString(5, "items:")
				writeComponentType(g, 6, f.Type)
			} else {
				writeComponentType(g, 5, f.Type)
			}
		}
		var required bool
		for _, f := range s.Fields {
			if !f.Optional {
				required = true
				break
			}
		}
		if required {
			g.WriteString(3, "required:")
			for _, f := range s.Fields {
				if !f.Optional {
					g.WriteString(4, "- ", f.Name)
				}
			}
		}
	}
}

func writeComponentType(g *generator, indentation int, typ string) {
	if typ == "uuid.UUID" {
		g.WriteString(indentation, "type: string")
		g.WriteString(indentation, "format: uuid")
	} else if typ == "time.Time" {
		g.WriteString(indentation, "type: string")
		g.WriteString(indentation, "format: date")
	} else if strings.Contains(typ, ".") {
		g.WriteString(indentation, "$ref: '#/components/schemas/", typ, "'")
	} else {
		g.WriteString(indentation, "type: ", typ)
	}
}
