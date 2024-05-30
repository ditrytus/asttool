package cmd

import (
	"cohesion/format"
	"cohesion/loader"
	"fmt"
	"github.com/spf13/cobra"
	"go/ast"
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
			loader := loader.NewDirPackageLoader(dir)
			pkgs, _, err := loader.Load()
			if err != nil {
				panic(err)
			}
			for _, pkg := range pkgs {
				fmt.Println(pkg.PkgPath)
				v := format.NewFormatVisitor(indent)
				for _, file := range pkg.Syntax {
					ast.Walk(v, file)
				}
				s := format.FormatOutput(v)
				fmt.Println(s)
				fmt.Println()
			}
		},
	}
)
