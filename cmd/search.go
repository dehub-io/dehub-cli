package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for packages",
	Long:  `Search for packages in configured repositories.`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		fmt.Printf("search command - searching for %s (not implemented yet)\n", query)
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
