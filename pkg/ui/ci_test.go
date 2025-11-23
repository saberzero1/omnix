package ui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestNewCIRunner(t *testing.T) {
	model := NewCIRunner(".", "default", []string{"x86_64-linux"}, []string{"ROOT", "subflake1"})

	assert.Equal(t, ".", model.flake)
	assert.Equal(t, "default", model.config)
	assert.Equal(t, []string{"x86_64-linux"}, model.systems)
	assert.Equal(t, 2, len(model.subflakes))
	assert.Equal(t, "ROOT", model.subflakes[0].Name)
	assert.Equal(t, StatusPending, model.subflakes[0].Status)
}

func TestCIRunnerSubflakeStart(t *testing.T) {
	model := NewCIRunner(".", "default", []string{"x86_64-linux"}, []string{"ROOT"})

	msg := SubflakeStartMsg{Index: 0, Name: "ROOT"}
	updatedModel, _ := model.Update(msg)
	m := updatedModel.(CIRunnerModel)

	assert.Equal(t, 0, m.currentIndex)
	assert.Equal(t, StatusRunning, m.subflakes[0].Status)
}

func TestCIRunnerSubflakeComplete(t *testing.T) {
	model := NewCIRunner(".", "default", []string{"x86_64-linux"}, []string{"ROOT"})

	// Start subflake
	startMsg := SubflakeStartMsg{Index: 0, Name: "ROOT"}
	updatedModel, _ := model.Update(startMsg)
	m := updatedModel.(CIRunnerModel)

	// Complete successfully
	completeMsg := SubflakeCompleteMsg{Index: 0, Error: ""}
	updatedModel, _ = m.Update(completeMsg)
	m = updatedModel.(CIRunnerModel)

	assert.Equal(t, StatusSuccess, m.subflakes[0].Status)
	assert.Greater(t, m.subflakes[0].Duration.Nanoseconds(), int64(0))
}

func TestCIRunnerSubflakeFailed(t *testing.T) {
	model := NewCIRunner(".", "default", []string{"x86_64-linux"}, []string{"ROOT"})

	// Start subflake
	startMsg := SubflakeStartMsg{Index: 0, Name: "ROOT"}
	updatedModel, _ := model.Update(startMsg)
	m := updatedModel.(CIRunnerModel)

	// Complete with error
	completeMsg := SubflakeCompleteMsg{Index: 0, Error: "build failed"}
	updatedModel, _ = m.Update(completeMsg)
	m = updatedModel.(CIRunnerModel)

	assert.Equal(t, StatusFailed, m.subflakes[0].Status)
	assert.Equal(t, "build failed", m.subflakes[0].Error)
}

func TestCIRunnerStepProgress(t *testing.T) {
	model := NewCIRunner(".", "default", []string{"x86_64-linux"}, []string{"ROOT"})

	// Start subflake
	startMsg := SubflakeStartMsg{Index: 0, Name: "ROOT"}
	updatedModel, _ := model.Update(startMsg)
	m := updatedModel.(CIRunnerModel)

	// Start step
	stepStartMsg := StepStartMsg{SubflakeIndex: 0, StepName: "build"}
	updatedModel, _ = m.Update(stepStartMsg)
	m = updatedModel.(CIRunnerModel)

	assert.Equal(t, 1, len(m.subflakes[0].Steps))
	assert.Equal(t, "build", m.subflakes[0].Steps[0].Name)
	assert.Equal(t, StatusRunning, m.subflakes[0].Steps[0].Status)

	// Complete step
	stepCompleteMsg := StepCompleteMsg{
		SubflakeIndex: 0,
		Output:        "Built 5 outputs",
		Error:         "",
	}
	updatedModel, _ = m.Update(stepCompleteMsg)
	m = updatedModel.(CIRunnerModel)

	assert.Equal(t, StatusSuccess, m.subflakes[0].Steps[0].Status)
	assert.Equal(t, "Built 5 outputs", m.subflakes[0].Steps[0].Output)
	assert.Greater(t, m.subflakes[0].Steps[0].Duration.Nanoseconds(), int64(0))
}

func TestCIRunnerStepSkipped(t *testing.T) {
	model := NewCIRunner(".", "default", []string{"x86_64-linux"}, []string{"ROOT"})

	// Start subflake
	startMsg := SubflakeStartMsg{Index: 0, Name: "ROOT"}
	updatedModel, _ := model.Update(startMsg)
	m := updatedModel.(CIRunnerModel)

	// Skip step
	skipMsg := StepSkipMsg{
		SubflakeIndex: 0,
		StepName:      "lockfile",
		Reason:        "Skipped (has override inputs)",
	}
	updatedModel, _ = m.Update(skipMsg)
	m = updatedModel.(CIRunnerModel)

	assert.Equal(t, 1, len(m.subflakes[0].Steps))
	assert.Equal(t, "lockfile", m.subflakes[0].Steps[0].Name)
	assert.Equal(t, StatusSkipped, m.subflakes[0].Steps[0].Status)
	assert.Equal(t, "Skipped (has override inputs)", m.subflakes[0].Steps[0].Output)
}

func TestCIRunnerView(t *testing.T) {
	model := NewCIRunner(".", "default", []string{"x86_64-linux"}, []string{"ROOT"})

	// Initial view should not panic
	view := model.View()
	assert.Contains(t, view, "Running CI")
	assert.Contains(t, view, "Flake: .")
	assert.Contains(t, view, "ROOT")
}

func TestCIRunnerDone(t *testing.T) {
	model := NewCIRunner(".", "default", []string{"x86_64-linux"}, []string{"ROOT"})

	doneMsg := DoneMsg{}
	updatedModel, cmd := model.Update(doneMsg)
	m := updatedModel.(CIRunnerModel)

	assert.True(t, m.done)
	assert.NotNil(t, cmd)
}

func TestCIRunnerError(t *testing.T) {
	model := NewCIRunner(".", "default", []string{"x86_64-linux"}, []string{"ROOT"})

	errorMsg := ErrorMsg{Err: assert.AnError}
	updatedModel, cmd := model.Update(errorMsg)
	m := updatedModel.(CIRunnerModel)

	assert.True(t, m.done)
	assert.Equal(t, assert.AnError, m.err)
	assert.NotNil(t, cmd)
}

func TestGetStatusIcon(t *testing.T) {
	model := NewCIRunner(".", "default", []string{"x86_64-linux"}, []string{"ROOT"})

	tests := []struct {
		status   StepStatus
		expected string
	}{
		{StatusPending, "○"},
		{StatusRunning, "◐"},
		{StatusSuccess, "✓"},
		{StatusFailed, "✗"},
		{StatusSkipped, "⊘"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, model.getStatusIcon(tt.status))
	}
}

func TestCIRunnerWindowSize(t *testing.T) {
	model := NewCIRunner(".", "default", []string{"x86_64-linux"}, []string{"ROOT"})

	sizeMsg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, _ := model.Update(sizeMsg)
	m := updatedModel.(CIRunnerModel)

	assert.Equal(t, 100, m.width)
	assert.Equal(t, 50, m.height)
}

func TestCIRunnerToggleOutput(t *testing.T) {
	model := NewCIRunner(".", "default", []string{"x86_64-linux"}, []string{"ROOT"})
	assert.False(t, model.showAllOutput)

	// Press 'o' to toggle
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}}
	updatedModel, _ := model.Update(keyMsg)
	m := updatedModel.(CIRunnerModel)

	assert.True(t, m.showAllOutput)

	// Press 'o' again to toggle back
	updatedModel, _ = m.Update(keyMsg)
	m = updatedModel.(CIRunnerModel)

	assert.False(t, m.showAllOutput)
}
