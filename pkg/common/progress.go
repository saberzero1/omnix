package common

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// ProgressIndicator shows progress for long-running operations
type ProgressIndicator struct {
	message string
	writer  io.Writer
	ticker  *time.Ticker
	done    chan bool
	mu      sync.Mutex
	active  bool
}

// NewProgressIndicator creates a new progress indicator
func NewProgressIndicator(message string) *ProgressIndicator {
	return &ProgressIndicator{
		message: message,
		writer:  os.Stderr,
		done:    make(chan bool),
	}
}

// NewProgressIndicatorWithWriter creates a progress indicator with a custom writer
func NewProgressIndicatorWithWriter(message string, writer io.Writer) *ProgressIndicator {
	return &ProgressIndicator{
		message: message,
		writer:  writer,
		done:    make(chan bool),
	}
}

// Start begins showing the progress indicator
func (p *ProgressIndicator) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.active {
		return
	}

	p.active = true
	p.ticker = time.NewTicker(500 * time.Millisecond)

	go func() {
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0

		for {
			select {
			case <-p.ticker.C:
				_, _ = fmt.Fprintf(p.writer, "\r%s %s", frames[i%len(frames)], p.message)
				i++
			case <-p.done:
				return
			}
		}
	}()
}

// Stop stops the progress indicator
func (p *ProgressIndicator) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.active {
		return
	}

	p.active = false
	if p.ticker != nil {
		p.ticker.Stop()
	}
	p.done <- true
	_, _ = fmt.Fprintf(p.writer, "\r") // Clear the line
}

// UpdateMessage changes the progress message
func (p *ProgressIndicator) UpdateMessage(message string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.message = message
}

// Complete stops the progress indicator and shows a completion message
func (p *ProgressIndicator) Complete(message string) {
	p.Stop()
	_, _ = fmt.Fprintf(p.writer, "\r✓ %s\n", message)
}

// Fail stops the progress indicator and shows a failure message
func (p *ProgressIndicator) Fail(message string) {
	p.Stop()
	_, _ = fmt.Fprintf(p.writer, "\r✗ %s\n", message)
}

// WithProgress runs a function with a progress indicator
func WithProgress(message string, fn func() error) error {
	p := NewProgressIndicator(message)
	p.Start()
	defer p.Stop()
	return fn()
}
