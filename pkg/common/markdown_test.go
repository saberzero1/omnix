package common

import (
	"strings"
	"testing"
)

func TestRenderMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		wantErr  bool
		contains string
	}{
		{
			name:     "simple text",
			markdown: "Hello, World!",
			wantErr:  false,
			contains: "Hello",
		},
		{
			name:     "header",
			markdown: "# My Header\n\nSome text",
			wantErr:  false,
			contains: "My Header",
		},
		{
			name:     "code block",
			markdown: "```go\nfunc main() {}\n```",
			wantErr:  false,
			contains: "main",
		},
		{
			name:     "list",
			markdown: "- Item 1\n- Item 2\n- Item 3",
			wantErr:  false,
			contains: "Item",
		},
		{
			name:     "empty string",
			markdown: "",
			wantErr:  false,
			contains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RenderMarkdown(tt.markdown)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderMarkdown() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.contains != "" && !strings.Contains(got, tt.contains) {
				t.Errorf("RenderMarkdown() output should contain %q, got: %q", tt.contains, got)
			}
		})
	}
}

func TestPrintMarkdown(t *testing.T) {
	// Test that PrintMarkdown doesn't panic
	markdown := "# Test\n\nThis is a test"

	err := PrintMarkdown(markdown)
	if err != nil {
		t.Errorf("PrintMarkdown() failed: %v", err)
	}
}

func TestGetMarkdownRenderer(t *testing.T) {
	// First call creates renderer
	r1, err := GetMarkdownRenderer()
	if err != nil {
		t.Fatalf("GetMarkdownRenderer() failed: %v", err)
	}
	if r1 == nil {
		t.Fatal("GetMarkdownRenderer() returned nil")
	}

	// Second call should return same instance
	r2, err := GetMarkdownRenderer()
	if err != nil {
		t.Fatalf("GetMarkdownRenderer() second call failed: %v", err)
	}
	if r1 != r2 {
		t.Error("GetMarkdownRenderer() should return same instance")
	}
}

func TestNewMarkdownRenderer(t *testing.T) {
	renderer, err := NewMarkdownRenderer()
	if err != nil {
		t.Fatalf("NewMarkdownRenderer() failed: %v", err)
	}

	if renderer == nil {
		t.Fatal("NewMarkdownRenderer() returned nil")
	}

	// Test rendering with the new renderer
	markdown := "**Bold text**"
	result, err := renderer.RenderMarkdown(markdown)
	if err != nil {
		t.Errorf("RenderMarkdown() failed: %v", err)
	}

	if result == "" {
		t.Error("RenderMarkdown() returned empty string for non-empty input")
	}
}

func TestRenderMarkdownToBuffer(t *testing.T) {
	markdown := "Test content"

	buf, err := RenderMarkdownToBuffer(markdown)
	if err != nil {
		t.Fatalf("RenderMarkdownToBuffer() failed: %v", err)
	}

	if buf == nil {
		t.Fatal("RenderMarkdownToBuffer() returned nil buffer")
	}

	if buf.Len() == 0 {
		t.Error("RenderMarkdownToBuffer() returned empty buffer")
	}
}

func TestMarkdownRendererPrintMarkdownTo(t *testing.T) {
	renderer, err := NewMarkdownRenderer()
	if err != nil {
		t.Fatalf("NewMarkdownRenderer() failed: %v", err)
	}

	// Create a string builder to capture output
	var buf strings.Builder
	markdown := "# Test Output"

	err = renderer.PrintMarkdownTo(&buf, markdown)
	if err != nil {
		t.Errorf("PrintMarkdownTo() failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Test Output") {
		t.Errorf("PrintMarkdownTo() output should contain 'Test Output', got: %q", output)
	}
}

func TestMarkdownWithSpecialCharacters(t *testing.T) {
	markdown := "Test with `code`, **bold**, and *italic*"

	got, err := RenderMarkdown(markdown)
	if err != nil {
		t.Fatalf("RenderMarkdown() failed: %v", err)
	}

	// Just verify it doesn't crash and returns something
	if got == "" {
		t.Error("RenderMarkdown() returned empty string for formatted text")
	}
}
