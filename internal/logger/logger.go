// Package logger provides a simple wrapper for the zap logging library.
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a simple wrapper for the zap logging library.
type Logger struct {
	zap *zap.Logger
}

// NewLogger creates a new Logger instance with the specified log level.
// The log level should be a string, such as "debug", "info", "warn", or "error".
func NewLogger(level string) (*Logger, error) {
	// convert the text logging level to zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}

	// create a new logger configuration
	config := zap.NewProductionConfig()
	// set the level
	config.Level = lvl
	// config.OutputPaths = []string{"stdout", "./logs/" + logFile}
	logger, err := config.Build(zap.AddCaller())

	if err != nil {
		return nil, err
	}

	return &Logger{zap: logger}, nil
}

// Debug logs a message at the debug level with optional fields.
func (l Logger) Debug(msg string, fields ...zap.Field) {
	l.writer().Debug(msg, fields...)
}

// Info logs a message at the info level with optional fields.
func (l Logger) Info(msg string, fields ...zap.Field) {
	l.writer().Info(msg, fields...)
}

// Warn logs a message at the warn level with optional fields.
func (l Logger) Warn(msg string, fields ...zapcore.Field) {
	l.writer().Warn(msg, fields...)
}

// writer returns the underlying zap.Logger instance.
// If the logger is not initialized, it returns a no-op logger.
func (l Logger) writer() *zap.Logger {
	var noOpLogger = zap.NewNop()
	if l.zap == nil {
		return noOpLogger
	}

	return l.zap
}
