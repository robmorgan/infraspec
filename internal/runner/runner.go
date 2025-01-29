package runner

import (
	"fmt"
	"os"
	"time"

	"github.com/cucumber/godog"
	"go.uber.org/zap"

	"github.com/robmorgan/infraspec/internal/config"
	"github.com/robmorgan/infraspec/internal/context"
	"github.com/robmorgan/infraspec/pkg/steps"
)

// Runner handles the execution of feature files
type Runner struct {
	context *context.TestContext
	cfg     *config.Config
}

func New(cfg *config.Config) (*Runner, error) {
	return &Runner{
		cfg: cfg,
	}, nil
}

// Run executes the specified feature file
func (r *Runner) Run(featurePath string) error {
	defer r.cfg.Logger.Sync()

	// Validate feature file exists
	if _, err := os.Stat(featurePath); os.IsNotExist(err) {
		return fmt.Errorf("feature file not found: %s", featurePath)
	}

	r.cfg.Logger.Info("Starting test execution",
		zap.String("feature", featurePath),
		zap.String("provider", r.cfg.Provider),
		zap.String("region", r.cfg.DefaultRegion),
	)

	suite := &godog.TestSuite{
		ScenarioInitializer: r.initializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{featurePath},
			TestingT: nil,
		},
	}

	start := time.Now()
	status := suite.Run()
	duration := time.Since(start)

	// Log test execution summary
	r.cfg.Logger.Info("Test execution completed",
		zap.Duration("duration", duration),
		zap.Int("status", status),
	)

	if err := r.cleanup(); err != nil {
		r.cfg.Logger.Error("Cleanup failed", zap.Error(err))
		return err
	}

	if status != 0 {
		return fmt.Errorf("test execution failed with status: %d", status)
	}

	return nil
}

// initializeScenario sets up the godog scenario context
func (r *Runner) initializeScenario(sc *godog.ScenarioContext) {
	// Initialize test context for each scenario
	sc.BeforeScenario(func(*godog.Scenario) {
		r.context = context.New(r.cfg)
	})

	// Register step definitions
	steps.RegisterSteps(r.context, sc)

	// Add hooks for logging
	sc.BeforeStep(func(st *godog.Step) {
		r.cfg.Logger.Debug("Executing step",
			zap.String("step", st.Text),
		)
	})

	sc.AfterStep(func(st *godog.Step, err error) {
		if err != nil {
			r.cfg.Logger.Error("Step failed",
				zap.String("step", st.Text),
				zap.Error(err),
			)
		} else {
			r.cfg.Logger.Debug("Step completed successfully",
				zap.String("step", st.Text),
			)
		}
	})

	sc.AfterScenario(func(sc *godog.Scenario, err error) {
		if err != nil {
			r.cfg.Logger.Error("Scenario failed",
				zap.String("scenario", sc.Name),
				zap.Error(err),
			)
		} else {
			r.cfg.Logger.Info("Scenario completed successfully",
				zap.String("scenario", sc.Name),
			)
		}
	})
}

// cleanup performs necessary cleanup after test execution
func (r *Runner) cleanup() error {
	if !r.cfg.Cleanup.Automatic {
		r.cfg.Logger.Info("Automatic cleanup disabled, skipping...")
		return nil
	}

	r.cfg.Logger.Info("Starting cleanup",
		zap.Int("timeout", r.cfg.Cleanup.Timeout),
	)

	done := make(chan error)
	go func() {
		done <- r.context.Cleanup()
	}()

	select {
	case err := <-done:
		if err != nil {
			return fmt.Errorf("cleanup failed: %w", err)
		}
		r.cfg.Logger.Info("Cleanup completed successfully")
		return nil
	case <-time.After(time.Duration(r.cfg.Cleanup.Timeout) * time.Second):
		return fmt.Errorf("cleanup timed out after %d seconds", r.cfg.Cleanup.Timeout)
	}
}
