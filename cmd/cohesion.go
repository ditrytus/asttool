package cmd

import (
	"cohesion/cohesion"
	"cohesion/loader"
	"fmt"
	"github.com/spf13/cobra"
	"go/ast"
)

var cohesionCmd = &cobra.Command{
	Use:   "cohesion",
	Short: "Print cohesion metrics for Go source code",
	Run: func(cmd *cobra.Command, args []string) {
		loader := loader.NewDirPackageLoader(dir)
		pkgs, fileSet, err := loader.Load()
		if err != nil {
			printErrorAndExit(err)
		}
		for _, pkg := range pkgs {
			fmt.Println(pkg.PkgPath)
			v, err := cohesion.NewCohesionVisitor(fileSet, pkg)
			if err != nil {
				printErrorAndExit(err)
			}
			for _, file := range pkg.Syntax {
				ast.Walk(v, file)
			}
			fmt.Printf("Connected components: %d\n", v.ConnectedComponents())
			fmt.Printf("Average degree: %f\n", v.AverageDegree())
			fmt.Printf("Density: %f\n", v.Density())
			fmt.Println()
		}
	},
}
