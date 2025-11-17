package common

import (
	"testing"

	"go.uber.org/zap"
)

func TestSetupLogging(t *testing.T) {
	tests := []struct {
		name      string
		verbosity LogLevel
		bare      bool
		wantErr   bool
	}{
		{
			name:      "error level",
			verbosity: ErrorLevel,
			bare:      false,
			wantErr:   false,
		},
		{
			name:      "warn level",
			verbosity: WarnLevel,
			bare:      false,
			wantErr:   false,
		},
		{
			name:      "info level",
			verbosity: InfoLevel,
			bare:      false,
			wantErr:   false,
		},
		{
			name:      "debug level",
			verbosity: DebugLevel,
			bare:      false,
			wantErr:   false,
		},
		{
			name:      "trace level",
			verbosity: TraceLevel,
			bare:      false,
			wantErr:   false,
		},
		{
			name:      "bare format",
			verbosity: InfoLevel,
			bare:      true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetupLogging(tt.verbosity, tt.bare)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetupLogging() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify logger is set
			logger := Logger()
			if logger == nil {
				t.Error("Logger() returned nil after SetupLogging()")
			}
		})
	}
}

func TestGetZapLevel(t *testing.T) {
	tests := []struct {
		name  string
		level LogLevel
		want  zap.AtomicLevel
	}{
		{
			name:  "error level",
			level: ErrorLevel,
			want:  zap.NewAtomicLevelAt(zap.ErrorLevel),
		},
		{
			name:  "warn level",
			level: WarnLevel,
			want:  zap.NewAtomicLevelAt(zap.WarnLevel),
		},
		{
			name:  "info level",
			level: InfoLevel,
			want:  zap.NewAtomicLevelAt(zap.InfoLevel),
		},
		{
			name:  "debug level",
			level: DebugLevel,
			want:  zap.NewAtomicLevelAt(zap.DebugLevel),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getZapLevel(tt.level)
			if got != tt.want.Level() {
				t.Errorf("getZapLevel() = %v, want %v", got, tt.want.Level())
			}
		})
	}
}

func TestLogger(t *testing.T) {
	// Reset logger
	logger = nil

	// First call should create a default logger
	l := Logger()
	if l == nil {
		t.Error("Logger() returned nil")
	}

	// Second call should return the same logger
	l2 := Logger()
	if l != l2 {
		t.Error("Logger() returned different instances")
	}
}

func TestSync(t *testing.T) {
	// Setup a logger
	if err := SetupLogging(InfoLevel, false); err != nil {
		t.Fatalf("SetupLogging() failed: %v", err)
	}

	// Sync should not return an error
	if err := Sync(); err != nil {
		// Note: Sync can fail on stderr on some systems, so we just log it
		t.Logf("Sync() returned error (may be expected): %v", err)
	}
}
