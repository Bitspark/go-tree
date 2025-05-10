package materialize

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/path/to/module", "/path/to/module"},
		{"C:\\path\\to\\module", "C:/path/to/module"},
		{"./relative/path", "relative/path"},
		{"../parent/path", "../parent/path"},
		{"path//with//double//slashes", "path/with/double/slashes"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := NormalizePath(tt.input)
			if got != tt.want {
				t.Errorf("NormalizePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestRelativizePath(t *testing.T) {
	base := filepath.Join("root", "base")

	tests := []struct {
		name       string
		base       string
		target     string
		want       string
		wantPrefix bool
	}{
		{"Child path", base, filepath.Join(base, "child"), "child", false},
		{"Sibling path", base, filepath.Join(filepath.Dir(base), "sibling"), "../sibling", false},
		{"Far path", base, filepath.Join("other", "far", "away"), "../../other/far/away", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RelativizePath(tt.base, tt.target)
			if tt.wantPrefix {
				if strings.HasPrefix(got, "..") {
					// This is good, we want a prefix
				} else {
					t.Errorf("RelativizePath(%q, %q) = %q, should start with '..'", tt.base, tt.target, got)
				}
			} else {
				expected := NormalizePath(tt.want)
				if got != expected {
					t.Errorf("RelativizePath(%q, %q) = %q, want %q", tt.base, tt.target, got, expected)
				}
			}
		})
	}
}

func TestIsLocalPath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/absolute/path", true},
		{"C:\\windows\\path", true},
		{"./relative/path", true},
		{"../parent/path", true},
		{"github.com/user/repo", false},
		{"golang.org/x/tools", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := IsLocalPath(tt.path)
			if got != tt.want {
				t.Errorf("IsLocalPath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestCreateUniqueModulePath(t *testing.T) {
	// Create a test environment
	env := &Environment{
		RootDir:     filepath.FromSlash("/test/root"),
		ModulePaths: make(map[string]string),
	}

	// Test different layout strategies
	tests := []struct {
		name           string
		modulePath     string
		layoutStrategy LayoutStrategy
		wantSuffix     string
	}{
		{"Flat layout", "github.com/user/repo", FlatLayout, "github.com_user_repo"},
		{"Hierarchical layout", "github.com/user/repo", HierarchicalLayout, filepath.FromSlash("github.com/user/repo")},
		{"GOPATH layout", "github.com/user/repo", GoPathLayout, filepath.FromSlash("src/github.com/user/repo")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateUniqueModulePath(env, tt.layoutStrategy, tt.modulePath)

			// For cross-platform testing, just verify the suffix matches
			if !strings.HasSuffix(got, tt.wantSuffix) {
				t.Errorf("CreateUniqueModulePath() = %v, want suffix %v", got, tt.wantSuffix)
			}
		})
	}

	// Test uniqueness with collision
	modPath := "github.com/user/repo"
	originalPath := CreateUniqueModulePath(env, FlatLayout, modPath)
	env.ModulePaths[modPath] = originalPath

	// Now get a new path which should be different
	newPath := CreateUniqueModulePath(env, FlatLayout, modPath)

	if newPath == originalPath {
		t.Errorf("CreateUniqueModulePath() didn't create a unique path: %v", newPath)
	}

	if !strings.Contains(newPath, originalPath+"_") {
		t.Errorf("CreateUniqueModulePath() = %v, expected to have original path plus suffix", newPath)
	}
}

func TestSanitizePathForFilename(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"github.com/user/repo", "github.com_user_repo"},
		{"C:\\windows\\path", "C_windows_path"},
		{"name:with:colons", "name_with_colons"},
		{"file?with*special\"chars", "file_with_special_chars"},
		{"multiple___underscores", "multiple_underscores"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := SanitizePathForFilename(tt.path)
			if got != tt.want {
				t.Errorf("SanitizePathForFilename(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
