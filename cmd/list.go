package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出已安装的依赖",
	Long:  `列出所有已安装的依赖包及其版本。`,
	Run: func(cmd *cobra.Command, args []string) {
		configPath := findConfig()

		// 尝试加载锁文件
		lock, err := loadLockFile(configPath)
		if err != nil {
			// 如果没有锁文件，显示 package.yaml 中的依赖
			cfg, err := loadProjectConfig(configPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
				os.Exit(1)
			}

			fmt.Println("依赖声明 (package.yaml):")
			if len(cfg.Dependencies) == 0 {
				fmt.Println("  (无)")
			}
			for pkg, version := range cfg.Dependencies {
				fmt.Printf("  %s: %s\n", pkg, version)
			}
			return
		}

		// 显示锁定的依赖
		fmt.Println("已安装依赖 (package.lock):")
		if len(lock.Dependencies) == 0 {
			fmt.Println("  (无)")
		}
		for _, dep := range lock.Dependencies {
			fmt.Printf("  %s@%s\n", dep.Name, dep.Version)
			if installVerbose {
				fmt.Printf("    SHA256: %s\n", dep.SHA256[:16]+"...")
			}
		}

		// 检查依赖目录
		depsDir := filepath.Join(filepath.Dir(configPath), ".dehub", "deps")
		if _, err := os.Stat(depsDir); err == nil {
			fmt.Println("\n依赖目录:")
			fmt.Printf("  %s\n", depsDir)
		}
	},
}

func init() {
	listCmd.Flags().BoolVarP(&installVerbose, "verbose", "v", false, "显示详细信息")
	rootCmd.AddCommand(listCmd)
}
