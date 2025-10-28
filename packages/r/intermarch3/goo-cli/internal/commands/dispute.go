package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"goo-cli/internal/config"
	"goo-cli/internal/gnokey"
	"goo-cli/internal/utils"
)

// NewDisputeCmd creates the dispute command
func NewDisputeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dispute",
		Short: "Manage disputes",
		Long:  "Create, query, and resolve disputes on proposed values",
	}

	cmd.AddCommand(NewDisputeCreateCmd())
	cmd.AddCommand(NewDisputeGetCmd())
	cmd.AddCommand(NewDisputeResolveCmd())

	return cmd
}

// NewDisputeCreateCmd creates a new dispute
func NewDisputeCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <request-id>",
		Short: "Create a dispute on a proposed value",
		Long:  "Challenge a proposed value by creating a dispute. Requires bond to be sent with the transaction.",
		Args:  cobra.ExactArgs(1),
		Example: `  goo dispute create req-001`,
		RunE: func(cmd *cobra.Command, args []string) error {
			requestID := args[0]

			cfg := config.Load()
			executor := gnokey.NewExecutor(cfg)

			utils.PrintWarning("Make sure to check the required bond amount before submitting!")

			// Execute transaction
			if err := executor.CallFunction("DisputeData", []string{requestID}, ""); err != nil {
				return err
			}

			utils.PrintSuccess("Dispute created successfully!")
			utils.PrintInfo(fmt.Sprintf("Request ID: %s", requestID))
			utils.PrintInfo("Voting period has started")
			utils.PrintWarning("Don't forget to add --send <bond>ugnot when executing the actual transaction")

			return nil
		},
	}

	return cmd
}

// NewDisputeGetCmd gets details of a dispute
func NewDisputeGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <request-id>",
		Short: "Get details of a dispute",
		Long:  "Retrieve details about a dispute including vote counts and timing",
		Args:  cobra.ExactArgs(1),
		Example: `  goo dispute get req-001`,
		RunE: func(cmd *cobra.Command, args []string) error {
			requestID := args[0]

			cfg := config.Load()
			executor := gnokey.NewExecutor(cfg)

			// Query the dispute
			result, err := executor.QueryFunction("GetDispute", []string{requestID})
			if err != nil {
				return err
			}

			utils.PrintSuccess(fmt.Sprintf("Dispute details for: %s", requestID))
			fmt.Println(result)

			return nil
		},
	}

	return cmd
}

// NewDisputeResolveCmd resolves a dispute
func NewDisputeResolveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve <request-id>",
		Short: "Resolve a dispute after voting period",
		Long:  "Finalize a dispute after the reveal period has ended",
		Args:  cobra.ExactArgs(1),
		Example: `  goo dispute resolve req-001`,
		RunE: func(cmd *cobra.Command, args []string) error {
			requestID := args[0]

			cfg := config.Load()
			executor := gnokey.NewExecutor(cfg)

			// Execute transaction
			if err := executor.CallFunction("ResolveDispute", []string{requestID}, ""); err != nil {
				return err
			}

			utils.PrintSuccess("Dispute resolution submitted!")
			utils.PrintInfo(fmt.Sprintf("Request ID: %s", requestID))

			return nil
		},
	}

	return cmd
}
