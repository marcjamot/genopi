package parser

import (
	"go/ast"
)

func readMethod(f *ast.File, x1 *ast.FuncDecl) (Method, bool) {
	if x1.Doc != nil {
		comments := make([]string, 0)
		for _, d := range x1.Doc.List {
			comments = append(comments, d.Text[3:])
		}
		return Method{
			Package:  f.Name.Name,
			Name:     x1.Name.Name,
			Comments: comments,
		}, true
	}

	return Method{}, false
}
