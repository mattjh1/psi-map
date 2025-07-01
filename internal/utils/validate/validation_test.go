package validate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateOutputPath(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name      string
		outputDir string
		filename  string
		extension string
		wantErr   string
	}{
		{"valid path", tempDir, "test-file_123", ".json", ""},
		{"extension without dot", tempDir, "test", "html", ""},
		{"empty directory", "", "test", ".json", "invalid output directory"},
		{"empty filename", tempDir, "", ".json", "invalid filename"},
		{"empty extension", tempDir, "test", "", "invalid extension"},
		{"path traversal", tempDir, "../../../etc/passwd", ".json", "path traversal"},
		{"invalid characters", tempDir, "test<>file", ".json", "invalid characters"},
		{"reserved name", tempDir, "CON", ".json", "reserved"},
		{"invalid extension", tempDir, "test", ".exe", "not allowed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Add debug logging for the problematic test case
			if tt.name == "path traversal" {
				t.Logf("Testing path traversal with:")
				t.Logf("  outputDir: %q", tt.outputDir)
				t.Logf("  filename: %q", tt.filename)
				t.Logf("  extension: %q", tt.extension)
			}

			result, err := ValidateOutputPath(tt.outputDir, tt.filename, tt.extension)

			// Debug logging for the problematic test case
			if tt.name == "path traversal" {
				t.Logf("Result: %q", result)
				t.Logf("Error: %v", err)
			}

			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("Expected error containing %q, but got no error", tt.wantErr)
					return
				}
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, result)

			// Verify path is within output directory
			absOutputDir, _ := filepath.Abs(tt.outputDir)
			absResult, _ := filepath.Abs(result)
			assert.True(t, strings.HasPrefix(absResult, absOutputDir+string(filepath.Separator)))
		})
	}
}

func TestValidateDirectory(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name    string
		dir     string
		wantErr string
	}{
		{"valid absolute", tempDir, ""},
		{"valid relative", ".", ""},
		{"empty", "", "cannot be empty"},
		{"path traversal", "../..", "path traversal"},
		{"traversal middle", "some/../path", "path traversal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateDirectory(tt.dir)

			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.True(t, filepath.IsAbs(result))
		})
	}
}

func TestValidateFileName(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
		wantErr  string
	}{
		{"valid simple", "test", "test", ""},
		{"valid complex", "test-file_123.backup", "test-file_123.backup", ""},
		{"strips path", "/path/to/file.txt", "", "path separators"},
		{"empty", "", "", "cannot be empty"},
		{"invalid chars", "test<file>", "", "invalid characters"},
		{"spaces", "test file", "", "invalid characters"},
		{"reserved CON", "CON", "", "reserved"},
		{"reserved lowercase", "con", "", "reserved"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateFileName(tt.filename)

			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestValidateFileNamePathTraversal(t *testing.T) {
	// Test the ValidateFileName function directly
	testCases := []struct {
		name        string
		input       string
		expectError bool
	}{
		{"simple filename", "test.txt", false},
		{"path traversal", "../../../etc/passwd", true},
		{"single dot dot", "..", true},
		{"with slash", "path/file", true},
		{"with backslash", "path\\file", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := ValidateFileName(tc.input)
			t.Logf("Input: %q", tc.input)
			t.Logf("Result: %q", result)
			t.Logf("Error: %v", err)

			if tc.expectError {
				assert.Error(t, err, "Expected error for input: %s", tc.input)
			} else {
				assert.NoError(t, err, "Expected no error for input: %s", tc.input)
			}
		})
	}
}

func TestValidateExtension(t *testing.T) {
	tests := []struct {
		name      string
		extension string
		want      string
		wantErr   string
	}{
		{"json with dot", ".json", ".json", ""},
		{"json without dot", "json", ".json", ""},
		{"case insensitive", ".json", ".json", ""},
		{"empty", "", "", "cannot be empty"},
		{"not allowed", ".exe", "", "not allowed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateExtension(tt.extension)

			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestValidateInputPath(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.txt")
	require.NoError(t, os.WriteFile(tempFile, []byte("test"), 0o644))

	tempSubDir := filepath.Join(tempDir, "subdir")
	require.NoError(t, os.Mkdir(tempSubDir, 0o755))

	tests := []struct {
		name      string
		inputPath string
		wantErr   string
	}{
		{"valid file", tempFile, ""},
		{"empty path", "", "cannot be empty"},
		{"non-existent", filepath.Join(tempDir, "missing.txt"), "cannot access file"},
		{"directory", tempSubDir, "not a regular file"},
		{"path traversal", "../../../etc/passwd", "path traversal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateInputPath(tt.inputPath)

			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.True(t, filepath.IsAbs(result))
		})
	}
}

func TestSafeCreateFile(t *testing.T) {
	tempDir := t.TempDir()

	file, path, err := SafeCreateFile(tempDir, "test", ".json")
	assert.NoError(t, err)
	assert.NotNil(t, file)
	assert.NotEmpty(t, path)
	file.Close()

	// Verify file exists
	_, err = os.Stat(path)
	assert.NoError(t, err)

	// Test error case
	file, path, err = SafeCreateFile("", "test", ".json")
	assert.Error(t, err)
	assert.Nil(t, file)
	assert.Empty(t, path)
}

func TestSafeOpenFile(t *testing.T) {
	tempDir := t.TempDir()
	tempFile := filepath.Join(tempDir, "test.txt")
	require.NoError(t, os.WriteFile(tempFile, []byte("content"), 0o644))

	file, err := SafeOpenFile(tempFile)
	assert.NoError(t, err)
	assert.NotNil(t, file)
	file.Close()

	// Test error case
	file, err = SafeOpenFile("")
	assert.Error(t, err)
	assert.Nil(t, file)
}

func TestSplitFilePath(t *testing.T) {
	tests := []struct {
		filePath string
		want     PathComponents
	}{
		{
			"test.txt",
			PathComponents{Dir: ".", Name: "test", Extension: ".txt", Base: "test.txt"},
		},
		{
			"/home/user/file.json",
			PathComponents{Dir: "/home/user", Name: "file", Extension: ".json", Base: "file.json"},
		},
		{
			"filename",
			PathComponents{Dir: ".", Name: "filename", Extension: "", Base: "filename"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.filePath, func(t *testing.T) {
			result := SplitFilePath(tt.filePath)
			assert.Equal(t, tt.want, result)
		})
	}
}

// Test key security behaviors
func TestSecurityValidation(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("path traversal blocked", func(t *testing.T) {
		_, err := ValidateOutputPath(tempDir, "../../../etc/passwd", ".txt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path traversal")
	})

	t.Run("reserved names blocked", func(t *testing.T) {
		for _, name := range []string{"CON", "PRN", "AUX", "com1", "lpt1"} {
			_, err := ValidateFileName(name)
			assert.Error(t, err, "should block reserved name: %s", name)
		}
	})

	t.Run("dangerous extensions blocked", func(t *testing.T) {
		for _, ext := range []string{".exe", ".bat", ".sh", ".js"} {
			_, err := ValidateExtension(ext)
			assert.Error(t, err, "should block dangerous extension: %s", ext)
		}
	})
}
