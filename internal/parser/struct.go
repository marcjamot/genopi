package parser

import (
	"fmt"
	"genopi/internal/common"
	"go/ast"
	"go/types"
	"log"
	"regexp"
)

var (
	basicTypes = make(map[string]struct{})
	jsonRegex  = regexp.MustCompile(`json:"([^"]+)"`)
)

func init() {
	for _, t := range types.Typ {
		basicTypes[t.Name()] = struct{}{}
	}
}

func readStruct(f *ast.File, x1 *ast.GenDecl) (common.Struct, bool) {
	var pack string
	var name string
	fields := make([]common.Field, 0)

	for _, spec := range x1.Specs {
		if x2, ok := spec.(*ast.TypeSpec); ok {
			if st, ok := x2.Type.(*ast.StructType); ok {
				pack = f.Name.Name
				name = x2.Name.Name

				for _, field := range st.Fields.List {
					switch x3 := field.Type.(type) {
					case *ast.Ident:
						fields = append(fields, readIdent(pack, field.Names[0].Name, field.Tag, x3))
					case *ast.StarExpr:
						if field, ok := readStar(pack, field.Names[0].Name, field.Tag, x3); ok {
							fields = append(fields, field)
						}
					case *ast.ArrayType:
						if field, ok := readArray(pack, field.Names[0].Name, field.Tag, x3); ok {
							fields = append(fields, field)
						}
					case *ast.MapType:
						// Does not handle maps
					}
				}

				return common.Struct{
					Package: pack,
					Name:    name,
					Fields:  fields,
				}, true
			}
		}
	}

	return common.Struct{}, false
}

func readIdent(pack, name string, lit *ast.BasicLit, ident *ast.Ident) common.Field {
	if lit != nil && jsonRegex.MatchString(lit.Value) {
		name = jsonRegex.FindStringSubmatch(lit.Value)[1]
	}

	typ := ident.Name
	if _, ok := basicTypes[ident.Name]; !ok {
		typ = fmt.Sprintf("%s.%s", pack, typ)
	}

	return common.Field{
		Name: name,
		Type: typ,
	}
}

func readStar(pack, name string, lit *ast.BasicLit, star *ast.StarExpr) (common.Field, bool) {
	var field common.Field
	if ident, ok := star.X.(*ast.Ident); ok {
		field = readIdent(pack, name, lit, ident)
	} else {
		log.Printf("skipping %s.%s: cannot handle pointers to anything but primitives or structs", pack, name)
		return common.Field{}, false
	}

	field.Optional = true
	return field, true
}

func readArray(pack, name string, lit *ast.BasicLit, arr *ast.ArrayType) (common.Field, bool) {
	var field common.Field
	if ident, ok := arr.Elt.(*ast.Ident); ok {
		field = readIdent(pack, name, lit, ident)
	} else {
		log.Printf("skipping %s.%s: cannot handle array of anything but primitives or structs", pack, name)
		return common.Field{}, false
	}

	field.Array = true
	return field, true
}
