package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/monitor"
)

var (
	releaseVersionFile string
	releaseGithubToken string
)

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Check for new AWS SDK releases",
	Long: `Check for new AWS SDK Go V2 releases on GitHub and compare against
the currently tracked version.

Examples:
  cloudmirror release
  cloudmirror release --version-file=.aws-sdk-version
  GITHUB_TOKEN=$TOKEN cloudmirror release`,
	Run: runRelease,
}

func init() {
	rootCmd.AddCommand(releaseCmd)

	releaseCmd.Flags().StringVar(&releaseVersionFile, "version-file", ".aws-sdk-version", "File storing current SDK version")
	releaseCmd.Flags().StringVar(&releaseGithubToken, "github-token", "", "GitHub token for API access (or GITHUB_TOKEN env)")
}

func runRelease(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	// Get token from environment if not provided
	token := releaseGithubToken
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	// Create monitor
	mon := monitor.NewSDKReleaseMonitor(token)

	// Check if there's a new version
	hasUpdate, latest, err := mon.HasNewVersion(ctx, releaseVersionFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking for new version: %v\n", err)
		os.Exit(1)
	}

	// Get current version
	current, err := mon.GetCurrentVersion(releaseVersionFile)
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error reading current version: %v\n", err)
		os.Exit(1)
	}
	if current == "" {
		current = "none"
	}

	// Output result
	result := struct {
		LatestVersion  string `json:"latest_version"`
		CurrentVersion string `json:"current_version"`
		HasUpdate      bool   `json:"has_update"`
		ReleaseURL     string `json:"release_url,omitempty"`
		PublishedAt    string `json:"published_at,omitempty"`
	}{
		LatestVersion:  latest.TagName,
		CurrentVersion: current,
		HasUpdate:      hasUpdate,
		ReleaseURL:     latest.HTMLURL,
		PublishedAt:    latest.PublishedAt.Format("2006-01-02T15:04:05Z"),
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling result: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(output))

	if !quiet && hasUpdate {
		fmt.Fprintf(os.Stderr, "\nNew SDK version available: %s (current: %s)\n", latest.TagName, current)
	}
}
