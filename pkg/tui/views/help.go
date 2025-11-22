package views

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Help displays help information
type Help struct {
	width  int
	height int
}

// NewHelp creates a new help view
func NewHelp() *Help {
	return &Help{}
}

// Init initializes the help view
func (h *Help) Init() tea.Cmd {
	return nil
}

// Update updates the help view
func (h *Help) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return h, nil
}

// View renders the help view
func (h *Help) View() string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Padding(1, 0)

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		PaddingTop(1)

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Width(15)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	content.WriteString(titleStyle.Render("Omnix TUI - Keyboard Shortcuts"))
	content.WriteString("\n\n")

	// Navigation section
	content.WriteString(sectionStyle.Render("Navigation"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("1"))
	content.WriteString(valueStyle.Render("Dashboard - Main overview"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("2"))
	content.WriteString(valueStyle.Render("Health - System health checks"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("3"))
	content.WriteString(valueStyle.Render("Info - System information"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("4"))
	content.WriteString(valueStyle.Render("Flake - Flake browser"))
	content.WriteString("\n\n")

	// Movement section
	content.WriteString(sectionStyle.Render("Movement"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("↑/k"))
	content.WriteString(valueStyle.Render("Move up"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("↓/j"))
	content.WriteString(valueStyle.Render("Move down"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("←/h"))
	content.WriteString(valueStyle.Render("Move left"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("→/l"))
	content.WriteString(valueStyle.Render("Move right"))
	content.WriteString("\n\n")

	// Actions section
	content.WriteString(sectionStyle.Render("Actions"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("r"))
	content.WriteString(valueStyle.Render("Refresh current view"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("?"))
	content.WriteString(valueStyle.Render("Toggle this help screen"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("q"))
	content.WriteString(valueStyle.Render("Quit the application"))
	content.WriteString("\n")
	content.WriteString(keyStyle.Render("Ctrl+C"))
	content.WriteString(valueStyle.Render("Quit the application"))
	content.WriteString("\n\n")

	// About section
	content.WriteString(sectionStyle.Render("About"))
	content.WriteString("\n")
	content.WriteString(valueStyle.Render("Omnix is a developer-friendly companion tool for Nix.\n"))
	content.WriteString(valueStyle.Render("Built with Bubble Tea (https://github.com/charmbracelet/bubbletea)\n"))

	return lipgloss.NewStyle().
		Padding(1, 2).
		Render(content.String())
}

// SetSize sets the size of the view
func (h *Help) SetSize(width, height int) {
	h.width = width
	h.height = height
}
