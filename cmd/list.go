package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed dependencies",
	Long:  `List all installed dependencies from package.lock.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("list command - not implemented yet")
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
