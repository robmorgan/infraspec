package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
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
			filename := args[0]
			fmt.Sprintln("testing %s", filename)

			s := primaryStyle.Render("InfraSpec") + " By Rob Morgan"
			fmt.Println(s)
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
