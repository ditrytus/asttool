package commands

import (
	"asttool"
	"asttool/cohesion"
	"github.com/spf13/cobra"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/packages"
)

func init() {
	cohesionCmd.AddCommand(deptsCmd)
	cohesionCmd.AddCommand(graphCmd)
}

var cohesionCmd = &cobra.Command{
	Use:   "cohesion",
	Short: "Print cohesion metrics for Go source code",
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
				return cohesion.FormatCohesionStats(visitor.(cohesion.Visitor))
			},
		).Run()
	},
}
