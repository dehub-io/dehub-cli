package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version 版本号
	Version = "0.1.0"
	// 用于全局的配置文件路径
	cfgFile string
	// 全局缓存目录
	cacheDir string
)

var rootCmd = &cobra.Command{
	Use:   "dehub",
	Short: "源码包管理器",
	Long: `dehub - 源码包管理器

受 Maven 启发，管理源码归档、处理依赖关系，
支持多级仓库层级架构。`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "package.yaml", "配置文件路径")
	rootCmd.PersistentFlags().StringVar(&cacheDir, "cache-dir", "", "缓存目录 (默认 ~/.dehub/cache)")
}
