package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FlakeBrowser displays flake information
type FlakeBrowser struct {
	width  int
	height int
}

// NewFlakeBrowser creates a new flake browser view
func NewFlakeBrowser() *FlakeBrowser {
	return &FlakeBrowser{}
}

// Init initializes the flake browser view
func (f *FlakeBrowser) Init() tea.Cmd {
	return nil
}

// Update updates the flake browser view
func (f *FlakeBrowser) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return f, nil
}

// View renders the flake browser view
func (f *FlakeBrowser) View() string {
	style := lipgloss.NewStyle().
		Padding(2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63"))

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86"))

	content := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).
		Render("\nFlake browser functionality will be implemented here.\n\n" +
			"This will allow you to explore Nix flake outputs,\n" +
			"including packages, devShells, apps, and more.\n")

	return style.Render(lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("Flake Browser"),
		content,
	))
}

// SetSize sets the size of the view
func (f *FlakeBrowser) SetSize(width, height int) {
	f.width = width
	f.height = height
}
