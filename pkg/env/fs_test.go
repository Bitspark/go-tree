package env

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestStandardModuleFSInitialization tests initialization of the standard module filesystem
func TestStandardModuleFSInitialization(t *testing.T) {
	fs := NewStandardModuleFS()

	// Verify it can be created
	if fs == nil {
		t.Errorf("Expected non-nil ModuleFS, got nil")
	}
}

// TestStandardModuleFSReadFile tests the ReadFile method
func TestStandardModuleFSReadFile(t *testing.T) {
	fs := NewStandardModuleFS()

	// Create a temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("test content")

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test reading existing file
	content, err := fs.ReadFile(testFile)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if string(content) != string(testContent) {
		t.Errorf("Expected content '%s', got '%s'", string(testContent), string(content))
	}

	// Test reading non-existent file
	_, err = fs.ReadFile(filepath.Join(tmpDir, "nonexistent.txt"))
	if err == nil {
		t.Errorf("Expected error for non-existent file, got nil")
	}
}

// TestStandardModuleFSWriteFile tests the WriteFile method
func TestStandardModuleFSWriteFile(t *testing.T) {
	fs := NewStandardModuleFS()

	// Create a temporary directory
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "write-test.txt")
	testContent := []byte("test write content")

	// Test writing to file
	err := fs.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify the file was created with correct content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("Failed to read written file: %v", err)
	}
	if string(content) != string(testContent) {
		t.Errorf("Expected content '%s', got '%s'", string(testContent), string(content))
	}

	// Test writing to a file in a non-existent directory
	nonExistentDir := filepath.Join(tmpDir, "nonexistent")
	nonExistentFile := filepath.Join(nonExistentDir, "test.txt")

	err = fs.WriteFile(nonExistentFile, testContent, 0644)
	if err == nil {
		t.Errorf("Expected error writing to non-existent directory, got nil")
	}
}

// TestStandardModuleFSMkdirAll tests the MkdirAll method
func TestStandardModuleFSMkdirAll(t *testing.T) {
	fs := NewStandardModuleFS()

	// Create a temporary base directory
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "test-dir", "nested-dir")

	// Test creating nested directories
	err := fs.MkdirAll(testDir, 0755)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify directories were created
	info, err := os.Stat(testDir)
	if err != nil {
		t.Errorf("Failed to stat created directory: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("Expected a directory, got a file")
	}

	// Test creating an already existing directory (should not error)
	err = fs.MkdirAll(testDir, 0755)
	if err != nil {
		t.Errorf("Expected no error for existing directory, got: %v", err)
	}
}

// TestStandardModuleFSRemoveAll tests the RemoveAll method
func TestStandardModuleFSRemoveAll(t *testing.T) {
	fs := NewStandardModuleFS()

	// Create a temporary directory with some files
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "test-remove-dir")
	nestedDir := filepath.Join(testDir, "nested-dir")

	// Create directory structure
	err := os.MkdirAll(nestedDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directories: %v", err)
	}

	// Create some files
	testFile1 := filepath.Join(testDir, "file1.txt")
	testFile2 := filepath.Join(nestedDir, "file2.txt")

	err = os.WriteFile(testFile1, []byte("file1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file1: %v", err)
	}

	err = os.WriteFile(testFile2, []byte("file2"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file2: %v", err)
	}

	// Test removing the directory and all contents
	err = fs.RemoveAll(testDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify directory no longer exists
	_, err = os.Stat(testDir)
	if !os.IsNotExist(err) {
		t.Errorf("Expected directory to be removed, but it still exists")
	}

	// Test removing a non-existent directory (should not error)
	err = fs.RemoveAll(testDir)
	if err != nil {
		t.Errorf("Expected no error for non-existent directory, got: %v", err)
	}
}

// TestStandardModuleFSStat tests the Stat method
func TestStandardModuleFSStat(t *testing.T) {
	fs := NewStandardModuleFS()

	// Create a temporary test file and directory
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test-stat.txt")
	testContent := []byte("test stat content")
	testNestedDir := filepath.Join(tmpDir, "test-stat-dir")

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = os.MkdirAll(testNestedDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Test stat on file
	fileInfo, err := fs.Stat(testFile)
	if err != nil {
		t.Errorf("Expected no error for file stat, got: %v", err)
	}
	if fileInfo.IsDir() {
		t.Errorf("Expected file to not be a directory")
	}
	if fileInfo.Size() != int64(len(testContent)) {
		t.Errorf("Expected file size %d, got %d", len(testContent), fileInfo.Size())
	}

	// Test stat on directory
	dirInfo, err := fs.Stat(testNestedDir)
	if err != nil {
		t.Errorf("Expected no error for directory stat, got: %v", err)
	}
	if !dirInfo.IsDir() {
		t.Errorf("Expected directory to be a directory")
	}

	// Test stat on non-existent file
	_, err = fs.Stat(filepath.Join(tmpDir, "nonexistent"))
	if !os.IsNotExist(err) {
		t.Errorf("Expected IsNotExist error, got: %v", err)
	}
}

// Helper function to safely remove a directory, ignoring errors
func safeRemoveAll(path string) {
	if err := os.RemoveAll(path); err != nil {
		// Ignore errors during cleanup in tests
		// This is especially important on Windows where files might still be locked
		_ = err
	}
}

// TestStandardModuleFSTempDir tests the TempDir method
func TestStandardModuleFSTempDir(t *testing.T) {
	fs := NewStandardModuleFS()

	// Create a temporary directory
	tmpDir, err := fs.TempDir("", "fs-test-")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify directory was created
	info, err := os.Stat(tmpDir)
	if err != nil {
		t.Errorf("Failed to stat created temp directory: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("Expected a directory, got a file")
	}

	// Clean up
	safeRemoveAll(tmpDir)

	// Test with custom base directory
	baseDir := t.TempDir()
	tmpDir, err = fs.TempDir(baseDir, "custom-")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify directory was created in the specified base
	if !strings.HasPrefix(tmpDir, baseDir) {
		t.Errorf("Expected temp dir to be under base dir '%s', got '%s'", baseDir, tmpDir)
	}

	// Clean up
	safeRemoveAll(tmpDir)
}
