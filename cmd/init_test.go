package cmd

import (
	"testing"
)

func TestInitCommandExists(t *testing.T) {
	initCmd, _, err := rootCmd.Find([]string{"init"})
	if err != nil {
		t.Fatal("init command should exist")
	}
	if initCmd == nil {
		t.Fatal("init command should not be nil")
	}
	if initCmd.Short == "" {
		t.Error("init command should have a short description")
	}
}

func TestInitCommandUsage(t *testing.T) {
	initCmd, _, _ := rootCmd.Find([]string{"init"})
	
	// init 命令使用位置参数 [name]，不需要 flag
	if initCmd.Use != "init [name]" {
		t.Errorf("init command usage should be 'init [name]', got %s", initCmd.Use)
	}
	
	// 检查 Example 存在
	if initCmd.Example == "" {
		t.Error("init command should have examples")
	}
}
