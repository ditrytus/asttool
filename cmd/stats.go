package cmd

import (
	"cohesion/loader"
	"cohesion/stats"
	"fmt"
	"github.com/spf13/cobra"
	"go/ast"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Print statistics about Go source code",
	Run: func(cmd *cobra.Command, args []string) {
		loader := loader.NewDirPackageLoader(dir)
		pkgs, _, err := loader.Load()
		if err != nil {
			panic(err)
		}
		for _, pkg := range pkgs {
			fmt.Println(pkg.PkgPath)
			v := stats.NewStatsVisitor()
			for _, file := range pkg.Syntax {
				ast.Walk(v, file)
			}
			s := v.Stats()
			fmt.Printf("Funcs: %d\n", s.FuncCount)
			fmt.Printf("Types: %d\n", s.TypeCount)
			fmt.Printf("Consts: %d\n", s.ConstCount)
			fmt.Printf("Vars: %d\n", s.VarCount)
			fmt.Println()
		}
	},
}
