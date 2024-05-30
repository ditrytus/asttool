package commands

import (
	"asttool"
	"asttool/format"
	"github.com/spf13/cobra"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/packages"
)

func init() {
	formatCmd.Flags().StringVarP(&indent, "indent", "i", "  ", "indentation string")
}

var (
	indent string

	formatCmd = &cobra.Command{
		Use:   "format",
		Short: "print Go source code AST in a formatted way",
		Run: func(cmd *cobra.Command, args []string) {
			asttool.NewAstTool(
				asttool.NewDirPackageLoader(dir),
				func(_ *token.FileSet, _ *packages.Package) ast.Visitor {
					return format.NewFormatVisitor(indent)
				},
				func(visitor ast.Visitor) string {
					return format.FormatOutput(visitor.(format.Visitor))
				},
			).Run()
		},
	}
)
