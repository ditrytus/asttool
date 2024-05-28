package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
	"hash/fnv"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type formatAstVisitor struct {
	b      *strings.Builder
	indent string
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

type cohesionAstVisitor struct {
	fileSet      *token.FileSet
	pkg          *packages.Package
	dependencies *simple.DirectedGraph
	typesInfo    *types.Info

	referencingObject types.Object
}

type objectNode struct {
	id int64
	types.Object
}

func (o objectNode) ID() int64 {
	return o.id
}

func newObjectNode(obj types.Object) objectNode {
	hash := fnv.New64()
	hash.Write([]byte(obj.Pkg().Path() + obj.Name() + strconv.Itoa(int(obj.Pos()))))
	return objectNode{int64(hash.Sum64()), obj}
}

func (c *cohesionAstVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch n := node.(type) {
	case *ast.FuncDecl:
		if obj, ok := c.typesInfo.Defs[n.Name]; ok {
			return c.childVisitor(obj)
		}
	case *ast.TypeSpec:
		if obj, ok := c.typesInfo.Defs[n.Name]; ok {
			return c.childVisitor(obj)
		}
	}
	if expr, ok := node.(ast.Expr); ok {
		info := &types.Info{
			Uses: make(map[*ast.Ident]types.Object),
		}
		err := types.CheckExpr(c.fileSet, c.pkg.Types, node.Pos(), expr, info)
		if err != nil {
			return c
		}
		c.addUsages(info)
		return nil
	}
	return c
}

func (c *cohesionAstVisitor) childVisitor(obj types.Object) *cohesionAstVisitor {
	return &cohesionAstVisitor{
		fileSet:      c.fileSet,
		pkg:          c.pkg,
		dependencies: c.dependencies,
		typesInfo:    c.typesInfo,

		referencingObject: obj,
	}
}

func (c *cohesionAstVisitor) debugPrintInfo(info *types.Info) {
	if info.Defs != nil {
		fmt.Println("Defs:")
		for ident, object := range info.Defs {
			if c.belongsToPackage(object) {
				fmt.Println(c.fileSet.Position(ident.Pos()), ident.Name, object)
			}
		}
	}
	if info.Uses != nil {
		fmt.Println("Uses:")
		for ident, object := range info.Uses {
			if c.belongsToPackage(object) {
				fmt.Println(c.fileSet.Position(ident.Pos()), ident.Name, object)
			}
		}
	}
	if info.Implicits != nil {
		fmt.Println("Implicits:")
		for ident, object := range info.Implicits {
			if c.belongsToPackage(object) {
				fmt.Println(c.fileSet.Position(ident.Pos()), ident, object)
			}
		}
	}
	fmt.Println()
}

func (c *cohesionAstVisitor) belongsToPackage(obj types.Object) bool {
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

func (c *cohesionAstVisitor) addUsages(info *types.Info) {
	if c.referencingObject == nil {
		return
	}
	if info.Uses == nil {
		return
	}
	for _, object := range info.Uses {
		if !c.belongsToPackage(object) {
			continue
		}
		from, ok := c.dependencies.NodeWithID(newObjectNode(c.referencingObject).ID())
		if ok {
			continue
		}
		to, ok := c.dependencies.NodeWithID(newObjectNode(object).ID())
		if ok {
			continue
		}
		if from.ID() == to.ID() {
			continue
		}
		edge := c.dependencies.NewEdge(from, to)
		c.dependencies.SetEdge(edge)
	}
}

func (c *cohesionAstVisitor) addDefinitions(info *types.Info) {
	for _, object := range info.Defs {
		if !c.belongsToPackage(object) {
			continue
		}
		node := newObjectNode(object)
		if _, ok := c.dependencies.NodeWithID(node.ID()); !ok {
			continue
		}
		c.dependencies.AddNode(node)
	}
}

func NewCohesionVisitor(
	fileSet *token.FileSet,
	pkg *packages.Package,
) (*cohesionAstVisitor, error) {
	c := &cohesionAstVisitor{
		fileSet:      fileSet,
		pkg:          pkg,
		dependencies: simple.NewDirectedGraph(),
		typesInfo:    pkg.TypesInfo,
	}
	c.addDefinitions(c.typesInfo)
	return c, nil
}

func NodeString(fileSet *token.FileSet, ident any) string {
	var buff bytes.Buffer
	format.Node(&buff, fileSet, ident)
	return buff.String()
}

func main() {
	dir := "/Users/jakubgruszecki/Documents/isbn"
	var pkgs []*packages.Package
	fileSet := token.NewFileSet()
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		conf := &packages.Config{Mode: packages.LoadSyntax, Fset: fileSet, Dir: path}
		dirPkgs, err := packages.Load(conf, path)
		if err != nil {
			return err
		}
		for _, pkg := range dirPkgs {
			if len(pkg.Errors) == 1 && pkg.Errors[0].Kind == packages.ListError {
				continue
			}
			pkgs = append(pkgs, pkg)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	for _, pkg := range pkgs {
		fmt.Println(pkg.PkgPath)
		if len(pkg.Errors) > 0 {
			packages.PrintErrors([]*packages.Package{pkg})
		}
		s := &statsAstVisitor{}
		p := &formatAstVisitor{&strings.Builder{}, ""}
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
		fmt.Println(c.FormatDependencies())
		fmt.Printf("Connected components: %d\n", c.ConnectedComponents())
		fmt.Printf("Average degree: %f\n", c.AverageDegree())
		fmt.Printf("Density: %f\n", c.Density())
	}
}

func (c *cohesionAstVisitor) FormatDependencies() string {
	var b strings.Builder
	nodes := c.dependencies.Nodes()
	for nodes.Next() {
		node := nodes.Node().(objectNode)
		b.WriteString(fmt.Sprintf("%s %s\n", c.fileSet.Position(node.Pos()), node.Name()))
		deps := c.dependencies.From(node.ID())
		for deps.Next() {
			dep := deps.Node().(objectNode)
			b.WriteString(fmt.Sprintf("\t%s %s\n", c.fileSet.Position(dep.Pos()), dep.Name()))
		}
	}
	return b.String()
}

func (c *cohesionAstVisitor) ConnectedComponents() int {
	return len(topo.ConnectedComponents(c.getUndirectedDependencies()))
}

func (c *cohesionAstVisitor) getUndirectedDependencies() *simple.UndirectedGraph {
	undirected := simple.NewUndirectedGraph()
	nodes := c.dependencies.Nodes()
	for nodes.Next() {
		undirected.AddNode(nodes.Node())
	}
	edges := c.dependencies.Edges()
	for edges.Next() {
		edge := edges.Edge()
		undirected.SetEdge(undirected.NewEdge(edge.From(), edge.To()))
	}
	return undirected
}

func (c *cohesionAstVisitor) AverageDegree() float64 {
	var totalDegree int
	nodes := c.dependencies.Nodes()
	totalNodes := nodes.Len()
	for nodes.Next() {
		totalDegree += c.dependencies.From(nodes.Node().ID()).Len()
	}
	return float64(totalDegree) / float64(totalNodes)
}

func (c *cohesionAstVisitor) Density() float64 {
	nodesCount := c.dependencies.Nodes().Len()
	maxEdges := nodesCount * (nodesCount - 1) / 2
	edgesCount := c.dependencies.Edges().Len()
	return float64(edgesCount) / float64(maxEdges)
}
