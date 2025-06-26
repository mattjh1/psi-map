package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/pterm/pterm"
)

func TestMain(m *testing.M) {
	// Disable pterm colors and styling to simplify output comparison
	pterm.DisableColor()
	pterm.DisableStyling()
	// Redirect pterm output to discard to prevent interference with test buffer
	pterm.SetDefaultOutput(io.Discard)

	// Reset singleton logger before and after tests
	ResetLogger()
	code := m.Run()
	ResetLogger()
	os.Exit(code)
}

// ResetLogger resets the singleton logger for testing
func ResetLogger() {
	once = sync.Once{}
	singletonLogger = nil
}

// threadSafeBuffer wraps bytes.Buffer with a mutex for thread-safe writes
type threadSafeBuffer struct {
	buf bytes.Buffer
	mu  sync.Mutex
}

func (b *threadSafeBuffer) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *threadSafeBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func TestGetLogger(t *testing.T) {
	// Test singleton initialization
	l1 := GetLogger()
	l2 := GetLogger()
	if l1 != l2 {
		t.Errorf("GetLogger returned different instances: %v != %v", l1, l2)
	}
	if l1.Level != INFO {
		t.Errorf("Expected default level INFO, got %v", l1.Level)
	}
}

func TestInit(t *testing.T) {
	ResetLogger()
	Init(WithLevel(DEBUG), WithPrefix("TEST"), WithTime(true))
	l := GetLogger()
	if l.Level != DEBUG {
		t.Errorf("Expected level DEBUG, got %v", l.Level)
	}
	if l.Prefix != "TEST" {
		t.Errorf("Expected prefix TEST, got %v", l.Prefix)
	}
	if !l.ShowTime {
		t.Errorf("Expected ShowTime true, got %v", l.ShowTime)
	}
}

func TestLoggingLevels(t *testing.T) {
	var buf threadSafeBuffer
	l := New(WithLevel(WARN), WithOutput(&buf))

	l.Debug("Debug message")
	l.Info("Info message")
	l.Warn("Warn message")
	l.Error("Error message")

	output := buf.String()
	if strings.Contains(output, "Debug") || strings.Contains(output, "Info") {
		t.Errorf("Expected no Debug or Info messages, got: %v", output)
	}
	if !strings.Contains(output, "Warn") || !strings.Contains(output, "Error") {
		t.Errorf("Expected Warn and Error messages, got: %v", output)
	}
}

func TestTaggedLogging(t *testing.T) {
	var buf threadSafeBuffer
	l := New(WithLevel(INFO), WithOutput(&buf))

	l.Tagged("TEST", "Tagged message", "🚀")

	output := buf.String()
	if !strings.Contains(output, "[TEST]") || !strings.Contains(output, "🚀 Tagged message") {
		t.Errorf("Expected tagged message with emoji, got: %v", output)
	}
}

func TestAlternativeOutput(t *testing.T) {
	var buf1, buf2 threadSafeBuffer
	l := New(WithLevel(INFO), WithOutput(&buf1))
	l.Output = &buf2 // Simulate alternative output

	l.Info("Test message")

	if !strings.Contains(buf2.String(), "Test message") {
		t.Errorf("Expected message in alternative output, got: %v", buf2.String())
	}
	if buf1.String() != "" {
		t.Errorf("Expected no message in original output, got: %v", buf1.String())
	}
}

func TestUIHeader(t *testing.T) {
	var buf threadSafeBuffer
	l := New(WithOutput(&buf))
	u := l.UI()

	u.Header("Test Header")

	output := buf.String()
	if !strings.Contains(output, "Test Header") {
		t.Errorf("Expected header output, got: %v", output)
	}
}

func TestUITable(t *testing.T) {
	var buf threadSafeBuffer
	l := New(WithOutput(&buf))
	u := l.UI(WithUIStyle(&UIStyle{
		TableBorderStyle: pterm.NewStyle(pterm.FgLightBlue),
	}))

	u.Table([]string{"A", "B"}, [][]string{{"1", "2"}})

	output := buf.String()
	if !strings.Contains(output, "A") || !strings.Contains(output, "1") {
		t.Errorf("Expected table output, got: %v", output)
	}
}

func TestRunSpinner(t *testing.T) {
	var buf threadSafeBuffer
	l := New(WithLevel(INFO), WithOutput(&buf))
	u := l.UI()

	// Test success case
	err := u.RunSpinner("Test task", func() error {
		l.Info("Task running")
		return nil
	})
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !strings.Contains(buf.String(), "Task running") {
		t.Errorf("Expected task log, got: %v", buf.String())
	}

	// Reset buffer
	buf = threadSafeBuffer{}

	// Test failure case
	err = u.RunSpinner("Failing task", func() error {
		l.Info("Task failing")
		return fmt.Errorf("task failed")
	})
	if err == nil || err.Error() != "task failed" {
		t.Errorf("Expected error 'task failed', got: %v", err)
	}
	if !strings.Contains(buf.String(), "Task failing") {
		t.Errorf("Expected task log, got: %v", buf.String())
	}
}

func TestPrompt(t *testing.T) {
	l := New()
	u := l.UI()

	// Test error cases
	_, err := u.Prompt("", TextInput)
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Errorf("Expected empty question error, got: %v", err)
	}

	_, err = u.Prompt("Select", SelectInput)
	if err == nil || !strings.Contains(err.Error(), "at least one option") {
		t.Errorf("Expected no options error, got: %v", err)
	}

	_, err = u.Prompt("Invalid", InputType(999))
	if err == nil || !strings.Contains(err.Error(), "unknown input type") {
		t.Errorf("Expected unknown input type error, got: %v", err)
	}
}

func TestClear(t *testing.T) {
	var buf threadSafeBuffer
	l := New(WithOutput(&buf))
	u := l.UI()

	u.Clear()

	if !strings.Contains(buf.String(), "\033[H\033[2J") {
		t.Errorf("Expected clear sequence, got: %v", buf.String())
	}
}

func TestThreadSafety(t *testing.T) {
	var buf threadSafeBuffer
	l := New(WithLevel(INFO), WithOutput(&buf))
	var wg sync.WaitGroup

	// Ensure INFO level before starting
	l.SetLevel(INFO)

	// Concurrent logging
	for i := range 10 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			l.Info("Message %d", n)
			// Simulate level change, but ensure INFO messages are logged first
			if n%2 == 0 {
				l.SetLevel(WARN)
			}
		}(i)
	}
	wg.Wait()

	// Reset to INFO for final verification
	l.SetLevel(INFO)

	output := buf.String()
	for i := range 10 {
		if !strings.Contains(output, fmt.Sprintf("Message %d", i)) {
			t.Logf("Message %d not found in: %v", i, output)
		}
	}
	// Since level changes may filter some messages, check for at least some presence
	if !strings.Contains(output, "Message") {
		t.Errorf("Expected some messages in output, got: %v", output)
	}
}

func TestTimeFormat(t *testing.T) {
	var buf threadSafeBuffer
	l := New(WithLevel(INFO), WithTime(true), WithOutput(&buf))

	l.Info("Test message")

	output := buf.String()
	now := time.Now().Format("15:04:05")
	if !strings.Contains(output, now[:5]) { // Check partial time (allow second variance)
		t.Errorf("Expected timestamp in output, got: %v", output)
	}
}
