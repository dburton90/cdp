package scanner

import (
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
			// fd returns /path/to/project/.git, we want /path/to/project
			projectPath := filepath.Dir(line)
			projects = append(projects, project.Project{
				Name: filepath.Base(projectPath),
				Path: projectPath,
			})
		}
	}

	return projects, nil
}

