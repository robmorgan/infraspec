package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/robmorgan/infraspec/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage InfraSpec configuration",
	Long:  `Manage InfraSpec configuration settings, including InfraSpec Cloud authentication tokens.`,
}

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a configuration value",
	Long:  `Set a configuration value. Currently supported keys: infraspec-cloud-token`,
	Args:  cobra.ExactArgs(2),
	RunE:  runConfigSet,
}

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration value",
	Long:  `Get a configuration value. Currently supported keys: infraspec-cloud-token`,
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigGet,
}

var configUnsetCmd = &cobra.Command{
	Use:   "unset [key]",
	Short: "Unset a configuration value",
	Long:  `Unset a configuration value. Currently supported keys: infraspec-cloud-token`,
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigUnset,
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configUnsetCmd)
	RootCmd.AddCommand(configCmd)
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	key := args[0]
	value := args[1]

	switch key {
	case "infraspec-cloud-token":
		userConfig, err := config.LoadUserConfig()
		if err != nil {
			return fmt.Errorf("failed to load user config: %w", err)
		}

		userConfig.InfraspecCloudToken = value

		if err := config.SaveUserConfig(userConfig); err != nil {
			return fmt.Errorf("failed to save user config: %w", err)
		}

		configPath, err := config.GetUserConfigPath()
		if err != nil {
			return fmt.Errorf("failed to get config path: %w", err)
		}

		fmt.Printf("✅ Successfully set infraspec-cloud-token\n")
		fmt.Printf("📁 Configuration saved to: %s\n", configPath)
		return nil
	default:
		return fmt.Errorf("unknown configuration key: %s. Supported keys: infraspec-cloud-token", key)
	}
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	key := args[0]

	switch key {
	case "infraspec-cloud-token":
		// Check environment variable first
		if token := os.Getenv(config.InfraspecCloudTokenEnvVar); token != "" {
			fmt.Printf("infraspec-cloud-token=%s (from environment variable %s)\n", maskToken(token), config.InfraspecCloudTokenEnvVar)
			return nil
		}

		// Fall back to config file
		userConfig, err := config.LoadUserConfig()
		if err != nil {
			return fmt.Errorf("failed to load user config: %w", err)
		}

		if userConfig.InfraspecCloudToken == "" {
			fmt.Printf("infraspec-cloud-token is not set\n")
			return nil
		}

		fmt.Printf("infraspec-cloud-token=%s\n", maskToken(userConfig.InfraspecCloudToken))
		return nil
	default:
		return fmt.Errorf("unknown configuration key: %s. Supported keys: infraspec-cloud-token", key)
	}
}

func runConfigUnset(cmd *cobra.Command, args []string) error {
	key := args[0]

	switch key {
	case "infraspec-cloud-token":
		userConfig, err := config.LoadUserConfig()
		if err != nil {
			return fmt.Errorf("failed to load user config: %w", err)
		}

		if userConfig.InfraspecCloudToken == "" {
			fmt.Printf("infraspec-cloud-token is not set\n")
			return nil
		}

		userConfig.InfraspecCloudToken = ""

		if err := config.SaveUserConfig(userConfig); err != nil {
			return fmt.Errorf("failed to save user config: %w", err)
		}

		configPath, err := config.GetUserConfigPath()
		if err != nil {
			return fmt.Errorf("failed to get config path: %w", err)
		}

		fmt.Printf("✅ Successfully unset infraspec-cloud-token\n")
		fmt.Printf("📁 Configuration saved to: %s\n", configPath)
		return nil
	default:
		return fmt.Errorf("unknown configuration key: %s. Supported keys: infraspec-cloud-token", key)
	}
}

// maskToken masks a token for display, showing only the first 4 characters
func maskToken(token string) string {
	if len(token) <= 4 {
		return strings.Repeat("*", len(token))
	}
	return token[:4] + strings.Repeat("*", len(token)-4)
}
