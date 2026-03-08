package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <package>",
	Short: "Add a dependency",
	Long:  `Add a dependency to package.yaml.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pkg := args[0]
		fmt.Printf("add command - adding %s (not implemented yet)\n", pkg)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
