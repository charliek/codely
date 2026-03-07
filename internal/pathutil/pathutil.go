// Package pathutil provides shared path manipulation utilities.
package pathutil

import (
	"os"
	"path/filepath"
	"strings"
)

// ContractHome replaces the user's home directory prefix with ~.
func ContractHome(path string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	if strings.HasPrefix(path, home) {
		return "~" + path[len(home):]
	}
	return path
}

// ExpandPath expands ~ to the user's home directory and cleans the path.
// NOTE: This does not enforce sandbox boundaries; it only normalizes the path.
func ExpandPath(path string) string {
	if path == "" {
		return path
	}

	// Expand ~ to home directory
	if path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		path = home + path[1:]
	}

	// Clean the path to resolve . and .. components
	path = filepath.Clean(path)

	return path
}
