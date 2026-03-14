package adapter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// DefaultAdapter dehub-server 标准适配器
type DefaultAdapter struct {
	serverURL  string
	configPath string
	httpClient *http.Client
}

// NewDefaultAdapter 创建默认适配器
func NewDefaultAdapter(serverURL, configPath string) (*DefaultAdapter, error) {
	return &DefaultAdapter{
		serverURL:  serverURL,
		configPath: configPath,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Login 登录（用户名密码方式）
func (a *DefaultAdapter) Login() error {
	fmt.Println("请输入用户名: ")
	var username string
	fmt.Scanln(&username)
	
	fmt.Println("请输入密码: ")
	var password string
	fmt.Scanln(&password)
	
	// 调用服务端登录 API
	reqBody := map[string]string{
		"username": username,
		"password": password,
	}
	
	resp, err := a.post("/api/v1/auth/login", reqBody)
	if err != nil {
		return err
	}
	
	// 保存 token
	return a.saveToken(resp.Token)
}

// Logout 登出
func (a *DefaultAdapter) Logout() error {
	credPath := a.getCredentialsPath()
	return os.Remove(credPath)
}

// GetAuthStatus 获取认证状态
func (a *DefaultAdapter) GetAuthStatus() (*AuthStatus, error) {
	cred, err := a.loadCredentials()
	if err != nil {
		return &AuthStatus{LoggedIn: false}, nil
	}
	
	return &AuthStatus{
		LoggedIn:   cred.Token != "",
		Username:   cred.Username,
		ServerType: "default",
	}, nil
}

// CreateNamespace 创建命名空间
func (a *DefaultAdapter) CreateNamespace(name, description string) (*Namespace, error) {
	reqBody := map[string]string{
		"name":        name,
		"description": description,
	}
	
	resp, err := a.postAuth("/api/v1/namespaces", reqBody)
	if err != nil {
		return nil, err
	}
	
	var ns Namespace
	if err := json.Unmarshal(resp, &ns); err != nil {
		return nil, err
	}
	
	return &ns, nil
}

// ListNamespaces 列出命名空间
func (a *DefaultAdapter) ListNamespaces() ([]*Namespace, error) {
	resp, err := a.get("/api/v1/namespaces")
	if err != nil {
		return nil, err
	}
	
	var namespaces []*Namespace
	if err := json.Unmarshal(resp, &namespaces); err != nil {
		return nil, err
	}
	
	return namespaces, nil
}

// GetNamespace 获取命名空间信息
func (a *DefaultAdapter) GetNamespace(name string) (*Namespace, error) {
	resp, err := a.get("/api/v1/namespaces/" + name)
	if err != nil {
		return nil, err
	}
	
	var ns Namespace
	if err := json.Unmarshal(resp, &ns); err != nil {
		return nil, err
	}
	
	return &ns, nil
}

// Publish 发布包
func (a *DefaultAdapter) Publish(pkgPath string, progress io.Writer) error {
	// 读取 package.yaml
	configData, err := os.ReadFile(filepath.Join(pkgPath, "package.yaml"))
	if err != nil {
		return err
	}

	var cfg struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
	}
	if err := yaml.Unmarshal(configData, &cfg); err != nil {
		return err
	}

	return fmt.Errorf("默认适配器不支持发布功能，请使用 GitHub 模式")
}

// ListPackages 列出包
func (a *DefaultAdapter) ListPackages(namespace string) ([]*Package, error) {
	url := "/api/v1/packages"
	if namespace != "" {
		url += "?namespace=" + namespace
	}
	
	resp, err := a.get(url)
	if err != nil {
		return nil, err
	}
	
	var packages []*Package
	if err := json.Unmarshal(resp, &packages); err != nil {
		return nil, err
	}
	
	return packages, nil
}

// GetPackage 获取包信息
func (a *DefaultAdapter) GetPackage(name string) (*Package, error) {
	resp, err := a.get("/api/v1/packages/" + name)
	if err != nil {
		return nil, err
	}
	
	var pkg Package
	if err := json.Unmarshal(resp, &pkg); err != nil {
		return nil, err
	}
	
	return &pkg, nil
}

// Install 安装依赖
func (a *DefaultAdapter) Install(dependencies map[string]string, targetDir string) error {
	return fmt.Errorf("默认适配器不支持安装功能，请使用 GitHub 模式")
}

// Search 搜索包
func (a *DefaultAdapter) Search(query string) ([]*Package, error) {
	resp, err := a.get("/api/v1/search?q=" + query)
	if err != nil {
		return nil, err
	}
	
	var packages []*Package
	if err := json.Unmarshal(resp, &packages); err != nil {
		return nil, err
	}
	
	return packages, nil
}

// HTTP 辅助方法

func (a *DefaultAdapter) get(path string) ([]byte, error) {
	req, err := http.NewRequest("GET", a.serverURL+path, nil)
	if err != nil {
		return nil, err
	}
	
	cred, _ := a.loadCredentials()
	if cred != nil && cred.Token != "" {
		req.Header.Set("Authorization", "Bearer "+cred.Token)
	}
	
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	return io.ReadAll(resp.Body)
}

func (a *DefaultAdapter) post(path string, body interface{}) (*authResponse, error) {
	jsonBody, _ := json.Marshal(body)
	
	req, err := http.NewRequest("POST", a.serverURL+path, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var authResp authResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, err
	}
	
	return &authResp, nil
}

func (a *DefaultAdapter) postAuth(path string, body interface{}) ([]byte, error) {
	jsonBody, _ := json.Marshal(body)
	
	req, err := http.NewRequest("POST", a.serverURL+path, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	
	cred, _ := a.loadCredentials()
	if cred != nil && cred.Token != "" {
		req.Header.Set("Authorization", "Bearer "+cred.Token)
	}
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	return io.ReadAll(resp.Body)
}

type authResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
}

func (a *DefaultAdapter) getCredentialsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".dehub", "credentials.yaml")
}

func (a *DefaultAdapter) loadCredentials() (*credentials, error) {
	data, err := os.ReadFile(a.getCredentialsPath())
	if err != nil {
		return nil, err
	}
	
	var cred credentials
	if err := yaml.Unmarshal(data, &cred); err != nil {
		return nil, err
	}
	
	return &cred, nil
}

func (a *DefaultAdapter) saveToken(token string) error {
	credPath := a.getCredentialsPath()
	if err := os.MkdirAll(filepath.Dir(credPath), 0700); err != nil {
		return err
	}
	
	cred := credentials{Token: token}
	data, _ := yaml.Marshal(cred)
	
	return os.WriteFile(credPath, data, 0600)
}
