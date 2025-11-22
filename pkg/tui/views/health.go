package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/saberzero1/omnix/pkg/health/checks"
)

// HealthCheck displays health check results
type HealthCheck struct {
	width   int
	height  int
	checks  []checks.NamedCheck
	loading bool
}

// NewHealthCheck creates a new health check view
func NewHealthCheck() *HealthCheck {
	return &HealthCheck{
		loading: true,
	}
}

// Init initializes the health check view
func (h *HealthCheck) Init() tea.Cmd {
	return nil
}

// Update updates the health check view
func (h *HealthCheck) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return h, nil
}

// View renders the health check view
func (h *HealthCheck) View() string {
	if h.loading && len(h.checks) == 0 {
		return lipgloss.NewStyle().
			Padding(2).
			Render("Loading health checks...")
	}

	var content strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Padding(1, 0)

	content.WriteString(titleStyle.Render("Nix Health Checks"))
	content.WriteString("\n\n")

	for _, namedCheck := range h.checks {
		content.WriteString(h.renderCheck(namedCheck))
		content.WriteString("\n")
	}

	return lipgloss.NewStyle().
		Padding(1, 2).
		Render(content.String())
}

// renderCheck renders a single health check
func (h *HealthCheck) renderCheck(namedCheck checks.NamedCheck) string {
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

// SetSize sets the size of the view
func (h *HealthCheck) SetSize(width, height int) {
	h.width = width
	h.height = height
}

// SetData sets the health check data
func (h *HealthCheck) SetData(checks []checks.NamedCheck) {
	h.checks = checks
	h.loading = false
}
