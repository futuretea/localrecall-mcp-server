package logging

import (
	"io"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	// stdioMode indicates if logging should be suppressed (for stdio MCP mode)
	stdioMode = false
)

// Initialize sets up the logging system with the specified level and output
func Initialize(level int, output io.Writer) {
	// Set global log level
	zerolog.SetGlobalLevel(getZerologLevel(level))

	// Configure output
	if output == nil {
		output = os.Stderr
	}

	// Use console writer for better readability
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: output}).
		With().
		Timestamp().
		Logger()
}

// SetStdioMode enables or disables stdio mode
// In stdio mode, all logging is suppressed to avoid interfering with MCP protocol
func SetStdioMode(enabled bool) {
	stdioMode = enabled
	if enabled {
		// Disable all logging in stdio mode
		zerolog.SetGlobalLevel(zerolog.Disabled)
	}
}

// getZerologLevel converts an integer level to zerolog.Level
func getZerologLevel(level int) zerolog.Level {
	switch {
	case level <= 0:
		return zerolog.PanicLevel
	case level == 1:
		return zerolog.FatalLevel
	case level == 2:
		return zerolog.ErrorLevel
	case level == 3:
		return zerolog.WarnLevel
	case level == 4:
		return zerolog.InfoLevel
	case level == 5:
		return zerolog.DebugLevel
	default:
		return zerolog.TraceLevel
	}
}

// Info logs an info message
func Info(format string, v ...interface{}) {
	log.Info().Msgf(format, v...)
}

// Warn logs a warning message
func Warn(format string, v ...interface{}) {
	log.Warn().Msgf(format, v...)
}

// Error logs an error message
func Error(format string, v ...interface{}) {
	log.Error().Msgf(format, v...)
}

// Debug logs a debug message
func Debug(format string, v ...interface{}) {
	log.Debug().Msgf(format, v...)
}

// Fatal logs a fatal message and exits
func Fatal(format string, v ...interface{}) {
	if stdioMode {
		os.Exit(1)
	}
	log.Fatal().Msgf(format, v...)
}
