package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"goo-cli/internal/config"
	"goo-cli/internal/gnokey"
	"goo-cli/internal/utils"
)

// NewQueryCmd creates the query command
func NewQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query oracle data",
		Long:  "Read-only queries for oracle state and parameters",
	}

	cmd.AddCommand(NewQueryResultCmd())
	cmd.AddCommand(NewQueryParamsCmd())
	cmd.AddCommand(NewQueryListCmd())

	return cmd
}

// NewQueryResultCmd queries the result of a request
func NewQueryResultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "result <request-id>",
		Short: "Get the final result of a request",
		Long:  "Query the winning value for a resolved request",
		Args:  cobra.ExactArgs(1),
		Example: `  goo query result req-001`,
		RunE: func(cmd *cobra.Command, args []string) error {
			requestID := args[0]

			cfg := config.Load()
			executor := gnokey.NewExecutor(cfg)

			// Query result
			result, err := executor.QueryFunction("RequestResult", []string{requestID})
			if err != nil {
				return err
			}

			utils.PrintSuccess(fmt.Sprintf("Result for request %s:", requestID))
			fmt.Println(result)

			return nil
		},
	}

	return cmd
}

// NewQueryParamsCmd queries oracle parameters
func NewQueryParamsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Get oracle parameters",
		Long:  "Query all oracle configuration parameters",
		Example: `  goo query params`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()
			executor := gnokey.NewExecutor(cfg)

			utils.PrintSection("Oracle Parameters")

			// Query each parameter
			params := []struct {
				name     string
				funcName string
			}{
				{"Bond", "GetBond"},
				{"Resolution Time", "GetResolutionTime"},
				{"Requester Reward", "GetRequesterReward"},
				{"Dispute Duration", "GetDisputeDuration"},
				{"Reveal Duration", "GetRevealDuration"},
				{"Vote Token Price", "GetVoteTokenPrice"},
			}

			for _, p := range params {
				result, err := executor.QueryFunction(p.funcName, []string{})
				if err != nil {
					utils.PrintError(fmt.Sprintf("Failed to query %s: %v", p.name, err))
					continue
				}
				utils.PrintKeyValue(p.name, result)
			}

			return nil
		},
	}

	return cmd
}

// NewQueryListCmd lists requests (limited functionality without indexer)
func NewQueryListCmd() *cobra.Command {
	var state string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List requests (limited without indexer)",
		Long:  "Attempt to list requests by parsing Render() output. Note: This is limited without a proper indexer.",
		Example: `  goo query list
  goo query list --state Proposed`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()
			executor := gnokey.NewExecutor(cfg)

			// Query the Render function
			result, err := executor.QueryFunction("Render", []string{""})
			if err != nil {
				return err
			}

			utils.PrintSuccess("Oracle State:")
			if state != "" {
				utils.PrintInfo(fmt.Sprintf("Filtering by state: %s", state))
			}
			fmt.Println(result)

			utils.PrintWarning("Note: Full list functionality requires an indexer")
			utils.PrintInfo("Consider using gnokey query to call Render() directly for formatted output")

			return nil
		},
	}

	cmd.Flags().StringVar(&state, "state", "", "Filter by state: Requested, Proposed, Disputed, Resolved")

	return cmd
}
