package utils

import (
	"fmt"
	"strings"
	"time"
)

// FormatUgnot formats ugnot amount for display
func FormatUgnot(amount int64) string {
	return fmt.Sprintf("%d ugnot", amount)
}

// FormatBool formats boolean for Gno function calls
func FormatBool(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
	}
	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd%dh", days, hours)
}

// FormatTimestamp formats a Unix timestamp
func FormatTimestamp(ts int64) string {
	t := time.Unix(ts, 0)
	return t.Format("2006-01-02 15:04:05 MST")
}

// FormatTimeUntil formats time remaining until a timestamp
func FormatTimeUntil(ts int64) string {
	t := time.Unix(ts, 0)
	duration := time.Until(t)
	if duration < 0 {
		return "expired"
	}
	return FormatDuration(duration)
}

// TruncateString truncates a string to maxLen with ellipsis
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// FormatAddress formats a Gno address for display
func FormatAddress(addr string) string {
	if len(addr) <= 16 {
		return addr
	}
	return addr[:8] + "..." + addr[len(addr)-6:]
}

// PrintKeyValue prints a key-value pair with proper alignment
func PrintKeyValue(key string, value interface{}) {
	fmt.Printf("  %-20s %v\n", key+":", value)
}

// PrintSection prints a section header
func PrintSection(title string) {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("  %s\n", strings.ToUpper(title))
	fmt.Println(strings.Repeat("=", 60))
}

// PrintSuccess prints a success message
func PrintSuccess(message string) {
	fmt.Printf("✓ %s\n", message)
}

// PrintError prints an error message
func PrintError(message string) {
	fmt.Printf("✗ %s\n", message)
}

// PrintWarning prints a warning message
func PrintWarning(message string) {
	fmt.Printf("⚠ %s\n", message)
}

// PrintInfo prints an info message
func PrintInfo(message string) {
	fmt.Printf("ℹ %s\n", message)
}
