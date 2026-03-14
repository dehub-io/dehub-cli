package cmd

import (
	"bytes"
	"testing"

	"github.com/dehub-io/dehub-cli/adapter"
)

func TestNamespaceCommandExists(t *testing.T) {
	nsCmd, _, err := rootCmd.Find([]string{"namespace"})
	if err != nil {
		t.Fatal("namespace command should exist")
	}
	if nsCmd == nil {
		t.Fatal("namespace command should not be nil")
	}
	
	// 检查子命令
	subCommands := nsCmd.Commands()
	cmdNames := make(map[string]bool)
	for _, cmd := range subCommands {
		cmdNames[cmd.Name()] = true
	}
	
	expected := []string{"create", "list"}
	for _, name := range expected {
		if !cmdNames[name] {
			t.Errorf("namespace command should have '%s' subcommand", name)
		}
	}
}

func TestNamespaceCreateSuccess(t *testing.T) {
	ResetAdapter()
	mock := adapter.NewMockAdapter()
	mock.AuthStatus = &adapter.AuthStatus{LoggedIn: true, Username: "test-user"}
	SetAdapter(mock)
	
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"namespace", "create", "test-ns"})
	
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("namespace create error: %v", err)
	}
}

func TestNamespaceListSuccess(t *testing.T) {
	ResetAdapter()
	mock := adapter.NewMockAdapter()
	mock.Namespaces = []*adapter.Namespace{
		{Name: "ns1", Status: "active"},
		{Name: "ns2", Status: "active"},
	}
	SetAdapter(mock)
	
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"namespace", "list"})
	
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("namespace list error: %v", err)
	}
}

func TestNamespaceListEmpty(t *testing.T) {
	ResetAdapter()
	mock := adapter.NewMockAdapter()
	mock.Namespaces = []*adapter.Namespace{}
	SetAdapter(mock)
	
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"namespace", "list"})
	
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("namespace list error: %v", err)
	}
}
