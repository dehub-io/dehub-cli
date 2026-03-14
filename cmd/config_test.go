package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestProjectConfigStruct(t *testing.T) {
	cfg := ProjectConfig{
		Name:         "test/pkg",
		Version:      "1.0.0",
		Platform:     "esp32",
		Board:        "esp32c3",
		Dependencies: map[string]string{"ns/pkg": "1.0.0"},
		Repositories: []Repository{{Name: "central", URL: "https://example.com"}},
	}

	if cfg.Name != "test/pkg" {
		t.Errorf("ProjectConfig.Name = %s", cfg.Name)
	}
	if len(cfg.Dependencies) != 1 {
		t.Errorf("len(Dependencies) = %d", len(cfg.Dependencies))
	}
}

func TestLockFileStruct(t *testing.T) {
	lock := LockFile{
		Version: 1,
		Dependencies: []LockedPackage{
			{Name: "test/pkg", Version: "1.0.0", SHA256: "abc123"},
		},
	}

	if lock.Version != 1 {
		t.Errorf("LockFile.Version = %d", lock.Version)
	}
	if len(lock.Dependencies) != 1 {
		t.Errorf("len(Dependencies) = %d", len(lock.Dependencies))
	}
}

func TestIndexStruct(t *testing.T) {
	index := Index{
		SchemaVersion: "1.0",
		Packages:      []string{"ns/pkg1", "ns/pkg2"},
		UpdatedAt:     "2024-01-01",
	}

	if index.SchemaVersion != "1.0" {
		t.Errorf("Index.SchemaVersion = %s", index.SchemaVersion)
	}
	if len(index.Packages) != 2 {
		t.Errorf("len(Packages) = %d", len(index.Packages))
	}
}

func TestPackageIndexStruct(t *testing.T) {
	pkgIndex := PackageIndex{
		SchemaVersion: "1.0",
		Package:       "test/pkg",
		Versions: []VersionInfo{
			{Version: "1.0.0", ArchiveURL: "https://example.com/pkg.tar.gz"},
		},
	}

	if pkgIndex.Package != "test/pkg" {
		t.Errorf("PackageIndex.Package = %s", pkgIndex.Package)
	}
	if len(pkgIndex.Versions) != 1 {
		t.Errorf("len(Versions) = %d", len(pkgIndex.Versions))
	}
}

func TestGetCacheDir(t *testing.T) {
	// 测试默认缓存目录
	cacheDir := getCacheDir()
	if cacheDir == "" {
		t.Error("getCacheDir returned empty string")
	}
	// 应该包含 .dehub
	if filepath.Base(filepath.Dir(cacheDir)) != ".dehub" {
		t.Errorf("getCacheDir = %s, should be under .dehub", cacheDir)
	}
}

func TestLoadProjectConfig(t *testing.T) {
	// 创建临时目录和配置文件
	tmpDir, err := os.MkdirTemp("", "dehub-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建测试配置
	cfgContent := `
name: test/pkg
version: 1.0.0
platform: esp32
dependencies:
  ns/lib1: "1.0.0"
`
	cfgPath := filepath.Join(tmpDir, "package.yaml")
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0644); err != nil {
		t.Fatalf("写入配置失败: %v", err)
	}

	cfg, err := loadProjectConfig(cfgPath)
	if err != nil {
		t.Fatalf("loadProjectConfig error = %v", err)
	}

	if cfg.Name != "test/pkg" {
		t.Errorf("cfg.Name = %s, want test/pkg", cfg.Name)
	}
	if cfg.Version != "1.0.0" {
		t.Errorf("cfg.Version = %s, want 1.0.0", cfg.Version)
	}
	if cfg.Platform != "esp32" {
		t.Errorf("cfg.Platform = %s, want esp32", cfg.Platform)
	}
	if len(cfg.Dependencies) != 1 {
		t.Errorf("len(Dependencies) = %d", len(cfg.Dependencies))
	}
}

func TestLoadProjectConfigDefaultRepo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dehub-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfgContent := `
name: test/pkg
version: 1.0.0
`
	cfgPath := filepath.Join(tmpDir, "package.yaml")
	os.WriteFile(cfgPath, []byte(cfgContent), 0644)

	cfg, err := loadProjectConfig(cfgPath)
	if err != nil {
		t.Fatalf("loadProjectConfig error = %v", err)
	}

	// 应该有默认仓库
	if len(cfg.Repositories) != 1 {
		t.Errorf("len(Repositories) = %d, want 1", len(cfg.Repositories))
	}
	if cfg.Repositories[0].Name != "central" {
		t.Errorf("Repositories[0].Name = %s, want central", cfg.Repositories[0].Name)
	}
}

func TestLoadProjectConfigNotExist(t *testing.T) {
	_, err := loadProjectConfig("/nonexistent/path/package.yaml")
	if err == nil {
		t.Error("loadProjectConfig should return error for nonexistent file")
	}
}

func TestLoadProjectConfigInvalidYAML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dehub-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfgPath := filepath.Join(tmpDir, "package.yaml")
	os.WriteFile(cfgPath, []byte("invalid: yaml: content:"), 0644)

	_, err = loadProjectConfig(cfgPath)
	if err == nil {
		t.Error("loadProjectConfig should return error for invalid YAML")
	}
}

func TestSaveLockFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dehub-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfgPath := filepath.Join(tmpDir, "package.yaml")
	os.WriteFile(cfgPath, []byte("name: test\nversion: 1.0.0"), 0644)

	lock := &LockFile{
		Version: 1,
		Dependencies: []LockedPackage{
			{Name: "test/pkg", Version: "1.0.0", SHA256: "abc123"},
		},
	}

	err = saveLockFile(cfgPath, lock)
	if err != nil {
		t.Fatalf("saveLockFile error = %v", err)
	}

	// 验证文件已创建
	lockPath := filepath.Join(tmpDir, "package.lock")
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		t.Error("package.lock should exist")
	}

	// 验证内容
	data, _ := os.ReadFile(lockPath)
	var loaded LockFile
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("解析锁文件失败: %v", err)
	}

	if loaded.Version != 1 {
		t.Errorf("loaded.Version = %d, want 1", loaded.Version)
	}
}

func TestLoadLockFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dehub-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建 package.yaml 和 package.lock
	cfgPath := filepath.Join(tmpDir, "package.yaml")
	os.WriteFile(cfgPath, []byte("name: test"), 0644)

	lockContent := `
version: 1
dependencies:
  - name: test/pkg
    version: 1.0.0
    sha256: abc123
`
	lockPath := filepath.Join(tmpDir, "package.lock")
	os.WriteFile(lockPath, []byte(lockContent), 0644)

	lock, err := loadLockFile(cfgPath)
	if err != nil {
		t.Fatalf("loadLockFile error = %v", err)
	}

	if lock.Version != 1 {
		t.Errorf("lock.Version = %d, want 1", lock.Version)
	}
	if len(lock.Dependencies) != 1 {
		t.Errorf("len(Dependencies) = %d", len(lock.Dependencies))
	}
}

func TestLoadLockFileNotExist(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dehub-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfgPath := filepath.Join(tmpDir, "package.yaml")
	os.WriteFile(cfgPath, []byte("name: test"), 0644)

	_, err = loadLockFile(cfgPath)
	if err == nil {
		t.Error("loadLockFile should return error when lock file doesn't exist")
	}
}

func TestFindConfig(t *testing.T) {
	// 保存当前工作目录
	cwd, _ := os.Getwd()

	tmpDir, err := os.MkdirTemp("", "dehub-test-*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建嵌套目录结构
	nestedDir := filepath.Join(tmpDir, "a", "b", "c")
	os.MkdirAll(nestedDir, 0755)

	// 在根目录创建 package.yaml
	cfgPath := filepath.Join(tmpDir, "package.yaml")
	os.WriteFile(cfgPath, []byte("name: test"), 0644)

	// 切换到嵌套目录
	os.Chdir(nestedDir)
	defer os.Chdir(cwd)

	// 保存原始值
	originalCfgFile := cfgFile
	cfgFile = "package.yaml"
	defer func() { cfgFile = originalCfgFile }()

	found := findConfig()
	if found == "" {
		t.Error("findConfig should find package.yaml")
	}
}
