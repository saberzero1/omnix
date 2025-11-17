package common

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/glamour"
)

// MarkdownRenderer provides markdown rendering functionality
type MarkdownRenderer struct {
	renderer *glamour.TermRenderer
}

// NewMarkdownRenderer creates a new markdown renderer
func NewMarkdownRenderer() (*MarkdownRenderer, error) {
	// Detect terminal capabilities and create renderer
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(0), // No word wrap, let terminal handle it
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create markdown renderer: %w", err)
	}

	return &MarkdownRenderer{renderer: r}, nil
}

// RenderMarkdown renders markdown to a string suitable for terminal display
func (m *MarkdownRenderer) RenderMarkdown(markdown string) (string, error) {
	out, err := m.renderer.Render(markdown)
	if err != nil {
		return "", fmt.Errorf("failed to render markdown: %w", err)
	}
	// Trim trailing newlines
	return strings.TrimSpace(out), nil
}

// PrintMarkdown prints markdown to stderr
func (m *MarkdownRenderer) PrintMarkdown(markdown string) error {
	rendered, err := m.RenderMarkdown(markdown)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(os.Stderr, rendered)
	return err
}

// PrintMarkdownTo prints markdown to the specified writer
func (m *MarkdownRenderer) PrintMarkdownTo(w io.Writer, markdown string) error {
	rendered, err := m.RenderMarkdown(markdown)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(w, rendered)
	return err
}

// Global renderer instance
var globalRenderer *MarkdownRenderer

// GetMarkdownRenderer returns a global markdown renderer instance
func GetMarkdownRenderer() (*MarkdownRenderer, error) {
	if globalRenderer == nil {
		var err error
		globalRenderer, err = NewMarkdownRenderer()
		if err != nil {
			return nil, err
		}
	}
	return globalRenderer, nil
}

// PrintMarkdown is a convenience function to print markdown using the global renderer
func PrintMarkdown(markdown string) error {
	r, err := GetMarkdownRenderer()
	if err != nil {
		return err
	}
	return r.PrintMarkdown(markdown)
}

// RenderMarkdown is a convenience function to render markdown using the global renderer
func RenderMarkdown(markdown string) (string, error) {
	r, err := GetMarkdownRenderer()
	if err != nil {
		return "", err
	}
	return r.RenderMarkdown(markdown)
}

// RenderMarkdownToBuffer renders markdown to a buffer
func RenderMarkdownToBuffer(markdown string) (*bytes.Buffer, error) {
	rendered, err := RenderMarkdown(markdown)
	if err != nil {
		return nil, err
	}
	return bytes.NewBufferString(rendered), nil
}
