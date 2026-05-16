# cd_project

Fast project directory navigator with shell integration and tab completion.

## Features

- 🚀 **Fast** - Uses `fd` for scanning when available (~10x faster), falls back to native Go
- 🔍 **Smart** - Case-insensitive substring matching, skips `node_modules`, `vendor`, etc.
- 🐚 **Shell Integration** - Native Bash and Zsh support with tab completion
- 📦 **Zero Dependencies** - Single Go binary, no runtime dependencies

## Installation

### Install

```bash
git clone https://github.com/dbarton/cd_project.git
cd cd_project
go install ./cmd/cd_project
# then move ./cd_project somewhere on your PATH
```

### 2. Set your project root

```bash
export CD_PROJECT_ROOT="$HOME/code"
# or multiple roots, colon-separated:
export CD_PROJECT_ROOT="$HOME/code:$HOME/work"
```

### 3. Wire up your shell

Add one of the following to your shell rc file. `cd_project init` prints the `cdp` function, tab completion, and the `cdp-refresh` alias.

**Bash** — add to `~/.bashrc`:

```bash
eval "$(cd_project init bash)"
```

**Zsh** — add to `~/.zshrc`:

```bash
eval "$(cd_project init zsh)"
```

Open a new shell (or `source` your rc file) and run `cdp <name>`. The first invocation will build the cache automatically.

## Usage

```bash
# Navigate to a project
cdp my-project

# Tab completion
cdp my<TAB>
# Output: my-app  my-api  my-project

# Refresh project cache after adding new repos
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

## CLI Reference

```
cd_project - Fast project directory navigator

Usage:
  cd_project --completion <query>   List matching projects
  cd_project --path <name>          Get project path
  cd_project --refresh              Rebuild project cache

Environment:
  CD_PROJECT_ROOT  Colon-separated list of root directories to scan
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
2. Results are cached in `~/.cd_project_folders` (CSV format)
3. `cdp <name>` looks up the project in cache and `cd`s to it
4. Tab completion uses the cache for instant suggestions

## Project Structure

```
cd_project/
├── cmd/cd_project/            # CLI entry point (cobra commands, shell init templates)
├── internal/
│   ├── cache/                 # Cache read/write
│   ├── matcher/               # Substring matching
│   ├── project/               # Project struct and name resolution
│   └── scanner/               # Directory scanning (native + fd)
└── README.md
```

## License

MIT
