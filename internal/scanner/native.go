package scanner

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/dbarton/cd_project/internal/project"
)

// NativeScanner uses Go's filepath.WalkDir for scanning.
type NativeScanner struct{}

// skipDirs are directories that never contain Git repos.
var skipDirs = map[string]bool{
	"node_modules": true,
	"vendor":       true,
	".cache":       true,
	"__pycache__":  true,
	".npm":         true,
	".cargo":       true,
	".venv":        true,
	"venv":         true,
	".tox":         true,
	"dist":         true,
	"build":        true,
	".gradle":      true,
	".m2":          true,
}

// Scan finds all Git repositories under the given root directories.
func (s *NativeScanner) Scan(roots []string) ([]project.Project, error) {
	var projects []project.Project

	for _, root := range roots {
		// Expand ~ and environment variables
		root = expandPath(root)

		// Verify root exists
		if _, err := os.Stat(root); err != nil {
			continue // Skip invalid roots
		}

		filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return filepath.SkipDir // Skip inaccessible directories
			}

			if !d.IsDir() {
				return nil
			}

			// Skip known non-project directories
			if skipDirs[d.Name()] {
				return filepath.SkipDir
			}

			// Skip hidden directories (except the root)
			if strings.HasPrefix(d.Name(), ".") && path != root {
				return filepath.SkipDir
			}

			// Check for .git directory
			gitPath := filepath.Join(path, ".git")
			if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
				projects = append(projects, project.Project{
					Name: filepath.Base(path),
					Path: path,
				})
				return filepath.SkipDir // Don't descend into Git repos
			}

			return nil
		})
	}

	return projects, nil
}

// expandPath expands ~ and environment variables in a path.
func expandPath(path string) string {
	// Expand environment variables
	path = os.ExpandEnv(path)

	// Expand ~
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = filepath.Join(home, path[1:])
		}
	}

	return path
}

