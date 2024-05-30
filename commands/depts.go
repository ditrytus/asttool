package commands

import (
	"asttool"
	"asttool/cohesion"
	"github.com/spf13/cobra"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/packages"
)

var deptsCmd = &cobra.Command{
	Use:   "dependencies",
	Short: "Print graph of internal dependencies of a package",
	Run: func(cmd *cobra.Command, args []string) {
		dir := args[0]
		asttool.NewAstTool(
			asttool.NewDirPackageLoader(dir),
			func(files *token.FileSet, pkg *packages.Package) ast.Visitor {
				visitor, err := cohesion.NewCohesionVisitor(files, pkg)
				if err != nil {
					panic(err)
				}
				return visitor
			},
			func(visitor ast.Visitor) string {
				return cohesion.FormatDependencies(visitor.(cohesion.Visitor))
			},
		).Run()
	},
}
