package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var (
	dir string

	rootCmd = &cobra.Command{
		Use:   "asttool",
		Short: "A set of tools for working with Go AST",
	}
)

func Execute() {
	printErrorAndExit(rootCmd.Execute())
}

func printErrorAndExit(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&dir, "dir", "d", ".", "directory to analyze")
	rootCmd.AddCommand(formatCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(cohesionCmd)
}
