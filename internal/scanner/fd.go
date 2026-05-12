package scanner

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dbarton/cd_project/internal/project"
)

// FdScanner uses the fd command for faster scanning.
type FdScanner struct{}

// Scan finds all Git repositories using fd command.
func (s *FdScanner) Scan(roots []string) ([]project.Project, error) {
	var projects []project.Project

	for _, root := range roots {
		// Expand environment variables and ~
		root = expandPath(root)

		// fd command: find .git directories
		// --type d: directories only
		// --hidden: include hidden dirs
		// --no-ignore: don't respect .gitignore
		// --prune: don't descend into matches (key for performance!)
		cmd := exec.Command(fdCommand(),
			"--type", "d",
			"--hidden",
			"--no-ignore",
			"--prune",
			"^\\.git$",
			root,
		)

		output, err := cmd.Output()
		if err != nil {
			continue // Fall through to next root
		}

		for _, line := range strings.Split(string(output), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			// fd returns /path/to/project/.git/ (with trailing slash)
			// We need to remove the trailing slash before calling Dir()
			line = strings.TrimSuffix(line, "/")
			// Now get the parent directory (the project folder)
			projectPath := filepath.Dir(line)
			if isInsideGitWorktree(projectPath) {
				continue
			}
			projects = append(projects, project.Project{
				Name: filepath.Base(projectPath),
				Path: projectPath,
			})
		}
	}

	return projects, nil
}

// isInsideGitWorktree reports whether projectPath is nested inside a git worktree
// or submodule root. Those have a .git FILE (not directory) at their root, which
// means the scanner should not have descended into them.
func isInsideGitWorktree(projectPath string) bool {
	dir := filepath.Dir(projectPath)
	for {
		fi, err := os.Stat(filepath.Join(dir, ".git"))
		if err == nil && !fi.IsDir() {
			return true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return false
		}
		dir = parent
	}
}

