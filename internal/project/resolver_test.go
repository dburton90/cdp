package project

import (
	"reflect"
	"testing"
)

func TestResolver_Resolve(t *testing.T) {
	tests := []struct {
		name     string
		projects []Project
		want     []Project
	}{
		{
			name: "no collisions",
			projects: []Project{
				{Name: "p1", Path: "/home/user/p1"},
				{Name: "p2", Path: "/home/user/p2"},
			},
			want: []Project{
				{Name: "p1", ResolvedName: "p1", Path: "/home/user/p1"},
				{Name: "p2", ResolvedName: "p2", Path: "/home/user/p2"},
			},
		},
		{
			name: "one level collision",
			projects: []Project{
				{Name: "search", Path: "/home/user/work/google/search"},
				{Name: "search", Path: "/home/user/personal/google/search"},
			},
			want: []Project{
				{Name: "search", ResolvedName: "work/google/search", Path: "/home/user/work/google/search"},
				{Name: "search", ResolvedName: "personal/google/search", Path: "/home/user/personal/google/search"},
			},
		},
		{
			name: "deep collision",
			projects: []Project{
				{Name: "app", Path: "/a/b/c/d/app"},
				{Name: "app", Path: "/x/y/c/d/app"},
			},
			want: []Project{
				{Name: "app", ResolvedName: "b/c/d/app", Path: "/a/b/c/d/app"},
				{Name: "app", ResolvedName: "y/c/d/app", Path: "/x/y/c/d/app"},
			},
		},
		{
			name: "identical paths (edge case)",
			projects: []Project{
				{Name: "path", Path: "/same/path"},
				{Name: "path", Path: "/same/path"},
			},
			want: []Project{
				{Name: "path", ResolvedName: "same/path", Path: "/same/path"},
				{Name: "path", ResolvedName: "same/path", Path: "/same/path"},
			},
		},
		{
			name: "mixed collisions",
			projects: []Project{
				{Name: "api", Path: "/work/api"},
				{Name: "api", Path: "/personal/api"},
				{Name: "web", Path: "/work/web"},
			},
			want: []Project{
				{Name: "api", ResolvedName: "work/api", Path: "/work/api"},
				{Name: "api", ResolvedName: "personal/api", Path: "/personal/api"},
				{Name: "web", ResolvedName: "web", Path: "/work/web"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewResolver()
			got := r.Resolve(tt.projects)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Resolver.Resolve() = %v, want %v", got, tt.want)
			}
		})
	}
}
