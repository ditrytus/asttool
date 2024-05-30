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
	graphCmd.Flags().StringVarP(
		&outputGraph,
		"output",
		"o",
		"graph.png",
		"output file path to save png file to",
	)
}

var (
	outputGraph string

	graphCmd = &cobra.Command{
		Use:   "graph",
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
					return cohesion.FormatCohesionGraph(visitor.(cohesion.Visitor), outputGraph)
				},
			).Run()
		},
	}
)
