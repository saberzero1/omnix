// Package common provides shared utilities for omnix
package common

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogLevel represents the logging verbosity level
type LogLevel int

const (
	// ErrorLevel logs only errors
	ErrorLevel LogLevel = iota
	// WarnLevel logs warnings and errors
	WarnLevel
	// InfoLevel logs info, warnings and errors (default)
	InfoLevel
	// DebugLevel logs debug, info, warnings and errors
	DebugLevel
	// TraceLevel logs everything
	TraceLevel
)

var logger *zap.Logger

// SetupLogging configures logging for the entire application
// verbosity: the log level (0=error, 1=warn, 2=info, 3=debug, 4=trace)
// bare: if true, only log messages without metadata
func SetupLogging(verbosity LogLevel, bare bool) error {
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stderr"}
	config.ErrorOutputPaths = []string{"stderr"}

	// Set encoding based on bare flag
	if bare {
		config.Encoding = "console"
		config.EncoderConfig.TimeKey = ""
		config.EncoderConfig.LevelKey = ""
		config.EncoderConfig.NameKey = ""
		config.EncoderConfig.CallerKey = ""
	} else {
		config.Encoding = "console"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Configure log level
	config.Level = zap.NewAtomicLevelAt(getZapLevel(verbosity))

	// Check for OMNIX_LOG environment variable
	if envLevel := os.Getenv("OMNIX_LOG"); envLevel != "" {
		var level zapcore.Level
		if err := level.UnmarshalText([]byte(envLevel)); err == nil {
			config.Level = zap.NewAtomicLevelAt(level)
		}
	}

	var err error
	logger, err = config.Build()
	if err != nil {
		return err
	}

	zap.ReplaceGlobals(logger)
	return nil
}

// getZapLevel converts our LogLevel to zap's Level
func getZapLevel(level LogLevel) zapcore.Level {
	switch level {
	case ErrorLevel:
		return zapcore.ErrorLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case DebugLevel:
		return zapcore.DebugLevel
	case TraceLevel:
		return zapcore.DebugLevel // Zap doesn't have trace, use debug
	default:
		return zapcore.InfoLevel
	}
}

// Logger returns the global logger instance
func Logger() *zap.Logger {
	if logger == nil {
		// Create a default logger if SetupLogging wasn't called
		logger, _ = zap.NewProduction()
	}
	return logger
}

// Sync flushes any buffered log entries
func Sync() error {
	if logger != nil {
		return logger.Sync()
	}
	return nil
}
