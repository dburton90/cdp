package cache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dbarton/cd_project/internal/project"
)

func TestFileCache_SaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "test_cache")

	c := NewFileCacheWithPath(cachePath)

	// Save projects
	projects := []project.Project{
		{Name: "project-a", Path: "/path/to/a"},
		{Name: "project-b", Path: "/path/to/b"},
	}

	if err := c.Save(projects); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load and verify
	loaded, err := c.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(loaded) != len(projects) {
		t.Errorf("Load() returned %d projects, want %d", len(loaded), len(projects))
	}

	for i, p := range loaded {
		if p.Name != projects[i].Name || p.Path != projects[i].Path {
			t.Errorf("Project[%d] = %v, want %v", i, p, projects[i])
		}
	}
}

func TestFileCache_LoadMissing(t *testing.T) {
	c := NewFileCacheWithPath("/nonexistent/path/that/does/not/exist")
	projects, err := c.Load()

	if err != nil {
		t.Errorf("Load() should not error for missing file, got %v", err)
	}
	if projects != nil && len(projects) != 0 {
		t.Errorf("Load() should return empty slice for missing file")
	}
}

func TestFileCache_MalformedLines(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "test_cache")

	// Write malformed content
	content := "good-project,/path/to/good\nbadline\n\nanother-good,/path/to/another"
	if err := os.WriteFile(cachePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	c := NewFileCacheWithPath(cachePath)
	projects, err := c.Load()

	if err != nil {
		t.Fatalf("Load() should handle malformed lines gracefully, got %v", err)
	}
	if len(projects) != 2 {
		t.Errorf("Load() returned %d projects, want 2 (skipping malformed)", len(projects))
	}
}

func TestFileCache_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "test_cache")

	// Write empty file
	if err := os.WriteFile(cachePath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	c := NewFileCacheWithPath(cachePath)
	projects, err := c.Load()

	if err != nil {
		t.Fatalf("Load() should handle empty file, got %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("Load() returned %d projects, want 0", len(projects))
	}
}

func TestFileCache_Path(t *testing.T) {
	c := NewFileCacheWithPath("/custom/path")
	if c.Path() != "/custom/path" {
		t.Errorf("Path() = %v, want /custom/path", c.Path())
	}
}

func TestNewFileCache(t *testing.T) {
	c := NewFileCache()
	home, _ := os.UserHomeDir()
	expected := filepath.Join(home, ".cd_project_folders")
	if c.Path() != expected {
		t.Errorf("NewFileCache().Path() = %v, want %v", c.Path(), expected)
	}
}

