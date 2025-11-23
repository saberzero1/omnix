package views

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDashboard(t *testing.T) {
	d := NewDashboard()
	require.NotNil(t, d)
}

func TestDashboard_Init(t *testing.T) {
	d := NewDashboard()
	cmd := d.Init()
	assert.Nil(t, cmd)
}

func TestDashboard_Update(t *testing.T) {
	d := NewDashboard()
	msg := tea.KeyMsg{}
	model, cmd := d.Update(msg)

	assert.NotNil(t, model)
	assert.Nil(t, cmd)
}

func TestDashboard_View(t *testing.T) {
	d := NewDashboard()
	output := d.View()

	assert.NotEmpty(t, output)
	assert.Contains(t, output, "Welcome to Omnix TUI")
}

func TestDashboard_SetSize(t *testing.T) {
	d := NewDashboard()
	d.SetSize(120, 40)

	assert.Equal(t, 120, d.width)
	assert.Equal(t, 40, d.height)
}

func TestNewHelp(t *testing.T) {
	h := NewHelp()
	require.NotNil(t, h)
}

func TestHelp_Init(t *testing.T) {
	h := NewHelp()
	cmd := h.Init()
	assert.Nil(t, cmd)
}

func TestHelp_Update(t *testing.T) {
	h := NewHelp()
	msg := tea.KeyMsg{}
	model, cmd := h.Update(msg)

	assert.NotNil(t, model)
	assert.Nil(t, cmd)
}

func TestHelp_View(t *testing.T) {
	h := NewHelp()
	output := h.View()

	assert.NotEmpty(t, output)
	assert.Contains(t, output, "Keyboard Shortcuts")
	assert.Contains(t, output, "Navigation")
}

func TestHelp_SetSize(t *testing.T) {
	h := NewHelp()
	h.SetSize(120, 40)

	assert.Equal(t, 120, h.width)
	assert.Equal(t, 40, h.height)
}

func TestNewFlakeBrowser(t *testing.T) {
	f := NewFlakeBrowser()
	require.NotNil(t, f)
}

func TestFlakeBrowser_Init(t *testing.T) {
	f := NewFlakeBrowser()
	cmd := f.Init()
	assert.Nil(t, cmd)
}

func TestFlakeBrowser_Update(t *testing.T) {
	f := NewFlakeBrowser()
	msg := tea.KeyMsg{}
	model, cmd := f.Update(msg)

	assert.NotNil(t, model)
	assert.Nil(t, cmd)
}

func TestFlakeBrowser_View(t *testing.T) {
	f := NewFlakeBrowser()
	output := f.View()

	assert.NotEmpty(t, output)
	assert.Contains(t, output, "Flake Browser")
}

func TestFlakeBrowser_SetSize(t *testing.T) {
	f := NewFlakeBrowser()
	f.SetSize(120, 40)

	assert.Equal(t, 120, f.width)
	assert.Equal(t, 40, f.height)
}

func TestNewHealthCheck(t *testing.T) {
	h := NewHealthCheck()
	require.NotNil(t, h)
	assert.True(t, h.loading)
}

func TestHealthCheck_Init(t *testing.T) {
	h := NewHealthCheck()
	cmd := h.Init()
	assert.Nil(t, cmd)
}

func TestHealthCheck_Update(t *testing.T) {
	h := NewHealthCheck()
	msg := tea.KeyMsg{}
	model, cmd := h.Update(msg)

	assert.NotNil(t, model)
	assert.Nil(t, cmd)
}

func TestHealthCheck_View_Loading(t *testing.T) {
	h := NewHealthCheck()
	output := h.View()

	assert.NotEmpty(t, output)
	assert.Contains(t, output, "Loading health checks")
}

func TestHealthCheck_SetSize(t *testing.T) {
	h := NewHealthCheck()
	h.SetSize(120, 40)

	assert.Equal(t, 120, h.width)
	assert.Equal(t, 40, h.height)
}

func TestNewSystemInfo(t *testing.T) {
	s := NewSystemInfo()
	require.NotNil(t, s)
	assert.True(t, s.loading)
}

func TestSystemInfo_Init(t *testing.T) {
	s := NewSystemInfo()
	cmd := s.Init()
	assert.Nil(t, cmd)
}

func TestSystemInfo_Update(t *testing.T) {
	s := NewSystemInfo()
	msg := tea.KeyMsg{}
	model, cmd := s.Update(msg)

	assert.NotNil(t, model)
	assert.Nil(t, cmd)
}

func TestSystemInfo_View_Loading(t *testing.T) {
	s := NewSystemInfo()
	output := s.View()

	assert.NotEmpty(t, output)
	assert.Contains(t, output, "Loading system information")
}

func TestSystemInfo_SetSize(t *testing.T) {
	s := NewSystemInfo()
	s.SetSize(120, 40)

	assert.Equal(t, 120, s.width)
	assert.Equal(t, 40, s.height)
}
