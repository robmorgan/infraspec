package runner

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFeatureStatus_String(t *testing.T) {
	tests := []struct {
		status   FeatureStatus
		expected string
	}{
		{StatusPending, "PENDING"},
		{StatusRunning, "RUNNING"},
		{StatusPassed, "PASS"},
		{StatusFailed, "FAIL"},
		{StatusTimeout, "TIMEOUT"},
		{StatusCanceled, "CANCELED"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

func TestParallelConfig(t *testing.T) {
	cfg := ParallelConfig{
		MaxWorkers: 4,
		Timeout:    5 * time.Minute,
	}

	assert.Equal(t, 4, cfg.MaxWorkers)
	assert.Equal(t, 5*time.Minute, cfg.Timeout)
}

func TestFeatureResult(t *testing.T) {
	result := FeatureResult{
		FeaturePath: "/path/to/test.feature",
		Status:      StatusPassed,
		Duration:    10 * time.Second,
		Error:       nil,
	}

	assert.Equal(t, "/path/to/test.feature", result.FeaturePath)
	assert.Equal(t, StatusPassed, result.Status)
	assert.Equal(t, 10*time.Second, result.Duration)
	assert.Nil(t, result.Error)
}

func TestAggregatedResults(t *testing.T) {
	results := &AggregatedResults{
		TotalFeatures:  5,
		PassedFeatures: 3,
		FailedFeatures: 2,
		TotalDuration:  1 * time.Minute,
		Results: []FeatureResult{
			{FeaturePath: "test1.feature", Status: StatusPassed},
			{FeaturePath: "test2.feature", Status: StatusPassed},
			{FeaturePath: "test3.feature", Status: StatusPassed},
			{FeaturePath: "test4.feature", Status: StatusFailed},
			{FeaturePath: "test5.feature", Status: StatusFailed},
		},
	}

	assert.Equal(t, 5, results.TotalFeatures)
	assert.Equal(t, 3, results.PassedFeatures)
	assert.Equal(t, 2, results.FailedFeatures)
	assert.Len(t, results.Results, 5)
}

func TestAggregateResults(t *testing.T) {
	pr := &ParallelRunner{}

	results := []FeatureResult{
		{FeaturePath: "test1.feature", Status: StatusPassed},
		{FeaturePath: "test2.feature", Status: StatusFailed},
		{FeaturePath: "test3.feature", Status: StatusPassed},
		{FeaturePath: "test4.feature", Status: StatusTimeout},
	}

	duration := 30 * time.Second
	agg := pr.aggregateResults(results, duration)

	assert.Equal(t, 4, agg.TotalFeatures)
	assert.Equal(t, 2, agg.PassedFeatures)
	assert.Equal(t, 2, agg.FailedFeatures)
	assert.Equal(t, duration, agg.TotalDuration)
}
