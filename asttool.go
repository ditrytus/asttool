package asttool

import (
	"fmt"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/packages"
)

type AstTool struct {
	loader         PackageLoader
	visitorFactory func(files *token.FileSet, pkg *packages.Package) ast.Visitor
	formatOutput   func(visitor ast.Visitor) string
}

func NewAstTool(
	loader PackageLoader,
	visitorFactory func(files *token.FileSet, pkg *packages.Package) ast.Visitor,
	formatOutput func(visitor ast.Visitor) string,
) *AstTool {
	return &AstTool{
		loader:         loader,
		visitorFactory: visitorFactory,
		formatOutput:   formatOutput,
	}
}

func (astTool *AstTool) Run() {
	pkgs, fileSet, err := astTool.loader.Load()
	if err != nil {
		panic(err)
	}
	for _, pkg := range pkgs {
		fmt.Println(pkg.PkgPath)
		v := astTool.visitorFactory(fileSet, pkg)
		for _, file := range pkg.Syntax {
			ast.Walk(v, file)
		}
		s := astTool.formatOutput(v)
		fmt.Println(s)
		fmt.Println()
	}
}
