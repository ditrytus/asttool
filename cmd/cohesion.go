package cmd

import (
	asttool "cohesion"
	"cohesion/cohesion"
	"cohesion/loader"
	"github.com/spf13/cobra"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/packages"
)

var cohesionCmd = &cobra.Command{
	Use:   "cohesion",
	Short: "Print cohesion metrics for Go source code",
	Run: func(cmd *cobra.Command, args []string) {
		asttool.NewAstTool(
			loader.NewDirPackageLoader(dir),
			func(files *token.FileSet, pkg *packages.Package) ast.Visitor {
				visitor, err := cohesion.NewCohesionVisitor(files, pkg)
				if err != nil {
					panic(err)
				}
				return visitor
			},
			func(visitor ast.Visitor) string {
				return cohesion.FormatCohesionStats(visitor.(cohesion.Visitor))
			},
		).Run()
	},
}
