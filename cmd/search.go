package cmd

import (
	"fmt"
	"strings"

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
		query := strings.ToLower(args[0])

		configPath := findConfig()
		cfg, err := loadProjectConfig(configPath)
		if err != nil {
			// 使用默认仓库
			cfg = &ProjectConfig{
				Repositories: []Repository{
					{Name: "central", URL: "https://dehub-io.github.io/dehub-repository/"},
				},
			}
		}

		fmt.Printf("搜索: %s\n\n", query)

		found := false
		for _, repo := range cfg.Repositories {
			index, err := fetchIndex(repo.URL)
			if err != nil {
				if installVerbose {
					fmt.Fprintf(cmd.OutOrStderr(), "获取索引失败 %s: %v\n", repo.Name, err)
				}
				continue
			}

			for _, pkg := range index.Packages {
				// 简单匹配：包名包含查询字符串
				if strings.Contains(strings.ToLower(pkg), query) {
					// 如果指定了 platform/board 过滤
					if searchPlatform != "" || searchBoard != "" {
						pkgIndex, err := fetchPackageIndex(repo.URL, pkg)
						if err != nil {
							continue
						}
						// TODO: 根据 platform/board 过滤
						_ = pkgIndex
					}

					fmt.Printf("  %s\n", pkg)
					if installVerbose {
						fmt.Printf("    仓库: %s\n", repo.Name)
					}
					found = true
				}
			}
		}

		if !found {
			fmt.Println("未找到匹配的包")
		}
	},
}

func init() {
	searchCmd.Flags().StringVar(&searchPlatform, "platform", "", "按平台过滤")
	searchCmd.Flags().StringVar(&searchBoard, "board", "", "按开发板过滤")
	searchCmd.Flags().BoolVarP(&installVerbose, "verbose", "v", false, "显示详细信息")
	rootCmd.AddCommand(searchCmd)
}
