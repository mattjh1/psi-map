package logger

import (
	"github.com/mattjh1/psi-map/internal/types/logger"
)

// New creates a new logger instance
func New(options ...logger.Option) *logger.Logger {
	return logger.New(options...)
}

// WithLevel sets the minimum log level
func WithLevel(level logger.LogLevel) logger.Option {
	return logger.WithLevel(level)
}

// WithPrefix sets a prefix for all log messages
func WithPrefix(prefix string) logger.Option {
	return logger.WithPrefix(prefix)
}

// WithTime enables/disables timestamp in logs
func WithTime(show bool) logger.Option {
	return logger.WithTime(show)
}

// WithPlainText enables/disables plain-text output (no colors)
func WithPlainText(plain bool) logger.Option {
	return logger.WithPlainText(plain)
}

// SetLevel safely updates the log level
func SetLevel(l *logger.Logger, level logger.LogLevel) {
	l.SetLevel(level)
}

// Debug logs debug information
func Debug(l *logger.Logger, message string, args ...any) {
	l.Debug(message, args...)
}

// Info logs general information
func Info(l *logger.Logger, message string, args ...any) {
	l.Info(message, args...)
}

// Warn logs warnings
func Warn(l *logger.Logger, message string, args ...any) {
	l.Warn(message, args...)
}

// Error logs errors
func Error(l *logger.Logger, message string, args ...any) {
	l.Error(message, args...)
}

// Success logs success messages
func Success(l *logger.Logger, message string, args ...any) {
	l.Success(message, args...)
}

// Tagged logs messages with a custom tag and optional emoji
func Tagged(l *logger.Logger, tag, message string, args ...any) {
	TaggedWithEmoji(l, tag, message, "", args...)
}

func TaggedWithEmoji(l *logger.Logger, tag, message, emoji string, args ...any) {
	l.Tagged(tag, message, emoji, args...)
}

// Exit terminates the program with the given exit code
func Exit(l *logger.Logger, code int) {
	l.Exit(code)
}
