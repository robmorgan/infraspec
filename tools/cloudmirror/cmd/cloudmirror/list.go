package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/analyzer"
	"github.com/robmorgan/infraspec/tools/cloudmirror/internal/models"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list [resource]",
	Short: "List AWS services",
	Long: `List AWS services from the SDK, implemented services, missing services,
or services in the allowlist.

Resources:
  services      List all AWS services available in the SDK
  implemented   List services implemented in InfraSpec
  missing       List AWS services not yet implemented
  allowlist     List services in the priority allowlist

Examples:
  cloudmirror list services
  cloudmirror list implemented
  cloudmirror list missing
  cloudmirror list allowlist`,
	Args: cobra.ExactArgs(1),
	Run:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) {
	resource := strings.ToLower(args[0])

	switch resource {
	case "services", "sdk":
		listSDKServices()
	case "implemented", "impl":
		listImplementedServices()
	case "missing":
		listMissingServices()
	case "allowlist", "allowed":
		listAllowlistServices()
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown resource '%s'. Use: services, implemented, missing, or allowlist\n", resource)
		os.Exit(1)
	}
}

func listSDKServices() {
	requireSDKPath()

	anal := analyzer.NewAnalyzer(sdkPath, servicesPath)
	services, err := anal.ListAWSServices()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing services: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Available AWS services in SDK:")
	for _, svc := range services {
		fmt.Printf("  - %s\n", svc)
	}
	fmt.Printf("\nTotal: %d services\n", len(services))
}

func listImplementedServices() {
	anal := analyzer.NewAnalyzer(sdkPath, servicesPath)
	services, err := anal.ListImplementedServices()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing implemented services: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Implemented services in InfraSpec:")
	for _, svc := range services {
		fmt.Printf("  - %s\n", svc)
	}
	fmt.Printf("\nTotal: %d services\n", len(services))
}

func listMissingServices() {
	requireSDKPath()

	anal := analyzer.NewAnalyzer(sdkPath, servicesPath)

	// Get all AWS services
	awsServices, err := anal.ListAWSServices()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing AWS services: %v\n", err)
		os.Exit(1)
	}

	// Get implemented services
	implServices, err := anal.ListImplementedServices()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing implemented services: %v\n", err)
		os.Exit(1)
	}

	// Build set of implemented services (normalized)
	implSet := make(map[string]bool)
	for _, svc := range implServices {
		implSet[strings.ToLower(svc)] = true
	}

	// Find missing services using centralized service name mappings
	var missing []string
	for _, svc := range awsServices {
		normalized := strings.ToLower(svc)
		infraspecName := models.GetInfraSpecServiceName(normalized)
		if !implSet[infraspecName] {
			missing = append(missing, svc)
		}
	}

	fmt.Printf("AWS services not yet implemented (%d of %d):\n", len(missing), len(awsServices))
	for _, svc := range missing {
		fmt.Printf("  - %s\n", svc)
	}
}

func listAllowlistServices() {
	services := models.GetServicesByPriority()

	fmt.Println("Allowed AWS Services (sorted by priority):")
	fmt.Println("==========================================")
	fmt.Println()

	for _, svc := range services {
		enabledStr := "✓"
		if !svc.Enabled {
			enabledStr = "✗"
		}
		fmt.Printf("  %s %-25s (priority: %3d) - %s\n", enabledStr, svc.Name, svc.Priority, svc.FullName)
	}

	fmt.Println()
	fmt.Printf("Total: %d services\n", len(services))
}
