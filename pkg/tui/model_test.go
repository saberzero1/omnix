package tui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	ctx := context.Background()
	m := New(ctx)

	require.NotNil(t, m)
	assert.Equal(t, DashboardView, m.currentView)
	assert.NotNil(t, m.dashboard)
	assert.NotNil(t, m.healthView)
	assert.NotNil(t, m.infoView)
	assert.NotNil(t, m.flakeView)
	assert.NotNil(t, m.helpView)
}

func TestDefaultKeyMap(t *testing.T) {
	keys := DefaultKeyMap()

	assert.NotNil(t, keys.Up)
	assert.NotNil(t, keys.Down)
	assert.NotNil(t, keys.Left)
	assert.NotNil(t, keys.Right)
	assert.NotNil(t, keys.Help)
	assert.NotNil(t, keys.Quit)
	assert.NotNil(t, keys.GoToDash)
	assert.NotNil(t, keys.GoToHealth)
	assert.NotNil(t, keys.GoToInfo)
	assert.NotNil(t, keys.GoToFlake)
	assert.NotNil(t, keys.Refresh)
}

func TestModel_Init(t *testing.T) {
	ctx := context.Background()
	m := New(ctx)

	cmd := m.Init()
	assert.NotNil(t, cmd)
}

func TestModel_Update_KeyNavigation(t *testing.T) {
	ctx := context.Background()
	m := New(ctx)

	tests := []struct {
		name     string
		key      string
		wantView View
	}{
		{"Dashboard", "1", DashboardView},
		{"Health", "2", HealthView},
		{"Info", "3", InfoView},
		{"Flake", "4", FlakeView},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			updatedModel, _ := m.Update(msg)
			model := updatedModel.(*Model)
			assert.Equal(t, tt.wantView, model.currentView)
		})
	}
}

func TestModel_Update_HelpToggle(t *testing.T) {
	ctx := context.Background()
	m := New(ctx)

	// Initially on dashboard
	assert.Equal(t, DashboardView, m.currentView)

	// Press ? to show help
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")}
	updatedModel, _ := m.Update(msg)
	model := updatedModel.(*Model)
	assert.Equal(t, HelpView, model.currentView)

	// Press ? again to go back to dashboard
	updatedModel, _ = model.Update(msg)
	model = updatedModel.(*Model)
	assert.Equal(t, DashboardView, model.currentView)
}

func TestModel_Update_Quit(t *testing.T) {
	ctx := context.Background()
	m := New(ctx)

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")}
	updatedModel, cmd := m.Update(msg)
	model := updatedModel.(*Model)

	assert.True(t, model.quitting)
	assert.NotNil(t, cmd)
}

func TestModel_Update_WindowSize(t *testing.T) {
	ctx := context.Background()
	m := New(ctx)

	msg := tea.WindowSizeMsg{Width: 120, Height: 40}
	updatedModel, _ := m.Update(msg)
	model := updatedModel.(*Model)

	assert.Equal(t, 120, model.width)
	assert.Equal(t, 40, model.height)
}

func TestModel_View(t *testing.T) {
	ctx := context.Background()
	m := New(ctx)

	// Test that each view renders without panic
	views := []View{DashboardView, HealthView, InfoView, FlakeView, HelpView}

	for _, view := range views {
		m.currentView = view
		output := m.View()
		assert.NotEmpty(t, output, "View %v should produce output", view)
	}
}

func TestModel_View_Quitting(t *testing.T) {
	ctx := context.Background()
	m := New(ctx)
	m.quitting = true

	output := m.View()
	assert.Empty(t, output, "Quitting view should be empty")
}

func TestModel_refreshCurrentView(t *testing.T) {
	ctx := context.Background()
	m := New(ctx)

	tests := []struct {
		name     string
		view     View
		wantCmd  bool
	}{
		{"Dashboard", DashboardView, false},
		{"Health", HealthView, true},
		{"Info", InfoView, true},
		{"Flake", FlakeView, true},
		{"Help", HelpView, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m.currentView = tt.view
			cmd := m.refreshCurrentView()
			if tt.wantCmd {
				assert.NotNil(t, cmd, "Should return a command for %v", tt.view)
			} else {
				assert.Nil(t, cmd, "Should not return a command for %v", tt.view)
			}
		})
	}
}
