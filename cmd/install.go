package cmd

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var installVerbose bool

var installCmd = &cobra.Command{
	Use:   "install [package]",
	Short: "安装依赖",
	Long:  `安装 package.yaml 中定义的所有依赖，或安装指定的包。`,
	Example: `  dehub install           安装所有依赖
  dehub install zyrthi/hal   安装指定包`,
	Run: func(cmd *cobra.Command, args []string) {
		configPath := findConfig()
		cfg, err := loadProjectConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
			os.Exit(1)
		}

		// 确保缓存目录存在
		cache := getCacheDir()
		if err := os.MkdirAll(cache, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "创建缓存目录失败: %v\n", err)
			os.Exit(1)
		}

		// 确保依赖目录存在
		depsDir := filepath.Join(filepath.Dir(configPath), ".dehub", "deps")
		if err := os.MkdirAll(depsDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "创建依赖目录失败: %v\n", err)
			os.Exit(1)
		}

		// 收集需要安装的包
		packages := cfg.Dependencies
		if len(args) > 0 {
			// 安装指定的包
			pkg := args[0]
			version := ""
			if len(args) > 1 {
				version = args[1]
			}
			if packages == nil {
				packages = make(map[string]string)
			}
			packages[pkg] = version
		}

		if len(packages) == 0 {
			fmt.Println("没有依赖需要安装")
			return
		}

		// 解析并安装依赖
		lock := &LockFile{Version: 1}
		for pkgName, versionConstraint := range packages {
			if installVerbose {
				fmt.Printf("解析: %s@%s\n", pkgName, versionConstraint)
			}

			pkgInfo, _, err := resolvePackage(cfg.Repositories, pkgName, versionConstraint)
			if err != nil {
				fmt.Fprintf(os.Stderr, "解析包失败 %s: %v\n", pkgName, err)
				continue
			}

			fmt.Printf("安装: %s@%s\n", pkgName, pkgInfo.Version)

			// 下载包
			archivePath, err := downloadPackage(cache, pkgInfo.ArchiveURL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "下载失败: %v\n", err)
				continue
			}

			// 获取并校验 SHA256
			sha256Hash, err := fetchSHA256(pkgInfo.SHA256URL)
			if err != nil {
				fmt.Fprintf(os.Stderr, "获取校验失败: %v\n", err)
				continue
			}

			// 校验文件
			if err := verifySHA256(archivePath, sha256Hash); err != nil {
				fmt.Fprintf(os.Stderr, "校验失败: %v\n", err)
				os.Remove(archivePath)
				continue
			}

			// 解压到依赖目录
			targetDir := filepath.Join(depsDir, pkgName, pkgInfo.Version)
			if err := extractArchive(archivePath, targetDir); err != nil {
				fmt.Fprintf(os.Stderr, "解压失败: %v\n", err)
				continue
			}

			fmt.Printf("  -> %s\n", targetDir)

			// 添加到锁文件
			lock.Dependencies = append(lock.Dependencies, LockedPackage{
				Name:    pkgName,
				Version: pkgInfo.Version,
				SHA256:  sha256Hash,
			})
		}

		// 保存锁文件
		if len(lock.Dependencies) > 0 {
			if err := saveLockFile(configPath, lock); err != nil {
				fmt.Fprintf(os.Stderr, "保存锁文件失败: %v\n", err)
			} else {
				fmt.Println("\n已生成: package.lock")
			}
		}
	},
}

func init() {
	installCmd.Flags().BoolVarP(&installVerbose, "verbose", "v", false, "详细输出")
	rootCmd.AddCommand(installCmd)
}

// resolvePackage 从仓库解析包版本
func resolvePackage(repos []Repository, pkgName, versionConstraint string) (*VersionInfo, string, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	for _, repo := range repos {
		url := strings.TrimSuffix(repo.URL, "/") + "/packages/" + pkgName + "/index.json"
		resp, err := client.Get(url)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			continue
		}

		var pkgIndex PackageIndex
		if err := json.NewDecoder(resp.Body).Decode(&pkgIndex); err != nil {
			continue
		}

		// 选择最新版本（简化版，暂不支持 semver 解析）
		if len(pkgIndex.Versions) > 0 {
			return &pkgIndex.Versions[0], repo.URL, nil
		}
	}

	return nil, "", fmt.Errorf("包不存在: %s", pkgName)
}

// downloadPackage 下载包到缓存
func downloadPackage(cache, url string) (string, error) {
	fileName := filepath.Base(url)
	cachePath := filepath.Join(cache, fileName)

	// 检查缓存
	if _, err := os.Stat(cachePath); err == nil {
		return cachePath, nil
	}

	// 下载
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("下载失败: 状态码 %d", resp.StatusCode)
	}

	out, err := os.Create(cachePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return cachePath, err
}

// fetchSHA256 获取 SHA256 校验值
func fetchSHA256(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// sha256sum 格式: "hash  filename"
	parts := strings.Fields(string(data))
	if len(parts) > 0 {
		return parts[0], nil
	}
	return "", fmt.Errorf("无效的 SHA256 格式")
}

// verifySHA256 校验文件 SHA256
func verifySHA256(filePath, expectedHash string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	actualHash := hex.EncodeToString(hash.Sum(nil))
	if actualHash != expectedHash {
		return fmt.Errorf("SHA256 不匹配: 期望 %s, 实际 %s", expectedHash, actualHash)
	}

	return nil
}

// extractArchive 解压 tar.gz 文件
func extractArchive(archivePath, targetDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(targetDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}
