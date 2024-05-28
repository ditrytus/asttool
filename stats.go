package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

type statsAstVisitor struct {
	funcCount  int
	typeCount  int
	constCount int
	varCount   int
}

func (v *statsAstVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch n := node.(type) {
	case *ast.FuncDecl:
		v.funcCount++
	case *ast.TypeSpec:
		v.typeCount++
	case *ast.GenDecl:
		switch n.Tok {
		case token.CONST:
			v.constCount++
		case token.VAR:
			v.varCount++
		}
	}
	return v
}

func FormatStatsVisitor(v *statsAstVisitor) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Funcs: %d\n", v.funcCount))
	b.WriteString(fmt.Sprintf("Types: %d\n", v.typeCount))
	b.WriteString(fmt.Sprintf("Consts: %d\n", v.constCount))
	b.WriteString(fmt.Sprintf("Vars: %d\n", v.varCount))
	return b.String()
}
