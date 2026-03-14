package cmd

import (
	"bytes"
	"testing"

	"github.com/dehub-io/dehub-cli/adapter"
)

func TestPublishCommandExists(t *testing.T) {
	publishCmd, _, err := rootCmd.Find([]string{"publish"})
	if err != nil {
		t.Fatal("publish command should exist")
	}
	if publishCmd == nil {
		t.Fatal("publish command should not be nil")
	}
	if publishCmd.Short == "" {
		t.Error("publish command should have a short description")
	}
}

func TestPublishCommandFlags(t *testing.T) {
	publishCmd, _, _ := rootCmd.Find([]string{"publish"})
	
	dirFlag := publishCmd.Flags().Lookup("dir")
	if dirFlag == nil {
		t.Error("publish command should have --dir flag")
	}
}

func TestPublishSuccess(t *testing.T) {
	ResetAdapter()
	mock := adapter.NewMockAdapter()
	mock.AuthStatus = &adapter.AuthStatus{LoggedIn: true, Username: "test-user"}
	SetAdapter(mock)
	
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"publish", "."})
	
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("publish command error: %v", err)
	}
}
