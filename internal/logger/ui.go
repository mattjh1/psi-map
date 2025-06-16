package logger

import (
	"github.com/mattjh1/psi-map/internal/types/logger"
)

// UI returns a UI instance for rendering CLI elements
func UI(l *logger.Logger) *logger.UI {
	return l.UI()
}

// Header prints a styled header
func Header(u *logger.UI, title string) {
	u.Header(title)
}

// Table prints data in a table format
func Table(u *logger.UI, headers []string, data [][]string) {
	u.Table(headers, data)
}

// RunSpinner runs a spinner for a task
func RunSpinner(u *logger.UI, text string, task func() error) error {
	return u.RunSpinner(text, task)
}

// Prompt creates an interactive prompt based on the input type
func Prompt(u *logger.UI, question string, inputType logger.InputType, options ...string) (any, error) {
	return u.Prompt(question, inputType, options...)
}
