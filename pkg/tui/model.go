package tui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/saberzero1/omnix/pkg/health"
	"github.com/saberzero1/omnix/pkg/health/checks"
	"github.com/saberzero1/omnix/pkg/nix"
	"github.com/saberzero1/omnix/pkg/tui/views"
)

// View represents different views in the TUI
type View int

const (
	DashboardView View = iota
	HealthView
	InfoView
	FlakeView
	HelpView
)

// Model is the main application model
type Model struct {
	ctx          context.Context
	currentView  View
	width        int
	height       int
	dashboard    *views.Dashboard
	healthView   *views.HealthCheck
	infoView     *views.SystemInfo
	flakeView    *views.FlakeBrowser
	helpView     *views.Help
	keys         KeyMap
	quitting     bool
	err          error
}

// KeyMap defines keyboard shortcuts
type KeyMap struct {
	Up         key.Binding
	Down       key.Binding
	Left       key.Binding
	Right      key.Binding
	Help       key.Binding
	Quit       key.Binding
	GoToDash   key.Binding
	GoToHealth key.Binding
	GoToInfo   key.Binding
	GoToFlake  key.Binding
	Refresh    key.Binding
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "move left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "move right"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		GoToDash: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "dashboard"),
		),
		GoToHealth: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "health"),
		),
		GoToInfo: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "info"),
		),
		GoToFlake: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "flake"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
	}
}

// New creates a new TUI model
func New(ctx context.Context) *Model {
	return &Model{
		ctx:         ctx,
		currentView: DashboardView,
		keys:        DefaultKeyMap(),
		dashboard:   views.NewDashboard(),
		healthView:  views.NewHealthCheck(),
		infoView:    views.NewSystemInfo(),
		flakeView:   views.NewFlakeBrowser(),
		helpView:    views.NewHelp(),
	}
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.dashboard.Init(),
		m.loadInitialData(),
	)
}

// loadInitialData loads initial data in the background
func (m *Model) loadInitialData() tea.Cmd {
	return tea.Batch(
		m.loadHealthData(),
		m.loadSystemInfo(),
	)
}

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			if m.currentView == HelpView {
				m.currentView = DashboardView
			} else {
				m.currentView = HelpView
			}
			return m, nil
		case key.Matches(msg, m.keys.GoToDash):
			m.currentView = DashboardView
			return m, nil
		case key.Matches(msg, m.keys.GoToHealth):
			m.currentView = HealthView
			return m, m.loadHealthData()
		case key.Matches(msg, m.keys.GoToInfo):
			m.currentView = InfoView
			return m, m.loadSystemInfo()
		case key.Matches(msg, m.keys.GoToFlake):
			m.currentView = FlakeView
			return m, nil
		case key.Matches(msg, m.keys.Refresh):
			return m, m.refreshCurrentView()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Update all views with new size
		m.dashboard.SetSize(msg.Width, msg.Height-4) // Reserve space for header/footer
		m.healthView.SetSize(msg.Width, msg.Height-4)
		m.infoView.SetSize(msg.Width, msg.Height-4)
		m.flakeView.SetSize(msg.Width, msg.Height-4)
		m.helpView.SetSize(msg.Width, msg.Height-4)
		return m, nil

	case healthDataMsg:
		m.healthView.SetData(msg.checks)
		return m, nil

	case systemInfoMsg:
		m.infoView.SetData(msg.info)
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil
	}

	// Update the current view
	switch m.currentView {
	case DashboardView:
		_, cmd = m.dashboard.Update(msg)
	case HealthView:
		_, cmd = m.healthView.Update(msg)
	case InfoView:
		_, cmd = m.infoView.Update(msg)
	case FlakeView:
		_, cmd = m.flakeView.Update(msg)
	case HelpView:
		_, cmd = m.helpView.Update(msg)
	}

	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the current view
func (m *Model) View() string {
	if m.quitting {
		return ""
	}

	var content string
	switch m.currentView {
	case DashboardView:
		content = m.dashboard.View()
	case HealthView:
		content = m.healthView.View()
	case InfoView:
		content = m.infoView.View()
	case FlakeView:
		content = m.flakeView.View()
	case HelpView:
		content = m.helpView.View()
	}

	// Add header and footer
	header := m.renderHeader()
	footer := m.renderFooter()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		footer,
	)
}

// renderHeader renders the header with navigation
func (m *Model) renderHeader() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		Padding(0, 1)

	navStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	selectedStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		Background(lipgloss.Color("236")).
		Padding(0, 1)

	title := titleStyle.Render("Omnix TUI")

	views := []struct {
		name string
		view View
	}{
		{"Dashboard", DashboardView},
		{"Health", HealthView},
		{"Info", InfoView},
		{"Flake", FlakeView},
	}

	nav := ""
	for i, v := range views {
		if i > 0 {
			nav += " "
		}
		if m.currentView == v.view {
			nav += selectedStyle.Render(v.name)
		} else {
			nav += navStyle.Render(v.name)
		}
	}

	headerStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderBottom(true).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	return headerStyle.Render(
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			title,
			"  ",
			nav,
		),
	)
}

// renderFooter renders the footer with keyboard shortcuts
func (m *Model) renderFooter() string {
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderTop(true).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1)

	help := "1-4: navigate • r: refresh • ?: help • q: quit"
	if m.err != nil {
		help = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(fmt.Sprintf("Error: %v", m.err))
	}

	return footerStyle.Render(help)
}

// refreshCurrentView refreshes data for the current view
func (m *Model) refreshCurrentView() tea.Cmd {
	switch m.currentView {
	case HealthView:
		return m.loadHealthData()
	case InfoView:
		return m.loadSystemInfo()
	case FlakeView:
		return m.loadFlakeData()
	default:
		return nil
	}
}

// loadHealthData loads health check data
func (m *Model) loadHealthData() tea.Cmd {
	return func() tea.Msg {
		ctx := m.ctx

		// Get Nix info
		nixInfo, err := nix.GetInfo(ctx)
		if err != nil {
			return errMsg{err: err}
		}

		// Run all health checks
		h := health.Default()
		checks := h.RunAllChecks(ctx, nixInfo)

		return healthDataMsg{checks: checks}
	}
}

// loadSystemInfo loads system information
func (m *Model) loadSystemInfo() tea.Cmd {
	return func() tea.Msg {
		ctx := m.ctx
		info, err := nix.GetInfo(ctx)
		if err != nil {
			return errMsg{err: err}
		}

		return systemInfoMsg{info: info}
	}
}

// loadFlakeData loads flake data
func (m *Model) loadFlakeData() tea.Cmd {
	return func() tea.Msg {
		// TODO: Implement flake data loading
		return nil
	}
}

// Message types
type healthDataMsg struct {
	checks []checks.NamedCheck
}

type systemInfoMsg struct {
	info *nix.Info
}

type errMsg struct {
	err error
}

// Run starts the TUI application
func Run(ctx context.Context) error {
	m := New(ctx)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
