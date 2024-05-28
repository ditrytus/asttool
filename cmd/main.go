package main

import (
	"cohesion/cohesion"
	"cohesion/format"
	"cohesion/stats"
	"fmt"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/packages"
	"os"
	"path/filepath"
)

func main() {
	dir := "/Users/jakubgruszecki/Documents/isbn"
	loader := NewDirPackageLoader(dir)
	pkgs, fileSet, err := loader.Load()
	if err != nil {
		panic(err)
	}
	for _, pkg := range pkgs {
		fmt.Println(pkg.PkgPath)
		if len(pkg.Errors) > 0 {
			packages.PrintErrors([]*packages.Package{pkg})
		}
		s := stats.NewStatsVisitor()
		p := format.NewFormatVisitor("  ")
		c, err := cohesion.NewCohesionVisitor(fileSet, pkg)
		if err != nil {
			panic(err)
		}
		for _, file := range pkg.Syntax {
			ast.Walk(s, file)
			ast.Walk(p, file)
			ast.Walk(c, file)
		}
		fmt.Println(p.String())
		fmt.Println(stats.FormatStatsVisitor(s))
		fmt.Println(cohesion.FormatDependencies(c))
		fmt.Printf("Connected components: %d\n", c.ConnectedComponents())
		fmt.Printf("Average degree: %f\n", c.AverageDegree())
		fmt.Printf("Density: %f\n", c.Density())
	}
}

type Config struct {
	Command      string
	Dir          string
	FormatIndent string
}

func DefaultConfig() Config {
	return Config{
		FormatIndent: "  ",
	}
}

type AstTool struct {
	loader PackageLoader
}

type VisitorProvider interface {
	Visitor() ast.Visitor
}

type Command int

type VisitorsFactory interface {
	NewStatsVisitor() stats.Visitor
	NewFormatVisitor() format.Visitor
	NewCohesionVisitor() (cohesion.Visitor, error)
}

type visitorsFactory struct {
	fileSet *token.FileSet
	pkg     *packages.Package
	indent  string
}

func (v visitorsFactory) NewStatsVisitor() stats.Visitor {
	return stats.NewStatsVisitor()
}

func (v visitorsFactory) NewFormatVisitor() format.Visitor {
	return format.NewFormatVisitor(v.indent)
}

func (v visitorsFactory) NewCohesionVisitor() (cohesion.Visitor, error) {
	return cohesion.NewCohesionVisitor(v.fileSet, v.pkg)
}

func NewVisitorsFactory(fileSet *token.FileSet, pkg *packages.Package, indent string) VisitorsFactory {
	return &visitorsFactory{fileSet: fileSet, pkg: pkg, indent: indent}
}

type PackageLoader interface {
	Load() ([]*packages.Package, *token.FileSet, error)
}

type dirPackageLoader struct {
	dir string
}

func NewDirPackageLoader(dir string) PackageLoader {
	return &dirPackageLoader{dir: dir}
}

func (d *dirPackageLoader) Load() ([]*packages.Package, *token.FileSet, error) {
	var pkgs []*packages.Package
	fileSet := token.NewFileSet()
	err := filepath.Walk(d.dir, func(path string, info os.FileInfo, err error) error {
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
	return pkgs, fileSet, err
}
