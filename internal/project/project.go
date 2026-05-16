// Package project defines the Project type representing a Git repository.
package project

import "fmt"

// Project represents a discovered Git repository.
type Project struct {
	Name         string // Base directory name (e.g., "my-project")
	ResolvedName string // Unique qualified name (e.g., "work/my-project")
	Path         string // Absolute path (e.g., "/home/user/code/my-project")
}

// String returns CSV representation for cache storage.
func (p Project) String() string {
	return fmt.Sprintf("%s,%s,%s", p.Name, p.ResolvedName, p.Path)
}

// Projects is a slice of Project with helper methods.
type Projects []Project

// Names returns just the project names.
func (ps Projects) Names() []string {
	names := make([]string, len(ps))
	for i, p := range ps {
		names[i] = p.Name
	}
	return names
}

