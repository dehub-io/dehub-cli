package adapter

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
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	"gopkg.in/yaml.v3"
)

// GitHub 配置常量
const (
	GitHubClientID    = "Ov23liz0ObyD0A0S2yQk"
	GitHubRepository  = "dehub-io/dehub-repository"
	GitHubIndexURL    = "https://dehub-io.github.io/dehub-repository"
	GitHubRawURL      = "https://raw.githubusercontent.com/dehub-io/dehub-repository/main"
)

// DehubServerGithubAdapter CLI 直连 GitHub 适配器
type DehubServerGithubAdapter struct {
	token      string
	configPath string
	httpClient *http.Client
}

// NewDehubServerGithubAdapter 创建适配器
func NewDehubServerGithubAdapter(serverURL, configPath string) *DehubServerGithubAdapter {
	return &DehubServerGithubAdapter{
		configPath: configPath,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// DeviceCodeResponse Device Flow 响应
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// TokenResponse Token 响应
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error"`
}

// Login 登录（Device Flow - GitHub 官方推荐的 CLI 方案）
func (a *DehubServerGithubAdapter) Login() error {
	client := &http.Client{Timeout: 30 * time.Second}

	// 1. 获取 Device Code
	fmt.Println("正在获取设备码...")
	var resp *http.Response
	var err error

	for i := 0; i < 10; i++ {
		resp, err = client.PostForm("https://github.com/login/device/code", url.Values{
			"client_id": {GitHubClientID},
			"scope":     {"repo"},
		})
		if err == nil {
			break
		}
		fmt.Printf("网络请求失败，重试 %d/10...\n", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return fmt.Errorf("获取设备码失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	values, _ := url.ParseQuery(string(body))

	deviceCode := values.Get("device_code")
	userCode := values.Get("user_code")
	verificationURI := values.Get("verification_uri")
	expiresIn := parseIntValue(values.Get("expires_in"))
	interval := parseIntValue(values.Get("interval"))
	if interval == 0 {
		interval = 5
	}

	// 2. 显示验证码并打开浏览器
	fmt.Println("\n========================================")
	if copyToClipboard(userCode) {
		fmt.Printf("验证码: %s（已复制到剪贴板）\n", userCode)
	} else {
		fmt.Printf("验证码: %s\n", userCode)
	}
	fmt.Println("========================================")

	openBrowser(verificationURI)
	fmt.Println("浏览器已打开，请在页面粘贴验证码 (Ctrl+V / Cmd+V)")

	// 3. 轮询等待授权
	for i := 0; i < expiresIn/interval; i++ {
		time.Sleep(time.Duration(interval) * time.Second)

		tokenResp, err := client.PostForm("https://github.com/login/oauth/access_token", url.Values{
			"client_id":   {GitHubClientID},
			"device_code": {deviceCode},
			"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
		})
		if err != nil {
			continue
		}

		body, _ := io.ReadAll(tokenResp.Body)
		tokenResp.Body.Close()

		tokenValues, _ := url.ParseQuery(string(body))
		accessToken := tokenValues.Get("access_token")
		errCode := tokenValues.Get("error")

		if accessToken != "" {
			a.token = accessToken

			// 获取用户信息
			username, _ := a.getGitHubUser()

			// 保存凭证
			cred := &credentials{
				Token:     a.token,
				Username:  username,
				ExpiresAt: time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339),
			}
			a.saveCredentials(cred)

			fmt.Println("登录成功！")
			return nil
		}

		switch errCode {
		case "authorization_pending":
			fmt.Printf("等待授权中... (%d/%d 秒)\n", (i+1)*interval, expiresIn)
		case "slow_down":
			interval += 5
			fmt.Printf("请求过快，等待间隔调整为 %d 秒\n", interval)
		case "expired_token":
			return fmt.Errorf("验证码已过期，请重新登录")
		case "access_denied":
			return fmt.Errorf("授权被拒绝")
		case "":
			// 无错误但也没有 token，继续等待
		default:
			return fmt.Errorf("登录失败: %s", errCode)
		}
	}

	return fmt.Errorf("登录超时")
}

// Logout 登出
func (a *DehubServerGithubAdapter) Logout() error {
	a.token = ""
	home, _ := os.UserHomeDir()
	credPath := filepath.Join(home, ".dehub", "credentials.yaml")
	os.Remove(credPath)
	return nil
}

// GetAuthStatus 获取认证状态
func (a *DehubServerGithubAdapter) GetAuthStatus() (*AuthStatus, error) {
	// 尝试加载已保存的凭证
	if a.token == "" {
		cred, err := a.loadCredentials()
		if err == nil && cred.Token != "" {
			a.token = cred.Token
		}
	}

	if a.token == "" {
		return &AuthStatus{LoggedIn: false, ServerType: "github"}, nil
	}

	username, err := a.getGitHubUser()
	if err != nil {
		return &AuthStatus{LoggedIn: false, ServerType: "github"}, nil
	}

	return &AuthStatus{
		LoggedIn:   true,
		Username:   username,
		ServerType: "github",
	}, nil
}

func (a *DehubServerGithubAdapter) getGitHubUser() (string, error) {
	req, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	req.Header.Set("Authorization", "Bearer "+a.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var user struct {
		Login string `json:"login"`
	}
	json.NewDecoder(resp.Body).Decode(&user)
	return user.Login, nil
}

// CreateNamespace 创建命名空间（创建 GitHub Issue）
func (a *DehubServerGithubAdapter) CreateNamespace(name, description string) (*Namespace, error) {
	if a.token == "" {
		cred, _ := a.loadCredentials()
		if cred != nil {
			a.token = cred.Token
		}
	}

	issueBody := map[string]interface{}{
		"title": fmt.Sprintf("命名空间申请: %s", name),
		"body": fmt.Sprintf(`## 命名空间申请

**命名空间名称**: %s

**用途说明**: %s

---
此 Issue 由 dehub CLI 自动创建。`, name, description),
		"labels": []string{"namespace-request"},
	}
	jsonBody, _ := json.Marshal(issueBody)

	req, _ := http.NewRequest("POST", fmt.Sprintf("https://api.github.com/repos/%s/issues", GitHubRepository), bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+a.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("创建 Issue 失败: HTTP %d", resp.StatusCode)
	}

	var issue struct {
		HTMLURL string `json:"html_url"`
	}
	json.NewDecoder(resp.Body).Decode(&issue)

	fmt.Printf("命名空间申请已创建: %s\n", issue.HTMLURL)

	return &Namespace{
		Name:        name,
		Status:      "pending",
		Description: description,
	}, nil
}

// GetNamespace 获取命名空间
func (a *DehubServerGithubAdapter) GetNamespace(name string) (*Namespace, error) {
	perms, err := a.fetchPermissions()
	if err != nil {
		return nil, err
	}

	cfg, ok := perms.Namespaces[name]
	if !ok {
		return nil, fmt.Errorf("命名空间不存在: %s", name)
	}

	return &Namespace{
		Name:        name,
		Owners:      cfg.Owners,
		Maintainers: cfg.Maintainers,
		Status:      cfg.Status,
		CreatedAt:   cfg.CreatedAt,
		Description: cfg.Description,
	}, nil
}

// ListNamespaces 列出命名空间
func (a *DehubServerGithubAdapter) ListNamespaces() ([]*Namespace, error) {
	perms, err := a.fetchPermissions()
	if err != nil {
		return nil, err
	}

	var namespaces []*Namespace
	for name, cfg := range perms.Namespaces {
		namespaces = append(namespaces, &Namespace{
			Name:        name,
			Owners:      cfg.Owners,
			Maintainers: cfg.Maintainers,
			Status:      cfg.Status,
			CreatedAt:   cfg.CreatedAt,
			Description: cfg.Description,
		})
	}

	return namespaces, nil
}

// PermissionsConfig 权限配置
type PermissionsConfig struct {
	Namespaces map[string]NamespaceConfig `yaml:"namespaces"`
}

// NamespaceConfig 命名空间配置
type NamespaceConfig struct {
	Owners      []string `yaml:"owners"`
	Maintainers []string `yaml:"maintainers"`
	Status      string   `yaml:"status"`
	CreatedAt   string   `yaml:"created_at"`
	Description string   `yaml:"description"`
}

func (a *DehubServerGithubAdapter) fetchPermissions() (*PermissionsConfig, error) {
	resp, err := http.Get(GitHubRawURL + "/permissions.yaml")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)

	var perms PermissionsConfig
	if err := yaml.Unmarshal(data, &perms); err != nil {
		return nil, err
	}

	return &perms, nil
}

// Publish 发布包（CLI 直连 GitHub API）
func (a *DehubServerGithubAdapter) Publish(pkgPath string, progress io.Writer) error {
	// 加载凭证
	if a.token == "" {
		cred, err := a.loadCredentials()
		if err != nil {
			return fmt.Errorf("请先登录: dehub login")
		}
		a.token = cred.Token
	}

	// 读取 package.yaml
	cfg, err := a.readPackageConfig(pkgPath)
	if err != nil {
		return fmt.Errorf("读取配置失败: %w", err)
	}

	fmt.Fprintf(progress, "包名: %s\n", cfg.Name)
	fmt.Fprintf(progress, "版本: %s\n\n", cfg.Version)

	// 解析命名空间
	parts := strings.Split(cfg.Name, "/")
	if len(parts) != 2 {
		return fmt.Errorf("无效的包名格式，应为 namespace/name")
	}
	namespace := parts[0]

	// 检查权限
	perms, err := a.fetchPermissions()
	if err != nil {
		return fmt.Errorf("获取权限配置失败: %w", err)
	}

	nsConfig, ok := perms.Namespaces[namespace]
	if !ok || nsConfig.Status != "active" {
		return fmt.Errorf("命名空间 '%s' 不存在或未激活", namespace)
	}

	username, _ := a.getGitHubUser()
	hasPermission := false
	for _, owner := range nsConfig.Owners {
		if owner == username {
			hasPermission = true
			break
		}
	}
	for _, maintainer := range nsConfig.Maintainers {
		if maintainer == username {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		return fmt.Errorf("无权限发布到命名空间 '%s'", namespace)
	}

	// 打包源码
	fmt.Fprintf(progress, "打包源码... ")
	archivePath, err := a.createArchive(pkgPath)
	if err != nil {
		return fmt.Errorf("打包失败: %w", err)
	}
	defer os.Remove(archivePath)
	fmt.Fprintf(progress, "✓\n")

	// 计算 SHA256
	fmt.Fprintf(progress, "计算 SHA256... ")
	sha256Hash, err := a.calculateSHA256(archivePath)
	if err != nil {
		return fmt.Errorf("计算校验失败: %w", err)
	}
	fmt.Fprintf(progress, "✓\n")

	// 创建 Draft Release
	tag := fmt.Sprintf("%s/%s", cfg.Name, cfg.Version)
	fmt.Fprintf(progress, "创建 Draft Release... ")
	releaseID, err := a.createDraftRelease(tag, cfg.Name, cfg.Version)
	if err != nil {
		return fmt.Errorf("创建 Release 失败: %w", err)
	}
	fmt.Fprintf(progress, "✓ (ID: %d)\n", releaseID)

	// 上传包文件
	fmt.Fprintf(progress, "上传 package.tar.gz... ")
	if err := a.uploadReleaseAsset(releaseID, "package.tar.gz", archivePath); err != nil {
		a.deleteRelease(releaseID)
		return fmt.Errorf("上传失败: %w", err)
	}
	fmt.Fprintf(progress, "✓\n")

	// 上传 SHA256 文件
	fmt.Fprintf(progress, "上传 sha256... ")
	sha256Content := fmt.Sprintf("%s  package.tar.gz\n", sha256Hash)
	if err := a.uploadReleaseAssetFromBytes(releaseID, "sha256", []byte(sha256Content)); err != nil {
		a.deleteRelease(releaseID)
		return fmt.Errorf("上传 SHA256 失败: %w", err)
	}
	fmt.Fprintf(progress, "✓\n")

	// 触发验证 workflow
	fmt.Fprintf(progress, "触发验证 workflow... ")
	runID, err := a.triggerVerifyWorkflow(releaseID, cfg.Name, cfg.Version)
	if err != nil {
		fmt.Fprintf(progress, "⚠ (可能需要手动触发)\n")
		fmt.Fprintf(progress, "\n发布请求已提交！\n")
		fmt.Fprintf(progress, "请手动检查: https://github.com/%s/actions\n", GitHubRepository)
		return nil
	}
	fmt.Fprintf(progress, "✓ (Run ID: %d)\n\n", runID)

	// 轮询 workflow 状态
	fmt.Fprintf(progress, "等待验证完成... ")
	result, err := a.waitForWorkflow(runID, progress)
	if err != nil {
		fmt.Fprintf(progress, "⚠ 轮询失败: %v\n", err)
		fmt.Fprintf(progress, "请手动检查: https://github.com/%s/actions/%d\n", GitHubRepository, runID)
		return nil
	}

	if result == "success" {
		fmt.Fprintf(progress, "\n✅ 发布成功！\n")
		fmt.Fprintf(progress, "包: %s@%s\n", cfg.Name, cfg.Version)
		fmt.Fprintf(progress, "下载: https://github.com/%s/releases/tag/%s\n", GitHubRepository, tag)
		return nil
	}

	// 失败，获取错误原因
	fmt.Fprintf(progress, "\n❌ 发布失败！\n")
	failReason := a.getWorkflowFailureReason(runID)
	if failReason != "" {
		fmt.Fprintf(progress, "原因: %s\n", failReason)
	}
	fmt.Fprintf(progress, "详情: https://github.com/%s/actions/runs/%d\n", GitHubRepository, runID)
	return fmt.Errorf("workflow 执行失败: %s", result)
}

// PackageConfig 包配置
type PackageConfig struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

func (a *DehubServerGithubAdapter) readPackageConfig(pkgPath string) (*PackageConfig, error) {
	data, err := os.ReadFile(filepath.Join(pkgPath, "package.yaml"))
	if err != nil {
		return nil, err
	}

	var cfg PackageConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (a *DehubServerGithubAdapter) createArchive(dir string) (string, error) {
	tmpFile, err := os.CreateTemp("", "dehub-*.tar.gz")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	gzw := gzip.NewWriter(tmpFile)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	extensions := map[string]bool{
		".yaml": true, ".yml": true,
		".c": true, ".h": true,
		".cpp": true, ".hpp": true,
		".s": true, ".S": true,
		".md": true, ".txt": true,
	}

	dirs := []string{"src", "include", "lib", "inc", "examples"}

	for _, d := range dirs {
		dirPath := filepath.Join(dir, d)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			continue
		}
		filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			relPath, _ := filepath.Rel(dir, path)
			return a.addFileToTar(tw, dir, relPath)
		})
	}

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if extensions[ext] {
			relPath, _ := filepath.Rel(dir, path)
			return a.addFileToTar(tw, dir, relPath)
		}
		return nil
	})

	return tmpFile.Name(), nil
}

func (a *DehubServerGithubAdapter) addFileToTar(tw *tar.Writer, baseDir, relPath string) error {
	path := filepath.Join(baseDir, relPath)
	info, err := os.Stat(path)
	if err != nil {
		return err
	}

	header, _ := tar.FileInfoHeader(info, "")
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

func (a *DehubServerGithubAdapter) calculateSHA256(filePath string) (string, error) {
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

func (a *DehubServerGithubAdapter) createDraftRelease(tag, pkgName, version string) (int64, error) {
	body := map[string]interface{}{
		"tag_name":   tag,
		"name":       tag,
		"draft":      true,
		"prerelease": false,
		"body":       fmt.Sprintf("Package: %s\nVersion: %s", pkgName, version),
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", fmt.Sprintf("https://api.github.com/repos/%s/releases", GitHubRepository), bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+a.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("GitHub API: %s", string(respBody))
	}

	var release struct {
		ID int64 `json:"id"`
	}
	json.NewDecoder(resp.Body).Decode(&release)
	return release.ID, nil
}

func (a *DehubServerGithubAdapter) uploadReleaseAsset(releaseID int64, filename, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, _ := file.Stat()

	req, _ := http.NewRequest("POST", fmt.Sprintf("https://uploads.github.com/repos/%s/releases/%d/assets?name=%s", GitHubRepository, releaseID, filename), file)
	req.Header.Set("Authorization", "Bearer "+a.token)
	req.Header.Set("Content-Type", "application/gzip")
	req.ContentLength = stat.Size()

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub upload: %s", string(respBody))
	}

	return nil
}

func (a *DehubServerGithubAdapter) uploadReleaseAssetFromBytes(releaseID int64, filename string, content []byte) error {
	req, _ := http.NewRequest("POST", fmt.Sprintf("https://uploads.github.com/repos/%s/releases/%d/assets?name=%s", GitHubRepository, releaseID, filename), bytes.NewReader(content))
	req.Header.Set("Authorization", "Bearer "+a.token)
	req.Header.Set("Content-Type", "text/plain")
	req.ContentLength = int64(len(content))

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub upload: %s", string(respBody))
	}

	return nil
}

func (a *DehubServerGithubAdapter) deleteRelease(releaseID int64) {
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("https://api.github.com/repos/%s/releases/%d", GitHubRepository, releaseID), nil)
	req.Header.Set("Authorization", "Bearer "+a.token)
	a.httpClient.Do(req)
}

func (a *DehubServerGithubAdapter) triggerVerifyWorkflow(releaseID int64, pkgName, version string) (int64, error) {
	body := map[string]interface{}{
		"ref": "main",
		"inputs": map[string]string{
			"release_id":   fmt.Sprintf("%d", releaseID),
			"package_name": pkgName,
			"version":      version,
		},
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", fmt.Sprintf("https://api.github.com/repos/%s/actions/workflows/verify-publish.yml/dispatches", GitHubRepository), bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+a.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("workflow dispatch: %s", string(respBody))
	}

	// 获取最新的 workflow run
	time.Sleep(2 * time.Second)
	req2, _ := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/actions/workflows/verify-publish.yml/runs?per_page=1", GitHubRepository), nil)
	req2.Header.Set("Authorization", "Bearer "+a.token)
	req2.Header.Set("Accept", "application/vnd.github.v3+json")

	resp2, err := a.httpClient.Do(req2)
	if err != nil {
		return 0, fmt.Errorf("获取 workflow 运行状态失败: %w", err)
	}
	defer resp2.Body.Close()

	var result struct {
		WorkflowRuns []struct {
			ID int64 `json:"id"`
		} `json:"workflow_runs"`
	}
	json.NewDecoder(resp2.Body).Decode(&result)

	if len(result.WorkflowRuns) > 0 {
		return result.WorkflowRuns[0].ID, nil
	}

	return 0, nil
}

// waitForWorkflow 轮询等待 workflow 完成
func (a *DehubServerGithubAdapter) waitForWorkflow(runID int64, progress io.Writer) (string, error) {
	maxWait := 5 * time.Minute
	interval := 5 * time.Second
	start := time.Now()

	for time.Since(start) < maxWait {
		req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/actions/runs/%d", GitHubRepository, runID), nil)
		req.Header.Set("Authorization", "Bearer "+a.token)
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		resp, err := a.httpClient.Do(req)
		if err != nil {
			time.Sleep(interval)
			continue
		}

		var run struct {
			Status     string `json:"status"`
			Conclusion string `json:"conclusion"`
		}
		json.NewDecoder(resp.Body).Decode(&run)
		resp.Body.Close()

		// status: queued, in_progress, completed
		// conclusion: success, failure, cancelled, timed_out, action_required
		if run.Status == "completed" {
			return run.Conclusion, nil
		}

		fmt.Fprintf(progress, ".")
		time.Sleep(interval)
	}

	return "", fmt.Errorf("等待超时")
}

// getWorkflowFailureReason 获取 workflow 失败原因
func (a *DehubServerGithubAdapter) getWorkflowFailureReason(runID int64) string {
	// 获取失败 job
	req, _ := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/actions/runs/%d/jobs?status=failed", GitHubRepository, runID), nil)
	req.Header.Set("Authorization", "Bearer "+a.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var result struct {
		Jobs []struct {
			Name   string `json:"name"`
			Steps  []struct {
				Name       string `json:"name"`
				Conclusion string `json:"conclusion"`
			} `json:"steps"`
		} `json:"jobs"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if len(result.Jobs) > 0 {
		job := result.Jobs[0]
		for _, step := range job.Steps {
			if step.Conclusion == "failure" {
				return fmt.Sprintf("Step '%s' failed", step.Name)
			}
		}
		return job.Name
	}

	return ""
}

// ListPackages 列出包（从 GitHub Pages）
func (a *DehubServerGithubAdapter) ListPackages(namespace string) ([]*Package, error) {
	resp, err := http.Get(GitHubIndexURL + "/index.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var index struct {
		Packages []string `json:"packages"`
	}
	json.NewDecoder(resp.Body).Decode(&index)

	var packages []*Package
	for _, pkgName := range index.Packages {
		if namespace != "" && !strings.HasPrefix(pkgName, namespace+"/") {
			continue
		}
		packages = append(packages, &Package{Name: pkgName})
	}

	return packages, nil
}

// GetPackage 获取包信息
func (a *DehubServerGithubAdapter) GetPackage(name string) (*Package, error) {
	resp, err := http.Get(fmt.Sprintf("%s/packages/%s/index.json", GitHubIndexURL, name))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("包不存在: %s", name)
	}

	var pkgIndex struct {
		Package  string `json:"package"`
		Versions []struct {
			Version     string `json:"version"`
			ArchiveURL  string `json:"archive_url"`
			ReleaseDate string `json:"release_date"`
		} `json:"versions"`
	}
	json.NewDecoder(resp.Body).Decode(&pkgIndex)

	pkg := &Package{Name: pkgIndex.Package}
	for _, v := range pkgIndex.Versions {
		// 构造 SHA256 URL
		sha256URL := strings.Replace(v.ArchiveURL, "package.tar.gz", "sha256", 1)
		pkg.Versions = append(pkg.Versions, &Version{
			Version:     v.Version,
			ArchiveURL:  v.ArchiveURL,
			SHA256URL:   sha256URL,
			ReleaseDate: v.ReleaseDate,
		})
	}

	return pkg, nil
}

// Install 安装依赖
func (a *DehubServerGithubAdapter) Install(dependencies map[string]string, targetDir string) error {
	if targetDir == "" {
		targetDir = ".dehub/deps"
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	cacheDir := getCacheDir()
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("创建缓存目录失败: %w", err)
	}

	for pkgName, versionConstraint := range dependencies {
		fmt.Printf("解析: %s@%s\n", pkgName, versionConstraint)

		pkg, err := a.GetPackage(pkgName)
		if err != nil {
			fmt.Printf("  获取包信息失败: %v\n", err)
			continue
		}

		var version *Version
		if versionConstraint == "" || versionConstraint == "latest" {
			if len(pkg.Versions) > 0 {
				version = pkg.Versions[0]
			}
		} else {
			for _, v := range pkg.Versions {
				if v.Version == versionConstraint {
					version = v
					break
				}
			}
		}

		if version == nil {
			fmt.Printf("  未找到版本: %s\n", versionConstraint)
			continue
		}

		fmt.Printf("安装: %s@%s\n", pkgName, version.Version)

		// 下载包
		archivePath, err := a.downloadPackage(cacheDir, version.ArchiveURL)
		if err != nil {
			fmt.Printf("  下载失败: %v\n", err)
			continue
		}

		// 获取并校验 SHA256
		if version.SHA256URL != "" {
			sha256Hash, err := a.fetchSHA256(version.SHA256URL)
			if err == nil {
				if err := a.verifySHA256(archivePath, sha256Hash); err != nil {
					fmt.Printf("  校验失败: %v\n", err)
					os.Remove(archivePath)
					continue
				}
			}
		}

		targetPath := filepath.Join(targetDir, pkgName, version.Version)
		if err := a.extractArchive(archivePath, targetPath); err != nil {
			fmt.Printf("  解压失败: %v\n", err)
			continue
		}

		fmt.Printf("  -> %s\n", targetPath)
	}

	return nil
}

func getCacheDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".dehub", "cache")
}

func (a *DehubServerGithubAdapter) downloadPackage(cacheDir, downloadURL string) (string, error) {
	fileName := filepath.Base(downloadURL)
	cachePath := filepath.Join(cacheDir, fileName)

	if _, err := os.Stat(cachePath); err == nil {
		return cachePath, nil
	}

	resp, err := http.Get(downloadURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(cachePath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return cachePath, err
}

func (a *DehubServerGithubAdapter) fetchSHA256(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	parts := strings.Fields(string(data))
	if len(parts) > 0 {
		return parts[0], nil
	}
	return "", fmt.Errorf("无效的 SHA256 格式")
}

func (a *DehubServerGithubAdapter) verifySHA256(filePath, expectedHash string) error {
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
		return fmt.Errorf("SHA256 不匹配")
	}

	return nil
}

func (a *DehubServerGithubAdapter) extractArchive(archivePath, targetDir string) error {
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

// Search 搜索包
func (a *DehubServerGithubAdapter) Search(query string) ([]*Package, error) {
	resp, err := http.Get(GitHubIndexURL + "/index.json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var index struct {
		Packages []string `json:"packages"`
	}
	json.NewDecoder(resp.Body).Decode(&index)

	query = strings.ToLower(query)
	var results []*Package
	for _, pkgName := range index.Packages {
		if strings.Contains(strings.ToLower(pkgName), query) {
			results = append(results, &Package{Name: pkgName})
		}
	}

	return results, nil
}

// saveCredentials 保存凭证
func (a *DehubServerGithubAdapter) saveCredentials(cred *credentials) error {
	home, _ := os.UserHomeDir()
	credDir := filepath.Join(home, ".dehub")
	if err := os.MkdirAll(credDir, 0700); err != nil {
		return err
	}

	credPath := filepath.Join(credDir, "credentials.yaml")
	data, _ := yaml.Marshal(cred)
	return os.WriteFile(credPath, data, 0600)
}

// loadCredentials 加载凭证
func (a *DehubServerGithubAdapter) loadCredentials() (*credentials, error) {
	home, _ := os.UserHomeDir()
	credPath := filepath.Join(home, ".dehub", "credentials.yaml")

	data, err := os.ReadFile(credPath)
	if err != nil {
		return nil, err
	}

	var cred credentials
	if err := yaml.Unmarshal(data, &cred); err != nil {
		return nil, err
	}

	return &cred, nil
}

// openBrowser 打开浏览器
func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}
	cmd.Start()
}

// copyToClipboard 复制文本到剪贴板
func copyToClipboard(text string) bool {
	return clipboard.WriteAll(text) == nil
}

// SetToken 设置 Token
func (a *DehubServerGithubAdapter) SetToken(token string) {
	a.token = token
}

// GetToken 获取 Token
func (a *DehubServerGithubAdapter) GetToken() string {
	return a.token
}

func parseIntValue(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}