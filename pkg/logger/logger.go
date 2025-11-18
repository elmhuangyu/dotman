package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	// Logger is the global logger instance
	Logger zerolog.Logger
)

// Init initializes the global logger with default configuration
func init() {
	// Default to info level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Configure console writer with colors
	output := zerolog.ConsoleWriter{
		Out:             os.Stdout,
		FormatTimestamp: func(i interface{}) string { return "" },
	}

	// Create logger
	Logger = log.Output(output)
}

// SetDebugMode enables debug level logging
func SetDebugMode() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	Logger.Debug().Msg("Debug mode enabled")
}

// GetLogger returns the global logger instance
func GetLogger() zerolog.Logger {
	return Logger
}
