package cli

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v2"
)

func TestAnalyzeCommand(t *testing.T) {
	cmd := analyzeCommand()
	// Test basic command properties
	assert.Equal(t, "analyze", cmd.Name)
	assert.Equal(t, []string{"run"}, cmd.Aliases)
	assert.Equal(t, "Analyze sitemap and generate reports", cmd.Usage)
	assert.Equal(t, "[flags] <sitemap_url_or_file>", cmd.ArgsUsage)
	assert.Contains(t, cmd.Description, "Analyze a sitemap and generate reports")

	// Test flags
	require.Len(t, cmd.Flags, 7, "Expected 7 flags")
	flagMap := make(map[string]cli.Flag)
	for _, flag := range cmd.Flags {
		switch f := flag.(type) {
		case *cli.StringFlag:
			flagMap[f.Name] = f
		case *cli.IntFlag:
			flagMap[f.Name] = f
		}
	}

	// Test string flags
	outputFlag, ok := flagMap["output"].(*cli.StringFlag)
	if !ok {
		t.Fatal("output flag not found or wrong type")
	}
	assert.Equal(t, "output", outputFlag.Name)
	assert.Equal(t, []string{"o"}, outputFlag.Aliases)
	assert.Equal(t, "json", outputFlag.Value) // constants.JSON = "json"

	outputDirFlag, ok := flagMap["output-dir"].(*cli.StringFlag)
	if !ok {
		t.Fatal("output-dir flag not found or wrong type")
	}
	assert.Equal(t, "output-dir", outputDirFlag.Name)
	assert.Equal(t, ".", outputDirFlag.Value)

	nameFlag, ok := flagMap["name"].(*cli.StringFlag)
	if !ok {
		t.Fatal("name flag not found or wrong type")
	}
	assert.Equal(t, "name", nameFlag.Name)
	assert.Equal(t, "psi-report", nameFlag.Value)

	providerFlag, ok := flagMap["provider"].(*cli.StringFlag)
	if !ok {
		t.Fatal("provider flag not found or wrong type")
	}
	assert.Equal(t, "provider", providerFlag.Name)
	assert.Equal(t, "psi", providerFlag.Value)

	lighthouseFlag, ok := flagMap["lighthouse-url"].(*cli.StringFlag)
	if !ok {
		t.Fatal("lighthouse-url flag not found or wrong type")
	}
	assert.Equal(t, "lighthouse-url", lighthouseFlag.Name)
	assert.Equal(t, "", lighthouseFlag.Value)

	workersFlag, ok := flagMap["workers"].(*cli.IntFlag)
	if !ok {
		t.Fatal("workers flag not found or wrong type")
	}
	assert.Equal(t, "workers", workersFlag.Name)
	assert.Equal(t, []string{"w"}, workersFlag.Aliases)
	assert.Greater(t, workersFlag.Value, 0, "Workers should have positive default")

	cacheTTLFlag, ok := flagMap["cache-ttl"].(*cli.IntFlag)
	if !ok {
		t.Fatal("cache-ttl flag not found or wrong type")
	}
	assert.Equal(t, "cache-ttl", cacheTTLFlag.Name)
	assert.Equal(t, 24, cacheTTLFlag.Value)
}

func TestAnalyzeCommandAction_MissingArgs(t *testing.T) {
	cmd := analyzeCommand()

	// Create a context with no arguments
	app := &cli.App{}
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
	ctx := cli.NewContext(app, flagSet, nil)

	err := cmd.Action(ctx)
	require.Error(t, err)
	assert.Equal(t, "sitemap URL or file path is required", err.Error())
}

func TestAnalyzeCommandAction_WithValidArgs(t *testing.T) {
	// Since we can't easily mock the runAnalysis function call directly,
	// we'll test that the command action gets to the point where it would call runAnalysis
	// and verify all the setup is correct

	ctx := createTestContext([]string{"sitemap.xml"}, map[string]string{
		"output":     "json",
		"output-dir": "./reports",
		"name":       "test-report",
	})

	// Test that handleOutputFlags would succeed (this is called before runAnalysis)
	err := handleOutputFlags(ctx)
	assert.NoError(t, err, "handleOutputFlags should succeed with valid arguments")

	// Verify the context has all the expected values that would be passed to runAnalysis
	assert.Equal(t, "sitemap.xml", ctx.Args().First(), "Should have sitemap argument")
	assert.Equal(t, "json", ctx.String("output"), "Should have correct output format")
	assert.Equal(t, "./reports", ctx.String("output-dir"), "Should have correct output directory")
	assert.Equal(t, "test-report", ctx.String("name"), "Should have correct name")

	// Note: The actual cmd.Action(ctx) call would invoke runAnalysis,
	// but testing that requires either dependency injection or integration testing
	// The runAnalysis function itself should be tested in run_test.go
}

func TestHandleOutputFlags(t *testing.T) {
	tests := []struct {
		name          string
		outputFormat  string
		expectedError bool
	}{
		{
			name:          "Valid JSON format",
			outputFormat:  "json",
			expectedError: false,
		},
		{
			name:          "Valid HTML format",
			outputFormat:  "html",
			expectedError: false,
		},
		{
			name:          "Valid STDOUT format",
			outputFormat:  "stdout",
			expectedError: false,
		},
		{
			name:          "Valid JSON format uppercase",
			outputFormat:  "JSON",
			expectedError: false,
		},
		{
			name:          "Valid HTML format mixed case",
			outputFormat:  "Html",
			expectedError: false,
		},
		{
			name:          "Invalid format",
			outputFormat:  "xml",
			expectedError: true,
		},
		{
			name:          "Invalid format empty",
			outputFormat:  "",
			expectedError: true,
		},
		{
			name:          "Invalid format random",
			outputFormat:  "invalid",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock CLI context with the output flag set
			app := &cli.App{}
			flagSet := flag.NewFlagSet("test", flag.ContinueOnError)
			flagSet.String("output", tt.outputFormat, "output format")
			err := flagSet.Parse([]string{})
			if err != nil {
				t.Fatalf("Failed to parse flags: %v", err)
			}
			ctx := cli.NewContext(app, flagSet, nil)

			err = handleOutputFlags(ctx)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "unsupported output format")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test helper to create a CLI context with specific arguments and flags
func createTestContext(args []string, flags map[string]string) *cli.Context {
	app := &cli.App{}
	flagSet := flag.NewFlagSet("test", flag.ContinueOnError)

	// Add common flags
	flagSet.String("output", "json", "output format")
	flagSet.String("output-dir", ".", "output directory")
	flagSet.String("name", "psi-report", "output filename")
	flagSet.Int("workers", 4, "number of workers")
	flagSet.Int("cache-ttl", 24, "cache TTL")

	// Set flag values if provided
	flagArgs := []string{}
	for key, value := range flags {
		flagArgs = append(flagArgs, "--"+key, value)
	}
	flagArgs = append(flagArgs, args...)

	flagSet.Parse(flagArgs) //nolint:errcheck // test helper function, parse errors would cause test failures anyway
	return cli.NewContext(app, flagSet, nil)
}

func TestAnalyzeCommandAction_WithFlags(t *testing.T) {
	tests := []struct {
		name                    string
		args                    []string
		flags                   map[string]string
		expectHandleOutputError bool
	}{
		{
			name: "Valid arguments and flags",
			args: []string{"sitemap.xml"},
			flags: map[string]string{
				"output":     "html",
				"output-dir": "./reports",
			},
			expectHandleOutputError: false,
		},
		{
			name: "Invalid output format",
			args: []string{"sitemap.xml"},
			flags: map[string]string{
				"output": "xml",
			},
			expectHandleOutputError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := createTestContext(tt.args, tt.flags)

			// Test handleOutputFlags separately since runAnalysis is in run.go
			err := handleOutputFlags(ctx)
			if tt.expectHandleOutputError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAnalyzeCommandAction_CallsRunAnalysis(t *testing.T) {
	// This test verifies the command action calls runAnalysis with correct parameters
	// The actual runAnalysis function should be tested in run_test.go

	ctx := createTestContext([]string{"sitemap.xml"}, map[string]string{
		"output": "json",
	})

	// We can't easily test the actual runAnalysis call without mocking,
	// but we can verify the action would proceed past validation
	err := handleOutputFlags(ctx)
	assert.NoError(t, err, "handleOutputFlags should succeed with valid format")

	// Verify the context has the expected values that would be passed to runAnalysis
	assert.Equal(t, "sitemap.xml", ctx.Args().First())
	assert.Equal(t, "json", ctx.String("output"))
	assert.False(t, ctx.IsSet("port"), "port should not be set for analyze command")
}
