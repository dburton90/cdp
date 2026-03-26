# Implementation Plan: cd_project

## Overview

This document provides a step-by-step implementation guide for the `cd_project` Go CLI tool. Each phase builds on the previous, with clear deliverables, code snippets, and complexity estimates.

**Total Estimated Time**: 8-12 hours for experienced Go developer

---

## Phase 1: Project Setup

**Complexity**: Low  
**Dependencies**: None  
**Estimated Time**: 30 minutes

### Files to Create

```
cd_project/
├── cmd/
│   └── cd_project/
│       └── main.go
├── internal/
│   ├── cache/
│   ├── matcher/
│   ├── project/
│   └── scanner/
├── scripts/
│   └── shell/
├── go.mod
├── Makefile
└── README.md
```

### Steps

1. **Initialize Go module**
   ```bash
   mkdir -p cd_project && cd cd_project
   go mod init github.com/user/cd_project
   ```

2. **Create directory structure**
   ```bash
   mkdir -p cmd/cd_project
   mkdir -p internal/{cache,matcher,project,scanner}
   mkdir -p scripts/shell
   ```

3. **Create basic main.go with flag parsing**

   ```go
   // cmd/cd_project/main.go
   package main

   import (
       "flag"
       "fmt"
       "os"
   )

   var Version = "dev"

   func main() {
       completion := flag.String("completion", "", "Return matching project names for query")
       path := flag.String("path", "", "Return absolute path for project name")
       refresh := flag.Bool("refresh", false, "Rescan directories and rebuild cache")
       version := flag.Bool("version", false, "Print version")

       flag.Usage = func() {
           fmt.Fprintf(os.Stderr, "cd_project - Fast project directory navigator\n\n")
           fmt.Fprintf(os.Stderr, "Usage:\n")
           fmt.Fprintf(os.Stderr, "  cd_project --completion <query>   List matching projects\n")
           fmt.Fprintf(os.Stderr, "  cd_project --path <name>          Get project path\n")
           fmt.Fprintf(os.Stderr, "  cd_project --refresh              Rebuild project cache\n")
           fmt.Fprintf(os.Stderr, "\nEnvironment:\n")
           fmt.Fprintf(os.Stderr, "  CD_PROJECT_ROOT  Colon-separated list of root directories\n")
       }
       
       flag.Parse()

       switch {
       case *version:
           fmt.Printf("cd_project %s\n", Version)
       case *completion != "":
           handleCompletion(*completion)
       case *path != "":
           handlePath(*path)
       case *refresh:
           handleRefresh()
       default:
           flag.Usage()
       }
   }

   func handleCompletion(query string) {
       // TODO: Implement in Phase 6
       fmt.Println("completion not implemented")
   }

   func handlePath(name string) {
       // TODO: Implement in Phase 6
       fmt.Println("path not implemented")
   }

   func handleRefresh() {
       // TODO: Implement in Phase 6
       fmt.Println("refresh not implemented")
   }
   ```

4. **Create Makefile**

   ```makefile
   BINARY_NAME=cd_project
   VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

   .PHONY: build install clean test

   build:
   	go build -ldflags "-X main.Version=$(VERSION)" -o $(BINARY_NAME) ./cmd/cd_project

   install: build
   	./scripts/install.sh

   clean:
   	rm -f $(BINARY_NAME)
   	rm -f ~/.cd_project_folders
   	rm -f ~/.cd_project.bash ~/.cd_project.zsh

   test:
   	go test -v ./...

   lint:
   	golangci-lint run
   ```

### Verification

```bash
go build ./cmd/cd_project
./cd_project --help
./cd_project --version
```

---

## Phase 2: Core Data Structures

**Complexity**: Low  
**Dependencies**: Phase 1  
**Estimated Time**: 20 minutes

### Files to Create

- `internal/project/project.go`

### Implementation

```go
// internal/project/project.go
package project

import "fmt"

// Project represents a discovered Git repository
type Project struct {
    Name string // Directory name (e.g., "my-project")
    Path string // Absolute path (e.g., "/home/user/code/my-project")
}

// String returns CSV representation for cache storage
func (p Project) String() string {
    return fmt.Sprintf("%s,%s", p.Name, p.Path)
}

// Projects is a slice of Project with helper methods
type Projects []Project

// Names returns just the project names
func (ps Projects) Names() []string {
    names := make([]string, len(ps))
    for i, p := range ps {
        names[i] = p.Name
    }
    return names
}
```

### Verification

```bash
go build ./...
```

---

## Phase 3: Cache Manager

**Complexity**: Medium
**Dependencies**: Phase 2
**Estimated Time**: 45 minutes

### Files to Create

- `internal/cache/cache.go`

### Interface Definition

```go
// internal/cache/cache.go
package cache

import "github.com/user/cd_project/internal/project"

// Cache defines the interface for project cache operations
type Cache interface {
    // Load reads projects from cache file
    Load() ([]project.Project, error)

    // Save writes projects to cache file
    Save(projects []project.Project) error

    // Path returns the cache file path
    Path() string
}
```

### FileCache Implementation

```go
// FileCache implements Cache using ~/.cd_project_folders
type FileCache struct {
    path string
}

// NewFileCache creates a cache using the default path
func NewFileCache() *FileCache {
    home, _ := os.UserHomeDir()
    return &FileCache{path: filepath.Join(home, ".cd_project_folders")}
}

// NewFileCacheWithPath creates a cache with a custom path (for testing)
func NewFileCacheWithPath(path string) *FileCache {
    return &FileCache{path: path}
}

// Path returns the cache file location
func (c *FileCache) Path() string {
    return c.path
}

// Load reads projects from the cache file
// Returns empty slice if file doesn't exist (not an error)
// Skips malformed lines gracefully
func (c *FileCache) Load() ([]project.Project, error) {
    data, err := os.ReadFile(c.path)
    if os.IsNotExist(err) {
        return nil, nil // Empty cache is valid
    }
    if err != nil {
        return nil, fmt.Errorf("reading cache: %w", err)
    }

    var projects []project.Project
    for _, line := range strings.Split(string(data), "\n") {
        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }
        parts := strings.SplitN(line, ",", 2)
        if len(parts) != 2 {
            continue // Skip malformed lines
        }
        projects = append(projects, project.Project{
            Name: parts[0],
            Path: parts[1],
        })
    }

    return projects, nil
}

// Save writes projects to the cache file
// Creates parent directories if needed
func (c *FileCache) Save(projects []project.Project) error {
    // Ensure directory exists
    dir := filepath.Dir(c.path)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return fmt.Errorf("creating cache directory: %w", err)
    }

    // Build content
    var lines []string
    for _, p := range projects {
        lines = append(lines, p.String())
    }
    content := strings.Join(lines, "\n")
    if len(lines) > 0 {
        content += "\n"
    }

    // Write atomically using temp file
    tmpPath := c.path + ".tmp"
    if err := os.WriteFile(tmpPath, []byte(content), 0644); err != nil {
        return fmt.Errorf("writing cache: %w", err)
    }
    if err := os.Rename(tmpPath, c.path); err != nil {
        os.Remove(tmpPath)
        return fmt.Errorf("renaming cache: %w", err)
    }

    return nil
}
```

### Error Handling

- File not found → Return empty slice (trigger refresh suggestion)
- Permission denied → Return error with helpful message
- Malformed lines → Skip silently, log in verbose mode
- Atomic writes → Use temp file + rename to prevent corruption

### Verification

```bash
go test ./internal/cache/...
```

---

## Phase 4: Project Scanner

**Complexity**: High
**Dependencies**: Phase 2
**Estimated Time**: 1.5 hours

### Files to Create

- `internal/scanner/scanner.go` - Interface and factory
- `internal/scanner/native.go` - Go stdlib implementation
- `internal/scanner/fd.go` - fd CLI integration

### Scanner Interface

```go
// internal/scanner/scanner.go
package scanner

import "github.com/user/cd_project/internal/project"

// Scanner discovers Git repositories in directory trees
type Scanner interface {
    // Scan finds all Git repos under the given root directories
    // Skips subdirectories once .git is found
    Scan(roots []string) ([]project.Project, error)
}

// NewScanner returns the best available scanner
// Prefers fd if available, falls back to native
func NewScanner() Scanner {
    if fdAvailable() {
        return &FdScanner{}
    }
    return &NativeScanner{}
}

// ScannerType returns "fd" or "native" based on what will be used
func ScannerType() string {
    if fdAvailable() {
        return "fd"
    }
    return "native"
}
```

### NativeScanner Implementation

```go
// internal/scanner/native.go
package scanner

import (
    "io/fs"
    "os"
    "path/filepath"

    "github.com/user/cd_project/internal/project"
)

// NativeScanner uses Go's filepath.WalkDir
type NativeScanner struct{}

// skipDirs are directories that never contain Git repos
var skipDirs = map[string]bool{
    "node_modules": true,
    "vendor":       true,
    ".cache":       true,
    "__pycache__":  true,
    ".npm":         true,
    ".cargo":       true,
}

func (s *NativeScanner) Scan(roots []string) ([]project.Project, error) {
    var projects []project.Project

    for _, root := range roots {
        // Expand ~ and environment variables
        root = os.ExpandEnv(root)
        if strings.HasPrefix(root, "~") {
            home, _ := os.UserHomeDir()
            root = filepath.Join(home, root[1:])
        }

        filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
            if err != nil {
                return filepath.SkipDir // Skip inaccessible directories
            }

            if !d.IsDir() {
                return nil
            }

            // Skip known non-project directories
            if skipDirs[d.Name()] {
                return filepath.SkipDir
            }

            // Check for .git directory
            gitPath := filepath.Join(path, ".git")
            if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
                projects = append(projects, project.Project{
                    Name: filepath.Base(path),
                    Path: path,
                })
                return filepath.SkipDir // Don't descend into Git repos
            }

            return nil
        })
    }

    return projects, nil
}
```

### FdScanner Implementation

```go
// internal/scanner/fd.go
package scanner

import (
    "os/exec"
    "path/filepath"
    "strings"

    "github.com/user/cd_project/internal/project"
)

// FdScanner uses the fd command for faster scanning
type FdScanner struct{}

// fdAvailable checks if fd is installed
func fdAvailable() bool {
    _, err := exec.LookPath("fd")
    return err == nil
}

func (s *FdScanner) Scan(roots []string) ([]project.Project, error) {
    var projects []project.Project

    for _, root := range roots {
        // Expand environment variables
        root = os.ExpandEnv(root)
        if strings.HasPrefix(root, "~") {
            home, _ := os.UserHomeDir()
            root = filepath.Join(home, root[1:])
        }

        // fd command: find .git directories
        // --type d: directories only
        // --hidden: include hidden dirs
        // --no-ignore: don't respect .gitignore
        // --prune: don't descend into matches
        cmd := exec.Command("fd",
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
```

### Key Design Decisions

1. **Smart traversal**: Skip `.git` subdirectories and common non-project dirs
2. **fd preference**: Auto-detect and use fd for ~10x performance
3. **Error resilience**: Log permission errors, don't fail entire scan
4. **Path expansion**: Handle `~` and `$HOME` in paths

### Verification

```bash
go test ./internal/scanner/...
# Manual test
export CD_PROJECT_ROOT="$HOME/code"
go run ./cmd/cd_project --refresh
```

---

## Phase 5: Matcher

**Complexity**: Medium
**Dependencies**: Phase 2
**Estimated Time**: 45 minutes

### Files to Create

- `internal/matcher/matcher.go`

### Interface Definition

```go
// internal/matcher/matcher.go
package matcher

import "github.com/user/cd_project/internal/project"

// Matcher provides project name matching functionality
type Matcher interface {
    // Match returns projects containing the given query (case-insensitive)
    // Sorted by relevance: exact match first, then prefix matches, then substring matches
    Match(query string, projects []project.Project) []project.Project

    // FindExact returns the single best match
    // Returns error if no match or ambiguous (multiple matches)
    FindExact(name string, projects []project.Project) (project.Project, error)
}
```

### SubstringMatcher Implementation

```go
// PrefixMatcher implements case-insensitive substring matching
type PrefixMatcher struct{}

// NewMatcher creates a new PrefixMatcher
func NewMatcher() *PrefixMatcher {
    return &PrefixMatcher{}
}

// Match returns all projects whose names contain query (case-insensitive)
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

    // Sort: exact matches first, then prefix matches, then substring matches
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

// FindExact finds a single project by exact name match
// Returns error if no match found or if multiple projects have the same name
func (m *PrefixMatcher) FindExact(name string, projects []project.Project) (project.Project, error) {
    lowerName := strings.ToLower(name)
    var matches []project.Project

    for _, p := range projects {
        if strings.EqualFold(p.Name, lowerName) {
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

// AmbiguousMatchError is returned when multiple projects match
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
```

### Verification

```bash
go test ./internal/matcher/...
```

---

## Phase 6: CLI Commands

**Complexity**: Medium
**Dependencies**: Phases 3, 4, 5
**Estimated Time**: 1 hour

### Files to Modify

- `cmd/cd_project/main.go`

### Helper Functions

```go
// cmd/cd_project/main.go

import (
    "github.com/user/cd_project/internal/cache"
    "github.com/user/cd_project/internal/matcher"
    "github.com/user/cd_project/internal/scanner"
)

// getRoots returns the project root directories from environment
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
```

### Command Implementations

```go
func handleCompletion(query string) {
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
    matches := m.Match(query, projects)

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
```

### Verification

```bash
export CD_PROJECT_ROOT="$HOME/code"
go build ./cmd/cd_project

# Test refresh
./cd_project --refresh

# Test completion
./cd_project --completion ""
./cd_project --completion "my"

# Test path
./cd_project --path "some-project"
```

---

## Phase 7: Shell Integration

**Complexity**: Medium
**Dependencies**: Phase 6
**Estimated Time**: 45 minutes

### Files to Create

- `scripts/shell/bash_completion.sh`
- `scripts/shell/zsh_completion.sh`

### Bash Integration

```bash
# scripts/shell/bash_completion.sh
# cd_project shell integration for Bash

# Main cdp function - wrapper around cd
cdp() {
    if [ -z "$1" ]; then
        echo "Usage: cdp <project-name>" >&2
        return 1
    fi

    local target
    target=$(cd_project --path "$1" 2>&1)
    local exit_code=$?

    if [ $exit_code -eq 0 ] && [ -n "$target" ]; then
        cd "$target" || return 1
    else
        echo "$target" >&2
        return 1
    fi
}

# Bash completion function
_cdp_completions() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local completions
    completions=$(cd_project --completion "$cur" 2>/dev/null)
    COMPREPLY=($(compgen -W "$completions" -- "$cur"))
}

# Register completion
complete -F _cdp_completions cdp

# Refresh helper
cdp-refresh() {
    cd_project --refresh
}
```

### Zsh Integration

```zsh
# scripts/shell/zsh_completion.sh
# cd_project shell integration for Zsh

# Main cdp function - wrapper around cd
cdp() {
    if [[ -z "$1" ]]; then
        echo "Usage: cdp <project-name>" >&2
        return 1
    fi

    local target
    target=$(cd_project --path "$1" 2>&1)
    local exit_code=$?

    if [[ $exit_code -eq 0 ]] && [[ -n "$target" ]]; then
        cd "$target" || return 1
    else
        echo "$target" >&2
        return 1
    fi
}

# Zsh completion function
_cdp() {
    local -a completions
    completions=(${(f)"$(cd_project --completion "${words[2]:-}" 2>/dev/null)"})
    _describe 'project' completions
}

# Register completion
compdef _cdp cdp

# Refresh helper
cdp-refresh() {
    cd_project --refresh
}
```

### Template Generation in Go (Optional)

Add to `cmd/cd_project/main.go`:

```go
//go:embed scripts/shell/bash_completion.sh
var bashTemplate string

//go:embed scripts/shell/zsh_completion.sh
var zshTemplate string

func handleShellInit(shell string) {
    switch shell {
    case "bash":
        fmt.Print(bashTemplate)
    case "zsh":
        fmt.Print(zshTemplate)
    default:
        fmt.Fprintf(os.Stderr, "Unknown shell: %s (supported: bash, zsh)\n", shell)
        os.Exit(1)
    }
}
```

### Verification

```bash
# Test bash completion
source scripts/shell/bash_completion.sh
cdp <TAB>

# Test zsh completion
source scripts/shell/zsh_completion.sh
cdp <TAB>
```

---

## Phase 8: Install Script

**Complexity**: Medium
**Dependencies**: Phase 7
**Estimated Time**: 1 hour

### Files to Create

- `scripts/install.sh`

### Install Script Implementation

```bash
#!/usr/bin/env bash
# cd_project installation script
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BINARY_NAME="cd_project"
INSTALL_DIR="${HOME}/.local/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Check if fd is available
check_fd() {
    if command -v fd &>/dev/null; then
        log_info "fd found - will use for faster scanning"
    else
        log_warn "fd not found - using native Go scanner"
        echo "  Install fd for ~10x faster scanning: https://github.com/sharkdp/fd"
    fi
}

# Generate shell configuration lines (printed for user to copy)
generate_shell_config() {
    local integration_file="$1"
    local project_roots="$2"
    echo ""
    echo "# cd_project integration"
    echo "export CD_PROJECT_ROOT=\"$project_roots\""
    echo "[ -f \"$integration_file\" ] && source \"$integration_file\""
}

main() {
    echo "========================================"
    echo "  cd_project Installation"
    echo "========================================"
    echo

    # Check for Go
    if ! command -v go &>/dev/null; then
        log_error "Go is required but not installed."
        exit 1
    fi

    check_fd
    echo

    # Prompt for project root
    read -p "Project root directories [$HOME/code]: " PROJECT_ROOTS
    PROJECT_ROOTS="${PROJECT_ROOTS:-$HOME/code}"

    # Prompt for shell selection
    echo
    echo "Select shell integration:"
    echo "  1) Bash only"
    echo "  2) Zsh only"
    echo "  3) Both"
    read -p "Choice [3]: " SHELL_CHOICE
    SHELL_CHOICE="${SHELL_CHOICE:-3}"

    # Build binary
    log_info "Building cd_project..."
    cd "$PROJECT_ROOT"
    go build -o "$BINARY_NAME" ./cmd/cd_project

    # Install binary
    mkdir -p "$INSTALL_DIR"
    cp "$BINARY_NAME" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    log_info "Installed binary to $INSTALL_DIR/$BINARY_NAME"

    # Install shell integration files (do not modify RC files)
    if [[ "$SHELL_CHOICE" == "1" || "$SHELL_CHOICE" == "3" ]]; then
        cp "$PROJECT_ROOT/scripts/shell/bash_completion.sh" "$HOME/.cd_project.bash"
        log_info "Installed $HOME/.cd_project.bash"
    fi

    if [[ "$SHELL_CHOICE" == "2" || "$SHELL_CHOICE" == "3" ]]; then
        cp "$PROJECT_ROOT/scripts/shell/zsh_completion.sh" "$HOME/.cd_project.zsh"
        log_info "Installed $HOME/.cd_project.zsh"
    fi

    # Run initial refresh
    log_info "Running initial project scan..."
    export CD_PROJECT_ROOT="$PROJECT_ROOTS"
    "$INSTALL_DIR/$BINARY_NAME" --refresh

    echo
    echo "========================================"
    echo "  Installation Complete!"
    echo "========================================"
    echo
    echo "Add the following to your shell configuration file:"
    echo

    if [[ "$SHELL_CHOICE" == "1" || "$SHELL_CHOICE" == "3" ]]; then
        echo "For ~/.bashrc:"
        echo "----------------------------------------"
        generate_shell_config "$HOME/.cd_project.bash" "$PROJECT_ROOTS"
        echo "----------------------------------------"
        echo
    fi

    if [[ "$SHELL_CHOICE" == "2" || "$SHELL_CHOICE" == "3" ]]; then
        echo "For ~/.zshrc:"
        echo "----------------------------------------"
        generate_shell_config "$HOME/.cd_project.zsh" "$PROJECT_ROOTS"
        echo "----------------------------------------"
        echo
    fi

    echo "Then restart your shell or run:"
    echo "  source ~/.bashrc   # for Bash"
    echo "  source ~/.zshrc    # for Zsh"
    echo
    echo "Usage:"
    echo "  cdp <project>      # cd to project"
    echo "  cdp <TAB>          # tab completion"
    echo "  cdp-refresh        # rescan projects"
}

main "$@"
```

**Note**: The install script never directly modifies `.bashrc` or `.zshrc` files. It only prints the configuration that users need to manually add.

### Verification

```bash
chmod +x scripts/install.sh
./scripts/install.sh
```

---

## Phase 9: Testing

**Complexity**: Medium
**Dependencies**: Phases 2-6
**Estimated Time**: 1.5 hours

### Files to Create

- `internal/project/project_test.go`
- `internal/cache/cache_test.go`
- `internal/scanner/scanner_test.go`
- `internal/matcher/matcher_test.go`
- `integration_test.go` (root level)

### Unit Test: Project

```go
// internal/project/project_test.go
package project

import "testing"

func TestProject_String(t *testing.T) {
    p := Project{Name: "my-project", Path: "/home/user/code/my-project"}
    want := "my-project,/home/user/code/my-project"
    if got := p.String(); got != want {
        t.Errorf("String() = %v, want %v", got, want)
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
```

### Unit Test: Cache

```go
// internal/cache/cache_test.go
package cache

import (
    "os"
    "path/filepath"
    "testing"

    "github.com/user/cd_project/internal/project"
)

func TestFileCache_SaveLoad(t *testing.T) {
    // Create temp file
    tmpDir := t.TempDir()
    cachePath := filepath.Join(tmpDir, "test_cache")

    c := NewFileCacheWithPath(cachePath)

    // Save projects
    projects := []project.Project{
        {Name: "project-a", Path: "/path/to/a"},
        {Name: "project-b", Path: "/path/to/b"},
    }

    if err := c.Save(projects); err != nil {
        t.Fatalf("Save() error = %v", err)
    }

    // Load and verify
    loaded, err := c.Load()
    if err != nil {
        t.Fatalf("Load() error = %v", err)
    }

    if len(loaded) != len(projects) {
        t.Errorf("Load() returned %d projects, want %d", len(loaded), len(projects))
    }

    for i, p := range loaded {
        if p.Name != projects[i].Name || p.Path != projects[i].Path {
            t.Errorf("Project[%d] = %v, want %v", i, p, projects[i])
        }
    }
}

func TestFileCache_LoadMissing(t *testing.T) {
    c := NewFileCacheWithPath("/nonexistent/path")
    projects, err := c.Load()

    if err != nil {
        t.Errorf("Load() should not error for missing file, got %v", err)
    }
    if projects != nil && len(projects) != 0 {
        t.Errorf("Load() should return empty slice for missing file")
    }
}

func TestFileCache_MalformedLines(t *testing.T) {
    tmpDir := t.TempDir()
    cachePath := filepath.Join(tmpDir, "test_cache")

    // Write malformed content
    content := "good-project,/path/to/good\nbadline\n\nanother-good,/path/to/another"
    os.WriteFile(cachePath, []byte(content), 0644)

    c := NewFileCacheWithPath(cachePath)
    projects, err := c.Load()

    if err != nil {
        t.Fatalf("Load() should handle malformed lines gracefully")
    }
    if len(projects) != 2 {
        t.Errorf("Load() returned %d projects, want 2 (skipping malformed)", len(projects))
    }
}
```

### Unit Test: Matcher

```go
// internal/matcher/matcher_test.go
package matcher

import (
    "testing"

    "github.com/user/cd_project/internal/project"
)

func TestPrefixMatcher_Match(t *testing.T) {
    m := NewMatcher()
    projects := []project.Project{
        {Name: "my-app", Path: "/my-app"},
        {Name: "my-api", Path: "/my-api"},
        {Name: "other", Path: "/other"},
        {Name: "MY-APP", Path: "/MY-APP-upper"},
        {Name: "preloliac", Path: "/preloliac"},
        {Name: "lol-project", Path: "/lol-project"},
    }

    tests := []struct {
        query     string
        wantCount int
    }{
        {"my", 3},          // my-app, my-api, MY-APP (case-insensitive)
        {"MY", 3},          // Same matches
        {"my-app", 2},      // my-app and MY-APP
        {"other", 1},
        {"nonexistent", 0},
        {"", 6},            // All projects
        {"lol", 2},         // preloliac and lol-project (substring matching)
        {"app", 2},         // my-app and MY-APP (substring match)
    }

    for _, tt := range tests {
        t.Run(tt.query, func(t *testing.T) {
            matches := m.Match(tt.query, projects)
            if len(matches) != tt.wantCount {
                t.Errorf("Match(%q) returned %d, want %d", tt.query, len(matches), tt.wantCount)
            }
        })
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
```

### Unit Test: Scanner

```go
// internal/scanner/scanner_test.go
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
        os.MkdirAll(gitDir, 0755)
    }

    // Create non-project directory
    os.MkdirAll(filepath.Join(tmpDir, "not-a-project"), 0755)

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
    os.MkdirAll(filepath.Join(projectDir, ".git"), 0755)
    os.MkdirAll(filepath.Join(projectDir, "subdir", "nested", ".git"), 0755)

    s := &NativeScanner{}
    found, _ := s.Scan([]string{tmpDir})

    // Should only find parent, not nested
    if len(found) != 1 {
        t.Errorf("Scan() found %d projects, want 1 (parent only)", len(found))
    }
}

func TestFdAvailable(t *testing.T) {
    // Just ensure it doesn't panic
    _ = fdAvailable()
}
```

### Integration Test

```go
// integration_test.go
package main

import (
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "testing"
)

func TestIntegration_FullWorkflow(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test in short mode")
    }

    // Build binary
    tmpDir := t.TempDir()
    binaryPath := filepath.Join(tmpDir, "cd_project")

    cmd := exec.Command("go", "build", "-o", binaryPath, "./cmd/cd_project")
    if err := cmd.Run(); err != nil {
        t.Fatalf("Failed to build binary: %v", err)
    }

    // Create test project structure
    projectDir := filepath.Join(tmpDir, "projects", "test-project")
    os.MkdirAll(filepath.Join(projectDir, ".git"), 0755)

    // Set environment
    os.Setenv("CD_PROJECT_ROOT", filepath.Join(tmpDir, "projects"))
    defer os.Unsetenv("CD_PROJECT_ROOT")

    // Test refresh
    cmd = exec.Command(binaryPath, "--refresh")
    output, err := cmd.CombinedOutput()
    if err != nil {
        t.Fatalf("--refresh failed: %v\n%s", err, output)
    }
    if !strings.Contains(string(output), "Found 1 project") {
        t.Errorf("--refresh output unexpected: %s", output)
    }

    // Test completion
    cmd = exec.Command(binaryPath, "--completion", "test")
    output, err = cmd.Output()
    if err != nil {
        t.Fatalf("--completion failed: %v", err)
    }
    if !strings.Contains(string(output), "test-project") {
        t.Errorf("--completion should include test-project: %s", output)
    }

    // Test path
    cmd = exec.Command(binaryPath, "--path", "test-project")
    output, err = cmd.Output()
    if err != nil {
        t.Fatalf("--path failed: %v", err)
    }
    if strings.TrimSpace(string(output)) != projectDir {
        t.Errorf("--path = %s, want %s", output, projectDir)
    }
}
```

### Running Tests

```bash
# Run all unit tests
go test ./...

# Run with coverage
go test -cover ./...

# Run integration tests
go test -v ./...

# Run benchmarks (if any)
go test -bench=. ./...
```

---

## Phase 10: Documentation

**Complexity**: Low
**Dependencies**: All previous phases
**Estimated Time**: 30 minutes

### Files to Create/Update

- `README.md`

### README Content

```markdown
# cd_project

Fast project directory navigator with shell integration and tab completion.

## Features

- 🚀 **Fast** - Uses `fd` for scanning when available, falls back to native Go
- 🔍 **Smart** - Case-insensitive substring matching with relevance sorting
- 🐚 **Shell Integration** - Native Bash and Zsh support with tab completion
- 📦 **Zero Dependencies** - Single Go binary, no runtime dependencies

## Installation

### From Source

```bash
git clone https://github.com/user/cd_project.git
cd cd_project
./scripts/install.sh
```

### Manual Installation

```bash
# Build
go build -o cd_project ./cmd/cd_project

# Copy to PATH
cp cd_project ~/.local/bin/

# Copy shell integration file
cp scripts/shell/bash_completion.sh ~/.cd_project.bash

# Add to your ~/.bashrc (manually):
# export CD_PROJECT_ROOT="$HOME/code"
# [ -f ~/.cd_project.bash ] && source ~/.cd_project.bash

# Initial scan
cd_project --refresh
```

## Usage

```bash
# Navigate to a project
cdp my-project

# Tab completion
cdp my<TAB>

# Refresh project cache
cdp-refresh
# or
cd_project --refresh
```

## Configuration

Set `CD_PROJECT_ROOT` to specify where to scan for projects:

```bash
# Single directory
export CD_PROJECT_ROOT="$HOME/code"

# Multiple directories (colon-separated)
export CD_PROJECT_ROOT="$HOME/code:$HOME/work:$HOME/personal"
```

## Performance

For large directory trees, install [fd](https://github.com/sharkdp/fd) for ~10x faster scanning:

```bash
# Ubuntu/Debian
sudo apt install fd-find

# macOS
brew install fd

# Arch
sudo pacman -S fd
```

## How It Works

1. `cd_project --refresh` scans `CD_PROJECT_ROOT` for directories containing `.git`
2. Results are cached in `~/.cd_project_folders`
3. `cdp <name>` looks up the project in cache and `cd`s to it
4. Tab completion uses the cache for instant suggestions

## License

MIT
```

---

## Implementation Summary

| Phase | Complexity | Est. Time | Key Deliverable |
|-------|------------|-----------|-----------------|
| 1. Project Setup | Low | 30 min | Go module, basic CLI |
| 2. Data Structures | Low | 20 min | Project struct |
| 3. Cache Manager | Medium | 45 min | CSV read/write |
| 4. Project Scanner | High | 1.5 hr | Native + fd scanners |
| 5. Matcher | Medium | 45 min | Substring matching |
| 6. CLI Commands | Medium | 1 hr | Full CLI implementation |
| 7. Shell Integration | Medium | 45 min | Bash/Zsh functions |
| 8. Install Script | Medium | 1 hr | Interactive installer |
| 9. Testing | Medium | 1.5 hr | Unit + integration tests |
| 10. Documentation | Low | 30 min | README |
| **Total** | | **~9 hours** | |

## Recommended Implementation Order

1. **Day 1 (4 hours)**: Phases 1-4 (Setup through Scanner)
2. **Day 2 (3 hours)**: Phases 5-6 (Matcher and CLI)
3. **Day 3 (2 hours)**: Phases 7-8 (Shell and Install)
4. **Day 4 (2 hours)**: Phases 9-10 (Testing and Docs)

## Success Criteria

- [ ] `cd_project --refresh` scans and caches projects
- [ ] `cd_project --completion <query>` returns matching names (substring)
- [ ] `cd_project --path <name>` returns absolute path
- [ ] `cdp <name>` changes directory (via shell function)
- [ ] Tab completion works in both Bash and Zsh
- [ ] All tests pass
- [ ] Install script works on fresh system

