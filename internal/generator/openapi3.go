package generator

import (
	"bytes"
	"fmt"
	"genopi/internal/common"
	"io/ioutil"
	"strings"
)

type generator struct {
	b *bytes.Buffer
}

func (g *generator) Indentation(indentation int) {
	for i := 0; i < indentation; i = i + 1 {
		g.b.WriteString("  ")
	}
}

func (g *generator) NewLine() {
	g.b.WriteRune('\n')
}

func (g *generator) WriteString(indentation int, ss ...string) {
	g.Indentation(indentation)
	for _, s := range ss {
		g.b.WriteString(s)
	}
	g.NewLine()
}

func (g *generator) String() string {
	return g.b.String()
}

func OpenAPI3(api common.Api) error {
	g := &generator{b: &bytes.Buffer{}}

	g.WriteString(0, `openapi: "3.0.0"`)
	g.WriteString(0, "info:")
	g.WriteString(1, "title: ", api.Status.Title)
	g.WriteString(1, "version: ", api.Status.Version)
	g.WriteString(0, "servers:")
	g.WriteString(1, "- url: ", api.Status.URL)

	paths(api, g)
	// components(router, g)

	return ioutil.WriteFile("api.yaml", []byte(g.String()), 0644)
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

	g.WriteString(0, "paths:")
	for k, v := range endpoints {
		g.WriteString(1, k, ":")

		for _, e := range v {
			handler(g, e)
		}
	}
}

func handler(g *generator, e common.Endpoint) {
	g.WriteString(2, fmt.Sprintf("%s:", strings.ToLower(e.Method)))
	g.WriteString(3, "summary: ", e.Name)
	g.WriteString(3, "responses:")
}
