package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/robmorgan/infraspec/internal/build"
	"github.com/robmorgan/infraspec/internal/config"
	"github.com/robmorgan/infraspec/internal/runner"
	"github.com/robmorgan/infraspec/internal/telemetry"
	"github.com/robmorgan/infraspec/pkg/awshelpers"
)

var (
	verbose      bool
	format       string
	virtualCloud bool
	parallel     int // Number of features to run in parallel (0 = sequential)
	timeout      int // Per-feature timeout in seconds (0 = no timeout)

	RootCmd = &cobra.Command{
		Use:     "infraspec [features...]",
		Short:   "InfraSpec tests infrastructure code in plain English.",
		Long:    `InfraSpec is a tool for testing your cloud infrastructure in plain English, no code required.`,
		Version: build.Version,
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			startTime := time.Now()

			cfg, err := config.LoadConfig("", virtualCloud)
			if err != nil {
				fmt.Printf("Failed to load config: %v\n", err)
				return
			}

			// Set parallel mode flag in config
			if parallel > 0 {
				cfg.ParallelMode = true
			}

			if verbose {
				cfg.Verbose = true
				config.Logging.Logger.Debug("Verbose mode enabled")
			}

			// Check Virtual Cloud API health if enabled
			if err := awshelpers.CheckVirtualCloudHealth(); err != nil {
				fmt.Printf("Virtual Cloud health check failed: %v\n", err)
				return
			}

			// Initialize telemetry
			tel := telemetry.New(telemetry.Config{
				Enabled: cfg.Telemetry.Enabled,
				UserID:  cfg.Telemetry.UserID,
				Logger:  zap.NewNop().Sugar(), // Discard all telemetry output
			})

			// Ensure telemetry is flushed on exit
			defer tel.Flush()

			// Track CLI start
			tel.TrackCLIStart(args)
			tel.TrackConfigLoaded("default")

			// Discover all feature files from provided paths
			var featureFiles []string
			for _, arg := range args {
				files, err := runner.DiscoverFeatureFiles(arg)
				if err != nil {
					log.Fatalf("Failed to discover features: %v", err)
				}
				featureFiles = append(featureFiles, files...)
			}

			// Remove duplicates
			featureFiles = runner.UniqueStrings(featureFiles)

			if parallel > 0 && len(featureFiles) > 1 {
				// Parallel execution mode
				runParallel(cfg, tel, featureFiles, startTime)
			} else {
				// Sequential execution mode
				runSequential(cfg, tel, featureFiles, startTime)
			}
		},
	}
)

// runParallel executes feature files in parallel.
func runParallel(cfg *config.Config, tel *telemetry.Client, featureFiles []string, startTime time.Time) {
	parallelCfg := runner.ParallelConfig{
		MaxWorkers: parallel,
		Timeout:    time.Duration(timeout) * time.Second,
	}

	pr := runner.NewParallelRunner(cfg, parallelCfg)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt, canceling tests...")
		pr.Cancel()
	}()

	// Print header
	if progress := pr.GetProgress(); progress != nil {
		progress.PrintHeader()
	} else {
		fmt.Printf("\nRunning %d feature(s) with %d worker(s)...\n\n", len(featureFiles), parallel)
	}

	ctx := context.Background()
	results, err := pr.RunParallel(ctx, featureFiles, format)
	if err != nil {
		log.Fatalf("Parallel execution failed: %v", err)
	}

	// Print summary
	runner.PrintParallelResults(results)

	// Track telemetry for each feature
	for _, r := range results.Results {
		if r.Status == runner.StatusPassed {
			tel.TrackTestComplete(r.FeaturePath, r.Duration, 0)
		} else {
			errMsg := ""
			if r.Error != nil {
				errMsg = r.Error.Error()
			}
			tel.TrackTestFailed(r.FeaturePath, r.Duration, errMsg)
		}
	}

	if results.FailedFeatures > 0 {
		os.Exit(1)
	}
}

// runSequential executes feature files sequentially (original behavior).
func runSequential(cfg *config.Config, tel *telemetry.Client, featureFiles []string, startTime time.Time) {
	var failed bool
	for _, featureFile := range featureFiles {
		featureStart := time.Now()
		tel.TrackTestRun(featureFile)

		if err := runner.New(cfg).RunWithFormat(featureFile, format); err != nil {
			tel.TrackTestFailed(featureFile, time.Since(featureStart), err.Error())
			log.Printf("Test execution failed for %s: %v", featureFile, err)
			failed = true
			continue
		}
		tel.TrackTestComplete(featureFile, time.Since(featureStart), 0)
	}

	if failed {
		os.Exit(1)
	}
}

func init() {
	// Global flags
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	RootCmd.PersistentFlags().StringVarP(&format, "format", "f", "default", "output format (default, text, pretty, junit, cucumber)")
	RootCmd.PersistentFlags().BoolVar(&virtualCloud, "virtual-cloud", false, "use InfraSpec Virtual Cloud to emulate AWS-compatible APIs")

	// Parallel execution flags
	RootCmd.PersistentFlags().IntVarP(&parallel, "parallel", "p", 0, "number of features to run in parallel (0 = sequential)")
	RootCmd.PersistentFlags().IntVar(&timeout, "timeout", 0, "per-feature timeout in seconds (0 = no timeout)")

	RootCmd.SetVersionTemplate(`{{printf "%s version %s\n" .Name .Version}}`)
}
