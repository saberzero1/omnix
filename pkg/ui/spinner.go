package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SpinnerModel wraps a spinner with a message
type SpinnerModel struct {
	spinner spinner.Model
	message string
	done    bool
	err     error
}

// NewSpinner creates a new spinner model
func NewSpinner(message string) SpinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return SpinnerModel{
		spinner: s,
		message: message,
	}
}

func (m SpinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m SpinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinner.TickMsg:
		if !m.done {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	case DoneMsg:
		m.done = true
		return m, tea.Quit
	case ErrorMsg:
		m.err = msg.Err
		m.done = true
		return m, tea.Quit
	default:
		return m, nil
	}
}

func (m SpinnerModel) View() string {
	if m.err != nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(fmt.Sprintf("✗ %s: %v", m.message, m.err))
	}
	if m.done {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(fmt.Sprintf("✓ %s", m.message))
	}
	return fmt.Sprintf("%s %s", m.spinner.View(), m.message)
}

// DoneMsg signals completion
type DoneMsg struct{}

// ErrorMsg signals an error
type ErrorMsg struct {
	Err error
}

// RunWithSpinner runs a function with a spinner display
func RunWithSpinner(message string, fn func() error) error {
	p := tea.NewProgram(NewSpinner(message))
	
	go func() {
		time.Sleep(100 * time.Millisecond) // Give spinner time to start
		err := fn()
		if err != nil {
			p.Send(ErrorMsg{Err: err})
		} else {
			p.Send(DoneMsg{})
		}
	}()

	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
