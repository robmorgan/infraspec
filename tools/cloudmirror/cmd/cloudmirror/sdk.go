package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/sdkcache"
)

var sdkListLimit int

var sdkCmd = &cobra.Command{
	Use:   "sdk",
	Short: "Manage the cached AWS SDK",
	Long: `Manage the cached AWS SDK Go V2 repository.

CloudMirror caches the AWS SDK in ~/.cloudmirror/aws-sdk-go-v2 for reuse
across commands. Use this subcommand to view, update, or manage the cache.

Examples:
  cloudmirror sdk status              # Show cached SDK info
  cloudmirror sdk update              # Update to latest
  cloudmirror sdk checkout v1.30.0    # Switch to specific version
  cloudmirror sdk list-versions       # List available versions
  cloudmirror sdk clean               # Remove cached SDK`,
}

var sdkStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show cached SDK information",
	Long:  `Display information about the cached AWS SDK, including path, version, and size.`,
	Run:   runSDKStatus,
}

var sdkUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update cached SDK to latest",
	Long:  `Fetch the latest changes from the AWS SDK repository and update the cache.`,
	Run:   runSDKUpdate,
}

var sdkCheckoutCmd = &cobra.Command{
	Use:   "checkout <version>",
	Short: "Switch to a specific SDK version",
	Long: `Switch the cached SDK to a specific version (git tag).

Examples:
  cloudmirror sdk checkout v1.30.0
  cloudmirror sdk checkout release-2024-01-15`,
	Args: cobra.ExactArgs(1),
	Run:  runSDKCheckout,
}

var sdkListVersionsCmd = &cobra.Command{
	Use:   "list-versions",
	Short: "List available SDK versions",
	Long:  `List available AWS SDK versions (git tags) that can be checked out.`,
	Run:   runSDKListVersions,
}

var sdkCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove cached SDK",
	Long:  `Remove the cached AWS SDK from ~/.cloudmirror/aws-sdk-go-v2.`,
	Run:   runSDKClean,
}

func init() {
	rootCmd.AddCommand(sdkCmd)

	sdkCmd.AddCommand(sdkStatusCmd)
	sdkCmd.AddCommand(sdkUpdateCmd)
	sdkCmd.AddCommand(sdkCheckoutCmd)
	sdkCmd.AddCommand(sdkListVersionsCmd)
	sdkCmd.AddCommand(sdkCleanCmd)

	sdkListVersionsCmd.Flags().IntVar(&sdkListLimit, "limit", 20, "Maximum number of versions to display")
}

func runSDKStatus(cmd *cobra.Command, args []string) {
	cache, err := getSDKCache()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	status, err := cache.Status()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting status: %v\n", err)
		os.Exit(1)
	}

	if !status.Exists {
		fmt.Println("SDK not cached.")
		fmt.Println("")
		fmt.Println("Run a command that requires the SDK (e.g., 'cloudmirror analyze')")
		fmt.Println("to automatically download it, or use 'cloudmirror sdk update'.")
		return
	}

	// Output as JSON if not quiet mode, otherwise human-readable
	if quiet {
		output, _ := json.MarshalIndent(status, "", "  ")
		fmt.Println(string(output))
		return
	}

	fmt.Println("AWS SDK Cache Status")
	fmt.Println("====================")
	fmt.Printf("Path:      %s\n", status.CacheDir)
	fmt.Printf("Version:   %s\n", status.Version)
	fmt.Printf("Commit:    %s\n", status.CommitSHA)
	fmt.Printf("Size:      %s\n", status.SizeHuman)
}

func runSDKUpdate(cmd *cobra.Command, args []string) {
	if !sdkcache.IsGitAvailable() {
		fmt.Fprintln(os.Stderr, "Error: git is not installed.")
		fmt.Fprintln(os.Stderr, "Please install git to use SDK management features.")
		os.Exit(1)
	}

	cache, err := getSDKCache()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// If not cached, clone first
	if !cache.HasCache() {
		if !quiet {
			fmt.Println("SDK not cached. Downloading...")
		}
		if _, err := cache.GetSDKPath(""); err != nil {
			fmt.Fprintf(os.Stderr, "Error downloading SDK: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Update existing cache
	if err := cache.Update(); err != nil {
		fmt.Fprintf(os.Stderr, "Error updating SDK: %v\n", err)
		os.Exit(1)
	}
}

func runSDKCheckout(cmd *cobra.Command, args []string) {
	version := args[0]

	if !sdkcache.IsGitAvailable() {
		fmt.Fprintln(os.Stderr, "Error: git is not installed.")
		os.Exit(1)
	}

	cache, err := getSDKCache()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !cache.HasCache() {
		fmt.Fprintln(os.Stderr, "SDK not cached. Run 'cloudmirror sdk update' first.")
		os.Exit(1)
	}

	if err := cache.Checkout(version); err != nil {
		fmt.Fprintf(os.Stderr, "Error checking out version: %v\n", err)
		os.Exit(1)
	}
}

func runSDKListVersions(cmd *cobra.Command, args []string) {
	if !sdkcache.IsGitAvailable() {
		fmt.Fprintln(os.Stderr, "Error: git is not installed.")
		os.Exit(1)
	}

	cache, err := getSDKCache()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if !cache.HasCache() {
		fmt.Fprintln(os.Stderr, "SDK not cached. Run 'cloudmirror sdk update' first.")
		os.Exit(1)
	}

	versions, err := cache.ListVersions(sdkListLimit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing versions: %v\n", err)
		os.Exit(1)
	}

	if len(versions) == 0 {
		fmt.Println("No versions found.")
		return
	}

	if !quiet {
		fmt.Printf("Available SDK versions (showing %d):\n", len(versions))
		fmt.Println()
	}

	for _, v := range versions {
		fmt.Println(v)
	}

	if !quiet && sdkListLimit > 0 && len(versions) == sdkListLimit {
		fmt.Println()
		fmt.Printf("Use --limit to show more versions.\n")
	}
}

func runSDKClean(cmd *cobra.Command, args []string) {
	cache, err := getSDKCache()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := cache.Clean(); err != nil {
		fmt.Fprintf(os.Stderr, "Error cleaning cache: %v\n", err)
		os.Exit(1)
	}
}
