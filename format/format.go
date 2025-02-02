package format

import (
	"fmt"
	"go/ast"
	"strings"
)

type Visitor interface {
	ast.Visitor
	fmt.Stringer
}

type formatAstVisitor struct {
	b      *strings.Builder
	indent string
}

func NewFormatVisitor(indent string) Visitor {
	return &formatAstVisitor{
		b:      &strings.Builder{},
		indent: indent,
	}
}

func (p *formatAstVisitor) String() string {
	return p.b.String()
}

func (p *formatAstVisitor) Visit(node ast.Node) (w ast.Visitor) {
	if node == nil {
		return nil
	}
	p.b.WriteString(p.indent)
	p.b.WriteString(fmt.Sprintf("%T %s", node, shortNodeString(node)))
	p.b.WriteString("\n")
	return &formatAstVisitor{b: p.b, indent: p.indent + "  "}
}

func shortNodeString(node ast.Node) any {
	switch n := node.(type) {
	case *ast.FuncDecl:
		return n.Name
	case *ast.TypeSpec:
		return n.Name
	case *ast.Ident:
		return n.Name
	case *ast.File:
		return n.Name
	}
	return ""
}
