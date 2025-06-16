package logger

import (
	"fmt"

	"github.com/pterm/pterm"
)

// UI provides methods for rendering CLI UI elements
type UI struct {
	Logger *Logger
}

// InputType defines the type of interactive prompt
type InputType int

const (
	TextInput InputType = iota
	ConfirmInput
	SelectInput
	MultiSelectInput
)

// UI returns a UI instance for rendering CLI elements
func (l *Logger) UI() *UI {
	return &UI{Logger: l}
}

// Header prints a styled header
func (u *UI) Header(title string) {
	pterm.DefaultHeader.WithFullWidth().WithBackgroundStyle(pterm.NewStyle(pterm.BgDarkGray)).Println(title)
}

// Table prints data in a table format
func (u *UI) Table(headers []string, data [][]string) {
	tableData := pterm.TableData{headers}
	tableData = append(tableData, data...)
	pterm.DefaultTable.WithHasHeader().WithData(tableData).Render()
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
	switch inputType {
	case TextInput:
		return pterm.DefaultInteractiveTextInput.Show(question)
	case ConfirmInput:
		return pterm.DefaultInteractiveConfirm.Show(question)
	case SelectInput:
		return pterm.DefaultInteractiveSelect.WithOptions(options).Show(question)
	case MultiSelectInput:
		return pterm.DefaultInteractiveMultiselect.WithOptions(options).Show(question)
	default:
		return nil, fmt.Errorf("unknown input type")
	}
}
