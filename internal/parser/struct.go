package parser

import (
	"fmt"
	"go/ast"
	"go/types"
	"log"
)

var basicTypes = make(map[string]struct{})

func init() {
	for _, t := range types.Typ {
		basicTypes[t.Name()] = struct{}{}
	}
}

func readStruct(f *ast.File, x1 *ast.GenDecl) (Struct, bool) {
	var pack string
	var name string
	fields := make([]Field, 0)

	for _, spec := range x1.Specs {
		if x2, ok := spec.(*ast.TypeSpec); ok {
			if st, ok := x2.Type.(*ast.StructType); ok {
				pack = f.Name.Name
				name = x2.Name.Name

				for _, field := range st.Fields.List {
					switch x3 := field.Type.(type) {
					case *ast.Ident:
						fields = append(fields, readIdent(pack, field.Names[0].Name, x3))
					case *ast.StarExpr:
						fields = append(fields, readStar(pack, field.Names[0].Name, x3))
					case *ast.ArrayType:
						fields = append(fields, readArray(pack, field.Names[0].Name, x3))
					case *ast.MapType:
						// Does not handle maps
					}
				}

				return Struct{
					Package: pack,
					Name:    name,
					Fields:  fields,
				}, true
			}
		}
	}

	return Struct{}, false
}

func readIdent(pack, name string, ident *ast.Ident) Field {
	typ := ident.Name
	if _, ok := basicTypes[ident.Name]; !ok {
		typ = fmt.Sprintf("%s.%s", pack, typ)
	}

	return Field{
		Name: name,
		Type: typ,
	}
}

func readStar(pack, name string, star *ast.StarExpr) Field {
	var field Field
	if ident, ok := star.X.(*ast.Ident); ok {
		field = readIdent(pack, name, ident)
	} else {
		log.Fatalf("cannot handle pointers to anything but primitives or structs")
	}

	field.Optional = true
	return field
}

func readArray(pack, name string, arr *ast.ArrayType) Field {
	var field Field
	if ident, ok := arr.Elt.(*ast.Ident); ok {
		field = readIdent(pack, name, ident)
	} else {
		log.Fatalf("cannot handle array of anything but primitives or structs")
	}

	field.Array = true
	return field
}
