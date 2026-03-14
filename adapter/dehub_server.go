package adapter

import (
	"io"
)

// DehubServer 定义 dehub 服务端标准接口
// 不同后端实现此接口
type DehubServer interface {
	// 认证相关
	Login() error
	Logout() error
	GetAuthStatus() (*AuthStatus, error)
	
	// 命名空间相关
	CreateNamespace(name, description string) (*Namespace, error)
	ListNamespaces() ([]*Namespace, error)
	GetNamespace(name string) (*Namespace, error)
	
	// 包相关
	Publish(pkgPath string, progress io.Writer) error
	ListPackages(namespace string) ([]*Package, error)
	GetPackage(name string) (*Package, error)
	
	// 依赖相关
	Install(dependencies map[string]string, targetDir string) error
	Search(query string) ([]*Package, error)
}

// AuthStatus 认证状态
type AuthStatus struct {
	LoggedIn   bool   `json:"logged_in"`
	Username   string `json:"username"`
	ServerType string `json:"server_type"` // "default" | "github"
}

// Namespace 命名空间
type Namespace struct {
	Name        string   `json:"name"`
	Owners      []string `json:"owners"`
	Maintainers []string `json:"maintainers"`
	Status      string   `json:"status"` // "pending" | "active" | "deprecated"
	CreatedAt   string   `json:"created_at"`
	Description string   `json:"description"`
}

// Package 包信息
type Package struct {
	Name        string        `json:"name"`
	Namespace   string        `json:"namespace"`
	Versions    []*Version    `json:"versions"`
	Description string        `json:"description"`
	License     string        `json:"license"`
}

// Version 版本信息
type Version struct {
	Version     string `json:"version"`
	ArchiveURL  string `json:"archive_url"`
	SHA256      string `json:"sha256"`
	SHA256URL   string `json:"sha256_url"`
	ReleaseDate string `json:"release_date"`
}

// credentials 本地凭证存储（公共类型）
type credentials struct {
	Token     string `yaml:"token"`
	Username  string `yaml:"username"`
	ExpiresAt string `yaml:"expires_at"`
}

// ServerType 服务端类型
type ServerType string

const (
	ServerTypeDefault ServerType = "default"
	ServerTypeGitHub  ServerType = "github"
)

// DetectServerType 检测服务端类型
// 由于采用 CLI 直连 GitHub 模式，默认返回 GitHub 类型
func DetectServerType(serverURL string) (ServerType, error) {
	// CLI 直连 GitHub 模式，无需服务端
	return ServerTypeGitHub, nil
}

// NewAdapter 根据服务端类型创建适配器
func NewAdapter(serverURL string, configPath string) (DehubServer, error) {
	serverType, err := DetectServerType(serverURL)
	if err != nil {
		return nil, err
	}
	
	switch serverType {
	case ServerTypeGitHub:
		return NewDehubServerGithubAdapter(serverURL, configPath), nil
	default:
		return NewDefaultAdapter(serverURL, configPath)
	}
}
