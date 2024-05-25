package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type printDefersVisitor struct {
	fset *token.FileSet
}

func (l printDefersVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch node.(type) {
	case *ast.DeferStmt:
		var buf bytes.Buffer
		printer.Fprint(&buf, l.fset, node)
		fmt.Println(buf.String())
	}
	return l
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

func main() {
	dir := "/Users/jakubgruszecki/Documents/sarama"
	fileSet := token.NewFileSet()
	pkgs := make(map[string]*ast.Package)
	if err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		dirPkg, err := parser.ParseDir(fileSet, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		for key, val := range dirPkg {
			pkgs[key] = val
		}
		return nil
	}); err != nil {
		panic(err)
	}
	var b strings.Builder
	for _, pkg := range pkgs {
		s := statsVisitor{}
		ast.Walk(&s, pkg)
		b.WriteString(fmt.Sprintf("Package %s\n", pkg.Name))
		b.WriteString(fmt.Sprintf("  Funcs: %d\n", s.funcCount))
		b.WriteString(fmt.Sprintf("  Types: %d\n", s.typeCount))
		b.WriteString(fmt.Sprintf("  Consts: %d\n", s.constCount))
		b.WriteString(fmt.Sprintf("  Vars: %d\n", s.varCount))

	}
	fmt.Println(b.String())
}
