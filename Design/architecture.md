# Architecture Decision Document: cd_project

## Project Overview

- **Project Name**: cd_project
- **Vision**: A fast Go-based CLI tool for navigating to Git repositories with shell integration and tab completion
- **Key Drivers**: Speed, simplicity, cross-shell compatibility (Bash/Zsh), minimal dependencies

## System Context

### Users
| User Type | Description | Scale |
|-----------|-------------|-------|
| Developer | Navigates between Git repositories frequently | Single user per install |

### External Systems
| System | Integration Type | Purpose |
|--------|------------------|---------|
| File System | Direct access | Scan for .git directories |
| Shell (Bash/Zsh) | Function integration | Tab completion, cd wrapper |
| `fd` (optional) | CLI subprocess | Faster directory scanning |

## Architecture Overview

### System Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         Shell Layer                              │
│  ┌──────────────┐  ┌──────────────┐  ┌─────────────────────┐    │
│  │ cdp function │  │  Completion  │  │  Shell RC Config    │    │
│  │  (cd wrapper)│  │   Function   │  │  (.bashrc/.zshrc)   │    │
│  └──────┬───────┘  └──────┬───────┘  └─────────────────────┘    │
└─────────┼─────────────────┼─────────────────────────────────────┘
          │                 │
          ▼                 ▼
┌─────────────────────────────────────────────────────────────────┐
│                     cd_project CLI (Go)                          │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    CLI Parser (cmd/)                      │   │
│  │   --completion <query> | --path <name> | --refresh        │   │
│  └──────────────────────────────────────────────────────────┘   │
│          │                    │                  │               │
│          ▼                    ▼                  ▼               │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐       │
│  │    Fuzzy     │    │    Cache     │    │   Project    │       │
│  │   Matcher    │    │   Manager    │    │   Scanner    │       │
│  └──────────────┘    └──────┬───────┘    └──────┬───────┘       │
│                             │                   │                │
└─────────────────────────────┼───────────────────┼────────────────┘
                              │                   │
                              ▼                   ▼
                    ┌──────────────┐    ┌──────────────┐
                    │ Cache File   │    │ File System  │
                    │ (~/.cd_      │    │ / fd CLI     │
                    │ project_     │    │              │
                    │ folders)     │    │              │
                    └──────────────┘    └──────────────┘
```

## Go Project Structure

```
cd_project/
├── cmd/
│   └── cd_project/
│       └── main.go           # Entry point, CLI argument parsing
├── internal/
│   ├── cache/
│   │   └── cache.go          # Cache read/write operations
│   ├── matcher/
│   │   └── matcher.go        # Fuzzy/prefix matching logic
│   ├── project/
│   │   └── project.go        # Project struct and utilities
│   └── scanner/
│       ├── scanner.go        # Directory scanning interface
│       ├── native.go         # Go-native filepath.Walk scanner
│       └── fd.go             # fd-based scanner (optional)
├── scripts/
│   ├── install.sh            # Interactive install script
│   └── shell/
│       ├── bash_completion.sh
│       └── zsh_completion.sh
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Components

### Component: CLI Parser (`cmd/cd_project/main.go`)
- **Type**: CLI Entry Point
- **Responsibility**: Parse arguments, route to appropriate handler
- **Interfaces**: Command-line flags
- **Design Notes**: Use standard library `flag` package (no external deps)

### Component: Cache Manager (`internal/cache/`)
- **Type**: Data Access Layer
- **Responsibility**: Read/write project cache, handle file locking
- **Interfaces**: `Load() []Project`, `Save([]Project) error`
- **File Format**: CSV with header `name,path`

### Component: Project Scanner (`internal/scanner/`)
- **Type**: Service
- **Responsibility**: Find all Git repositories under root directories
- **Interfaces**: `Scanner` interface with `Scan(roots []string) []Project`
- **Scaling Strategy**: Concurrent scanning with worker pool

### Component: Substring Matcher (`internal/matcher/`)
- **Type**: Utility
- **Responsibility**: Match user input to project names (substring matching)
- **Interfaces**: `Match(query string, projects []Project) []Project`

---

## Data Structures

### Project Representation

```go
// internal/project/project.go

// Project represents a discovered Git repository
type Project struct {
    Name string // Directory name (e.g., "my-project")
    Path string // Absolute path (e.g., "/home/user/code/my-project")
}

// String returns CSV representation
func (p Project) String() string {
    return fmt.Sprintf("%s,%s", p.Name, p.Path)
}
```

### Cache Format

**File**: `~/.cd_project_folders`
**Format**: CSV without header (simple, grep-able)

```csv
my-project,/home/user/code/my-project
another-repo,/home/user/work/another-repo
duplicate-name,/home/user/personal/duplicate-name
duplicate-name,/home/user/work/duplicate-name
```

**Design Decision**: Allow duplicate names in cache. Resolution happens at query time.

---

## Core Interfaces

### Scanner Interface

```go
// internal/scanner/scanner.go

// Scanner discovers Git repositories
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
```

### Cache Interface

```go
// internal/cache/cache.go

type Cache interface {
    // Load reads projects from cache file
    Load() ([]project.Project, error)

    // Save writes projects to cache file
    Save(projects []project.Project) error

    // Path returns the cache file path
    Path() string
}

// FileCache implements Cache using ~/.cd_project_folders
type FileCache struct {
    path string
}

func NewFileCache() *FileCache {
    home, _ := os.UserHomeDir()
    return &FileCache{path: filepath.Join(home, ".cd_project_folders")}
}
```

### Matcher Interface

```go
// internal/matcher/matcher.go

type Matcher interface {
    // Match returns projects containing the given query (case-insensitive)
    // Sorted by relevance (exact match first, then prefix, then substring)
    Match(query string, projects []project.Project) []project.Project

    // FindExact returns the single best match or error if ambiguous
    FindExact(name string, projects []project.Project) (project.Project, error)
}
```

---

## CLI Commands

### Command: `--completion <query>`
- **Purpose**: Return matching project names for shell completion (substring match)
- **Output**: Newline-separated list of project names
- **Example**: `cd_project --completion lol` → `lol-project\npreloliac`

### Command: `--path <name>`
- **Purpose**: Return absolute path for a project name
- **Output**: Single path or error
- **Duplicate Handling**: If multiple matches, print error with options to stderr, exit 1
- **Example**: `cd_project --path my-project` → `/home/user/code/my-project`

### Command: `--refresh`
- **Purpose**: Rescan directories and rebuild cache
- **Output**: Count of projects found
- **Example**: `cd_project --refresh` → `Found 42 projects`

### Command: (no args)
- **Purpose**: Interactive mode or usage help
- **Output**: Usage information

---

## Shell Integration Design

### Bash Integration (`~/.cd_project.bash`)

```bash
# cd_project shell integration for Bash

# Main cdp function - wrapper around cd
cdp() {
    if [ -z "$1" ]; then
        echo "Usage: cdp <project-name>" >&2
        return 1
    fi

    local target
    target=$(cd_project --path "$1")

    if [ $? -eq 0 ] && [ -n "$target" ]; then
        cd "$target" || return 1
    else
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
```

### Zsh Integration (`~/.cd_project.zsh`)

```zsh
# cd_project shell integration for Zsh

# Main cdp function - wrapper around cd
cdp() {
    if [[ -z "$1" ]]; then
        echo "Usage: cdp <project-name>" >&2
        return 1
    fi

    local target
    target=$(cd_project --path "$1")

    if [[ $? -eq 0 ]] && [[ -n "$target" ]]; then
        cd "$target" || return 1
    else
        return 1
    fi
}

# Zsh completion function
_cdp() {
    local -a completions
    completions=(${(f)"$(cd_project --completion "${words[2]}" 2>/dev/null)"})
    _describe 'project' completions
}

# Register completion
compdef _cdp cdp
```

### Shell RC Integration

Add to `~/.bashrc` or `~/.zshrc`:
```bash
# cd_project integration
export CD_PROJECT_ROOT="$HOME/code:$HOME/work"  # Colon-separated roots
[ -f ~/.cd_project.bash ] && source ~/.cd_project.bash
```

---

## Build Script Design (`scripts/install.sh`)

### Interactive Flow

```
┌─────────────────────────────────────────────────┐
│           cd_project Installation               │
└─────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────┐
│ 1. Detect shells (bash/zsh presence)            │
└─────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────┐
│ 2. Prompt: Install for which shell?             │
│    [1] Bash only                                │
│    [2] Zsh only                                 │
│    [3] Both                                     │
└─────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────┐
│ 3. Prompt: Project root directories?            │
│    Default: $HOME/code                          │
│    (colon-separated for multiple)               │
└─────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────┐
│ 4. Compile Go binary                            │
│    go build -o cd_project ./cmd/cd_project      │
└─────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────┐
│ 5. Install binary to ~/.local/bin or /usr/local │
└─────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────┐
│ 6. Generate shell integration files             │
│    ~/.cd_project.bash and/or ~/.cd_project.zsh  │
└─────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────┐
│ 7. Append source line to shell RC files         │
│    (with backup and duplicate check)            │
└─────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────┐
│ 8. Run initial refresh                          │
│    cd_project --refresh                         │
└─────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────┐
│ 9. Print success message with usage             │
└─────────────────────────────────────────────────┘
```

### Key Script Features

```bash
#!/usr/bin/env bash
set -euo pipefail

# Check if fd is available for faster scanning
check_fd() {
    if command -v fd &>/dev/null; then
        echo "✓ fd found - will use for faster scanning"
    else
        echo "ℹ fd not found - using native Go scanner"
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
```

**Note**: The install script never directly modifies `.bashrc` or `.zshrc` files. Instead, it prints the configuration that users need to manually add to their shell configuration files. This approach:
- Avoids unexpected modifications to user configuration files
- Gives users full control over their shell setup
- Prevents potential conflicts with existing configurations

---

## Edge Cases

### 1. Duplicate Project Names

**Scenario**: Multiple repos with same directory name
```
~/code/my-project/.git
~/work/my-project/.git
```

**Solution**:
- Store all duplicates in cache
- On `--path` with ambiguity, print error with all options:
  ```
  Multiple projects named 'my-project':
    1. /home/user/code/my-project
    2. /home/user/work/my-project
  Use full path or rename directory.
  ```
- On `--completion`, return all matches (shell shows all)

### 2. Symlinks

**Strategy**: Follow symlinks but track visited inodes to avoid cycles

```go
// internal/scanner/native.go
type NativeScanner struct {
    visited map[uint64]bool  // inode tracking
}

func (s *NativeScanner) shouldVisit(info os.FileInfo) bool {
    stat, ok := info.Sys().(*syscall.Stat_t)
    if !ok {
        return true  // can't get inode, visit anyway
    }
    if s.visited[stat.Ino] {
        return false  // already visited
    }
    s.visited[stat.Ino] = true
    return true
}
```

### 3. Permission Errors

**Strategy**: Log and continue, don't fail entire scan

```go
func (s *NativeScanner) Scan(roots []string) ([]project.Project, error) {
    var projects []project.Project
    var scanErrors []error

    for _, root := range roots {
        err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
            if err != nil {
                // Log but continue
                scanErrors = append(scanErrors, fmt.Errorf("access denied: %s", path))
                return filepath.SkipDir
            }
            // ... rest of logic
        })
    }

    // Return projects even if some errors occurred
    return projects, nil
}
```

### 4. Case-Insensitive Substring Matching

**Strategy**: Lowercase comparison for matching, preserve original case in output.
Results are sorted by relevance: exact matches first, then prefix matches, then substring matches.

```go
func (m *PrefixMatcher) Match(query string, projects []project.Project) []project.Project {
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
            return iExact
        }

        iPrefix := strings.HasPrefix(iLower, lowerQuery)
        jPrefix := strings.HasPrefix(jLower, lowerQuery)
        if iPrefix != jPrefix {
            return iPrefix
        }

        return iLower < jLower
    })

    return matches
}
```

### 5. Missing CD_PROJECT_ROOT

**Strategy**: Graceful fallback with helpful error

```go
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

### 6. Empty or Corrupted Cache

**Strategy**: Regenerate automatically

```go
func (c *FileCache) Load() ([]project.Project, error) {
    data, err := os.ReadFile(c.path)
    if os.IsNotExist(err) {
        return nil, nil  // Empty cache, not an error
    }
    if err != nil {
        return nil, err
    }

    var projects []project.Project
    for _, line := range strings.Split(string(data), "\n") {
        line = strings.TrimSpace(line)
        if line == "" {
            continue
        }
        parts := strings.SplitN(line, ",", 2)
        if len(parts) != 2 {
            continue  // Skip malformed lines
        }
        projects = append(projects, project.Project{
            Name: parts[0],
            Path: parts[1],
        })
    }

    return projects, nil
}
```

---

## Performance Considerations

### 1. fd Integration for Faster Scanning

`fd` is significantly faster than Go's `filepath.Walk` for large directory trees.

```go
// internal/scanner/fd.go

type FdScanner struct{}

func fdAvailable() bool {
    _, err := exec.LookPath("fd")
    return err == nil
}

func (s *FdScanner) Scan(roots []string) ([]project.Project, error) {
    var projects []project.Project

    for _, root := range roots {
        // fd command: find .git directories, print parent
        // --type d: directories only
        // --hidden: include hidden dirs (.git)
        // --no-ignore: don't respect .gitignore
        // --prune: don't descend into matches (key for performance!)
        cmd := exec.Command("fd",
            "--type", "d",
            "--hidden",
            "--no-ignore",
            "--prune",
            "^.git$",
            root,
        )

        output, err := cmd.Output()
        if err != nil {
            continue  // Fall through to next root
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

### 2. Native Scanner with Smart Traversal

```go
// internal/scanner/native.go

type NativeScanner struct {
    visited map[uint64]bool
}

func (s *NativeScanner) Scan(roots []string) ([]project.Project, error) {
    s.visited = make(map[uint64]bool)
    var projects []project.Project

    for _, root := range roots {
        filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
            if err != nil {
                return filepath.SkipDir
            }

            if !d.IsDir() {
                return nil
            }

            // Skip common non-project directories for speed
            name := d.Name()
            if name == "node_modules" || name == "vendor" || name == ".cache" {
                return filepath.SkipDir
            }

            // Check for .git
            gitPath := filepath.Join(path, ".git")
            if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
                projects = append(projects, project.Project{
                    Name: filepath.Base(path),
                    Path: path,
                })
                return filepath.SkipDir  // Don't descend into git repos
            }

            return nil
        })
    }

    return projects, nil
}
```

### 3. Caching Strategy

| Operation | Cache Behavior |
|-----------|----------------|
| `--completion` | Read from cache only, never refresh |
| `--path` | Read from cache only, never refresh |
| `--refresh` | Full rescan, overwrite cache |
| Startup | Load cache lazily on first access |

**Rationale**: Tab completion must be instantaneous. Users explicitly trigger refresh when needed.

### 4. Concurrent Scanning (Native Scanner)

```go
func (s *NativeScanner) ScanConcurrent(roots []string) ([]project.Project, error) {
    var wg sync.WaitGroup
    projectsChan := make(chan project.Project, 100)

    // Scan each root concurrently
    for _, root := range roots {
        wg.Add(1)
        go func(r string) {
            defer wg.Done()
            s.scanRoot(r, projectsChan)
        }(root)
    }

    // Collect results
    go func() {
        wg.Wait()
        close(projectsChan)
    }()

    var projects []project.Project
    for p := range projectsChan {
        projects = append(projects, p)
    }

    return projects, nil
}
```

### 5. Performance Benchmarks (Expected)

| Scenario | fd Scanner | Native Scanner |
|----------|-----------|----------------|
| 100 repos, 10K dirs | ~50ms | ~200ms |
| 500 repos, 100K dirs | ~200ms | ~2s |
| 1000 repos, 500K dirs | ~500ms | ~10s |

**Recommendation**: Suggest `fd` installation in README for users with large directory trees.

---

## Technology Stack

### Languages & Frameworks
| Component | Language | Framework | Rationale |
|-----------|----------|-----------|-----------|
| CLI | Go 1.21+ | stdlib only | Single binary, fast startup, no deps |
| Shell Integration | Bash/Zsh | Native | Universal compatibility |
| Install Script | Bash | Native | Works everywhere |

### Dependencies
| Dependency | Type | Purpose |
|------------|------|---------|
| None (stdlib) | Go modules | Keep binary small, fast compile |
| fd (optional) | External CLI | Faster scanning |

---

## Build & Distribution

### Makefile

```makefile
BINARY_NAME=cd_project
VERSION?=$(shell git describe --tags --always --dirty)

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

### Installation Methods

1. **From source**: `make install`
2. **Go install**: `go install github.com/user/cd_project/cmd/cd_project@latest`
3. **Release binary**: Download from GitHub releases

---

## Architecture Decisions

### ADR-001: Single Binary, No External Go Dependencies
- **Status**: Accepted
- **Context**: Tool should be fast to start and easy to distribute
- **Decision**: Use only Go stdlib
- **Alternatives**: Cobra for CLI, viper for config
- **Consequences**: Simpler build, slightly more CLI boilerplate

### ADR-002: CSV Cache Format
- **Status**: Accepted
- **Context**: Need simple, human-readable cache
- **Decision**: Plain CSV without headers
- **Alternatives**: JSON, SQLite, binary format
- **Consequences**: Easy to debug/edit, grep-able, slightly slower for huge caches

### ADR-003: fd as Optional Accelerator
- **Status**: Accepted
- **Context**: Native Go scanning is slow for large trees
- **Decision**: Auto-detect and use fd if available
- **Alternatives**: Require fd, only native
- **Consequences**: Best of both worlds - works everywhere, fast where available

### ADR-004: Environment Variable for Roots
- **Status**: Accepted
- **Context**: Need to configure which directories to scan
- **Decision**: Use `CD_PROJECT_ROOT` env var, colon-separated
- **Alternatives**: Config file, CLI flags, XDG config
- **Consequences**: Simple, follows Unix conventions, easy to override

---

## Risks and Mitigations

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| Large directory trees slow scanning | M | M | Use fd, skip known non-project dirs |
| Stale cache | L | H | Easy refresh command, consider file watcher later |
| Permission errors | L | M | Log and continue, don't fail |
| Symlink cycles | H | L | Track visited inodes |
| Shell compatibility issues | M | L | Test on multiple shell versions |

---

## Future Considerations

1. **File watcher**: Auto-refresh cache when directories change
2. **Fuzzy matching**: Implement fuzzy search (e.g., "mp" matches "my-project")
3. **Frecency sorting**: Track usage, show most-used projects first
4. **Config file**: For advanced users who need more options
5. **Fish shell support**: Add Fish completion
6. **Interactive picker**: TUI for selecting from multiple matches

