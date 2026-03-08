package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dehub",
	Short: "Dehub is a source package manager",
	Long: `Dehub is a source code package manager inspired by Maven.
It manages source archives, handles dependencies, and supports
multi-level repository hierarchies.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
