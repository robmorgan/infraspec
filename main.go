package main

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/spf13/cobra"

	"github.com/robmorgan/infraspec/internal/config"
	"github.com/robmorgan/infraspec/internal/runner"
)

var (
	verbose bool

	primaryStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#0099cc"))
	secondaryStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff69b4"))

	cucumberOpts = &godog.Options{
		//Output: os.Stdout,
		Output:      colors.Colored(os.Stdout),
		Format:      "pretty",
		Concurrency: 1,
		Randomize:   0,
	}

	rootCmd = &cobra.Command{
		Use:   "infraspec [feature file]",
		Short: "InfraSpec tests infrastructure code using Gherkin syntax.",
		Long:  `InfraSpec is a tool for running infrastructure tests written in pure Gherkin syntax.`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			s := primaryStyle.Render("InfraSpec") + " By Rob Morgan"
			fmt.Println(s)

			cfg, err := config.DefaultConfig()
			if err != nil {
				fmt.Printf("Failed to load config: %v\n", err)
				return
			}

			runner, err := runner.New(cfg)
			if err != nil {
				fmt.Printf("Failed to create runner: %v\n", err)
				return
			}

			// parse args as features path
			// opts.Paths = []string{args[0]}

			// status := tspec.TestSuite{
			// 	Name:                 "tspec",
			// 	TestSuiteInitializer: tspec.InitializeTestSuite,
			// 	ScenarioInitializer:  tspec.InitializeScenario,
			// 	Options:              &opts,
			// }.Run()

			//logger.Infof("Exit code is: %d", status)

			if err := runner.Run(args[0]); err != nil {
				log.Fatalf("Test execution failed: %v", err)
			}
		},
	}
)

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
