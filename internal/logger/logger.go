package logger

import (
	"fmt"
	"io"
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
	Output   io.Writer // Optional output destination (e.g., file, network)
	mu       sync.RWMutex
}

// UI provides methods for rendering CLI UI elements
type UI struct {
	Logger *Logger
	Style  *UIStyle // Optional styling for UI elements
}

// UIStyle defines custom styling for UI elements
type UIStyle struct {
	TableBorderStyle *pterm.Style
	HeaderBgColor    pterm.Color
}

// InputType defines the type of interactive prompt
type InputType int

const (
	TextInput InputType = iota
	ConfirmInput
	SelectInput
	MultiSelectInput
)

// Option is a functional option for configuring Logger
type Option func(*Logger)

// UIOption is a functional option for configuring UI
type UIOption func(*UI)

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

// WithOutput sets an alternative output destination for logs
func WithOutput(w io.Writer) Option {
	return func(l *Logger) { l.Output = w }
}

// WithUIStyle sets custom styling for UI elements
func WithUIStyle(style *UIStyle) UIOption {
	return func(u *UI) { u.Style = style }
}

// singletonLogger holds the single logger instance
var (
	singletonLogger *Logger
	once            sync.Once
)

// Init initializes the singleton logger with the given options
func Init(options ...Option) {
	once.Do(func() {
		singletonLogger = &Logger{Level: INFO, Output: os.Stdout}
		for _, opt := range options {
			opt(singletonLogger)
		}
	})
}

// GetLogger returns the singleton logger instance
func GetLogger() *Logger {
	if singletonLogger == nil {
		// Initialize with default options if not already initialized
		Init()
	}
	return singletonLogger
}

// New creates a new logger instance (non-singleton, for testing or special cases)
func New(options ...Option) *Logger {
	l := &Logger{Level: INFO, Output: os.Stdout}
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
	formattedMsg := l.formatMessage(msg)

	// Output to CLI using pterm
	printer.Println(formattedMsg)

	// Output to alternative destination if specified and not stdout
	if l.Output != nil && l.Output != os.Stdout {
		fmt.Fprintln(l.Output, formattedMsg)
	}
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

// Warning logs warnings (alias for Warn for compatibility)
func (l *Logger) Warning(message string, args ...any) {
	l.Warn(message, args...)
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
	formattedMsg := l.formatMessage(msg)

	// Create a custom printer for the tag
	printer := pterm.PrefixPrinter{
		Prefix: pterm.Prefix{
			Text:  tag,
			Style: pterm.NewStyle(tagColor(tag), pterm.FgLightWhite),
		},
	}
	printer.Println(formattedMsg)

	// Output to alternative destination if specified
	if l.Output != nil && l.Output != os.Stdout {
		fmt.Fprintln(l.Output, fmt.Sprintf("[%s] %s", tag, formattedMsg))
	}
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

// UI returns a UI instance for rendering CLI elements
func (l *Logger) UI(options ...UIOption) *UI {
	ui := &UI{Logger: l}
	for _, opt := range options {
		opt(ui)
	}
	return ui
}

// Header prints a styled header
func (u *UI) Header(title string) {
	header := pterm.DefaultHeader.WithFullWidth()
	if u.Style != nil && u.Style.HeaderBgColor != 0 {
		header = header.WithBackgroundStyle(pterm.NewStyle(u.Style.HeaderBgColor))
	} else {
		header = header.WithBackgroundStyle(pterm.NewStyle(pterm.BgDarkGray))
	}
	header.Println(title)
}

// Section prints a section header with styling
func (u *UI) Section(title string) {
	section := pterm.DefaultSection.WithLevel(2)
	section.Println(title)
}

// Table prints data in a table format
func (u *UI) Table(headers []string, data [][]string) {
	table := pterm.DefaultTable.WithHasHeader().WithData(append(pterm.TableData{headers}, data...))
	if u.Style != nil && u.Style.TableBorderStyle != nil {
		table = table.WithBoxed(true).WithStyle(u.Style.TableBorderStyle)
	}
	table.Render()
}

// RunSpinner runs a spinner for a task
func (u *UI) RunSpinner(text string, task func() error) error {
	spinner, _ := pterm.DefaultSpinner.WithText(text).Start()
	err := task()
	if err != nil {
		spinner.Fail("Failed: " + err.Error())
		return err
	}
	spinner.Success("Completed")
	return nil
}

// Prompt creates an interactive prompt based on the input type
func (u *UI) Prompt(question string, inputType InputType, options ...string) (any, error) {
	if question == "" {
		return nil, fmt.Errorf("prompt question cannot be empty")
	}
	switch inputType {
	case TextInput:
		return pterm.DefaultInteractiveTextInput.Show(question)
	case ConfirmInput:
		return pterm.DefaultInteractiveConfirm.Show(question)
	case SelectInput:
		if len(options) == 0 {
			return nil, fmt.Errorf("select input requires at least one option")
		}
		return pterm.DefaultInteractiveSelect.WithOptions(options).Show(question)
	case MultiSelectInput:
		if len(options) == 0 {
			return nil, fmt.Errorf("multi-select input requires at least one option")
		}
		return pterm.DefaultInteractiveMultiselect.WithOptions(options).Show(question)
	default:
		return nil, fmt.Errorf("unknown input type: %v", inputType)
	}
}

// Clear clears the terminal screen
func (u *UI) Clear() {
	pterm.Print("\033[H\033[2J")
}
