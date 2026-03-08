package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new project",
	Long:  `Initialize a new project with a package.yaml file.`,
	Run: func(cmd *cobra.Command, args []string) {
		if _, err := os.Stat("package.yaml"); err == nil {
			fmt.Println("package.yaml already exists")
			return
		}

		content := `# package.yaml - Project dependencies
name: my-project
version: "0.1.0"

# Optional, for search only
platform: ""
board: ""

dependencies: {}

repositories:
  - name: central
    url: https://dehub-io.github.io/dehub-repository/
`
		if err := os.WriteFile("package.yaml", []byte(content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return
		}
		fmt.Println("Created package.yaml")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
