package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CIRunnerModel manages the UI for CI runs
type CIRunnerModel struct {
	flake         string
	config        string
	systems       []string
	subflakes     []SubflakeProgress
	currentIndex  int
	done          bool
	err           error
	width         int
	height        int
	spinner       spinner.Model
	startTime     time.Time
	showAllOutput bool
}

// SubflakeProgress tracks progress for a single subflake
type SubflakeProgress struct {
	Name      string
	Status    StepStatus
	Steps     []StepProgress
	StartTime time.Time
	Duration  time.Duration
	Error     string
}

// StepProgress tracks progress for a single step
type StepProgress struct {
	Name      string
	Status    StepStatus
	Output    string
	Error     string
	StartTime time.Time
	Duration  time.Duration
}

// StepStatus represents the status of a step
type StepStatus int

const (
	StatusPending StepStatus = iota
	StatusRunning
	StatusSuccess
	StatusFailed
	StatusSkipped
)

// NewCIRunner creates a new CI runner UI model
func NewCIRunner(flake, config string, systems []string, subflakeNames []string) CIRunnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	subflakes := make([]SubflakeProgress, len(subflakeNames))
	for i, name := range subflakeNames {
		subflakes[i] = SubflakeProgress{
			Name:   name,
			Status: StatusPending,
			Steps:  []StepProgress{},
		}
	}

	return CIRunnerModel{
		flake:        flake,
		config:       config,
		systems:      systems,
		subflakes:    subflakes,
		spinner:      s,
		startTime:    time.Now(),
		currentIndex: -1,
	}
}

func (m CIRunnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m CIRunnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.done = true
			return m, tea.Quit
		case "o":
			m.showAllOutput = !m.showAllOutput
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		if !m.done {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case SubflakeStartMsg:
		if msg.Index >= 0 && msg.Index < len(m.subflakes) {
			m.currentIndex = msg.Index
			m.subflakes[msg.Index].Status = StatusRunning
			m.subflakes[msg.Index].StartTime = time.Now()
		}
		return m, nil

	case SubflakeCompleteMsg:
		if msg.Index >= 0 && msg.Index < len(m.subflakes) {
			m.subflakes[msg.Index].Duration = time.Since(m.subflakes[msg.Index].StartTime)
			if msg.Error != "" {
				m.subflakes[msg.Index].Status = StatusFailed
				m.subflakes[msg.Index].Error = msg.Error
			} else {
				m.subflakes[msg.Index].Status = StatusSuccess
			}
		}
		return m, nil

	case StepStartMsg:
		if msg.SubflakeIndex >= 0 && msg.SubflakeIndex < len(m.subflakes) {
			step := StepProgress{
				Name:      msg.StepName,
				Status:    StatusRunning,
				StartTime: time.Now(),
			}
			m.subflakes[msg.SubflakeIndex].Steps = append(m.subflakes[msg.SubflakeIndex].Steps, step)
		}
		return m, nil

	case StepCompleteMsg:
		if msg.SubflakeIndex >= 0 && msg.SubflakeIndex < len(m.subflakes) {
			steps := m.subflakes[msg.SubflakeIndex].Steps
			if len(steps) > 0 {
				lastStep := &steps[len(steps)-1]
				lastStep.Duration = time.Since(lastStep.StartTime)
				lastStep.Output = msg.Output
				if msg.Error != "" {
					lastStep.Status = StatusFailed
					lastStep.Error = msg.Error
				} else {
					lastStep.Status = StatusSuccess
				}
			}
		}
		return m, nil

	case StepSkipMsg:
		if msg.SubflakeIndex >= 0 && msg.SubflakeIndex < len(m.subflakes) {
			step := StepProgress{
				Name:   msg.StepName,
				Status: StatusSkipped,
				Output: msg.Reason,
			}
			m.subflakes[msg.SubflakeIndex].Steps = append(m.subflakes[msg.SubflakeIndex].Steps, step)
		}
		return m, nil

	case DoneMsg:
		m.done = true
		return m, tea.Quit

	case ErrorMsg:
		m.err = msg.Err
		m.done = true
		return m, tea.Quit
	}

	return m, nil
}

func (m CIRunnerModel) View() string {
	if m.width == 0 {
		m.width = 80
	}

	var b strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("86")).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderBottom(true).
		BorderForeground(lipgloss.Color("240")).
		Width(m.width-2).
		Padding(0, 1)

	header := fmt.Sprintf("Running CI • Flake: %s", m.flake)
	if m.config != "" {
		header += fmt.Sprintf(" • Config: %s", m.config)
	}
	header += fmt.Sprintf(" • Systems: %s", strings.Join(m.systems, ", "))

	b.WriteString(headerStyle.Render(header))
	b.WriteString("\n\n")

	// Subflakes progress
	for i, sf := range m.subflakes {
		b.WriteString(m.renderSubflake(i, sf))
		b.WriteString("\n")
	}

	// Footer with elapsed time and controls
	if !m.done {
		elapsed := time.Since(m.startTime).Round(time.Second)
		footerStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderTop(true).
			BorderForeground(lipgloss.Color("240")).
			Width(m.width-2).
			Padding(0, 1)

		footer := fmt.Sprintf("Elapsed: %s • Press 'o' to toggle output details • 'q' to quit", elapsed)
		b.WriteString("\n")
		b.WriteString(footerStyle.Render(footer))
	} else if m.err != nil {
		// Error message
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(fmt.Sprintf("✗ Error: %v", m.err)))
		b.WriteString("\n")
	} else {
		// Success summary
		elapsed := time.Since(m.startTime).Round(time.Second)
		successStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)
		b.WriteString("\n")
		b.WriteString(successStyle.Render(fmt.Sprintf("✓ All CI steps passed in %s", elapsed)))
		b.WriteString("\n")
	}

	return b.String()
}

func (m CIRunnerModel) renderSubflake(index int, sf SubflakeProgress) string {
	var b strings.Builder

	// Subflake header
	statusIcon := m.getStatusIcon(sf.Status)
	statusStyle := m.getStatusStyle(sf.Status)

	nameStyle := lipgloss.NewStyle().Bold(true)
	name := nameStyle.Render(sf.Name)

	if index == m.currentIndex && sf.Status == StatusRunning {
		name = fmt.Sprintf("%s %s", m.spinner.View(), name)
	} else {
		name = fmt.Sprintf("%s %s", statusIcon, name)
	}

	durationStr := ""
	if sf.Duration > 0 {
		durationStr = fmt.Sprintf(" (%s)", sf.Duration.Round(time.Millisecond))
	}

	b.WriteString(statusStyle.Render(name + durationStr))
	b.WriteString("\n")

	// Show error if failed
	if sf.Status == StatusFailed && sf.Error != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Padding(0, 2)
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %s", sf.Error)))
		b.WriteString("\n")
	}

	// Steps
	for _, step := range sf.Steps {
		b.WriteString(m.renderStep(step))
	}

	return b.String()
}

func (m CIRunnerModel) renderStep(step StepProgress) string {
	var b strings.Builder

	statusIcon := m.getStatusIcon(step.Status)
	statusStyle := m.getStatusStyle(step.Status)

	stepIndent := "  "
	stepName := fmt.Sprintf("%s%s %s", stepIndent, statusIcon, step.Name)

	durationStr := ""
	if step.Duration > 0 {
		durationStr = fmt.Sprintf(" (%s)", step.Duration.Round(time.Millisecond))
	}

	b.WriteString(statusStyle.Render(stepName + durationStr))
	b.WriteString("\n")

	// Show error if failed
	if step.Status == StatusFailed && step.Error != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Padding(0, 4)
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error: %s", step.Error)))
		b.WriteString("\n")
	}

	// Show output if requested and available
	if m.showAllOutput && step.Output != "" {
		outputStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Padding(0, 4)

		// Limit output length for display
		output := step.Output
		if len(output) > 500 {
			output = output[:500] + "... (truncated)"
		}
		b.WriteString(outputStyle.Render(output))
		b.WriteString("\n")
	}

	return b.String()
}

func (m CIRunnerModel) getStatusIcon(status StepStatus) string {
	switch status {
	case StatusPending:
		return "○"
	case StatusRunning:
		return "◐"
	case StatusSuccess:
		return "✓"
	case StatusFailed:
		return "✗"
	case StatusSkipped:
		return "⊘"
	default:
		return "?"
	}
}

func (m CIRunnerModel) getStatusStyle(status StepStatus) lipgloss.Style {
	switch status {
	case StatusPending:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	case StatusRunning:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
	case StatusSuccess:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	case StatusFailed:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	case StatusSkipped:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	default:
		return lipgloss.NewStyle()
	}
}

// Message types for CI runner

// SubflakeStartMsg signals the start of a subflake run
type SubflakeStartMsg struct {
	Index int
	Name  string
}

// SubflakeCompleteMsg signals the completion of a subflake run
type SubflakeCompleteMsg struct {
	Index int
	Error string
}

// StepStartMsg signals the start of a CI step
type StepStartMsg struct {
	SubflakeIndex int
	StepName      string
}

// StepCompleteMsg signals the completion of a CI step
type StepCompleteMsg struct {
	SubflakeIndex int
	Output        string
	Error         string
}

// StepSkipMsg signals that a step was skipped
type StepSkipMsg struct {
	SubflakeIndex int
	StepName      string
	Reason        string
}
