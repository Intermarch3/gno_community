package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"goo-cli/internal/config"
	"goo-cli/internal/gnokey"
	"goo-cli/internal/utils"
)

// NewProposeCmd creates the propose command
func NewProposeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "propose",
		Short: "Propose values for requests",
		Long:  "Propose a value for a data request or resolve a request",
	}

	cmd.AddCommand(NewProposeValueCmd())
	cmd.AddCommand(NewProposeResolveCmd())

	return cmd
}

// NewProposeValueCmd proposes a value for a request
func NewProposeValueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "value <request-id> <value>",
		Short: "Propose a value for a request",
		Long:  "Propose a value for a data request. Requires bond to be sent with the transaction.",
		Args:  cobra.ExactArgs(2),
		Example: `  goo propose value 0000001 3500`,
		RunE: func(cmd *cobra.Command, args []string) error {
			requestID := args[0]
			value := args[1]

			keyOverride, _ := cmd.Flags().GetString("key")
			verbose, _ := cmd.Flags().GetBool("verbose")
			cfg := config.LoadWithKeyOverride(keyOverride)
			executor := gnokey.NewExecutor(cfg, verbose)

			// Query the required bond amount from contract
			utils.PrintInfo("Querying required bond amount from contract...")
			bond, err := executor.QueryInt64("GetBond")
			if err != nil {
				return fmt.Errorf("failed to query bond amount: %w", err)
			}

			utils.PrintInfo(fmt.Sprintf("Bond required: %d ugnot", bond))

			// Execute transaction with bond
			funcArgs := []string{requestID, value}
			sendAmount := fmt.Sprintf("%dugnot", bond)

			if err := executor.CallFunction("ProposeValue", funcArgs, sendAmount); err != nil {
				return err
			}

			utils.PrintSuccess("Value proposed successfully!")
			utils.PrintInfo(fmt.Sprintf("Request ID: %s", requestID))
			utils.PrintInfo(fmt.Sprintf("Proposed Value: %s", value))
			utils.PrintInfo(fmt.Sprintf("Bond sent: %d ugnot", bond))

			return nil
		},
	}

	return cmd
}

// NewProposeResolveCmd resolves a non-disputed request
func NewProposeResolveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve <request-id>",
		Short: "Resolve a non-disputed request",
		Long:  "Finalize a request that has not been disputed after the resolution time",
		Args:  cobra.ExactArgs(1),
		Example: `  goo propose resolve 0000001`,
		RunE: func(cmd *cobra.Command, args []string) error {
			requestID := args[0]

			keyOverride, _ := cmd.Flags().GetString("key")
			verbose, _ := cmd.Flags().GetBool("verbose")
			cfg := config.LoadWithKeyOverride(keyOverride)
			executor := gnokey.NewExecutor(cfg, verbose)

			// Execute transaction
			if err := executor.CallFunction("ResolveRequest", []string{requestID}, ""); err != nil {
				return err
			}

			utils.PrintSuccess("Request resolution submitted!")
			utils.PrintInfo(fmt.Sprintf("Request ID: %s", requestID))

			return nil
		},
	}

	return cmd
}
