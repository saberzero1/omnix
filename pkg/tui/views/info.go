package views

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/saberzero1/omnix/pkg/nix"
)

// SystemInfo displays system information
type SystemInfo struct {
	width   int
	height  int
	info    *nix.Info
	loading bool
}

// NewSystemInfo creates a new system info view
func NewSystemInfo() *SystemInfo {
	return &SystemInfo{
		loading: true,
	}
}

// Init initializes the system info view
func (s *SystemInfo) Init() tea.Cmd {
	return nil
}

// Update updates the system info view
func (s *SystemInfo) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return s, nil
}

// View renders the system info view
func (s *SystemInfo) View() string {
	if s.loading && s.info == nil {
		return lipgloss.NewStyle().
			Width(s.width).
			Padding(2).
			Render("Loading system information...")
	}

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
		Width(20)

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	content.WriteString(titleStyle.Render("Nix System Information"))
	content.WriteString("\n\n")

	// Nix Version Section
	if s.info != nil {
		content.WriteString(sectionStyle.Render("Nix Version"))
		content.WriteString("\n")
		content.WriteString(keyStyle.Render("Version: "))
		content.WriteString(valueStyle.Render(s.info.Version.String()))
		content.WriteString("\n")
	}

	// Nix Config Section
	if s.info != nil {
		content.WriteString(sectionStyle.Render("Nix Configuration"))
		content.WriteString("\n")

		if s.info.Config.System.Value != "" {
			content.WriteString(keyStyle.Render("System: "))
			content.WriteString(valueStyle.Render(s.info.Config.System.Value))
			content.WriteString("\n")
		}

		if len(s.info.Config.Substituters.Value) > 0 {
			content.WriteString(keyStyle.Render("Substituters: "))
			content.WriteString(valueStyle.Render(strings.Join(s.info.Config.Substituters.Value, ", ")))
			content.WriteString("\n")
		}

		if s.info.Config.MaxJobs.Value > 0 {
			content.WriteString(keyStyle.Render("Max Jobs: "))
			content.WriteString(valueStyle.Render(fmt.Sprintf("%d", s.info.Config.MaxJobs.Value)))
			content.WriteString("\n")
		}

		if s.info.Config.Cores.Value > 0 {
			content.WriteString(keyStyle.Render("Cores: "))
			content.WriteString(valueStyle.Render(fmt.Sprintf("%d", s.info.Config.Cores.Value)))
			content.WriteString("\n")
		}

		if len(s.info.Config.ExperimentalFeatures.Value) > 0 {
			content.WriteString(keyStyle.Render("Experimental: "))
			content.WriteString(valueStyle.Render(strings.Join(s.info.Config.ExperimentalFeatures.Value, ", ")))
			content.WriteString("\n")
		}
	}

	// Environment Section
	if s.info != nil && s.info.Env != nil {
		content.WriteString(sectionStyle.Render("Environment"))
		content.WriteString("\n")

		if s.info.Env.User != "" {
			content.WriteString(keyStyle.Render("User: "))
			content.WriteString(valueStyle.Render(s.info.Env.User))
			content.WriteString("\n")
		}

		content.WriteString(keyStyle.Render("OS: "))
		content.WriteString(valueStyle.Render(s.info.Env.OS.String()))
		content.WriteString("\n")

		if len(s.info.Env.Groups) > 0 {
			content.WriteString(keyStyle.Render("Groups: "))
			content.WriteString(valueStyle.Render(strings.Join(s.info.Env.Groups, ", ")))
			content.WriteString("\n")
		}
	}

	return lipgloss.NewStyle().
		Width(s.width).
		Padding(1, 2).
		Render(content.String())
}

// SetSize sets the size of the view
func (s *SystemInfo) SetSize(width, height int) {
	s.width = width
	s.height = height
}

// SetData sets the system info data
func (s *SystemInfo) SetData(info *nix.Info) {
	s.info = info
	s.loading = false
}
