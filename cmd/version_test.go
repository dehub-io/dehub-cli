package cmd

import (
	"bytes"
	"testing"
)

func TestVersionCommandExists(t *testing.T) {
	versionCmd, _, err := rootCmd.Find([]string{"version"})
	if err != nil {
		t.Fatal("version command should exist")
	}
	if versionCmd == nil {
		t.Fatal("version command should not be nil")
	}
	if versionCmd.Short == "" {
		t.Error("version command should have a short description")
	}
}

func TestVersionCommand(t *testing.T) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"version"})
	
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("version command error: %v", err)
	}
	
	output := buf.String()
	if output == "" {
		t.Error("version command should produce output")
	}
}
