package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"goo-cli/internal/config"
	"goo-cli/internal/utils"
)

// NewConfigCmd creates the config command
func NewConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
		Long:  "Initialize and manage the CLI configuration file",
	}

	cmd.AddCommand(NewConfigInitCmd())
	cmd.AddCommand(NewConfigShowCmd())

	return cmd
}

// NewConfigInitCmd initializes the config file
func NewConfigInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration file",
		Long:  "Create a new configuration file with default values at ~/.goo/config.yaml",
		Example: `  goo config init`,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath, err := config.GetConfigPath()
			if err != nil {
				return err
			}

			// Check if config already exists
			if _, err := os.Stat(configPath); err == nil {
				return fmt.Errorf("config file already exists at %s", configPath)
			}

			// Create config with default values
			cfg := config.DefaultConfig()

			// Ask for Google API key (optional)
			fmt.Println()
			fmt.Println("🔍 AI-Powered Proposal Configuration (Optional)")
			fmt.Println("═══════════════════════════════════════════════")
			fmt.Println()
			fmt.Println("The CLI can use Google Gemini AI to automatically research and propose values.")
			fmt.Println("This feature requires a free Google API key.")
			fmt.Println()
			fmt.Print("Enter Google API Key (leave empty to skip): ")

			reader := bufio.NewReader(os.Stdin)
			apiKey, _ := reader.ReadString('\n')
			apiKey = strings.TrimSpace(apiKey)

			if apiKey != "" {
				cfg.GoogleAPIKey = apiKey
				utils.PrintSuccess("Google API key configured")
			} else {
				utils.PrintInfo("Skipped - you can add it later in ~/.goo/config.yaml")
			}

			// Save config
			if err := config.Save(cfg); err != nil {
				return err
			}

			fmt.Println()
			utils.PrintSuccess(fmt.Sprintf("Config file created at %s", configPath))
			fmt.Println()
			fmt.Println("Configuration:")
			fmt.Printf("  Key Name:      %s\n", cfg.KeyName)
			fmt.Printf("  Realm Path:    %s\n", cfg.RealmPath)
			fmt.Printf("  Chain ID:      %s\n", cfg.ChainID)
			fmt.Printf("  Remote:        %s\n", cfg.Remote)
			fmt.Printf("  Gas Fee:       %s\n", cfg.GasFee)
			fmt.Printf("  Gas Wanted:    %d\n", cfg.GasWanted)
			if cfg.GoogleAPIKey != "" {
				maskedKey := cfg.GoogleAPIKey
				if len(maskedKey) > 8 {
					maskedKey = maskedKey[:8] + "..."
				}
				fmt.Printf("  Google API Key: %s\n", maskedKey)
			} else {
				fmt.Printf("  Google API Key: (not configured)\n")
			}
			fmt.Println()
			fmt.Println("Edit this file to customize your settings.")

			if cfg.GoogleAPIKey == "" {
				fmt.Println()
				fmt.Println("💡 To enable AI-powered proposals:")
				fmt.Println("  1. Get a free API key: https://makersuite.google.com/app/apikey")
				fmt.Println("  2. Edit ~/.goo/config.yaml and add:")
				fmt.Println("     google_api_key: your-api-key-here")
				fmt.Println("  3. Use: goo propose value <id> --search")
			}

			return nil
		},
	}

	return cmd
}

// NewConfigShowCmd displays the current configuration
func NewConfigShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long:  "Display the current CLI configuration",
		Example: `  goo config show`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()

			utils.PrintSection("Current Configuration")
			utils.PrintKeyValue("Key Name", cfg.KeyName)
			utils.PrintKeyValue("Realm Path", cfg.RealmPath)
			utils.PrintKeyValue("Chain ID", cfg.ChainID)
			utils.PrintKeyValue("Remote", cfg.Remote)
			utils.PrintKeyValue("Gas Fee", cfg.GasFee)
			utils.PrintKeyValue("Gas Wanted", cfg.GasWanted)
			
			if cfg.GoogleAPIKey != "" {
				maskedKey := cfg.GoogleAPIKey
				if len(maskedKey) > 8 {
					maskedKey = maskedKey[:8] + "..."
				}
				utils.PrintKeyValue("Google API Key", maskedKey)
			} else {
				utils.PrintKeyValue("Google API Key", "(not configured)")
			}
			fmt.Println()

			configPath, err := config.GetConfigPath()
			if err == nil {
				utils.PrintInfo(fmt.Sprintf("Config file: %s", configPath))
			}

			return nil
		},
	}

	return cmd
}
