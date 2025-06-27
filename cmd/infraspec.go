package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/robmorgan/infraspec/internal/build"
	"github.com/robmorgan/infraspec/internal/config"
	"github.com/robmorgan/infraspec/internal/runner"
	"github.com/robmorgan/infraspec/internal/telemetry"
)

var (
	verbose          bool
	disableTelemetry bool

	primaryStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#0099cc"))
	secondaryStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff69b4"))

	rootCmd = &cobra.Command{
		Use:     "infraspec [features...]",
		Short:   "InfraSpec tests infrastructure code using Gherkin syntax.",
		Long:    `InfraSpec is a tool for running infrastructure tests written in pure Gherkin syntax.`,
		Version: build.Version,
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			startTime := time.Now()

			s := primaryStyle.Render("InfraSpec") + " By Rob Morgan"
			fmt.Println(s)

			cfg, err := config.DefaultConfig()
			if err != nil {
				fmt.Printf("Failed to load config: %v\n", err)
				return
			}

			if verbose {
				cfg.Verbose = true
				cfg.Logger.Debug("Verbose mode enabled")
			}

			// Disable telemetry if requested
			if disableTelemetry {
				cfg.Telemetry.Enabled = false
			}

			// Initialize telemetry
			tel := telemetry.New(telemetry.Config{
				Enabled: cfg.Telemetry.Enabled,
				UserID:  cfg.Telemetry.UserID,
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
			if err := runner.New(cfg).Run(featureFile); err != nil {
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
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&disableTelemetry, "disable-telemetry", false, "disable telemetry collection")

	rootCmd.SetVersionTemplate(`{{printf "%s version %s\n" .Name .Version}}`)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
