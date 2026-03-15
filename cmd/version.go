package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:    "version",
	Short:  "显示版本信息",
	Hidden: true, // 隐藏，使用 --version 即可
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "dehub version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
