package project

import "testing"

func TestProject_String(t *testing.T) {
	p := Project{Name: "my-project", ResolvedName: "work/my-project", Path: "/home/user/code/my-project"}
	want := "my-project,work/my-project,/home/user/code/my-project"
	if got := p.String(); got != want {
		t.Errorf("String() = %v, want %v", got, want)
	}
}

func TestProject_StringWithComma(t *testing.T) {
	// Note: paths with commas would break CSV parsing
	// This test documents the current behavior
	p := Project{Name: "my-project", ResolvedName: "my-project", Path: "/path/with,comma"}
	got := p.String()
	if got != "my-project,my-project,/path/with,comma" {
		t.Errorf("String() = %v", got)
	}
}

func TestProjects_Names(t *testing.T) {
	ps := Projects{
		{Name: "a", Path: "/a"},
		{Name: "b", Path: "/b"},
	}
	names := ps.Names()
	if len(names) != 2 || names[0] != "a" || names[1] != "b" {
		t.Errorf("Names() = %v, want [a b]", names)
	}
}

func TestProjects_NamesEmpty(t *testing.T) {
	ps := Projects{}
	names := ps.Names()
	if len(names) != 0 {
		t.Errorf("Names() = %v, want empty slice", names)
	}
}

