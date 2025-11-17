package cmd

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/robmorgan/infraspec/internal/build"
	"github.com/robmorgan/infraspec/internal/config"
	"github.com/robmorgan/infraspec/internal/runner"
	"github.com/robmorgan/infraspec/internal/telemetry"
)

var (
	verbose      bool
	format       string
	virtualCloud bool

	RootCmd = &cobra.Command{
		Use:     "infraspec [features...]",
		Short:   "InfraSpec tests infrastructure code in plain English.",
		Long:    `InfraSpec is a tool for testing your cloud infrastructure in plain English, no code required.`,
		Version: build.Version,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			startTime := time.Now()

			cfg, err := config.LoadConfig("", virtualCloud)
			if err != nil {
				fmt.Printf("Failed to load config: %v\n", err)
				return
			}

			if verbose {
				cfg.Verbose = true
				config.Logging.Logger.Debug("Verbose mode enabled")
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

			// Track test run start
			featureFile := args[0]
			tel.TrackTestRun(featureFile)

			// Run tests
			if err := runner.New(cfg).RunWithFormat(featureFile, format); err != nil {
				// Track test failure
				tel.TrackTestFailed(featureFile, time.Since(startTime), err.Error())
				log.Fatalf("Test execution failed: %v", err)
			}

			// Track successful completion
			tel.TrackTestComplete(featureFile, time.Since(startTime), 0) // You can pass actual step count from runner
		},
	}
)

func init() {
	// Global flags
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	RootCmd.PersistentFlags().StringVarP(&format, "format", "f", "default", "output format (default, text, pretty, junit, cucumber)")
	RootCmd.PersistentFlags().BoolVar(&virtualCloud, "virtual-cloud", false, "use InfraSpec Virtual Cloud to emulate AWS-compatible APIs")
	RootCmd.PersistentFlags().BoolVar(&virtualCloud, "vc", false, "Alias for --virtual-cloud")

	RootCmd.SetVersionTemplate(`{{printf "%s version %s\n" .Name .Version}}`)
}
