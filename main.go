package main

import (
	"fmt"
	"log"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/robmorgan/infraspec/internal/config"
	"github.com/robmorgan/infraspec/internal/runner"
)

var (
	verbose bool

	primaryStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#0099cc"))
	secondaryStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ff69b4"))

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
