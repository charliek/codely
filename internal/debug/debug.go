// Package debug provides an optional file-based debug logger.
// By default all calls are no-ops. Call Enable to start writing
// timestamped log lines to a file.
package debug

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/charliek/codely/internal/pathutil"
)

var (
	mu      sync.Mutex
	file    *os.File
	enabled bool
)

// Enable opens the log file for writing (truncated). The path is expanded
// via pathutil.ExpandPath so "~" is supported.
func Enable(path string) error {
	mu.Lock()
	defer mu.Unlock()

	path = pathutil.ExpandPath(path)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("creating debug log directory: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("opening debug log: %w", err)
	}

	file = f
	enabled = true
	return nil
}

// Close closes the log file.
func Close() {
	mu.Lock()
	defer mu.Unlock()
	if file != nil {
		file.Close()
		file = nil
	}
	enabled = false
}

// Log writes a timestamped line to the debug log if enabled.
func Log(format string, a ...any) {
	mu.Lock()
	defer mu.Unlock()
	if !enabled || file == nil {
		return
	}
	ts := time.Now().Format("15:04:05.000")
	fmt.Fprintf(file, "%s  %s\n", ts, fmt.Sprintf(format, a...))
}

// Enabled returns whether debug logging is active.
func Enabled() bool {
	mu.Lock()
	defer mu.Unlock()
	return enabled
}
