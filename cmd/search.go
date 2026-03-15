package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var searchPlatform string
var searchBoard string

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "搜索包",
	Long:  `在配置的仓库中搜索包。`,
	Example: `  dehub search hal
  dehub search oled --platform esp32
  dehub search sensor --board any`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		query := args[0]
		
		adapter := getAdapter()
		
		fmt.Printf("搜索: %s\n\n", query)
		
		packages, err := adapter.Search(query)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "搜索失败: %v\n", err)
			return
		}
		
		if len(packages) == 0 {
			fmt.Println("未找到匹配的包")
			return
		}
		
		for _, pkg := range packages {
			fmt.Printf("  %s", pkg.Name)
			if pkg.Description != "" {
				fmt.Printf(" - %s", pkg.Description)
			}
			fmt.Println()
			if installVerbose && len(pkg.Versions) > 0 {
				fmt.Printf("    最新版本: %s\n", pkg.Versions[0].Version)
			}
		}
		
		fmt.Printf("\n共找到 %d 个包\n", len(packages))
	},
}

func init() {
	searchCmd.Flags().StringVar(&searchPlatform, "platform", "", "按平台过滤")
	searchCmd.Flags().StringVar(&searchBoard, "board", "", "按开发板过滤")
	searchCmd.Flags().BoolVarP(&installVerbose, "verbose", "v", false, "显示详细信息")
	rootCmd.AddCommand(searchCmd)
}
