package common

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestProgressIndicator_StartStop(t *testing.T) {
	buf := &bytes.Buffer{}
	p := NewProgressIndicatorWithWriter("Testing", buf)

	p.Start()
	time.Sleep(600 * time.Millisecond) // Wait for at least one tick
	p.Stop()

	output := buf.String()
	if !strings.Contains(output, "Testing") {
		t.Errorf("Expected output to contain 'Testing', got: %s", output)
	}
}

func TestProgressIndicator_Complete(t *testing.T) {
	buf := &bytes.Buffer{}
	p := NewProgressIndicatorWithWriter("Testing", buf)

	p.Start()
	time.Sleep(200 * time.Millisecond)
	p.Complete("Done")

	output := buf.String()
	if !strings.Contains(output, "✓ Done") {
		t.Errorf("Expected output to contain '✓ Done', got: %s", output)
	}
}

func TestProgressIndicator_Fail(t *testing.T) {
	buf := &bytes.Buffer{}
	p := NewProgressIndicatorWithWriter("Testing", buf)

	p.Start()
	time.Sleep(200 * time.Millisecond)
	p.Fail("Failed")

	output := buf.String()
	if !strings.Contains(output, "✗ Failed") {
		t.Errorf("Expected output to contain '✗ Failed', got: %s", output)
	}
}

func TestProgressIndicator_UpdateMessage(t *testing.T) {
	buf := &bytes.Buffer{}
	p := NewProgressIndicatorWithWriter("Initial", buf)

	p.Start()
	time.Sleep(200 * time.Millisecond)
	p.UpdateMessage("Updated")
	time.Sleep(600 * time.Millisecond)
	p.Stop()

	output := buf.String()
	if !strings.Contains(output, "Updated") {
		t.Errorf("Expected output to contain 'Updated', got: %s", output)
	}
}

func TestProgressIndicator_MultipleStartStop(t *testing.T) {
	buf := &bytes.Buffer{}
	p := NewProgressIndicatorWithWriter("Testing", buf)

	// Starting multiple times should be safe
	p.Start()
	p.Start()
	time.Sleep(200 * time.Millisecond)
	p.Stop()
	p.Stop()

	// Should not panic
}

func TestWithProgress(t *testing.T) {
	executed := false
	err := WithProgress("Testing", func() error {
		executed = true
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !executed {
		t.Error("Expected function to be executed")
	}
}

func TestWithProgress_Error(t *testing.T) {
	testErr := fmt.Errorf("test error")
	err := WithProgress("Testing", func() error {
		return testErr
	})

	if err != testErr {
		t.Errorf("Expected error %v, got: %v", testErr, err)
	}
}
