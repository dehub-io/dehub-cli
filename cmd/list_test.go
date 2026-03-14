package cmd

import (
	"bytes"
	"os"
	"testing"

	"github.com/dehub-io/dehub-cli/adapter"
)

func TestListCommandExists(t *testing.T) {
	listCmd, _, err := rootCmd.Find([]string{"list"})
	if err != nil {
		t.Fatal("list command should exist")
	}
	if listCmd == nil {
		t.Fatal("list command should not be nil")
	}
	if listCmd.Short == "" {
		t.Error("list command should have a short description")
	}
}

func TestListCommandWithConfig(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	// 创建 package.yaml
	configContent := `name: test-project
version: "0.1.0"
dependencies:
  test/pkg1: "1.0.0"
  test/pkg2: "2.0.0"
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
	rootCmd.SetArgs([]string{"list"})
	
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("list command error: %v", err)
	}
}

func TestListCommandWithLockFile(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	// 创建 package.yaml
	configContent := `name: test-project
version: "0.1.0"
dependencies:
  test/pkg: "1.0.0"
`
	if err := os.WriteFile("package.yaml", []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	// 创建 package.lock
	lockContent := `version: 1
dependencies:
  - name: test/pkg
    version: 1.0.0
    sha256: abc123def456
`
	if err := os.WriteFile("package.lock", []byte(lockContent), 0644); err != nil {
		t.Fatal(err)
	}

	ResetAdapter()
	mock := adapter.NewMockAdapter()
	SetAdapter(mock)
	
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"list"})
	
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("list command error: %v", err)
	}
}

func TestListCommandEmptyDeps(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldDir)

	// 创建空依赖的 package.yaml
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
	rootCmd.SetArgs([]string{"list"})
	
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("list command error: %v", err)
	}
}
