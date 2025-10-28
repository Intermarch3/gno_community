package gnokey

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"goo-cli/internal/config"
	"goo-cli/internal/utils"
)

// TxExecutor handles execution of gnokey transactions
type TxExecutor struct {
	KeyName   string
	RealmPath string
	ChainID   string
	Remote    string
	GasFee    string
	GasWanted int64
}

// NewExecutor creates a new TxExecutor from config
func NewExecutor(cfg *config.Config) *TxExecutor {
	return &TxExecutor{
		KeyName:   cfg.KeyName,
		RealmPath: cfg.RealmPath,
		ChainID:   cfg.ChainID,
		Remote:    cfg.Remote,
		GasFee:    cfg.GasFee,
		GasWanted: cfg.GasWanted,
	}
}

// CallFunction executes a function call (transaction)
func (e *TxExecutor) CallFunction(funcName string, args []string, sendCoins string) error {
	// Build command arguments
	cmdArgs := []string{
		"maketx", "call",
		"--pkgpath", e.RealmPath,
		"--func", funcName,
		"--gas-fee", e.GasFee,
		"--gas-wanted", fmt.Sprintf("%d", e.GasWanted),
		"--broadcast",
		"--chainid", e.ChainID,
		"--remote", e.Remote,
	}

	// Add function arguments
	for _, arg := range args {
		cmdArgs = append(cmdArgs, "--args", arg)
	}

	// Add coins if specified
	if sendCoins != "" {
		cmdArgs = append(cmdArgs, "--send", sendCoins)
	}

	// Add key name
	cmdArgs = append(cmdArgs, e.KeyName)

	// Print the command being executed
	fmt.Println("Executing:")
	printCommand("gnokey", cmdArgs)
	fmt.Println()

	// Execute the command with inherited stdin/stdout/stderr for interactive password input
	cmd := exec.Command("gnokey", cmdArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}

	return nil
}

// QueryFunction executes a query (read-only call)
func (e *TxExecutor) QueryFunction(funcName string, args []string) (string, error) {
	// Build the query path
	queryPath := fmt.Sprintf("%s.%s", e.RealmPath, funcName)
	if len(args) > 0 {
		formattedArgs := formatArgs(args)
		queryPath += fmt.Sprintf("(%s)", strings.Join(formattedArgs, ","))
	}

	// Build command arguments
	cmdArgs := []string{
		"query", "vm/qeval",
		"--remote", e.Remote,
		"--data", queryPath,
	}

	// Print the command being executed
	fmt.Println("Executing:")
	printCommand("gnokey", cmdArgs)
	fmt.Println()

	// Execute the command
	cmd := exec.Command("gnokey", cmdArgs...)
	output, err := cmd.CombinedOutput()

	// Print output
	fmt.Println(string(output))

	if err != nil {
		return "", fmt.Errorf("query failed: %w", err)
	}

	return string(output), nil
}

// formatArgs formats arguments for Gno function calls
func formatArgs(args []string) []string {
	formatted := make([]string, len(args))
	for i, arg := range args {
		formatted[i] = fmt.Sprintf("\"%s\"", arg)
	}
	return formatted
}

// printCommand prints a command with proper quoting for display
func printCommand(name string, args []string) {
	fmt.Print(name)
	for _, arg := range args {
		// Quote arguments that contain spaces or special characters
		if strings.ContainsAny(arg, " \t\n\"'") {
			fmt.Printf(" \"%s\"", strings.ReplaceAll(arg, "\"", "\\\""))
		} else {
			fmt.Printf(" %s", arg)
		}
	}
	fmt.Println()
}

// SaveVoteLocally saves vote data to local storage
func SaveVoteLocally(requestID, value, salt, hash string) error {
	// TODO: Implement actual file I/O
	utils.PrintInfo(fmt.Sprintf("Vote data would be saved to: ~/.goo/votes/%s.json", requestID))
	return nil
}

// LoadVoteLocally loads vote data from local storage
func LoadVoteLocally(requestID string) (value, salt string, err error) {
	// TODO: Implement actual file I/O
	utils.PrintWarning("Vote data loading not yet implemented")
	return "3500", "random-salt-here", nil
}
