package stats

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"
)

type Stats struct {
	FuncCount  int
	TypeCount  int
	ConstCount int
	VarCount   int
}

type Visitor interface {
	ast.Visitor
	Stats() Stats
}

type statsAstVisitor struct {
	s Stats
}

func NewStatsVisitor() Visitor {
	return &statsAstVisitor{}
}

func (v *statsAstVisitor) Stats() Stats {
	return v.s
}

func (v *statsAstVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch n := node.(type) {
	case *ast.FuncDecl:
		v.s.FuncCount++
	case *ast.TypeSpec:
		v.s.TypeCount++
	case *ast.GenDecl:
		switch n.Tok {
		case token.CONST:
			v.s.ConstCount++
		case token.VAR:
			v.s.VarCount++
		default:
			return v
		}
	}
	return v
}

func FormatStatsVisitor(v Visitor) string {
	s := v.Stats()
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Funcs: %d\n", s.FuncCount))
	b.WriteString(fmt.Sprintf("Types: %d\n", s.TypeCount))
	b.WriteString(fmt.Sprintf("Consts: %d\n", s.ConstCount))
	b.WriteString(fmt.Sprintf("Vars: %d\n", s.VarCount))
	return b.String()
}
