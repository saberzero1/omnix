package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProgressModel wraps a progress bar with a message
type ProgressModel struct {
	progress progress.Model
	message  string
	current  int
	total    int
	done     bool
	err      error
}

// NewProgress creates a new progress bar model.
// Total must be greater than 0 to avoid division by zero errors.
func NewProgress(message string, total int) ProgressModel {
	p := progress.New(progress.WithDefaultGradient())
	if total <= 0 {
		total = 1 // Prevent division by zero
	}
	return ProgressModel{
		progress: p,
		message:  message,
		total:    total,
	}
}

func (m ProgressModel) Init() tea.Cmd {
	return nil
}

func (m ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case ProgressMsg:
		m.current = msg.Current
		if m.current >= m.total {
			m.done = true
			return m, tea.Quit
		}
		return m, nil
	case DoneMsg:
		m.done = true
		m.current = m.total
		return m, tea.Quit
	case ErrorMsg:
		m.err = msg.Err
		m.done = true
		return m, tea.Quit
	default:
		return m, nil
	}
}

func (m ProgressModel) View() string {
	if m.err != nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(fmt.Sprintf("✗ %s: %v", m.message, m.err))
	}
	if m.done {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(fmt.Sprintf("✓ %s (%d/%d)", m.message, m.total, m.total))
	}

	if m.total == 0 {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render(fmt.Sprintf("%s (no work to do)", m.message))
	}

	percent := float64(m.current) / float64(m.total)
	return fmt.Sprintf("%s %s %d/%d",
		m.message,
		m.progress.ViewAs(percent),
		m.current,
		m.total,
	)
}

// ProgressMsg signals a progress update with the current step count.
type ProgressMsg struct {
	Current int
}
