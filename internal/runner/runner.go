package runner

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cucumber/godog"
	"go.uber.org/zap"

	"github.com/robmorgan/infraspec/internal/config"
	"github.com/robmorgan/infraspec/internal/contexthelpers"
	"github.com/robmorgan/infraspec/pkg/steps"
	"github.com/robmorgan/infraspec/pkg/steps/terraform"
)

// Runner handles the execution of feature files
type Runner struct {
	cfg *config.Config
}

func New(cfg *config.Config) *Runner {
	return &Runner{
		cfg: cfg,
	}
}

// Run executes the specified feature file
func (r *Runner) Run(featurePath string) error {
	defer r.cfg.Logger.Sync()

	// Validate feature file exists
	if _, err := os.Stat(featurePath); os.IsNotExist(err) {
		return fmt.Errorf("feature file not found: %s", featurePath)
	}

	r.cfg.Logger.Infof("Starting test execution using: %s", featurePath)

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
	r.cfg.Logger.Debugf("\nTest execution completed in %s with status: %d", duration, status)

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
	sc.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		// embed the config
		ctx = context.WithValue(ctx, contexthelpers.ConfigCtxKey{}, r.cfg)

		// embed the uri
		return context.WithValue(ctx, contexthelpers.UriCtxKey{}, sc.Uri), nil
	})

	// Register step definitions
	steps.RegisterSteps(sc)

	// Add hooks for logging
	sc.StepContext().Before(func(ctx context.Context, st *godog.Step) (context.Context, error) {
		r.cfg.Logger.Debug("Executing step", st, st.Text)
		return ctx, nil
	})

	sc.StepContext().After(func(ctx context.Context, st *godog.Step, status godog.StepResultStatus, err error) (context.Context, error) {
		if err != nil {
			r.cfg.Logger.Error("Step failed", "step", st.Text, "error", err)
		} else {
			r.cfg.Logger.Debug("Step completed successfully", "step", st.Text)
		}
		return ctx, nil
	})

	sc.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
		if err != nil {
			r.cfg.Logger.Error("Scenario failed", "scenario", sc.Name, "error", err)
		} else {
			r.cfg.Logger.Debugf("Scenario completed successfully: %s", sc.Name)
		}

		// If a Terraform configuration was applied, destroy it
		if contexthelpers.GetTerraformHasApplied(ctx) {
			r.cfg.Logger.Debug("Terraform has been applied, destroying resources")
			terraform.NewTerraformDestroyStep(ctx)
		}
		return ctx, nil
	})
}

// cleanup performs necessary cleanup after test execution
// TODO - this might be necessary if we've invoked tools like Terraform or need to cleanup resources
func (r *Runner) cleanup() error {
	if !r.cfg.Cleanup.Automatic {
		r.cfg.Logger.Debug("Automatic cleanup disabled, skipping...")
		return nil
	}

	r.cfg.Logger.Info("Starting cleanup",
		zap.Int("timeout", r.cfg.Cleanup.Timeout),
	)

	// done := make(chan error)
	// go func() {
	// 	done <- r.context.Cleanup()
	// }()

	// select {
	// case err := <-done:
	// 	if err != nil {
	// 		return fmt.Errorf("cleanup failed: %w", err)
	// 	}
	// 	r.cfg.Logger.Info("Cleanup completed successfully")
	// 	return nil
	// case <-time.After(time.Duration(r.cfg.Cleanup.Timeout) * time.Second):
	// 	return fmt.Errorf("cleanup timed out after %d seconds", r.cfg.Cleanup.Timeout)
	// }

	r.cfg.Logger.Debug("Cleanup completed successfully")
	return nil
}
