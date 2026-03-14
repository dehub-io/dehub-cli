package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/dehub-io/dehub-cli/adapter"
)

func TestInstallCommandExists(t *testing.T) {
	installCmd, _, err := rootCmd.Find([]string{"install"})
	if err != nil {
		t.Fatal("install command should exist")
	}
	if installCmd == nil {
		t.Fatal("install command should not be nil")
	}
	if installCmd.Short == "" {
		t.Error("install command should have a short description")
	}
}

func TestInstallCommandFlags(t *testing.T) {
	installCmd, _, _ := rootCmd.Find([]string{"install"})
	
	verboseFlag := installCmd.Flags().Lookup("verbose")
	if verboseFlag == nil {
		t.Error("install command should have --verbose flag")
	}
}

func TestInstallCommandWithConfig(t *testing.T) {
	// 创建临时目录和配置文件
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	// 创建 package.yaml
	configContent := `name: test-project
version: "0.1.0"
dependencies: {}
`
	if err := os.WriteFile("package.yaml", []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	ResetAdapter()
	mock := adapter.NewMockAdapter()
	SetAdapter(mock)
	
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"install"})
	
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("install command error: %v", err)
	}
}

func TestInstallCommandWithEmptyDeps(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	// 无依赖的配置
	configContent := `name: test-project
version: "0.1.0"
dependencies: {}
`
	if err := os.WriteFile("package.yaml", []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	ResetAdapter()
	mock := adapter.NewMockAdapter()
	SetAdapter(mock)
	
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"install"})
	
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("install command error: %v", err)
	}
}

func TestResolvePackageNotFound(t *testing.T) {
	repos := []Repository{{Name: "test", URL: "https://example.com/"}}
	
	_, _, err := resolvePackage(repos, "nonexistent/pkg", "1.0.0")
	if err == nil {
		t.Error("expected error for nonexistent package")
	}
}

func TestDownloadPackageInvalidURL(t *testing.T) {
	tmpDir := t.TempDir()
	
	_, err := downloadPackage(tmpDir, "://invalid-url")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestFetchSHA256InvalidURL(t *testing.T) {
	_, err := fetchSHA256("://invalid-url")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestVerifySHA256FileNotExist(t *testing.T) {
	err := verifySHA256("/nonexistent/file.tar.gz", "abc123")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestExtractArchiveInvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	invalidFile := tmpDir + "/invalid.tar.gz"
	os.WriteFile(invalidFile, []byte("not a tar.gz"), 0644)
	
	err := extractArchive(invalidFile, tmpDir+"/output")
	if err == nil {
		t.Error("expected error for invalid archive")
	}
}
