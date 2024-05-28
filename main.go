package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/packages"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
	"os"
	"path/filepath"
	"strings"
)

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
