package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNativeScanner_Scan(t *testing.T) {
	// Create test directory structure
	tmpDir := t.TempDir()

	// Create mock projects
	projects := []string{"project-a", "project-b"}
	for _, name := range projects {
		projectDir := filepath.Join(tmpDir, name)
		gitDir := filepath.Join(projectDir, ".git")
		if err := os.MkdirAll(gitDir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Create non-project directory
	if err := os.MkdirAll(filepath.Join(tmpDir, "not-a-project"), 0755); err != nil {
		t.Fatal(err)
	}

	// Scan
	s := &NativeScanner{}
	found, err := s.Scan([]string{tmpDir})

	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(found) != 2 {
		t.Errorf("Scan() found %d projects, want 2", len(found))
	}
}

func TestNativeScanner_SkipsNestedGit(t *testing.T) {
	tmpDir := t.TempDir()

	// Create project with nested .git (shouldn't happen, but test anyway)
	projectDir := filepath.Join(tmpDir, "parent-project")
	if err := os.MkdirAll(filepath.Join(projectDir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(projectDir, "subdir", "nested", ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	s := &NativeScanner{}
	found, _ := s.Scan([]string{tmpDir})

	// Should only find parent, not nested
	if len(found) != 1 {
		t.Errorf("Scan() found %d projects, want 1 (parent only)", len(found))
	}
	if len(found) > 0 && found[0].Name != "parent-project" {
		t.Errorf("Found project %s, want parent-project", found[0].Name)
	}
}

func TestNativeScanner_SkipsNodeModules(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a project
	projectDir := filepath.Join(tmpDir, "my-project")
	if err := os.MkdirAll(filepath.Join(projectDir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create node_modules with a .git inside (shouldn't be found)
	nodeModules := filepath.Join(projectDir, "node_modules", "some-package")
	if err := os.MkdirAll(filepath.Join(nodeModules, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	s := &NativeScanner{}
	found, _ := s.Scan([]string{tmpDir})

	// Should only find the main project
	if len(found) != 1 {
		t.Errorf("Scan() found %d projects, want 1", len(found))
	}
}

func TestNativeScanner_SkipsWorktrees(t *testing.T) {
	tmpDir := t.TempDir()

	// Real git repo (has .git directory)
	realRepo := filepath.Join(tmpDir, "real-repo")
	if err := os.MkdirAll(filepath.Join(realRepo, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	// Git worktree root (.git is a file, not a directory)
	worktreeRoot := filepath.Join(tmpDir, "my-worktree")
	if err := os.MkdirAll(worktreeRoot, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(worktreeRoot, ".git"), []byte("gitdir: /repo/.git/worktrees/my-worktree"), 0644); err != nil {
		t.Fatal(err)
	}

	// Repo nested inside the worktree — should NOT be found
	nestedRepo := filepath.Join(worktreeRoot, "tools", "nested")
	if err := os.MkdirAll(filepath.Join(nestedRepo, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	s := &NativeScanner{}
	found, _ := s.Scan([]string{tmpDir})

	if len(found) != 1 {
		t.Errorf("Scan() found %d projects, want 1", len(found))
		for _, p := range found {
			t.Logf("  found: %s at %s", p.Name, p.Path)
		}
	}
	if len(found) > 0 && found[0].Name != "real-repo" {
		t.Errorf("Found project %s, want real-repo", found[0].Name)
	}
}

func TestNativeScanner_InvalidRoot(t *testing.T) {
	s := &NativeScanner{}
	found, err := s.Scan([]string{"/nonexistent/path/that/does/not/exist"})

	if err != nil {
		t.Errorf("Scan() should not error on invalid root, got %v", err)
	}
	if len(found) != 0 {
		t.Errorf("Scan() found %d projects, want 0", len(found))
	}
}

func TestFdAvailable(t *testing.T) {
	// Just ensure it doesn't panic
	_ = FdAvailable()
}

func TestScannerType(t *testing.T) {
	typ := ScannerType()
	if typ != "fd" && typ != "native" {
		t.Errorf("ScannerType() = %s, want 'fd' or 'native'", typ)
	}
}

func TestNewScanner(t *testing.T) {
	s := NewScanner()
	if s == nil {
		t.Error("NewScanner() returned nil")
	}
}

