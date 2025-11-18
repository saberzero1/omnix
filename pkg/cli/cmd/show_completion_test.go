package cmd

import (
	"bytes"
	"testing"
)

func TestNewShowCmd(t *testing.T) {
	cmd := NewShowCmd()

	if cmd == nil {
		t.Fatal("NewShowCmd() returned nil")
	}

	if cmd.Use != "show [FLAKE]" {
		t.Errorf("expected Use 'show [FLAKE]', got '%s'", cmd.Use)
	}
}

func TestShowCommand_Help(t *testing.T) {
	cmd := NewShowCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected help output, got empty string")
	}
}

func TestNewCompletionCmd(t *testing.T) {
	cmd := NewCompletionCmd()

	if cmd == nil {
		t.Fatal("NewCompletionCmd() returned nil")
	}

	if cmd.Use != "completion [bash|zsh|fish|powershell]" {
		t.Errorf("expected Use 'completion [bash|zsh|fish|powershell]', got '%s'", cmd.Use)
	}
}

func TestCompletionCommand_Help(t *testing.T) {
	cmd := NewCompletionCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected help output, got empty string")
	}
}

func TestCompletionCommand_Bash(t *testing.T) {
	cmd := NewCompletionCmd()

	// For completion commands, the output goes directly to os.Stdout
	// which we can't easily capture in tests. Just verify it executes without error.
	cmd.SetArgs([]string{"bash"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestCompletionCommand_Zsh(t *testing.T) {
	cmd := NewCompletionCmd()

	cmd.SetArgs([]string{"zsh"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestCompletionCommand_Fish(t *testing.T) {
	cmd := NewCompletionCmd()

	cmd.SetArgs([]string{"fish"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
}

func TestCompletionCommand_PowerShell(t *testing.T) {
	cmd := NewCompletionCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"powershell"})

	// PowerShell completion goes directly to stdout, so we can't easily capture it in tests
	// Just check that it doesn't error
	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	// The output may be empty in the buffer since it goes to stdout directly
	// That's okay for this test
}

func TestCompletionCommand_InvalidShell(t *testing.T) {
	cmd := NewCompletionCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"invalid"})

	err := cmd.Execute()
	if err == nil {
		t.Error("expected error for invalid shell, got nil")
	}
}
