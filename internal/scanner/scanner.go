// Package scanner provides functionality to discover Git repositories.
package scanner

import (
	"os/exec"

	"github.com/dbarton/cd_project/internal/project"
)

// Scanner discovers Git repositories in directory trees.
type Scanner interface {
	// Scan finds all Git repos under the given root directories.
	// Skips subdirectories once .git is found.
	Scan(roots []string) ([]project.Project, error)
}

// NewScanner returns the best available scanner.
// Prefers fd if available, falls back to native.
func NewScanner() Scanner {
	if FdAvailable() {
		return &FdScanner{}
	}
	return &NativeScanner{}
}

// ScannerType returns "fd" or "native" based on what will be used.
func ScannerType() string {
	if FdAvailable() {
		return "fd"
	}
	return "native"
}

// FdAvailable checks if fd is installed.
// Checks for both "fd" and "fdfind" (Debian/Ubuntu package name).
func FdAvailable() bool {
	return fdCommand() != ""
}

// fdCommand returns the fd command name if available.
// On Debian/Ubuntu, the package is called fd-find and installs as "fdfind".
func fdCommand() string {
	if _, err := exec.LookPath("fd"); err == nil {
		return "fd"
	}
	if _, err := exec.LookPath("fdfind"); err == nil {
		return "fdfind"
	}
	return ""
}

