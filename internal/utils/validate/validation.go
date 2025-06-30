package validate

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

const (
	// DefaultDirPermissions sets the default permissions for created directories
	DefaultDirPermissions = 0o755
)

var (
	ErrInvalidPath      = fmt.Errorf("invalid file path")
	ErrPathTraversal    = fmt.Errorf("path traversal attempt detected")
	ErrInvalidFileName  = fmt.Errorf("invalid file name")
	ErrInvalidExtension = fmt.Errorf("invalid file extension")
)

// ValidateOutputPath validates and sanitizes output paths for file creation
func ValidateOutputPath(outputDir, name, extension string) (string, error) {
	// Validate output directory
	cleanOutputDir, err := ValidateDirectory(outputDir)
	if err != nil {
		return "", fmt.Errorf("invalid output directory: %w", err)
	}

	// Validate and sanitize filename
	cleanName, err := ValidateFileName(name)
	if err != nil {
		return "", fmt.Errorf("invalid filename: %w", err)
	}

	// Validate extension
	cleanExt, err := ValidateExtension(extension)
	if err != nil {
		return "", fmt.Errorf("invalid extension: %w", err)
	}

	// Construct the full path
	fullPath := filepath.Join(cleanOutputDir, cleanName+cleanExt)

	// Final validation - ensure the resolved path is within the output directory
	absOutputDir, err := filepath.Abs(cleanOutputDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve output directory: %w", err)
	}

	absFullPath, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve output path: %w", err)
	}

	// Normalize paths to ensure consistent separator handling
	absOutputDir = filepath.Clean(absOutputDir)
	absFullPath = filepath.Clean(absFullPath)

	// Check if the resolved path is within the allowed directory
	// Use filepath.Rel to check if the path escapes the directory
	relPath, err := filepath.Rel(absOutputDir, absFullPath)
	if err != nil {
		return "", fmt.Errorf("path traversal detected: %w", ErrPathTraversal)
	}

	// If the relative path starts with "..", it means it's outside the directory
	if strings.HasPrefix(relPath, "..") || strings.Contains(relPath, string(filepath.Separator)+"..") {
		return "", fmt.Errorf("path traversal detected: %w", ErrPathTraversal)
	}

	return fullPath, nil
}

// ValidateDirectory validates and sanitizes directory paths
func ValidateDirectory(dir string) (string, error) {
	if dir == "" {
		return "", fmt.Errorf("directory cannot be empty: %w", ErrInvalidPath)
	}

	// Check for path traversal attempts BEFORE calling filepath.Clean()
	if strings.Contains(dir, "..") {
		return "", fmt.Errorf("directory contains path traversal: %w", ErrPathTraversal)
	}

	// Clean the path
	cleanDir := filepath.Clean(dir)

	// Convert to absolute path for consistency
	absDir, err := filepath.Abs(cleanDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve directory path: %w", err)
	}

	return absDir, nil
}

// ValidateFileName validates and sanitizes filenames
func ValidateFileName(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("filename cannot be empty: %w", ErrInvalidFileName)
	}

	// Check for path traversal attempts BEFORE calling filepath.Base()
	if strings.Contains(name, "..") {
		return "", fmt.Errorf("path traversal detected in filename: %w", ErrInvalidFileName)
	}

	// Check for path separators (shouldn't be in a filename)
	if strings.ContainsAny(name, "/\\") {
		return "", fmt.Errorf("filename cannot contain path separators: %w", ErrInvalidFileName)
	}

	// Remove any path components (this should now be safe)
	cleanName := filepath.Base(name)

	// Check for invalid characters
	validName := regexp.MustCompile(`^[a-zA-Z0-9_\-.]+$`)
	if !validName.MatchString(cleanName) {
		return "", fmt.Errorf("filename contains invalid characters: %w", ErrInvalidFileName)
	}

	// Check for reserved names on Windows
	reservedNames := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
	upperName := strings.ToUpper(cleanName)
	if slices.Contains(reservedNames, upperName) {
		return "", fmt.Errorf("filename is reserved: %w", ErrInvalidFileName)
	}

	return cleanName, nil
}

// ValidateExtension validates file extensions
func ValidateExtension(ext string) (string, error) {
	if ext == "" {
		return "", fmt.Errorf("extension cannot be empty: %w", ErrInvalidExtension)
	}

	// Ensure extension starts with a dot
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	// Whitelist allowed extensions
	allowedExtensions := map[string]bool{
		".json": true,
		".html": true,
		".xml":  true,
		".txt":  true,
	}

	if !allowedExtensions[strings.ToLower(ext)] {
		return "", fmt.Errorf("extension not allowed: %w", ErrInvalidExtension)
	}

	return ext, nil
}

// ValidateInputPath validates paths for reading files (like sitemaps)
func ValidateInputPath(inputPath string) (string, error) {
	if inputPath == "" {
		return "", fmt.Errorf("input path cannot be empty: %w", ErrInvalidPath)
	}

	// Clean the path
	cleanPath := filepath.Clean(inputPath)

	// Check for path traversal
	if strings.Contains(cleanPath, "..") {
		return "", fmt.Errorf("input path contains path traversal: %w", ErrPathTraversal)
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve input path: %w", err)
	}

	// Check if file exists and is readable
	info, err := os.Stat(absPath)
	if err != nil {
		return "", fmt.Errorf("cannot access file: %w", err)
	}

	// Ensure it's a regular file
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("path is not a regular file: %s", absPath)
	}

	return absPath, nil
}

// SafeCreateFile creates a file safely with proper validation
func SafeCreateFile(outputDir, name, extension string) (*os.File, string, error) {
	validPath, err := ValidateOutputPath(outputDir, name, extension)
	if err != nil {
		return nil, "", err
	}

	// Ensure the directory exists
	dir := filepath.Dir(validPath)
	if err := os.MkdirAll(dir, DefaultDirPermissions); err != nil {
		return nil, "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the file
	file, err := os.Create(validPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create file: %w", err)
	}

	return file, validPath, nil
}

// SafeOpenFile opens a file safely with proper validation
func SafeOpenFile(inputPath string) (*os.File, error) {
	validPath, err := ValidateInputPath(inputPath)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(validPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// PathComponents represents the components of a file path
type PathComponents struct {
	Dir       string
	Name      string
	Extension string
	Base      string
}

// SplitFilePath splits a file path into its components
func SplitFilePath(filePath string) PathComponents {
	dir := filepath.Dir(filePath)
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)

	return PathComponents{
		Dir:       dir,
		Name:      name,
		Extension: ext,
		Base:      base,
	}
}

// SafeCreateFileFromPath creates a file safely using an existing validated path
func SafeCreateFileFromPath(validatedPath string) (*os.File, error) {
	// Ensure the directory exists
	dir := filepath.Dir(validatedPath)
	if err := os.MkdirAll(dir, DefaultDirPermissions); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Create the file
	file, err := os.Create(validatedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	return file, nil
}
