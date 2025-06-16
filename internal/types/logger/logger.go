package logger

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/pterm/pterm"
)

// LogLevel represents the logging level
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// Logger is the main logger struct
type Logger struct {
	Level    LogLevel
	Prefix   string
	ShowTime bool
	mu       sync.RWMutex // Protects Level field
}

// Option is a functional option for configuring Logger
type Option func(*Logger)

// WithLevel sets the minimum log level
func WithLevel(level LogLevel) Option {
	return func(l *Logger) { l.Level = level }
}

// WithPrefix sets a prefix for all log messages
func WithPrefix(prefix string) Option {
	return func(l *Logger) { l.Prefix = prefix }
}

// WithTime enables/disables timestamp in logs
func WithTime(show bool) Option {
	return func(l *Logger) { l.ShowTime = show }
}

// WithPlainText enables/disables plain-text output (no colors)
func WithPlainText(plain bool) Option {
	return func(l *Logger) {
		if plain {
			pterm.DisableColor()
		}
	}
}

// New creates a new logger instance
func New(options ...Option) *Logger {
	l := &Logger{Level: INFO}
	for _, opt := range options {
		opt(l)
	}
	return l
}

// SetLevel safely updates the log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.Level = level
}

// formatMessage formats the message with prefix and timestamp
func (l *Logger) formatMessage(message string) string {
	var parts []string
	if l.ShowTime {
		parts = append(parts, time.Now().Format("15:04:05"))
	}
	if l.Prefix != "" {
		parts = append(parts, l.Prefix)
	}
	parts = append(parts, message)
	return strings.Join(parts, " ")
}

// log handles logging with level filtering
func (l *Logger) log(level LogLevel, printer *pterm.PrefixPrinter, message string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.Level > level {
		return
	}
	msg := fmt.Sprintf(message, args...)
	printer.Println(l.formatMessage(msg))
}

// Debug logs debug information (gray, low priority)
func (l *Logger) Debug(message string, args ...any) {
	l.log(DEBUG, &pterm.Debug, message, args...)
}

// Info logs general information (blue with info icon)
func (l *Logger) Info(message string, args ...any) {
	l.log(INFO, &pterm.Info, message, args...)
}

// Warn logs warnings (yellow with warning icon)
func (l *Logger) Warn(message string, args ...any) {
	l.log(WARN, &pterm.Warning, message, args...)
}

// Error logs errors (red with error icon)
func (l *Logger) Error(message string, args ...any) {
	l.log(ERROR, &pterm.Error, message, args...)
}

// Success logs success messages (green with checkmark)
func (l *Logger) Success(message string, args ...any) {
	l.log(INFO, &pterm.Success, message, args...)
}

// Tagged logs messages with a custom tag and optional emoji
func (l *Logger) Tagged(tag, message, emoji string, args ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.Level > INFO {
		return
	}
	msg := fmt.Sprintf(message, args...)
	if emoji != "" {
		msg = emoji + " " + msg
	}
	originalPrefix := pterm.Info.Prefix
	pterm.Info.Prefix = pterm.Prefix{
		Text:  tag,
		Style: pterm.NewStyle(tagColor(tag), pterm.FgLightWhite),
	}
	pterm.Info.Println(l.formatMessage(msg))
	pterm.Info.Prefix = originalPrefix
}

// tagColor maps tags to background colors
func tagColor(tag string) pterm.Color {
	switch strings.ToUpper(tag) {
	case "SERVER":
		return pterm.BgBlue
	case "CACHE":
		return pterm.BgMagenta
	case "ANALYZE":
		return pterm.BgGreen
	case "PSI":
		return pterm.BgYellow
	case "STEP":
		return pterm.BgLightBlue
	default:
		return pterm.BgCyan
	}
}

// Exit terminates the program with the given exit code
func (l *Logger) Exit(code int) {
	os.Exit(code)
}
