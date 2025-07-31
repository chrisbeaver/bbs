package sysop

import (
	"fmt"
	"strings"

	"golang.org/x/term"

	"bbs/internal/database"
)

// SysopMenu implements the main sysop menu functionality
type SysopMenu struct {
	db          *database.DB
	colorScheme ColorScheme
}

// NewSysopMenu creates a new sysop menu
func NewSysopMenu(db *database.DB, colorScheme ColorScheme) *SysopMenu {
	return &SysopMenu{
		db:          db,
		colorScheme: colorScheme,
	}
}

// Execute runs the sysop menu
func (sm *SysopMenu) Execute(term *term.Terminal) bool {
	for {
		// Clear screen and show menu
		term.Write([]byte(ClearScreen + HideCursor))

		// Header
		header := sm.colorScheme.Colorize("System Operator Menu", "primary")
		centeredHeader := sm.colorScheme.CenterText(header, 79)
		term.Write([]byte(centeredHeader + "\n"))

		separator := sm.colorScheme.DrawSeparator(len("System Operator Menu"), "═")
		centeredSeparator := sm.colorScheme.CenterText(separator, 79)
		term.Write([]byte(centeredSeparator + "\n\n"))

		// Menu options
		options := []string{
			"1) User Management",
			"2) Bulletin Management",
			"3) System Statistics",
			"4) System Configuration",
			"5) Database Maintenance",
			"Q) Return to main menu",
		}

		for _, option := range options {
			coloredOption := sm.colorScheme.Colorize(option, "text")
			centeredOption := sm.colorScheme.CenterText(coloredOption, 79)
			term.Write([]byte(centeredOption + "\n"))
		}

		// Prompt
		prompt := sm.colorScheme.Colorize("\nEnter your choice: ", "accent")
		centeredPrompt := sm.colorScheme.CenterText(prompt, 79)
		term.Write([]byte(centeredPrompt))
		term.Write([]byte(ShowCursor))

		// Get input
		input, err := term.ReadLine()
		if err != nil {
			return true
		}

		term.Write([]byte(HideCursor))

		switch strings.ToLower(strings.TrimSpace(input)) {
		case "1":
			// User Management
			userEditor := NewUserEditor(sm.db, sm.colorScheme)
			userEditor.Execute(term)
		case "2":
			// Bulletin Management
			bulletinEditor := NewBulletinEditor(sm.db, sm.colorScheme)
			bulletinEditor.Execute(term)
		case "3":
			sm.showSystemStats(term)
		case "4":
			sm.showMessage(term, "System configuration not yet implemented.", "secondary")
		case "5":
			sm.showMessage(term, "Database maintenance not yet implemented.", "secondary")
		case "q", "quit":
			term.Write([]byte(ShowCursor))
			return true
		default:
			sm.showMessage(term, "Invalid choice. Press any key to continue...", "error")
		}
	}
}

// showSystemStats displays basic system statistics
func (sm *SysopMenu) showSystemStats(term *term.Terminal) {
	term.Write([]byte(ClearScreen))

	header := sm.colorScheme.Colorize("System Statistics", "primary")
	centeredHeader := sm.colorScheme.CenterText(header, 79)
	term.Write([]byte(centeredHeader + "\n"))

	separator := sm.colorScheme.DrawSeparator(len("System Statistics"), "═")
	centeredSeparator := sm.colorScheme.CenterText(separator, 79)
	term.Write([]byte(centeredSeparator + "\n\n"))

	// Get users count
	users, err := sm.db.GetAllUsers(1000)
	if err != nil {
		sm.showMessage(term, "Error retrieving user statistics: "+err.Error(), "error")
		return
	}

	// Get bulletins count
	bulletins, err := sm.db.GetBulletins(1000)
	if err != nil {
		sm.showMessage(term, "Error retrieving bulletin statistics: "+err.Error(), "error")
		return
	}

	// Count active users
	activeUsers := 0
	totalCalls := 0
	for _, user := range users {
		if user.IsActive {
			activeUsers++
		}
		totalCalls += user.TotalCalls
	}

	// Display statistics
	stats := []string{
		"Total Users: " + fmt.Sprintf("%d", len(users)),
		"Active Users: " + fmt.Sprintf("%d", activeUsers),
		"Inactive Users: " + fmt.Sprintf("%d", len(users)-activeUsers),
		"Total Bulletins: " + fmt.Sprintf("%d", len(bulletins)),
		"Total System Calls: " + fmt.Sprintf("%d", totalCalls),
	}

	for _, stat := range stats {
		coloredStat := sm.colorScheme.Colorize(stat, "text")
		centeredStat := sm.colorScheme.CenterText(coloredStat, 79)
		term.Write([]byte(centeredStat + "\n"))
	}

	sm.showMessage(term, "\nPress any key to continue...", "text")
}

// showMessage displays a message and waits for key press
func (sm *SysopMenu) showMessage(term *term.Terminal, message, colorType string) {
	term.Write([]byte(HideCursor))

	coloredMsg := sm.colorScheme.Colorize(message, colorType)
	centeredMsg := sm.colorScheme.CenterText(coloredMsg, 79)
	term.Write([]byte("\n" + centeredMsg))

	term.Write([]byte(ShowCursor))
	term.ReadLine()
	term.Write([]byte(HideCursor))
}
