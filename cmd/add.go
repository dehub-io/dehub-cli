package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var addCmd = &cobra.Command{
	Use:   "add <package>[@version]",
	Short: "添加依赖",
	Long:  `添加一个依赖到 package.yaml 并安装。`,
	Example: `  dehub add zyrthi/hal
  dehub add zyrthi/hal@1.0.0
  dehub add zyrthi/hal@^1.0.0`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pkgSpec := args[0]

		// 解析包名和版本
		pkgName := pkgSpec
		version := ""
		for i := 0; i < len(pkgSpec); i++ {
			if pkgSpec[i] == '@' {
				pkgName = pkgSpec[:i]
				version = pkgSpec[i+1:]
				break
			}
		}

		configPath := findConfig()
		cfg, err := loadProjectConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
			os.Exit(1)
		}

		// 初始化 dependencies
		if cfg.Dependencies == nil {
			cfg.Dependencies = make(map[string]string)
		}

		// 设置版本
		if version == "" {
			version = "latest"
		}

		cfg.Dependencies[pkgName] = version

		// 写回配置文件
		data, err := yaml.Marshal(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "序列化失败: %v\n", err)
			os.Exit(1)
		}

		if err := os.WriteFile(configPath, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "写入配置失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("已添加依赖: %s@%s\n", pkgName, version)
		fmt.Printf("配置文件: %s\n", configPath)

		// 安装依赖
		fmt.Println("\n正在安装...")
		installDeps(cfg, configPath)
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func installDeps(cfg *ProjectConfig, configPath string) {
	// 确保缓存目录存在
	cache := getCacheDir()
	os.MkdirAll(cache, 0755)

	// 确保依赖目录存在
	depsDir := filepath.Join(filepath.Dir(configPath), ".dehub", "deps")
	os.MkdirAll(depsDir, 0755)

	for pkgName, versionConstraint := range cfg.Dependencies {
		pkgInfo, _, err := resolvePackage(cfg.Repositories, pkgName, versionConstraint)
		if err != nil {
			fmt.Fprintf(os.Stderr, "解析失败 %s: %v\n", pkgName, err)
			continue
		}

		fmt.Printf("安装: %s@%s\n", pkgName, pkgInfo.Version)

		archivePath, err := downloadPackage(cache, pkgInfo.ArchiveURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "下载失败: %v\n", err)
			continue
		}

		sha256Hash, err := fetchSHA256(pkgInfo.SHA256URL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "获取校验失败: %v\n", err)
			continue
		}

		if err := verifySHA256(archivePath, sha256Hash); err != nil {
			fmt.Fprintf(os.Stderr, "校验失败: %v\n", err)
			continue
		}

		targetDir := filepath.Join(depsDir, pkgName, pkgInfo.Version)
		if err := extractArchive(archivePath, targetDir); err != nil {
			fmt.Fprintf(os.Stderr, "解压失败: %v\n", err)
			continue
		}

		fmt.Printf("  -> %s\n", targetDir)
	}
}
