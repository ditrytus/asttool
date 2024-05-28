package stats

import (
	"go/ast"
	"go/token"
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
