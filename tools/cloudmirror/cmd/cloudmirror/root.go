package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/sdkcache"
)

const (
	version = "0.3.0"
	banner  = `
   _____ _                 _ __  __ _
  / ____| |               | |  \/  (_)
 | |    | | ___  _   _  __| | \  / |_ _ __ _ __ ___  _ __
 | |    | |/ _ \| | | |/ _' | |\/| | | '__| '__/ _ \| '__|
 | |____| | (_) | |_| | (_| | |  | | | |  | | | (_) | |
  \_____|_|\___/ \__,_|\__,_|_|  |_|_|_|  |_|  \___/|_|

  AWS API Parity Analyzer for InfraSpec
`
)

// Global flags shared across commands
var (
	sdkPath        string
	servicesPath   string
	verbose        bool
	quiet          bool
	updateSDK      bool
	sdkVersion     string
	noAutoDownload bool
)

// sdkCacheInstance is the global SDK cache instance
var sdkCacheInstance *sdkcache.SDKCache

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "cloudmirror",
	Short: "AWS API Parity Analyzer for InfraSpec",
	Long: `CloudMirror is a comprehensive tool for analyzing AWS API parity,
generating service scaffolds, monitoring SDK releases, and AI-powered
code generation for the InfraSpec API emulator.

Commands:
  analyze    Analyze AWS API coverage for implemented services
  scaffold   Generate service scaffold for a new AWS service
  compare    Compare two SDK versions for breaking changes
  list       List AWS services (sdk, implemented, missing, allowlist)
  sdk        Manage the cached AWS SDK
  release    Check for new AWS SDK releases
  generate   AI-powered implementation generation
  validate   Run Terraform validation tests
  check      Analyze code for pattern compliance

SDK Auto-Download:
  CloudMirror automatically downloads the AWS SDK to ~/.cloudmirror/aws-sdk-go-v2
  when needed. Use --no-auto-download to disable this behavior, or --update-sdk
  to force an update before running a command.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Print banner unless quiet mode or help/completion commands
		if !quiet && cmd.Name() != "help" && cmd.Name() != "completion" {
			fmt.Fprint(os.Stderr, banner)
			fmt.Fprintf(os.Stderr, "  Version: %s\n\n", version)
		}

		// Auto-detect paths if not specified
		if servicesPath == "" {
			servicesPath = findServicesPath()
		}
		if sdkPath == "" {
			sdkPath = findSDKPath()
		}

		// Handle --update-sdk flag
		if updateSDK && sdkPath != "" && sdkCacheInstance != nil {
			// Only update if using cached SDK
			if strings.Contains(sdkPath, sdkcache.DefaultCacheDir) {
				if err := sdkCacheInstance.Update(); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to update SDK: %v\n", err)
				}
			}
		}
	},
	Version: version,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global persistent flags
	rootCmd.PersistentFlags().StringVar(&sdkPath, "sdk-path", "", "Path to AWS SDK Go V2 source (auto-detected or downloaded if not specified)")
	rootCmd.PersistentFlags().StringVar(&servicesPath, "services-path", "", "Path to InfraSpec services directory (default: internal/emulator/services)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress banner and informational output")

	// SDK cache flags
	rootCmd.PersistentFlags().BoolVar(&updateSDK, "update-sdk", false, "Update cached SDK before running command")
	rootCmd.PersistentFlags().StringVar(&sdkVersion, "sdk-version", "", "Use specific SDK version (git tag, e.g., v1.30.0)")
	rootCmd.PersistentFlags().BoolVar(&noAutoDownload, "no-auto-download", false, "Disable automatic SDK download")
}

// findServicesPath attempts to locate the services directory
func findServicesPath() string {
	candidates := []string{
		"internal/emulator/services",
		"../internal/emulator/services",
		"../../internal/emulator/services",
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			abs, _ := filepath.Abs(candidate)
			return abs
		}
	}
	return "internal/emulator/services"
}

// findSDKPath attempts to locate the AWS SDK Go V2 source.
// It checks common locations first, then falls back to auto-download if enabled.
func findSDKPath() string {
	// Try common locations (prefer full clones with model files)
	candidates := []string{
		os.Getenv("AWS_SDK_GO_V2_PATH"),
		"/tmp/aws-sdk-go-v2",
		"../aws-sdk-go-v2",
		"../../aws-sdk-go-v2",
	}

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if hasModelsDir(candidate) {
			return candidate
		}
	}

	// Check SDK cache first (preferred - has full clone with all models)
	cache, err := sdkcache.NewSDKCache(verbose, quiet)
	if err == nil {
		sdkCacheInstance = cache
		if cache.HasCache() {
			// Handle version checkout if specified
			if sdkVersion != "" {
				if err := cache.Checkout(sdkVersion); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: Failed to checkout version %s: %v\n", sdkVersion, err)
				}
			}
			return cache.GetSDKDir()
		}
	}

	// Try GOPATH module cache as fallback (may have incomplete models)
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = filepath.Join(os.Getenv("HOME"), "go")
	}

	modPath := filepath.Join(gopath, "pkg/mod/github.com/aws")
	entries, err := os.ReadDir(modPath)
	if err == nil {
		var sdkVersions []string
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), "aws-sdk-go-v2@") {
				sdkVersions = append(sdkVersions, entry.Name())
			}
		}
		sort.Sort(sort.Reverse(sort.StringSlice(sdkVersions)))

		for _, v := range sdkVersions {
			candidate := filepath.Join(modPath, v)
			if hasModelsDir(candidate) {
				return candidate
			}
		}
	}

	// Auto-download if enabled
	if !noAutoDownload {
		if !sdkcache.IsGitAvailable() {
			// Git not available, can't auto-download
			return ""
		}

		if cache == nil {
			cache, err = sdkcache.NewSDKCache(verbose, quiet)
			if err != nil {
				return ""
			}
			sdkCacheInstance = cache
		}

		// Download SDK
		sdkDir, err := cache.GetSDKPath(sdkVersion)
		if err != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr, "Warning: Failed to download SDK: %v\n", err)
			}
			return ""
		}
		return sdkDir
	}

	return ""
}

// hasModelsDir checks if the SDK path contains the Smithy model files
func hasModelsDir(sdkPath string) bool {
	modelsPath := filepath.Join(sdkPath, "codegen", "sdk-codegen", "aws-models")
	info, err := os.Stat(modelsPath)
	return err == nil && info.IsDir()
}

// requireSDKPath validates that SDK path is available and exits if not
func requireSDKPath() {
	if sdkPath == "" {
		fmt.Fprintln(os.Stderr, "Error: Could not find or download AWS SDK with Smithy model files.")
		fmt.Fprintln(os.Stderr, "")
		if noAutoDownload {
			fmt.Fprintln(os.Stderr, "Auto-download is disabled. You can either:")
			fmt.Fprintln(os.Stderr, "  1. Remove --no-auto-download to allow automatic SDK download")
			fmt.Fprintln(os.Stderr, "  2. Manually clone the SDK and specify --sdk-path:")
		} else if !sdkcache.IsGitAvailable() {
			fmt.Fprintln(os.Stderr, "Git is not installed. Please either:")
			fmt.Fprintln(os.Stderr, "  1. Install git to enable automatic SDK download")
			fmt.Fprintln(os.Stderr, "  2. Manually download the SDK and specify --sdk-path:")
		} else {
			fmt.Fprintln(os.Stderr, "Please clone the full SDK and specify the path:")
		}
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "  git clone --depth 1 https://github.com/aws/aws-sdk-go-v2.git /tmp/aws-sdk-go-v2")
		fmt.Fprintln(os.Stderr, "  cloudmirror --sdk-path=/tmp/aws-sdk-go-v2 ...")
		os.Exit(1)
	}
}

// getSDKCache returns the global SDK cache instance, creating it if necessary
func getSDKCache() (*sdkcache.SDKCache, error) {
	if sdkCacheInstance != nil {
		return sdkCacheInstance, nil
	}

	cache, err := sdkcache.NewSDKCache(verbose, quiet)
	if err != nil {
		return nil, err
	}
	sdkCacheInstance = cache
	return cache, nil
}
