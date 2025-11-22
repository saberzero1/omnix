package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Dashboard is the main dashboard view
type Dashboard struct {
	width  int
	height int
}

// NewDashboard creates a new dashboard
func NewDashboard() *Dashboard {
	return &Dashboard{}
}

// Init initializes the dashboard
func (d *Dashboard) Init() tea.Cmd {
	return nil
}

// Update updates the dashboard
func (d *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return d, nil
}

// View renders the dashboard
func (d *Dashboard) View() string {
	style := lipgloss.NewStyle().
		Padding(2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Render("Welcome to Omnix TUI!")

	content := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Render("\nOmnix is a developer-friendly companion tool for Nix.\n\n" +
			"Use the keyboard shortcuts to navigate:\n\n" +
			"  1 - Dashboard (current view)\n" +
			"  2 - Health Checks\n" +
			"  3 - System Info\n" +
			"  4 - Flake Browser\n\n" +
			"  r - Refresh current view\n" +
			"  ? - Toggle help\n" +
			"  q - Quit\n")

	return style.Render(lipgloss.JoinVertical(lipgloss.Left, title, content))
}

// SetSize sets the size of the dashboard
func (d *Dashboard) SetSize(width, height int) {
	d.width = width
	d.height = height
}
