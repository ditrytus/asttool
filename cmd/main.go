package main

import (
	asttool "cohesion"
	"fmt"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/packages"
	"os"
	"path/filepath"
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
		s := asttool.NewStatsVisitor()
		p := asttool.NewFormatVisitor("  ")
		c, err := asttool.NewCohesionVisitor(fileSet, pkg)
		if err != nil {
			panic(err)
		}
		for _, file := range pkg.Syntax {
			ast.Walk(s, file)
			ast.Walk(p, file)
			ast.Walk(c, file)
		}
		fmt.Println(p.String())
		fmt.Println(asttool.FormatStatsVisitor(s))
		fmt.Println(c.FormatDependencies())
		fmt.Printf("Connected components: %d\n", c.ConnectedComponents())
		fmt.Printf("Average degree: %f\n", c.AverageDegree())
		fmt.Printf("Density: %f\n", c.Density())
	}
}
