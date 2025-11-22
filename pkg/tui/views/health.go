package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/saberzero1/omnix/pkg/health/checks"
	"github.com/saberzero1/omnix/pkg/ui"
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
			Width(h.width).
			Padding(2).
			Render("Loading health checks...")
	}

	// Use the shared UI rendering with full width
	return ui.RenderHealthChecks(h.checks, h.width)
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
