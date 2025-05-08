package testing

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MockFileInfo implements os.FileInfo for testing
type MockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

// Name returns the base name of the file
func (fi *MockFileInfo) Name() string { return fi.name }

// Size returns the length in bytes
func (fi *MockFileInfo) Size() int64 { return fi.size }

// Mode returns the file mode bits
func (fi *MockFileInfo) Mode() os.FileMode { return fi.mode }

// ModTime returns the modification time
func (fi *MockFileInfo) ModTime() time.Time { return fi.modTime }

// IsDir returns whether the file is a directory
func (fi *MockFileInfo) IsDir() bool { return fi.isDir }

// Sys returns the underlying data source (always nil for mocks)
func (fi *MockFileInfo) Sys() interface{} { return nil }

// MockModuleFS implements toolkit.ModuleFS for testing
type MockModuleFS struct {
	// Mock file contents
	Files map[string][]byte

	// Mock directories
	Directories map[string]bool

	// Track operations
	Operations []string

	// Error to return for specific operations
	Errors map[string]error
}

// NewMockModuleFS creates a new mock filesystem
func NewMockModuleFS() *MockModuleFS {
	return &MockModuleFS{
		Files:       make(map[string][]byte),
		Directories: make(map[string]bool),
		Operations:  make([]string, 0),
		Errors:      make(map[string]error),
	}
}

// ReadFile reads a file from the filesystem
func (fs *MockModuleFS) ReadFile(path string) ([]byte, error) {
	fs.Operations = append(fs.Operations, "ReadFile:"+path)

	if err, ok := fs.Errors["ReadFile:"+path]; ok {
		return nil, err
	}

	data, ok := fs.Files[path]
	if !ok {
		return nil, os.ErrNotExist
	}

	return data, nil
}

// WriteFile writes data to a file
func (fs *MockModuleFS) WriteFile(path string, data []byte, perm os.FileMode) error {
	fs.Operations = append(fs.Operations, "WriteFile:"+path)

	if err, ok := fs.Errors["WriteFile:"+path]; ok {
		return err
	}

	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if dir != "." && dir != "/" {
		if !fs.directoryExists(dir) {
			return os.ErrNotExist
		}
	}

	fs.Files[path] = data
	return nil
}

// MkdirAll creates a directory with all necessary parents
func (fs *MockModuleFS) MkdirAll(path string, perm os.FileMode) error {
	fs.Operations = append(fs.Operations, "MkdirAll:"+path)

	if err, ok := fs.Errors["MkdirAll:"+path]; ok {
		return err
	}

	fs.Directories[path] = true

	// Also create parent directories
	parts := strings.Split(path, string(filepath.Separator))
	current := ""

	for _, part := range parts {
		if part == "" {
			continue
		}

		if current == "" {
			current = part
		} else {
			current = filepath.Join(current, part)
		}

		fs.Directories[current] = true
	}

	return nil
}

// RemoveAll removes a path and any children
func (fs *MockModuleFS) RemoveAll(path string) error {
	fs.Operations = append(fs.Operations, "RemoveAll:"+path)

	if err, ok := fs.Errors["RemoveAll:"+path]; ok {
		return err
	}

	// Remove the directory
	delete(fs.Directories, path)

	// Remove all files and subdirectories
	for filePath := range fs.Files {
		if strings.HasPrefix(filePath, path+string(filepath.Separator)) {
			delete(fs.Files, filePath)
		}
	}

	for dirPath := range fs.Directories {
		if strings.HasPrefix(dirPath, path+string(filepath.Separator)) {
			delete(fs.Directories, dirPath)
		}
	}

	return nil
}

// Stat returns file info
func (fs *MockModuleFS) Stat(path string) (os.FileInfo, error) {
	fs.Operations = append(fs.Operations, "Stat:"+path)

	if err, ok := fs.Errors["Stat:"+path]; ok {
		return nil, err
	}

	// Check if it's a directory
	if isDir := fs.Directories[path]; isDir {
		return &MockFileInfo{
			name:    filepath.Base(path),
			size:    0,
			mode:    os.ModeDir | 0755,
			modTime: time.Now(),
			isDir:   true,
		}, nil
	}

	// Check if it's a file
	data, ok := fs.Files[path]
	if !ok {
		return nil, os.ErrNotExist
	}

	return &MockFileInfo{
		name:    filepath.Base(path),
		size:    int64(len(data)),
		mode:    0644,
		modTime: time.Now(),
		isDir:   false,
	}, nil
}

// TempDir creates a temporary directory
func (fs *MockModuleFS) TempDir(dir, pattern string) (string, error) {
	fs.Operations = append(fs.Operations, "TempDir:"+dir+"/"+pattern)

	if err, ok := fs.Errors["TempDir"]; ok {
		return "", err
	}

	// Create a fake temporary path
	tempPath := filepath.Join(dir, pattern+"-mock-12345")
	fs.Directories[tempPath] = true

	return tempPath, nil
}

// directoryExists checks if a directory exists in the mock filesystem
func (fs *MockModuleFS) directoryExists(path string) bool {
	return fs.Directories[path]
}
