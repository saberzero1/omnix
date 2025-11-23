package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/saberzero1/omnix/pkg/health/checks"
)

// RenderHealthChecks renders health check results with fancy styling
func RenderHealthChecks(namedChecks []checks.NamedCheck, width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Padding(1, 0)

	content.WriteString(titleStyle.Render("Nix Health Checks"))
	content.WriteString("\n\n")

	for _, namedCheck := range namedChecks {
		content.WriteString(RenderHealthCheck(namedCheck))
		content.WriteString("\n")
	}

	// Use full width if specified, otherwise use padding
	if width > 0 {
		return lipgloss.NewStyle().
			Width(width).
			Render(content.String())
	}

	return lipgloss.NewStyle().
		Padding(1, 2).
		Render(content.String())
}

// RenderHealthCheck renders a single health check with color-coded status
func RenderHealthCheck(namedCheck checks.NamedCheck) string {
	var icon string
	var color lipgloss.Color

	if namedCheck.Check.Result.IsGreen() {
		icon = "✓"
		color = lipgloss.Color("42") // Green
	} else {
		icon = "✗"
		color = lipgloss.Color("196") // Red
	}

	iconStyle := lipgloss.NewStyle().
		Foreground(color).
		Bold(true)

	titleStyle := lipgloss.NewStyle().
		Bold(true)

	detailStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		PaddingLeft(3)

	title := fmt.Sprintf("%s %s", iconStyle.Render(icon), titleStyle.Render(namedCheck.Check.Title))

	if namedCheck.Check.Info != "" {
		return title + "\n" + detailStyle.Render(namedCheck.Check.Info)
	}

	resultStr := namedCheck.Check.Result.String()
	if resultStr != "" && resultStr != "✅ Passed" {
		return title + "\n" + detailStyle.Render(resultStr)
	}

	return title
}
