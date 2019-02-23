package generator

import (
	"bytes"
	"fmt"
	"genopi/internal/common"
	"io/ioutil"
	"log"
	"sort"
	"strconv"
	"strings"
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

	return ioutil.WriteFile("api.yaml", []byte(g.String()), 0644)
}

func paths(api common.Api, g *generator) {
	structs := make(map[string]common.Struct)
	for _, s := range api.Structs {
		structs[fmt.Sprintf("%s.%s", s.Package, s.Name)] = s
	}

	endpoints := make(map[string][]common.Endpoint)
	for _, e := range api.Endpoints {
		es, ok := endpoints[e.Path]
		if !ok {
			es = make([]common.Endpoint, 0)
		}
		es = append(es, e)
		endpoints[e.Path] = es
	}

	g.WriteString(0, "paths:")
	for k, v := range endpoints {
		g.WriteString(1, k, ":")

		for _, e := range v {
			g.WriteString(2, fmt.Sprintf("%s:", strings.ToLower(e.Method)))
			g.WriteString(3, "summary: ", e.Name)
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

			if e.Body != nil {
				if _, ok := structs[*e.Body]; ok {
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
				if r.Type != nil {
					if _, ok := structs[*r.Type]; ok {
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

	sort.Slice(api.Structs, func(i, j int) bool {
		return strings.Compare(api.Structs[i].Name, api.Structs[j].Name) <= 0
	})

	for _, s := range api.Structs {
		name := fmt.Sprintf("%s.%s", s.Package, s.Name)
		g.WriteString(2, name, ":")
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
		g.WriteString(3, "required:")
		for _, f := range s.Fields {
			if !f.Optional {
				g.WriteString(4, "- ", f.Name)
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
