// Package matcher provides project name matching functionality.
package matcher

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dbarton/cd_project/internal/project"
)

// Matcher provides project name matching functionality.
type Matcher interface {
	// Match returns projects matching the given prefix (case-insensitive).
	// Sorted by relevance: exact match first, then alphabetically.
	Match(prefix string, projects []project.Project) []project.Project

	// FindExact returns the single best match.
	// Returns error if no match or ambiguous (multiple matches).
	FindExact(name string, projects []project.Project) (project.Project, error)
}

// PrefixMatcher implements case-insensitive prefix matching.
type PrefixMatcher struct{}

// NewMatcher creates a new PrefixMatcher.
func NewMatcher() *PrefixMatcher {
	return &PrefixMatcher{}
}

// Match returns all projects whose names start with prefix (case-insensitive).
func (m *PrefixMatcher) Match(prefix string, projects []project.Project) []project.Project {
	if prefix == "" {
		// Return all projects sorted alphabetically
		result := make([]project.Project, len(projects))
		copy(result, projects)
		sort.Slice(result, func(i, j int) bool {
			return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
		})
		return result
	}

	lowerPrefix := strings.ToLower(prefix)
	var matches []project.Project

	for _, p := range projects {
		if strings.HasPrefix(strings.ToLower(p.Name), lowerPrefix) {
			matches = append(matches, p)
		}
	}

	// Sort: exact matches first, then alphabetically
	sort.Slice(matches, func(i, j int) bool {
		iExact := strings.EqualFold(matches[i].Name, prefix)
		jExact := strings.EqualFold(matches[j].Name, prefix)
		if iExact != jExact {
			return iExact // Exact match comes first
		}
		return strings.ToLower(matches[i].Name) < strings.ToLower(matches[j].Name)
	})

	return matches
}

// FindExact finds a single project by exact name match.
// Returns error if no match found or if multiple projects have the same name.
func (m *PrefixMatcher) FindExact(name string, projects []project.Project) (project.Project, error) {
	var matches []project.Project

	for _, p := range projects {
		if strings.EqualFold(p.Name, name) {
			matches = append(matches, p)
		}
	}

	switch len(matches) {
	case 0:
		return project.Project{}, fmt.Errorf("project not found: %s", name)
	case 1:
		return matches[0], nil
	default:
		return project.Project{}, &AmbiguousMatchError{
			Name:    name,
			Matches: matches,
		}
	}
}

// AmbiguousMatchError is returned when multiple projects match.
type AmbiguousMatchError struct {
	Name    string
	Matches []project.Project
}

func (e *AmbiguousMatchError) Error() string {
	var paths []string
	for _, m := range e.Matches {
		paths = append(paths, m.Path)
	}
	return fmt.Sprintf("multiple projects named '%s':\n  %s\nUse full path or rename directory.",
		e.Name, strings.Join(paths, "\n  "))
}

