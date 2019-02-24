package parser

import (
	"go/ast"

	"github.com/plankt/genopi/internal/common"
)

func readMethod(f *ast.File, x1 *ast.FuncDecl) (common.Method, bool) {
	if x1.Doc != nil {
		comments := make([]string, 0)
		for _, d := range x1.Doc.List {
			if len(d.Text) >= 3 {
				comments = append(comments, d.Text[3:])
			}
		}
		return common.Method{
			Package:  f.Name.Name,
			Name:     x1.Name.Name,
			Comments: comments,
		}, true
	}

	return common.Method{}, false
}
