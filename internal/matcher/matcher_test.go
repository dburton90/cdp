package matcher

import (
	"testing"

	"github.com/dbarton/cd_project/internal/project"
)

func TestPrefixMatcher_Match(t *testing.T) {
	m := NewMatcher()
	projects := []project.Project{
		{Name: "my-app", Path: "/my-app"},
		{Name: "my-api", Path: "/my-api"},
		{Name: "other", Path: "/other"},
		{Name: "MY-APP", Path: "/MY-APP-upper"},
	}

	tests := []struct {
		prefix    string
		wantCount int
	}{
		{"my", 3},         // my-app, my-api, MY-APP (case-insensitive)
		{"MY", 3},         // Same matches
		{"my-app", 2},     // my-app and MY-APP
		{"other", 1},      // other only
		{"nonexistent", 0}, // no matches
		{"", 4},           // All projects when empty
	}

	for _, tt := range tests {
		t.Run(tt.prefix, func(t *testing.T) {
			matches := m.Match(tt.prefix, projects)
			if len(matches) != tt.wantCount {
				t.Errorf("Match(%q) returned %d, want %d", tt.prefix, len(matches), tt.wantCount)
			}
		})
	}
}

func TestPrefixMatcher_MatchSorting(t *testing.T) {
	m := NewMatcher()
	projects := []project.Project{
		{Name: "zebra", Path: "/zebra"},
		{Name: "alpha", Path: "/alpha"},
		{Name: "beta", Path: "/beta"},
	}

	// Empty prefix should return alphabetically sorted
	matches := m.Match("", projects)
	if len(matches) != 3 {
		t.Fatalf("Match('') returned %d, want 3", len(matches))
	}
	if matches[0].Name != "alpha" || matches[1].Name != "beta" || matches[2].Name != "zebra" {
		t.Errorf("Match('') not sorted alphabetically: %v", matches)
	}
}

func TestPrefixMatcher_FindExact(t *testing.T) {
	m := NewMatcher()
	projects := []project.Project{
		{Name: "unique", Path: "/unique"},
		{Name: "duplicate", Path: "/path1"},
		{Name: "duplicate", Path: "/path2"},
	}

	// Test unique match
	p, err := m.FindExact("unique", projects)
	if err != nil {
		t.Errorf("FindExact(unique) error = %v", err)
	}
	if p.Name != "unique" {
		t.Errorf("FindExact(unique) = %v, want unique", p.Name)
	}

	// Test case-insensitive match
	p, err = m.FindExact("UNIQUE", projects)
	if err != nil {
		t.Errorf("FindExact(UNIQUE) error = %v", err)
	}
	if p.Name != "unique" {
		t.Errorf("FindExact(UNIQUE) = %v, want unique", p.Name)
	}

	// Test ambiguous match
	_, err = m.FindExact("duplicate", projects)
	if err == nil {
		t.Error("FindExact(duplicate) should return error for ambiguous")
	}
	if _, ok := err.(*AmbiguousMatchError); !ok {
		t.Errorf("FindExact(duplicate) should return AmbiguousMatchError, got %T", err)
	}

	// Test no match
	_, err = m.FindExact("nonexistent", projects)
	if err == nil {
		t.Error("FindExact(nonexistent) should return error")
	}
}

func TestAmbiguousMatchError_Error(t *testing.T) {
	err := &AmbiguousMatchError{
		Name: "test",
		Matches: []project.Project{
			{Name: "test", Path: "/path1"},
			{Name: "test", Path: "/path2"},
		},
	}

	msg := err.Error()
	if msg == "" {
		t.Error("AmbiguousMatchError.Error() should not be empty")
	}
	// Should contain the project name
	if !contains(msg, "test") {
		t.Errorf("Error message should contain project name: %s", msg)
	}
	// Should contain paths
	if !contains(msg, "/path1") || !contains(msg, "/path2") {
		t.Errorf("Error message should contain paths: %s", msg)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

