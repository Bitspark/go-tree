package materialize

import "os"

// safeRemoveAll is a helper function for tests to safely remove a directory,
// ignoring any errors that might occur during cleanup.
// This is especially important on Windows where files might be locked.
func safeRemoveAll(path string) {
	if err := os.RemoveAll(path); err != nil {
		// Ignore errors during cleanup in tests
		_ = err
	}
}
