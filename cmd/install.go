package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install dependencies",
	Long:  `Install all dependencies defined in package.yaml.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("install command - not implemented yet")
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
