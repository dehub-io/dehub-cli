package cmd

import (
	"bytes"
	"testing"

	"github.com/dehub-io/dehub-cli/adapter"
)

func TestAuthCommandsExist(t *testing.T) {
	// 检查 auth 命令及其子命令存在
	authCmd, _, err := rootCmd.Find([]string{"auth"})
	if err != nil {
		t.Fatal("auth command should exist")
	}
	
	subCommands := authCmd.Commands()
	cmdNames := make(map[string]bool)
	for _, cmd := range subCommands {
		cmdNames[cmd.Name()] = true
	}
	
	expected := []string{"login", "logout", "whoami"}
	for _, name := range expected {
		if !cmdNames[name] {
			t.Errorf("auth command should have '%s' subcommand", name)
		}
	}
}

func TestAuthLoginSuccess(t *testing.T) {
	ResetAdapter()
	mock := adapter.NewMockAdapter()
	SetAdapter(mock)
	
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"auth", "login"})
	
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("auth login error: %v", err)
	}
}

func TestAuthLogoutSuccess(t *testing.T) {
	ResetAdapter()
	mock := adapter.NewMockAdapter()
	mock.AuthStatus = &adapter.AuthStatus{LoggedIn: true, Username: "test-user"}
	SetAdapter(mock)
	
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"auth", "logout"})
	
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("auth logout error: %v", err)
	}
}

func TestAuthWhoamiLoggedIn(t *testing.T) {
	ResetAdapter()
	mock := adapter.NewMockAdapter()
	mock.AuthStatus = &adapter.AuthStatus{LoggedIn: true, Username: "test-user"}
	SetAdapter(mock)
	
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"auth", "whoami"})
	
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("auth whoami error: %v", err)
	}
	
	output := buf.String()
	if output == "" {
		t.Error("whoami should produce output")
	}
}

func TestAuthWhoamiNotLoggedIn(t *testing.T) {
	ResetAdapter()
	mock := adapter.NewMockAdapter()
	mock.AuthStatus = &adapter.AuthStatus{LoggedIn: false}
	SetAdapter(mock)
	
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"auth", "whoami"})
	
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("auth whoami error: %v", err)
	}
	
	output := buf.String()
	if output == "" {
		t.Error("whoami should produce output")
	}
}
