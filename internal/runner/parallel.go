package runner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robmorgan/infraspec/internal/config"
)

// FeatureStatus represents the execution status of a feature.
type FeatureStatus int

const (
	StatusPending FeatureStatus = iota
	StatusRunning
	StatusPassed
	StatusFailed
	StatusTimeout
	StatusCanceled
)

func (s FeatureStatus) String() string {
	switch s {
	case StatusPending:
		return "PENDING"
	case StatusRunning:
		return "RUNNING"
	case StatusPassed:
		return "PASS"
	case StatusFailed:
		return "FAIL"
	case StatusTimeout:
		return "TIMEOUT"
	case StatusCanceled:
		return "CANCELED"
	default:
		return "UNKNOWN"
	}
}

// ParallelConfig holds configuration for parallel execution.
type ParallelConfig struct {
	MaxWorkers int           // Maximum concurrent feature executions
	Timeout    time.Duration // Per-feature timeout (0 = no timeout)
}

// FeatureResult captures the result of a single feature file execution.
type FeatureResult struct {
	FeaturePath string        // Path to the .feature file
	Status      FeatureStatus // Overall status
	Duration    time.Duration // Execution duration
	Error       error         // Error if failed
}

// AggregatedResults combines results from all parallel executions.
type AggregatedResults struct {
	TotalFeatures  int
	PassedFeatures int
	FailedFeatures int
	TotalDuration  time.Duration
	Results        []FeatureResult
}

// ParallelRunner orchestrates parallel feature execution.
type ParallelRunner struct {
	cfg         *config.Config
	parallelCfg ParallelConfig
	progress    *ProgressTracker
	cancel      context.CancelFunc
	wg          sync.WaitGroup
}

// NewParallelRunner creates a new parallel runner.
func NewParallelRunner(cfg *config.Config, parallelCfg ParallelConfig) *ParallelRunner {
	return &ParallelRunner{
		cfg:         cfg,
		parallelCfg: parallelCfg,
	}
}

// RunParallel executes multiple feature files in parallel.
func (pr *ParallelRunner) RunParallel(ctx context.Context, featurePaths []string, format string) (*AggregatedResults, error) {
	ctx, pr.cancel = context.WithCancel(ctx)
	defer pr.cancel()

	startTime := time.Now()
	totalFeatures := len(featurePaths)

	// Initialize progress tracker
	pr.progress = NewProgressTracker(totalFeatures, pr.parallelCfg.MaxWorkers)

	// Create semaphore channel for limiting concurrent workers
	semaphore := make(chan struct{}, pr.parallelCfg.MaxWorkers)

	// Create channel for results
	resultsChan := make(chan FeatureResult, totalFeatures)

	// Launch workers for each feature file
	for i, featurePath := range featurePaths {
		pr.wg.Add(1)
		go func(index int, fp string) {
			defer pr.wg.Done()

			// Acquire semaphore slot
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				resultsChan <- FeatureResult{
					FeaturePath: fp,
					Status:      StatusCanceled,
					Error:       ctx.Err(),
				}
				return
			}

			// Update progress to running
			pr.progress.UpdateStatus(index, fp, StatusRunning)

			// Run the feature
			result := pr.runSingleFeature(ctx, fp, format)

			// Update progress with result
			pr.progress.UpdateStatus(index, fp, result.Status)

			resultsChan <- result
		}(i, featurePath)
	}

	// Close results channel when all workers done
	go func() {
		pr.wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	var results []FeatureResult
	for result := range resultsChan {
		results = append(results, result)
	}

	// Sort results by original order (by feature path)
	sortedResults := make([]FeatureResult, totalFeatures)
	resultMap := make(map[string]FeatureResult)
	for _, r := range results {
		resultMap[r.FeaturePath] = r
	}
	for i, fp := range featurePaths {
		sortedResults[i] = resultMap[fp]
	}

	return pr.aggregateResults(sortedResults, time.Since(startTime)), nil
}

// runSingleFeature executes a single feature file with isolation.
func (pr *ParallelRunner) runSingleFeature(ctx context.Context, featurePath, format string) FeatureResult {
	startTime := time.Now()
	result := FeatureResult{
		FeaturePath: featurePath,
		Status:      StatusRunning,
	}

	// Apply per-feature timeout if configured
	if pr.parallelCfg.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, pr.parallelCfg.Timeout)
		defer cancel()
	}

	// Create a channel to receive the result
	done := make(chan error, 1)

	go func() {
		// Create isolated runner
		runner := New(pr.cfg)
		done <- runner.RunWithFormat(featurePath, format)
	}()

	// Wait for completion or cancellation
	select {
	case err := <-done:
		result.Duration = time.Since(startTime)
		if err != nil {
			result.Status = StatusFailed
			result.Error = err
		} else {
			result.Status = StatusPassed
		}
	case <-ctx.Done():
		result.Duration = time.Since(startTime)
		if ctx.Err() == context.DeadlineExceeded {
			result.Status = StatusTimeout
			result.Error = fmt.Errorf("feature execution timed out after %v", pr.parallelCfg.Timeout)
		} else {
			result.Status = StatusCanceled
			result.Error = ctx.Err()
		}
	}

	return result
}

// aggregateResults combines individual results into summary.
func (pr *ParallelRunner) aggregateResults(results []FeatureResult, totalDuration time.Duration) *AggregatedResults {
	agg := &AggregatedResults{
		TotalFeatures: len(results),
		TotalDuration: totalDuration,
		Results:       results,
	}

	for _, r := range results {
		if r.Status == StatusPassed {
			agg.PassedFeatures++
		} else {
			agg.FailedFeatures++
		}
	}

	return agg
}

// Cancel stops all running feature executions.
func (pr *ParallelRunner) Cancel() {
	if pr.cancel != nil {
		pr.cancel()
	}
}

// GetProgress returns the current progress tracker.
func (pr *ParallelRunner) GetProgress() *ProgressTracker {
	return pr.progress
}
