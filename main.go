package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/importer"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/types/typeutil"
	"gonum.org/v1/gonum/graph/simple"
	"hash/fnv"
	"strings"
)

type printVisitor struct {
	b      *strings.Builder
	indent string
}

func (p *printVisitor) Visit(node ast.Node) (w ast.Visitor) {
	if node == nil {
		return nil
	}
	p.b.WriteString(p.indent)
	p.b.WriteString(fmt.Sprintf("%T %s", node, shortNodeString(node)))
	p.b.WriteString("\n")
	return &printVisitor{b: p.b, indent: p.indent + "  "}
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

type statsVisitor struct {
	funcCount  int
	typeCount  int
	constCount int
	varCount   int
}

func (v *statsVisitor) Visit(node ast.Node) (w ast.Visitor) {
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

func FormatStatsVisitor(v *statsVisitor) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Funcs: %d\n", v.funcCount))
	b.WriteString(fmt.Sprintf("Types: %d\n", v.typeCount))
	b.WriteString(fmt.Sprintf("Consts: %d\n", v.constCount))
	b.WriteString(fmt.Sprintf("Vars: %d\n", v.varCount))
	return b.String()
}

type cohesionVisitor struct {
	fileSet      *token.FileSet
	pkg          *packages.Package
	dependencies *simple.DirectedGraph
	typesInfo    *types.Info
}

type typeNode struct {
	types.Type
}

func (t typeNode) ID() int64 {
	return int64(typeutil.MakeHasher().Hash(t.Type))
}

type objectNode struct {
	id int64
	types.Object
}

func (o objectNode) ID() int64 {
	return o.id
}

func newObjectNode(obj types.Object) objectNode {
	hash := fnv.New64a()
	hash.Sum([]byte(obj.Id()))
	return objectNode{int64(hash.Sum64()), obj}
}

func (c *cohesionVisitor) Visit(node ast.Node) (w ast.Visitor) {
	if expr, ok := node.(ast.Expr); ok {
		info := &types.Info{
			Defs:       make(map[*ast.Ident]types.Object),
			Uses:       make(map[*ast.Ident]types.Object),
			Types:      make(map[ast.Expr]types.TypeAndValue),
			Implicits:  make(map[ast.Node]types.Object),
			Selections: make(map[*ast.SelectorExpr]*types.Selection),
		}
		err := types.CheckExpr(c.fileSet, c.pkg.Types, node.Pos(), expr, info)
		if err != nil {
			return c
		}
		fmt.Println(NodeString(c.fileSet, node))
		c.printInfo(info)
		return nil
	}
	return c
}

func (c *cohesionVisitor) printInfo(info *types.Info) {
	fmt.Println("Defs:")
	for ident, object := range info.Defs {
		if c.BelongsToPackage(object) {
			fmt.Println(c.fileSet.Position(ident.Pos()), ident.Name, object)
		}
	}
	fmt.Println("Uses:")
	for ident, object := range info.Uses {
		if c.BelongsToPackage(object) {
			fmt.Println(c.fileSet.Position(ident.Pos()), ident.Name, object)
		}
	}
	fmt.Println("Implicits:")
	for ident, object := range info.Implicits {
		if c.BelongsToPackage(object) {
			fmt.Println(c.fileSet.Position(ident.Pos()), ident, object)
		}
	}
	fmt.Println()
}

func (c *cohesionVisitor) BelongsToPackage(obj types.Object) bool {
	if obj == nil || obj.Pkg() == nil {
		return false
	}
	switch obj.(type) {
	case *types.Var, *types.Const:
		return obj.Parent() == c.pkg.Types.Scope()
	case *types.Func, *types.TypeName:
		return obj.Pkg().Path() == c.pkg.PkgPath
	}
	return false
}

func NewCohesionVisitor(
	fileSet *token.FileSet,
	pkg *packages.Package,
) (*cohesionVisitor, error) {
	typesConfig := types.Config{Importer: importer.Default()}
	info := &types.Info{
		Defs: make(map[*ast.Ident]types.Object),
		Uses: make(map[*ast.Ident]types.Object),
	}
	if _, err := typesConfig.Check(pkg.PkgPath, fileSet, pkg.Syntax, info); err != nil {
		return nil, err
	}
	c := &cohesionVisitor{
		fileSet:      fileSet,
		pkg:          pkg,
		dependencies: simple.NewDirectedGraph(),
		typesInfo:    info,
	}
	c.printInfo(info)
	return c, nil
}

func NodeString(fileSet *token.FileSet, ident any) string {
	var buff bytes.Buffer
	format.Node(&buff, fileSet, ident)
	return buff.String()
}

func main() {
	dir := "/Users/jakubgruszecki/Documents/isbn/step_6"
	fileSet := token.NewFileSet()
	conf := &packages.Config{Mode: packages.LoadAllSyntax, Fset: fileSet, Dir: dir}
	pkgs, err := packages.Load(conf, dir)
	if err != nil {
		panic(err)
	}
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			panic(pkg.Errors[0])
		}
		fmt.Println(pkg.PkgPath)
		s := &statsVisitor{}
		p := &printVisitor{&strings.Builder{}, ""}
		c, err := NewCohesionVisitor(fileSet, pkg)
		if err != nil {
			panic(err)
		}
		for _, file := range pkg.Syntax {
			ast.Walk(s, file)
			ast.Walk(p, file)
			ast.Walk(c, file)
		}
		fmt.Println(p.b.String())
		fmt.Println(FormatStatsVisitor(s))
	}
}
