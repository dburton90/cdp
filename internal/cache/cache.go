// Package cache provides caching functionality for discovered projects.
package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dbarton/cd_project/internal/project"
)

// Cache defines the interface for project cache operations.
type Cache interface {
	// Load reads projects from cache file.
	Load() ([]project.Project, error)

	// Save writes projects to cache file.
	Save(projects []project.Project) error

	// Path returns the cache file path.
	Path() string

	// Exists returns true if the cache file exists.
	Exists() bool
}

// FileCache implements Cache using ~/.cd_project_folders.
type FileCache struct {
	path string
}

// NewFileCache creates a cache using the default path.
func NewFileCache() *FileCache {
	home, _ := os.UserHomeDir()
	return &FileCache{path: filepath.Join(home, ".cd_project_folders")}
}

// NewFileCacheWithPath creates a cache with a custom path (for testing).
func NewFileCacheWithPath(path string) *FileCache {
	return &FileCache{path: path}
}

// Path returns the cache file location.
func (c *FileCache) Path() string {
	return c.path
}

// Exists returns true if the cache file exists.
func (c *FileCache) Exists() bool {
	_, err := os.Stat(c.path)
	return err == nil
}

// Load reads projects from the cache file.
// Returns empty slice if file doesn't exist (not an error).
// Skips malformed lines gracefully.
func (c *FileCache) Load() ([]project.Project, error) {
	data, err := os.ReadFile(c.path)
	if os.IsNotExist(err) {
		return nil, nil // Empty cache is valid
	}
	if err != nil {
		return nil, fmt.Errorf("reading cache: %w", err)
	}

	var projects []project.Project
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ",", 3)
		if len(parts) < 2 {
			continue // Skip malformed lines
		}

		p := project.Project{
			Name: parts[0],
		}
		if len(parts) == 3 {
			p.ResolvedName = parts[1]
			p.Path = parts[2]
		} else {
			// Backward compatibility: 2 parts (Name, Path)
			p.ResolvedName = parts[0]
			p.Path = parts[1]
		}
		projects = append(projects, p)
	}

	return projects, nil
}

// Save writes projects to the cache file.
// Creates parent directories if needed.
// Uses atomic write (temp file + rename) to prevent corruption.
func (c *FileCache) Save(projects []project.Project) error {
	// Ensure directory exists
	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating cache directory: %w", err)
	}

	// Build content
	var lines []string
	for _, p := range projects {
		lines = append(lines, p.String())
	}
	content := strings.Join(lines, "\n")
	if len(lines) > 0 {
		content += "\n"
	}

	// Write atomically using temp file
	tmpPath := c.path + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing cache: %w", err)
	}
	if err := os.Rename(tmpPath, c.path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("renaming cache: %w", err)
	}

	return nil
}

