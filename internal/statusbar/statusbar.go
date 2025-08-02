package statusbar

import (
	"fmt"
	"strings"
	"time"

	"bbs/internal/config"
)

// StatusBar represents a terminal status bar
type StatusBar struct {
	username   string
	systemName string
	startTime  time.Time
	width      int
	isActive   bool
}

// New creates a new status bar instance
func New(username string, cfg *config.Config) *StatusBar {
	return &StatusBar{
		username:   username,
		systemName: cfg.BBS.SystemName,
		startTime:  time.Now(),
		width:      cfg.BBS.MaxLineLength,
		isActive:   true,
	}
}

// Render generates the status bar string with ANSI escape codes
func (sb *StatusBar) Render() string {
	if !sb.isActive {
		return ""
	}

	duration := time.Since(sb.startTime)
	durationStr := formatDuration(duration)

	// ANSI color codes
	const (
		blue         = "\033[44m" // Blue background
		white        = "\033[37m" // White text
		brightGreen  = "\033[92m" // Bright green text
		brightYellow = "\033[93m" // Bright yellow text
		reset        = "\033[0m"  // Reset all formatting
		clearLine    = "\033[2K"  // Clear entire line
	)

	// Calculate available space for each section
	leftSection := fmt.Sprintf(" %s", sb.username)
	rightSection := fmt.Sprintf("%s ", durationStr)
	centerSection := sb.systemName

	// Calculate padding for center alignment
	usedSpace := len(leftSection) + len(rightSection) + len(centerSection)
	if usedSpace >= sb.width {
		// Truncate if too long
		centerSection = truncateString(centerSection, sb.width-len(leftSection)-len(rightSection)-2)
		usedSpace = len(leftSection) + len(rightSection) + len(centerSection)
	}

	totalPadding := sb.width - usedSpace
	leftPadding := totalPadding / 2
	rightPadding := totalPadding - leftPadding

	// Build the status bar
	statusBar := fmt.Sprintf("%s%s%s%s%s%s%s%s%s%s%s",
		clearLine,          // Clear the line first
		blue,               // Blue background
		white, leftSection, // White username on left
		strings.Repeat(" ", leftPadding), // Left padding
		brightGreen, centerSection,       // Bright green system name in center
		strings.Repeat(" ", rightPadding), // Right padding
		brightYellow, rightSection,        // Bright yellow duration on right
		reset, // Reset formatting
	)

	return statusBar
}

// GetPositionCode returns the ANSI escape code to position cursor at bottom of screen
func (sb *StatusBar) GetPositionCode(terminalHeight int) string {
	return fmt.Sprintf("\033[%d;1H", terminalHeight)
}

// Update refreshes the status bar (useful for duration updates)
func (sb *StatusBar) Update() string {
	return sb.Render()
}

// SetActive enables or disables the status bar
func (sb *StatusBar) SetActive(active bool) {
	sb.isActive = active
}

// Clear returns ANSI codes to clear the status bar area
func (sb *StatusBar) Clear(terminalHeight int) string {
	return fmt.Sprintf("\033[%d;1H\033[2K", terminalHeight)
}

// formatDuration formats a duration into HH:MM:SS format
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
