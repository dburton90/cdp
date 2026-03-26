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
	// Match returns projects containing the given query (case-insensitive).
	// Sorted by relevance: exact match first, then prefix matches, then substring matches.
	Match(query string, projects []project.Project) []project.Project

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

// Match returns all projects whose names contain the query (case-insensitive).
// Sorted by relevance: exact match first, then prefix matches, then substring matches.
func (m *PrefixMatcher) Match(query string, projects []project.Project) []project.Project {
	if query == "" {
		// Return all projects sorted alphabetically
		result := make([]project.Project, len(projects))
		copy(result, projects)
		sort.Slice(result, func(i, j int) bool {
			return strings.ToLower(result[i].Name) < strings.ToLower(result[j].Name)
		})
		return result
	}

	lowerQuery := strings.ToLower(query)
	var matches []project.Project

	for _, p := range projects {
		if strings.Contains(strings.ToLower(p.Name), lowerQuery) {
			matches = append(matches, p)
		}
	}

	// Sort: exact matches first, then prefix matches, then substring matches, then alphabetically
	sort.Slice(matches, func(i, j int) bool {
		iLower := strings.ToLower(matches[i].Name)
		jLower := strings.ToLower(matches[j].Name)

		iExact := iLower == lowerQuery
		jExact := jLower == lowerQuery
		if iExact != jExact {
			return iExact // Exact match comes first
		}

		iPrefix := strings.HasPrefix(iLower, lowerQuery)
		jPrefix := strings.HasPrefix(jLower, lowerQuery)
		if iPrefix != jPrefix {
			return iPrefix // Prefix match comes before substring match
		}

		return iLower < jLower
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

