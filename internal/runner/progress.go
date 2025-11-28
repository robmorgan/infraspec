package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/term"
)

// FeatureProgress tracks the state of a single feature execution.
type FeatureProgress struct {
	Index       int
	Path        string
	Status      FeatureStatus
	StartTime   time.Time
	DisplayName string
}

// ProgressTracker manages and displays progress for parallel feature execution.
type ProgressTracker struct {
	totalFeatures int
	maxWorkers    int
	features      map[int]*FeatureProgress
	mu            sync.Mutex
	isTTY         bool
	completedCnt  int
	startTime     time.Time
}

// NewProgressTracker creates a new progress tracker.
func NewProgressTracker(totalFeatures, maxWorkers int) *ProgressTracker {
	return &ProgressTracker{
		totalFeatures: totalFeatures,
		maxWorkers:    maxWorkers,
		features:      make(map[int]*FeatureProgress),
		isTTY:         term.IsTerminal(int(os.Stdout.Fd())),
		startTime:     time.Now(),
	}
}

// UpdateStatus updates the status of a feature and prints progress.
func (pt *ProgressTracker) UpdateStatus(index int, path string, status FeatureStatus) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	// Create or update feature progress
	fp, exists := pt.features[index]
	if !exists {
		fp = &FeatureProgress{
			Index:       index,
			Path:        path,
			DisplayName: shortenPath(path),
			StartTime:   time.Now(),
		}
		pt.features[index] = fp
	}

	oldStatus := fp.Status
	fp.Status = status

	// Track completion
	if status != StatusRunning && status != StatusPending && oldStatus == StatusRunning {
		pt.completedCnt++
	}

	// Print progress update
	pt.printProgress(fp)
}

// printProgress prints the current progress state.
func (pt *ProgressTracker) printProgress(fp *FeatureProgress) {
	statusStr := formatStatus(fp.Status)
	elapsed := ""
	if fp.Status != StatusRunning && fp.Status != StatusPending {
		elapsed = fmt.Sprintf(" (%s)", time.Since(fp.StartTime).Round(time.Millisecond))
	}

	if pt.isTTY {
		// Simple line output for TTY (could be enhanced with ANSI codes for live updates)
		fmt.Printf("[%d/%d] %s %s%s\n",
			fp.Index+1,
			pt.totalFeatures,
			fp.DisplayName,
			statusStr,
			elapsed,
		)
	} else {
		// Plain output for non-TTY (CI environments)
		fmt.Printf("[%d/%d] %s %s%s\n",
			fp.Index+1,
			pt.totalFeatures,
			fp.DisplayName,
			statusStr,
			elapsed,
		)
	}
}

// PrintHeader prints the initial progress header.
func (pt *ProgressTracker) PrintHeader() {
	fmt.Printf("\nRunning %d feature(s) with %d worker(s)...\n\n",
		pt.totalFeatures,
		pt.maxWorkers,
	)
}

// formatStatus returns a formatted status string.
func formatStatus(status FeatureStatus) string {
	switch status {
	case StatusPending:
		return "[PENDING]"
	case StatusRunning:
		return "[RUNNING]"
	case StatusPassed:
		return "[PASS]"
	case StatusFailed:
		return "[FAIL]"
	case StatusTimeout:
		return "[TIMEOUT]"
	case StatusCanceled:
		return "[CANCELED]"
	default:
		return "[UNKNOWN]"
	}
}

// shortenPath returns a shortened display version of a file path.
// Keeps the last 2 directory components + filename for readability.
func shortenPath(path string) string {
	// Get the path relative to working directory if possible
	if wd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(wd, path); err == nil && len(rel) < len(path) {
			return rel
		}
	}

	// Otherwise, keep last few components
	parts := splitPath(path)
	if len(parts) <= 3 {
		return path
	}
	return filepath.Join(parts[len(parts)-3:]...)
}

// splitPath splits a path into its components.
func splitPath(path string) []string {
	var parts []string
	for {
		dir, file := filepath.Split(path)
		if file != "" {
			parts = append([]string{file}, parts...)
		}
		if dir == "" || dir == path {
			break
		}
		path = filepath.Clean(dir)
	}
	return parts
}
