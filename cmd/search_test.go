package cmd

import (
	"bytes"
	"testing"

	"github.com/dehub-io/dehub-cli/adapter"
)

func TestSearchCommandExists(t *testing.T) {
	searchCmd, _, err := rootCmd.Find([]string{"search"})
	if err != nil {
		t.Fatal("search command should exist")
	}
	if searchCmd == nil {
		t.Fatal("search command should not be nil")
	}
}

func TestSearchSuccess(t *testing.T) {
	ResetAdapter()
	mock := adapter.NewMockAdapter()
	mock.SearchResults = []*adapter.Package{
		{Name: "test/pkg1"},
		{Name: "test/pkg2"},
	}
	SetAdapter(mock)
	
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"search", "test"})
	
	_ = rootCmd.Execute()
}

func TestSearchNoResults(t *testing.T) {
	ResetAdapter()
	mock := adapter.NewMockAdapter()
	mock.SearchResults = []*adapter.Package{}
	SetAdapter(mock)
	
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"search", "nonexistent"})
	
	_ = rootCmd.Execute()
}

func TestSearchError(t *testing.T) {
	ResetAdapter()
	mock := adapter.NewMockAdapter()
	mock.SearchError = adapter.MockError("search failed")
	SetAdapter(mock)
	
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"search", "test"})
	
	_ = rootCmd.Execute()
}
