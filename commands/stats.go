package commands

import (
	"asttool"
	"asttool/stats"
	"github.com/spf13/cobra"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/packages"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Print statistics about Go source code",
	Run: func(cmd *cobra.Command, args []string) {
		asttool.NewAstTool(
			asttool.NewDirPackageLoader(dir),
			func(_ *token.FileSet, _ *packages.Package) ast.Visitor {
				return stats.NewStatsVisitor()
			},
			func(visitor ast.Visitor) string {
				return stats.FormatOutputStats(visitor.(stats.Visitor))
			},
		).Run()
	},
}
