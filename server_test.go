package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestValidatePathValidPaths tests that valid paths within DEV_ROOT are accepted
func TestValidatePathValidPaths(t *testing.T) {
	// Create a temporary directory to use as DEV_ROOT
	tmpDir, err := os.MkdirTemp("", "devroot-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	// Create a subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Create a file
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Set DEV_ROOT environment variable
	originalDevRoot := os.Getenv("DEV_ROOT")
	t.Cleanup(func() { _ = os.Setenv("DEV_ROOT", originalDevRoot) })
	if err := os.Setenv("DEV_ROOT", tmpDir); err != nil {
		t.Fatalf("Failed to set DEV_ROOT: %v", err)
	}

	tests := []struct {
		name        string
		path        string
		wantErr     bool
		errContains string
	}{
		{
			name:    "root of DEV_ROOT",
			path:    tmpDir,
			wantErr: false,
		},
		{
			name:    "subdirectory",
			path:    subDir,
			wantErr: false,
		},
		{
			name:    "file in DEV_ROOT",
			path:    testFile,
			wantErr: false,
		},
		{
			name:    "relative path within DEV_ROOT",
			path:    filepath.Join(tmpDir, "subdir", "..", "test.txt"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validatePath(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validatePath(%q) expected error, got nil", tt.path)
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("validatePath(%q) error = %v, want error containing %q", tt.path, err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("validatePath(%q) unexpected error: %v", tt.path, err)
				}
				if result == "" {
					t.Errorf("validatePath(%q) returned empty result", tt.path)
				}
			}
		})
	}
}

// TestValidatePathTraversalAttacks tests that path traversal attacks are blocked
func TestValidatePathTraversalAttacks(t *testing.T) {
	// Create a temporary directory to use as DEV_ROOT
	tmpDir, err := os.MkdirTemp("", "devroot-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	// Create a subdirectory for testing
	subDir := filepath.Join(tmpDir, "allowed")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Set DEV_ROOT environment variable
	originalDevRoot := os.Getenv("DEV_ROOT")
	t.Cleanup(func() { _ = os.Setenv("DEV_ROOT", originalDevRoot) })
	if err := os.Setenv("DEV_ROOT", subDir); err != nil {
		t.Fatalf("Failed to set DEV_ROOT: %v", err)
	}

	// Create a file outside DEV_ROOT
	outsideFile := filepath.Join(tmpDir, "secret.txt")
	if err := os.WriteFile(outsideFile, []byte("secret"), 0644); err != nil {
		t.Fatalf("Failed to create outside file: %v", err)
	}

	tests := []struct {
		name        string
		path        string
		wantErr     bool
		errContains string
	}{
		{
			name:        "escape with double dot",
			path:        filepath.Join(subDir, "..", "secret.txt"),
			wantErr:     true,
			errContains: "outside allowed DEV_ROOT",
		},
		{
			name:        "escape with multiple double dots",
			path:        filepath.Join(subDir, "a", "b", "..", "..", "..", "secret.txt"),
			wantErr:     true,
			errContains: "outside allowed DEV_ROOT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validatePath(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validatePath(%q) expected error, got nil", tt.path)
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("validatePath(%q) error = %v, want error containing %q", tt.path, err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("validatePath(%q) unexpected error: %v", tt.path, err)
				}
			}
		})
	}
}

// TestValidatePathSymlinkAttacks tests that symlink escape attempts are blocked
func TestValidatePathSymlinkAttacks(t *testing.T) {
	// Create a temporary directory to use as DEV_ROOT
	tmpDir, err := os.MkdirTemp("", "devroot-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	// Create allowed directory
	allowedDir := filepath.Join(tmpDir, "allowed")
	if err := os.Mkdir(allowedDir, 0755); err != nil {
		t.Fatalf("Failed to create allowed dir: %v", err)
	}

	// Create a file outside the allowed directory
	outsideDir := filepath.Join(tmpDir, "outside")
	if err := os.Mkdir(outsideDir, 0755); err != nil {
		t.Fatalf("Failed to create outside dir: %v", err)
	}
	outsideFile := filepath.Join(outsideDir, "secret.txt")
	if err := os.WriteFile(outsideFile, []byte("secret"), 0644); err != nil {
		t.Fatalf("Failed to create outside file: %v", err)
	}

	// Create a symlink inside allowedDir pointing to outside
	symlinkPath := filepath.Join(allowedDir, "link-to-outside")
	if err := os.Symlink(outsideFile, symlinkPath); err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}

	// Set DEV_ROOT environment variable
	originalDevRoot := os.Getenv("DEV_ROOT")
	t.Cleanup(func() { _ = os.Setenv("DEV_ROOT", originalDevRoot) })
	if err := os.Setenv("DEV_ROOT", allowedDir); err != nil {
		t.Fatalf("Failed to set DEV_ROOT: %v", err)
	}

	// Test that accessing the symlink is blocked
	_, err = validatePath(symlinkPath)
	if err == nil {
		t.Fatalf("validatePath(%q) should reject symlink pointing outside DEV_ROOT", symlinkPath)
	}
	if !strings.Contains(err.Error(), "outside allowed DEV_ROOT") {
		t.Errorf("validatePath(%q) error = %v, want error containing 'outside allowed DEV_ROOT'", symlinkPath, err)
	}
}

// TestValidatePathNonExistent tests behavior with non-existent paths
func TestValidatePathNonExistent(t *testing.T) {
	// Create a temporary directory to use as DEV_ROOT
	tmpDir, err := os.MkdirTemp("", "devroot-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	// Set DEV_ROOT environment variable
	originalDevRoot := os.Getenv("DEV_ROOT")
	t.Cleanup(func() { _ = os.Setenv("DEV_ROOT", originalDevRoot) })
	if err := os.Setenv("DEV_ROOT", tmpDir); err != nil {
		t.Fatalf("Failed to set DEV_ROOT: %v", err)
	}

	// Test non-existent path within DEV_ROOT
	nonExistentPath := filepath.Join(tmpDir, "does-not-exist", "file.txt")
	_, err = validatePath(nonExistentPath)
	if err == nil {
		t.Errorf("validatePath(%q) should reject non-existent path", nonExistentPath)
	}
	if !strings.Contains(err.Error(), "evaluate symlinks") {
		t.Errorf("validatePath(%q) error = %v, want error containing 'evaluate symlinks'", nonExistentPath, err)
	}
}

// TestValidatePathBrokenSymlink tests behavior with broken symlinks
func TestValidatePathBrokenSymlink(t *testing.T) {
	// Create a temporary directory to use as DEV_ROOT
	tmpDir, err := os.MkdirTemp("", "devroot-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	// Create a broken symlink
	brokenSymlink := filepath.Join(tmpDir, "broken-link")
	if err := os.Symlink("/non/existent/target", brokenSymlink); err != nil {
		t.Fatalf("Failed to create broken symlink: %v", err)
	}

	// Set DEV_ROOT environment variable
	originalDevRoot := os.Getenv("DEV_ROOT")
	t.Cleanup(func() { _ = os.Setenv("DEV_ROOT", originalDevRoot) })
	if err := os.Setenv("DEV_ROOT", tmpDir); err != nil {
		t.Fatalf("Failed to set DEV_ROOT: %v", err)
	}

	// Test that broken symlink is rejected
	_, err = validatePath(brokenSymlink)
	if err == nil {
		t.Errorf("validatePath(%q) should reject broken symlink", brokenSymlink)
	}
	if !strings.Contains(err.Error(), "evaluate symlinks") {
		t.Errorf("validatePath(%q) error = %v, want error containing 'evaluate symlinks'", brokenSymlink, err)
	}
}

// TestValidatePathDotDotPrefixedFiles tests that files starting with ".." inside DEV_ROOT are accepted
// This tests the fix for the bug where relPath[0:2] == ".." incorrectly rejected valid paths
func TestValidatePathDotDotPrefixedFiles(t *testing.T) {
	// Create a temporary directory to use as DEV_ROOT
	tmpDir, err := os.MkdirTemp("", "devroot-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) })

	// Create a file whose name starts with ".." inside DEV_ROOT
	// This is a valid filename on Unix systems
	dotdotFile := filepath.Join(tmpDir, "..hidden")
	if err := os.WriteFile(dotdotFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create ..hidden file: %v", err)
	}

	// Create a directory whose name starts with ".." inside DEV_ROOT
	dotdotDir := filepath.Join(tmpDir, "..special")
	if err := os.Mkdir(dotdotDir, 0755); err != nil {
		t.Fatalf("Failed to create ..special dir: %v", err)
	}

	// Create a file inside the "..special" directory
	dotdotNestedFile := filepath.Join(dotdotDir, "nested.txt")
	if err := os.WriteFile(dotdotNestedFile, []byte("nested"), 0644); err != nil {
		t.Fatalf("Failed to create nested file: %v", err)
	}

	// Set DEV_ROOT environment variable
	originalDevRoot := os.Getenv("DEV_ROOT")
	t.Cleanup(func() { _ = os.Setenv("DEV_ROOT", originalDevRoot) })
	if err := os.Setenv("DEV_ROOT", tmpDir); err != nil {
		t.Fatalf("Failed to set DEV_ROOT: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "file starting with .. at DEV_ROOT",
			path:    dotdotFile,
			wantErr: false,
		},
		{
			name:    "directory starting with .. at DEV_ROOT",
			path:    dotdotDir,
			wantErr: false,
		},
		{
			name:    "file inside directory starting with ..",
			path:    dotdotNestedFile,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validatePath(tt.path)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validatePath(%q) expected error, got nil", tt.path)
				}
			} else {
				if err != nil {
					t.Errorf("validatePath(%q) unexpected error: %v", tt.path, err)
				}
				if result == "" {
					t.Errorf("validatePath(%q) returned empty result", tt.path)
				}
			}
		})
	}
}

// TestValidatePathEmptyDevRoot tests behavior when DEV_ROOT is not set
func TestValidatePathEmptyDevRoot(t *testing.T) {
	// Save and clear DEV_ROOT
	originalDevRoot := os.Getenv("DEV_ROOT")
	t.Cleanup(func() { _ = os.Setenv("DEV_ROOT", originalDevRoot) })
	if err := os.Unsetenv("DEV_ROOT"); err != nil {
		t.Fatalf("Failed to unset DEV_ROOT: %v", err)
	}

	// Get current directory as the expected DEV_ROOT
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Test that a path within current directory is allowed
	result, err := validatePath(cwd)
	if err != nil {
		t.Errorf("validatePath(%q) unexpected error when DEV_ROOT not set: %v", cwd, err)
	}
	if result == "" {
		t.Errorf("validatePath(%q) returned empty result", cwd)
	}
}
