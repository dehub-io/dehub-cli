package cmd

import (
	"fmt"
	"os"

	"github.com/dehub-io/dehub-cli/adapter"

	"github.com/spf13/cobra"
)

var (
	// Version 版本号
	Version = "0.1.0"
	// 用于全局的配置文件路径
	cfgFile string
	// 全局缓存目录
	cacheDir string
	// 服务端 URL
	serverURL string
	// 全局适配器实例
	serverAdapter adapter.DehubServer
)

var rootCmd = &cobra.Command{
	Use:     "dehub",
	Short:   "源码包管理器",
	Long:    `dehub - 源码包管理器，管理源码归档、处理依赖关系。`,
	Version: Version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "package.yaml", "配置文件路径")
	rootCmd.PersistentFlags().StringVar(&cacheDir, "cache-dir", "", "缓存目录 (默认 ~/.dehub/cache)")
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "https://dehub.io", "服务端地址")

	// 隐藏 completion 命令
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	// 中文化 help 命令
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:   "help [command]",
		Short: "显示帮助信息",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				return
			}
			targetCmd, _, err := rootCmd.Find(args)
			if err != nil {
				fmt.Fprintf(os.Stderr, "未知命令: %s\n", args[0])
				os.Exit(1)
			}
			targetCmd.Help()
		},
	})

	// 自定义帮助模板
	rootCmd.SetHelpTemplate(`{{.Long}}

用法:
  {{.UseLine}}

命令:{{range .Commands}}{{if (not .Hidden)}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

参数:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}

使用 "{{.UseLine}} [command] --help" 查看命令详情。
`)

	// 自定义错误提示
	rootCmd.SetUsageTemplate(`用法: {{.UseLine}}

命令:{{range .Commands}}{{if (not .Hidden)}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}

参数:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}
`)

	// 中文化错误信息
	cobra.AddTemplateFunc("Error", func(err error) string {
		return fmt.Sprintf("错误: %v", err)
	})
}

// getAdapter 获取或初始化适配器
func getAdapter() adapter.DehubServer {
	if serverAdapter != nil {
		return serverAdapter
	}
	
	var err error
	serverAdapter, err = adapter.NewAdapter(serverURL, cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化适配器失败: %v\n", err)
		os.Exit(1)
	}
	
	return serverAdapter
}

// SetAdapter 设置适配器（用于测试）
func SetAdapter(a adapter.DehubServer) {
	serverAdapter = a
}

// ResetAdapter 重置适配器（用于测试）
func ResetAdapter() {
	serverAdapter = nil
}
