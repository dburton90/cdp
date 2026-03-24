package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dbarton/cd_project/internal/cache"
	"github.com/dbarton/cd_project/internal/matcher"
	"github.com/dbarton/cd_project/internal/scanner"
)

var Version = "dev"

func main() {
	// Custom flag handling to support --completion "" (empty string)
	var completionFlag string
	var completionSet bool

	flag.Func("completion", "Return matching project names for prefix", func(s string) error {
		completionFlag = s
		completionSet = true
		return nil
	})

	path := flag.String("path", "", "Return absolute path for project name")
	refresh := flag.Bool("refresh", false, "Rescan directories and rebuild cache")
	version := flag.Bool("version", false, "Print version")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "cd_project - Fast project directory navigator\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  cd_project --completion <prefix>  List matching projects\n")
		fmt.Fprintf(os.Stderr, "  cd_project --path <name>          Get project path\n")
		fmt.Fprintf(os.Stderr, "  cd_project --refresh              Rebuild project cache\n")
		fmt.Fprintf(os.Stderr, "\nEnvironment:\n")
		fmt.Fprintf(os.Stderr, "  CD_PROJECT_ROOT  Colon-separated list of root directories to scan\n")
	}

	flag.Parse()

	switch {
	case *version:
		fmt.Printf("cd_project %s\n", Version)
	case completionSet:
		handleCompletion(completionFlag)
	case *path != "":
		handlePath(*path)
	case *refresh:
		handleRefresh()
	default:
		flag.Usage()
	}
}

// getRoots returns the project root directories from environment.
func getRoots() ([]string, error) {
	rootEnv := os.Getenv("CD_PROJECT_ROOT")
	if rootEnv == "" {
		return nil, fmt.Errorf(
			"CD_PROJECT_ROOT not set.\n" +
				"Add to your shell RC:\n" +
				"  export CD_PROJECT_ROOT=\"$HOME/code\"\n" +
				"Use colon-separated paths for multiple roots:\n" +
				"  export CD_PROJECT_ROOT=\"$HOME/code:$HOME/work\"")
	}

	roots := strings.Split(rootEnv, ":")
	var validRoots []string

	for _, root := range roots {
		root = strings.TrimSpace(root)
		root = os.ExpandEnv(root)
		if info, err := os.Stat(root); err == nil && info.IsDir() {
			validRoots = append(validRoots, root)
		}
	}

	if len(validRoots) == 0 {
		return nil, fmt.Errorf("no valid directories in CD_PROJECT_ROOT: %s", rootEnv)
	}

	return validRoots, nil
}

func handleCompletion(prefix string) {
	c := cache.NewFileCache()
	projects, err := c.Load()
	if err != nil {
		os.Exit(1)
	}

	if len(projects) == 0 {
		// Silent exit for completion - no cache yet
		os.Exit(0)
	}

	m := matcher.NewMatcher()
	matches := m.Match(prefix, projects)

	// Output unique names only (for completion)
	seen := make(map[string]bool)
	for _, p := range matches {
		if !seen[p.Name] {
			fmt.Println(p.Name)
			seen[p.Name] = true
		}
	}
}

func handlePath(name string) {
	c := cache.NewFileCache()
	projects, err := c.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading cache: %v\n", err)
		os.Exit(1)
	}

	if len(projects) == 0 {
		fmt.Fprintf(os.Stderr, "No projects cached. Run: cd_project --refresh\n")
		os.Exit(1)
	}

	m := matcher.NewMatcher()
	project, err := m.FindExact(name, projects)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	fmt.Println(project.Path)
}

func handleRefresh() {
	roots, err := getRoots()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Scanning %d root(s) using %s scanner...\n",
		len(roots), scanner.ScannerType())

	s := scanner.NewScanner()
	projects, err := s.Scan(roots)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning: %v\n", err)
		os.Exit(1)
	}

	c := cache.NewFileCache()
	if err := c.Save(projects); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving cache: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Found %d projects\n", len(projects))
	fmt.Fprintf(os.Stderr, "Cache saved to: %s\n", c.Path())
}

