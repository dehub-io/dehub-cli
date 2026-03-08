package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [name]",
	Short: "初始化项目",
	Long:  `在当前目录初始化一个新项目，生成 package.yaml 文件。`,
	Example: `  dehub init
  dehub init my-project
  dehub init --config custom.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		projectName := "my-project"
		if len(args) > 0 {
			projectName = args[0]
		}

		configPath := findConfig()
		if _, err := os.Stat(configPath); err == nil {
			fmt.Printf("package.yaml 已存在: %s\n", configPath)
			return
		}

		content := fmt.Sprintf(`# package.yaml - 项目依赖配置
# 由 "dehub init" 生成

name: %s
version: "0.1.0"

# 可选字段，仅用于搜索
platform: ""
board: ""

# 依赖声明
dependencies: {}

# 仓库配置（按优先级排序）
repositories:
  - name: central
    url: https://dehub-io.github.io/dehub-repository/
`, projectName)

		if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "错误: %v\n", err)
			return
		}

		fmt.Printf("已创建: %s\n", configPath)
		fmt.Println("\n下一步:")
		fmt.Println("  dehub add <package>   添加依赖")
		fmt.Println("  dehub install         安装依赖")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
