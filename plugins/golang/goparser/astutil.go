package goparser

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/blueprint-uservices/blueprint/blueprint/pkg/blueprint/stringutil"
)

// Returns a string representation of an expr and its internals
// Useful for debugging golang code parsers.
func ExprStr(e ast.Expr) string {
	if e == nil {
		return "nil"
	}
	switch v := e.(type) {
	case *ast.BadExpr:
		return "*ast.BadExpr{}"
	case *ast.Ident:
		return fmt.Sprintf("*ast.Ident{Name: \"%v\"}", v.Name)
	case *ast.Ellipsis:
		return fmt.Sprintf("*ast.Ellipsis{\n%v\n}", exprField("Elt", v.Elt))
	case *ast.BasicLit:
		return fmt.Sprintf("*ast.BasicLit{\n  Kind: %v\n  Value: %v\n}", v.Kind, v.Value)
	case *ast.FuncLit:
		return "*ast.FuncLit{...}"
	case *ast.CompositeLit:
		return fmt.Sprintf("*ast.CompositeLit{\n%v\n%v\n}", exprField("Type", v.Type), exprFields("Elts", v.Elts))
	case *ast.ParenExpr:
		return fmt.Sprintf("*ast.ParenExpr{\n%v\n}", exprField("X", v.X))
	case *ast.SelectorExpr:
		return fmt.Sprintf("*ast.SelectorExpr{\n%v\n%v\n}", exprField("X", v.X), exprField("Sel", v.Sel))
	case *ast.IndexExpr:
		return fmt.Sprintf("*ast.IndexExpr{\n%v\n%v\n}", exprField("X", v.X), exprField("Index", v.Index))
	case *ast.IndexListExpr:
		return fmt.Sprintf("*ast.IndexListExpr{\n%v\n%v\n}", exprField("X", v.X), exprFields("Indices", v.Indices))
	case *ast.SliceExpr:
		return "*ast.SliceExpr{...}"
	case *ast.TypeAssertExpr:
		return "*ast.TypeAssertExpr{...}"
	case *ast.CallExpr:
		return fmt.Sprintf("*ast.CallExpr{\n%v\n}", exprField("Fun", v.Fun))
	case *ast.StarExpr:
		return "*ast.StarExpr{...}"
	case *ast.UnaryExpr:
		return "*ast.UnaryExpr{...}"
	case *ast.BinaryExpr:
		return "*ast.BinaryExpr{...}"
	case *ast.KeyValueExpr:
		return "*ast.KeyValueExpr{...}"
	}
	return "/* unknown expr */"
}

func exprField(name string, e ast.Expr) string {
	return stringutil.Indent(fmt.Sprintf("%v: %v", name, ExprStr(e)), 2)
}
func exprFields(name string, es []ast.Expr) string {
	if len(es) == 0 {
		return stringutil.Indent(fmt.Sprintf("%v: []", name), 2)
	}

	b := strings.Builder{}
	b.WriteString(name + ": [")

	for _, e := range es {
		b.WriteString("\n")
		b.WriteString(stringutil.Indent(ExprStr(e), 2))
		b.WriteString(",")
	}
	b.WriteString("\n]")

	return stringutil.Indent(b.String(), 2)
}
