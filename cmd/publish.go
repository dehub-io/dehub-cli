package cmd

import (
	"archive/tar"
	"bytes"
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

	"github.com/spf13/cobra"
)

var publishRepo string
var publishToken string

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "发布包到仓库",
	Long: `将当前目录打包并发布到 GitHub Release。

需要设置 GITHUB_TOKEN 环境变量或使用 --token 参数。`,
	Example: `  dehub publish
  dehub publish --repo dehub-io/dehub-repository
  GITHUB_TOKEN=xxx dehub publish`,
	Run: func(cmd *cobra.Command, args []string) {
		// 加载包配置
		configPath := findConfig()
		cfg, err := loadProjectConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
			os.Exit(1)
		}

		// 获取 GitHub Token
		token := publishToken
		if token == "" {
			token = os.Getenv("GITHUB_TOKEN")
		}
		if token == "" {
			fmt.Fprintln(os.Stderr, "错误: 需要设置 GITHUB_TOKEN 环境变量或使用 --token 参数")
			os.Exit(1)
		}

		// 获取仓库名
		repo := publishRepo
		if repo == "" {
			repo = os.Getenv("DEHUB_REPOSITORY")
			if repo == "" {
				repo = "dehub-io/dehub-repository"
			}
		}

		pkgName := cfg.Name
		version := cfg.Version

		fmt.Printf("发布: %s@%s\n", pkgName, version)
		fmt.Printf("仓库: %s\n\n", repo)

		// 打包
		pkgDir := filepath.Dir(configPath)
		archivePath := filepath.Join(pkgDir, "package.tar.gz")
		sha256Path := filepath.Join(pkgDir, "sha256")

		fmt.Println("打包中...")
		if err := createArchive(pkgDir, archivePath); err != nil {
			fmt.Fprintf(os.Stderr, "打包失败: %v\n", err)
			os.Exit(1)
		}
		defer os.Remove(archivePath)

		// 计算 SHA256
		fmt.Println("计算校验和...")
		hash, err := calculateSHA256(archivePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "计算失败: %v\n", err)
			os.Exit(1)
		}
		os.WriteFile(sha256Path, []byte(hash+"  package.tar.gz\n"), 0644)
		defer os.Remove(sha256Path)

		// 创建 Release
		tag := fmt.Sprintf("%s/%s", pkgName, version)
		fmt.Printf("创建 Release: %s\n", tag)

		releaseID, err := createGitHubRelease(repo, tag, token)
		if err != nil {
			fmt.Fprintf(os.Stderr, "创建 Release 失败: %v\n", err)
			os.Exit(1)
		}

		// 上传文件
		fmt.Println("上传 package.tar.gz...")
		if err := uploadReleaseAsset(repo, releaseID, "package.tar.gz", archivePath, token); err != nil {
			fmt.Fprintf(os.Stderr, "上传失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("上传 sha256...")
		if err := uploadReleaseAsset(repo, releaseID, "sha256", sha256Path, token); err != nil {
			fmt.Fprintf(os.Stderr, "上传失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("\n发布成功!")
		fmt.Printf("下载地址: https://github.com/%s/releases/download/%s/package.tar.gz\n", repo, tag)
	},
}

func init() {
	publishCmd.Flags().StringVar(&publishRepo, "repo", "", "目标仓库 (默认 dehub-io/dehub-repository)")
	publishCmd.Flags().StringVar(&publishToken, "token", "", "GitHub Token")
	rootCmd.AddCommand(publishCmd)
}

func createArchive(dir, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzw := gzip.NewWriter(file)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// 要打包的文件
	files := []string{"package.yaml", "package.yml"}
	// 扫描常见源码目录
	sourceDirs := []string{"src", "include", "lib", "inc"}

	// 添加 package.yaml
	for _, f := range files {
		path := filepath.Join(dir, f)
		if _, err := os.Stat(path); err == nil {
			if err := addFileToTar(tw, dir, f); err != nil {
				return err
			}
		}
	}

	// 添加源码目录
	for _, srcDir := range sourceDirs {
		srcPath := filepath.Join(dir, srcDir)
		if _, err := os.Stat(srcPath); err == nil {
			filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}
				relPath, _ := filepath.Rel(dir, path)
				return addFileToTar(tw, dir, relPath)
			})
		}
	}

	// 添加 .c .h 文件（如果有）
	extensions := []string{".c", ".h", ".cpp", ".hpp", ".s", ".S"}
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		for _, ext := range extensions {
			if strings.HasSuffix(path, ext) {
				relPath, _ := filepath.Rel(dir, path)
				return addFileToTar(tw, dir, relPath)
			}
		}
		return nil
	})

	return nil
}

func addFileToTar(tw *tar.Writer, baseDir, relPath string) error {
	path := filepath.Join(baseDir, relPath)
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	header.Name = relPath

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(tw, file)
	return err
}

func calculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func createGitHubRelease(repo, tag, token string) (int64, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases", repo)

	body := map[string]interface{}{
		"tag_name":   tag,
		"name":       tag,
		"draft":      false,
		"prerelease": false,
	}
	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("GitHub API 错误: %s", string(respBody))
	}

	var release struct {
		ID int64 `json:"id"`
	}
	json.NewDecoder(resp.Body).Decode(&release)

	return release.ID, nil
}

func uploadReleaseAsset(repo string, releaseID int64, filename, filePath, token string) error {
	url := fmt.Sprintf("https://uploads.github.com/repos/%s/releases/%d/assets?name=%s",
		repo, releaseID, filename)

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, _ := file.Stat()

	req, err := http.NewRequest("POST", url, file)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Content-Type", "application/gzip")
	req.ContentLength = stat.Size()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("上传失败: %s", string(respBody))
	}

	return nil
}
