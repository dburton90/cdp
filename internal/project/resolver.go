package project

import (
	"path/filepath"
	"strings"
)

// Resolver handles uniquely naming projects by prepending parent directories.
type Resolver struct{}

// NewResolver returns a new Resolver instance.
func NewResolver() *Resolver {
	return &Resolver{}
}

// Resolve takes a slice of projects and populates their ResolvedName fields uniquely.
func (r *Resolver) Resolve(projects []Project) []Project {
	if len(projects) == 0 {
		return projects
	}

	// Group projects by their current ResolvedName (initially the base Name)
	groups := make(map[string][]int)
	for i, p := range projects {
		// Initialize ResolvedName with the base Name if not already set
		if projects[i].ResolvedName == "" {
			projects[i].ResolvedName = p.Name
		}
		name := projects[i].ResolvedName
		groups[name] = append(groups[name], i)
	}

	for {
		hasCollisions := false
		newGroups := make(map[string][]int)
		canExpandAny := false

		for name, indices := range groups {
			if len(indices) == 1 {
				// No collision for this name
				newGroups[name] = indices
				continue
			}

			// Collision detected
			hasCollisions = true
			for _, idx := range indices {
				p := &projects[idx]
				
				// Try to prepend the parent directory
				parts := strings.Split(filepath.ToSlash(p.Path), "/")
				relParts := strings.Split(p.ResolvedName, "/")
				
				if len(parts) > len(relParts) {
					// Prepend the next parent
					nextParent := parts[len(parts)-len(relParts)-1]
					if nextParent != "" {
						p.ResolvedName = nextParent + "/" + p.ResolvedName
						canExpandAny = true
					}
				}
				
				newGroups[p.ResolvedName] = append(newGroups[p.ResolvedName], idx)
			}
		}

		if !hasCollisions || !canExpandAny {
			break
		}
		groups = newGroups
	}

	return projects
}
