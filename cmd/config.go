package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// ProjectConfig package.yaml 项目配置结构
type ProjectConfig struct {
	Name         string            `yaml:"name"`
	Version      string            `yaml:"version"`
	Platform     string            `yaml:"platform,omitempty"`
	Board        string            `yaml:"board,omitempty"`
	Dependencies map[string]string `yaml:"dependencies"`
	Repositories []Repository      `yaml:"repositories"`
}

// Repository 仓库配置
type Repository struct {
	Name string `yaml:"name"`
	URL  string `yaml:"url"`
}

// LockFile package.lock 锁文件结构
type LockFile struct {
	Version      int             `yaml:"version"`
	Dependencies []LockedPackage `yaml:"dependencies"`
}

// LockedPackage 锁定的包
type LockedPackage struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	SHA256  string `yaml:"sha256"`
}

// Index 仓库全局索引
type Index struct {
	SchemaVersion string   `json:"schema_version"`
	Packages      []string `json:"packages"`
	UpdatedAt     string   `json:"updated_at"`
}

// PackageIndex 单包版本索引
type PackageIndex struct {
	SchemaVersion string        `json:"schema_version"`
	Package       string        `json:"package"`
	Versions      []VersionInfo `json:"versions"`
}

// VersionInfo 版本信息
type VersionInfo struct {
	Version      string            `json:"version"`
	ArchiveURL   string            `json:"archive_url"`
	SHA256URL    string            `json:"sha256_url"`
	ReleaseDate  string            `json:"release_date"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
}

// getCacheDir 获取缓存目录
func getCacheDir() string {
	if cacheDir != "" {
		return cacheDir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".dehub", "cache")
}

// findConfig 查找配置文件
func findConfig() string {
	if cfgFile != "package.yaml" {
		return cfgFile
	}

	cwd, _ := os.Getwd()
	configPath := filepath.Join(cwd, cfgFile)
	if _, err := os.Stat(configPath); err == nil {
		return configPath
	}

	dir := cwd
	for {
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		configPath := filepath.Join(parent, cfgFile)
		if _, err := os.Stat(configPath); err == nil {
			return configPath
		}
		dir = parent
	}

	return cfgFile
}

// loadProjectConfig 加载项目配置
func loadProjectConfig(path string) (*ProjectConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg ProjectConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// 设置默认仓库
	if len(cfg.Repositories) == 0 {
		cfg.Repositories = []Repository{
			{Name: "central", URL: "https://dehub-io.github.io/dehub-repository/"},
		}
	}

	return &cfg, nil
}

// loadLockFile 加载锁文件
func loadLockFile(path string) (*LockFile, error) {
	lockPath := filepath.Join(filepath.Dir(path), "package.lock")
	data, err := os.ReadFile(lockPath)
	if err != nil {
		return nil, err
	}

	var lock LockFile
	if err := yaml.Unmarshal(data, &lock); err != nil {
		return nil, err
	}

	return &lock, nil
}

// saveLockFile 保存锁文件
func saveLockFile(path string, lock *LockFile) error {
	lockPath := filepath.Join(filepath.Dir(path), "package.lock")
	data, err := yaml.Marshal(lock)
	if err != nil {
		return err
	}

	header := "# dehub lock file - 自动生成，请勿手动修改\n\n"
	return os.WriteFile(lockPath, append([]byte(header), data...), 0644)
}

// fetchIndex 从仓库获取全局索引
func fetchIndex(repoURL string) (*Index, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	url := repoURL + "index.json"

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("获取索引失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("获取索引失败: 状态码 %d", resp.StatusCode)
	}

	var index Index
	if err := json.NewDecoder(resp.Body).Decode(&index); err != nil {
		return nil, fmt.Errorf("解析索引失败: %w", err)
	}

	return &index, nil
}

// fetchPackageIndex 从仓库获取包索引
func fetchPackageIndex(repoURL, pkgName string) (*PackageIndex, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	url := repoURL + "packages/" + pkgName + "/index.json"

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("获取包索引失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("包不存在: %s", pkgName)
	}

	var pkgIndex PackageIndex
	if err := json.NewDecoder(resp.Body).Decode(&pkgIndex); err != nil {
		return nil, fmt.Errorf("解析包索引失败: %w", err)
	}

	return &pkgIndex, nil
}
