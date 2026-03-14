package cmd

import (
	"bytes"
	"testing"

	"github.com/dehub-io/dehub-cli/adapter"
)

func TestSetAndGetAdapter(t *testing.T) {
	// 重置
	ResetAdapter()
	
	// 设置 mock 适配器
	mock := adapter.NewMockAdapter()
	SetAdapter(mock)
	
	// 获取应该返回同一个适配器
	got := getAdapter()
	if got != mock {
		t.Error("getAdapter should return the same adapter")
	}
	
	// 清理
	ResetAdapter()
}

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
}

func TestRootCommand(t *testing.T) {
	if rootCmd == nil {
		t.Fatal("rootCmd should not be nil")
	}
	if rootCmd.Use != "dehub" {
		t.Errorf("rootCmd.Use = %s, want dehub", rootCmd.Use)
	}
	if rootCmd.Short != "源码包管理器" {
		t.Errorf("rootCmd.Short = %s", rootCmd.Short)
	}
}

func TestRootCmdHasSubCommands(t *testing.T) {
	commands := rootCmd.Commands()
	if len(commands) == 0 {
		t.Error("rootCmd should have sub commands")
	}
	
	// 检查关键命令存在
	cmdNames := make(map[string]bool)
	for _, cmd := range commands {
		cmdNames[cmd.Name()] = true
	}
	
	expected := []string{"auth", "init", "install", "list", "namespace", "publish", "search", "version"}
	for _, name := range expected {
		if !cmdNames[name] {
			t.Errorf("rootCmd should have '%s' command", name)
		}
	}
}

func TestRootCmdFlags(t *testing.T) {
	configFlag := rootCmd.PersistentFlags().Lookup("config")
	if configFlag == nil {
		t.Error("rootCmd should have --config flag")
	}
	if configFlag.DefValue != "package.yaml" {
		t.Errorf("config default = %s, want package.yaml", configFlag.DefValue)
	}
	
	serverFlag := rootCmd.PersistentFlags().Lookup("server")
	if serverFlag == nil {
		t.Error("rootCmd should have --server flag")
	}
	
	cacheFlag := rootCmd.PersistentFlags().Lookup("cache-dir")
	if cacheFlag == nil {
		t.Error("rootCmd should have --cache-dir flag")
	}
}

func TestHelpCommand(t *testing.T) {
	// 执行 help 命令
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"help"})
	
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("help command error: %v", err)
	}
	
	output := buf.String()
	if output == "" {
		t.Error("help command should produce output")
	}
}

func TestVersionFlag(t *testing.T) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"--version"})
	
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("version flag error: %v", err)
	}
	
	output := buf.String()
	if output == "" {
		t.Error("--version should produce output")
	}
}
