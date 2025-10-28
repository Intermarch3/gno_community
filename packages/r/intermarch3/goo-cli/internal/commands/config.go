package commands

import (
	"fmt"

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
			if err := config.InitConfig(); err != nil {
				return err
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
