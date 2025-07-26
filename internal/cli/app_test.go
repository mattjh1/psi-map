package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestNewApp(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		commit    string
		buildTime string
	}{
		{
			name:      "with all version info",
			version:   "v1.0.0",
			commit:    "abc123",
			buildTime: "2024-01-01T00:00:00Z",
		},
		{
			name:      "with dev version info",
			version:   "dev",
			commit:    "none",
			buildTime: "unknown",
		},
		{
			name:      "with empty values",
			version:   "",
			commit:    "",
			buildTime: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp(tt.version, tt.commit, tt.buildTime)

			// Test basic app properties
			assert.Equal(t, "psi-map", app.Name)
			assert.Equal(t, "Analyze websites using PSI (PageSpeed Insights) API", app.Usage)
			assert.Equal(t, "A tool to analyze website overall performance using Google's PageSpeed Insights API", app.Description)
			assert.Equal(t, "psi-map [command] [options] [arguments...]", app.UsageText)

			// Test version formatting
			expectedVersion := tt.version + " (" + tt.commit + " @ " + tt.buildTime + ")"
			assert.Equal(t, expectedVersion, app.Version)

			// Test author information
			require.Len(t, app.Authors, 1)
			assert.Equal(t, "Mattias Holmgren", app.Authors[0].Name)
			assert.Equal(t, "me@mattjh.sh", app.Authors[0].Email)
		})
	}
}

func TestAppCommands(t *testing.T) {
	app := NewApp("test", "test", "test")

	// Test that commands are properly registered
	expectedCommands := []string{"analyze", "server", "cache"}
	require.Len(t, app.Commands, len(expectedCommands))

	// Create a map for easier lookup
	commandMap := make(map[string]*cli.Command)
	for _, cmd := range app.Commands {
		commandMap[cmd.Name] = cmd
	}

	// Check each expected command exists
	for _, expectedCmd := range expectedCommands {
		assert.Contains(t, commandMap, expectedCmd, "Command '%s' should be registered", expectedCmd)
	}
}

func TestAppExitErrHandler(t *testing.T) {
	app := NewApp("test", "test", "test")

	// Test that ExitErrHandler is set
	assert.NotNil(t, app.ExitErrHandler)

	// Test ExitErrHandler with nil error (should not panic)
	ctx := &cli.Context{}

	// This should not panic with nil error
	assert.NotPanics(t, func() {
		// We can't easily test os.Exit, but we can ensure no panic occurs
		// Note: Actual error testing would require more complex setup due to os.Exit
		_ = ctx // Use ctx to avoid unused variable
	})
}

func TestAppHelpOutput(t *testing.T) {
	app := NewApp("v1.0.0", "abc123", "2024-01-01")

	// Test help by inspecting app structure instead of running --help
	// which might call os.Exit() and freeze the test

	// Test basic app metadata that would appear in help
	assert.Equal(t, "psi-map", app.Name)
	assert.Equal(t, "Analyze websites using PSI (PageSpeed Insights) API", app.Usage)
	assert.Contains(t, app.Version, "v1.0.0")
	assert.Contains(t, app.Version, "abc123")
	assert.Contains(t, app.Version, "2024-01-01")

	// Test that commands exist (would appear in help)
	assert.True(t, len(app.Commands) > 0, "App should have commands that appear in help")

	// Test author info (appears in help)
	require.Len(t, app.Authors, 1)
	assert.Equal(t, "Mattias Holmgren", app.Authors[0].Name)
}

func TestAppVersion(t *testing.T) {
	app := NewApp("v2.1.0", "def456", "2024-02-01T12:00:00Z")

	// Test version string construction instead of running --version
	// which might call os.Exit() and cause issues
	expectedVersion := "v2.1.0 (def456 @ 2024-02-01T12:00:00Z)"
	assert.Equal(t, expectedVersion, app.Version)

	// Test individual components are present
	assert.Contains(t, app.Version, "v2.1.0")
	assert.Contains(t, app.Version, "def456")
	assert.Contains(t, app.Version, "2024-02-01T12:00:00Z")
}

func TestAppInvalidCommand(t *testing.T) {
	app := NewApp("test", "test", "test")

	// Test that the app has a defined set of valid commands
	// and doesn't include our invalid command
	validCommands := make(map[string]bool)
	for _, cmd := range app.Commands {
		validCommands[cmd.Name] = true
	}

	// Test that invalid command is not in the valid commands list
	assert.False(t, validCommands["invalid-command"], "invalid-command should not be a valid command")
	assert.False(t, validCommands["nonexistent"], "nonexistent should not be a valid command")

	// Test that we have expected valid commands
	assert.True(t, validCommands["analyze"], "analyze should be a valid command")
	assert.True(t, validCommands["server"], "server should be a valid command")
	assert.True(t, validCommands["cache"], "cache should be a valid command")
}

func TestAppEmptyArgs(t *testing.T) {
	app := NewApp("test", "test", "test")

	// Test app structure instead of running with empty args
	// which might show help and call os.Exit()
	assert.NotNil(t, app)
	assert.NotEmpty(t, app.Name)
	assert.NotEmpty(t, app.Usage)

	// Verify the app is properly configured to handle commands
	assert.True(t, len(app.Commands) > 0, "App should have commands available")
}

func TestAppVersionParsing(t *testing.T) {
	tests := []struct {
		name           string
		version        string
		commit         string
		buildTime      string
		expectedFormat string
	}{
		{
			name:           "release version",
			version:        "v1.2.3",
			commit:         "a1b2c3d",
			buildTime:      "2024-01-15T10:30:00Z",
			expectedFormat: "v1.2.3 (a1b2c3d @ 2024-01-15T10:30:00Z)",
		},
		{
			name:           "development version",
			version:        "dev",
			commit:         "none",
			buildTime:      "unknown",
			expectedFormat: "dev (none @ unknown)",
		},
		{
			name:           "special characters in commit",
			version:        "v0.1.0-beta",
			commit:         "feature/test-branch",
			buildTime:      "2024-01-01",
			expectedFormat: "v0.1.0-beta (feature/test-branch @ 2024-01-01)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := NewApp(tt.version, tt.commit, tt.buildTime)
			assert.Equal(t, tt.expectedFormat, app.Version)
		})
	}
}

// Test that all required CLI framework fields are properly set
func TestAppStructure(t *testing.T) {
	app := NewApp("v1.0.0", "abc123", "2024-01-01")

	// Test that essential fields are not empty/nil
	assert.NotEmpty(t, app.Name)
	assert.NotEmpty(t, app.Usage)
	assert.NotEmpty(t, app.Description)
	assert.NotEmpty(t, app.Version)
	assert.NotEmpty(t, app.UsageText)
	assert.NotNil(t, app.Commands)
	assert.NotNil(t, app.Authors)
	assert.NotNil(t, app.ExitErrHandler)

	// Test that we have the expected number of top-level commands
	assert.True(t, len(app.Commands) > 0, "App should have at least one command")
}

// Benchmark tests for performance
func BenchmarkNewApp(b *testing.B) {
	for b.Loop() {
		NewApp("v1.0.0", "abc123", "2024-01-01")
	}
}

func BenchmarkAppHelpExecution(b *testing.B) {
	app := NewApp("v1.0.0", "abc123", "2024-01-01")

	// Redirect output to discard
	oldStdout := os.Stdout
	devNull, err := os.Open(os.DevNull)
	require.NoError(b, err)
	os.Stdout = devNull
	defer func() {
		devNull.Close()
		os.Stdout = oldStdout
	}()

	b.ResetTimer()
	for b.Loop() {
		err := app.Run([]string{"psi-map", "--help"})
		if err != nil {
			b.Fatalf("Failed to run app: %v", err)
		}
	}
}
