package sysop

import (
	"fmt"

	"bbs/internal/database"
	"bbs/internal/menu"
	"bbs/internal/modules"
	"bbs/internal/modules/sysop/user_editor"
)

// handleCreateUser creates a new user account using the user editor
func handleCreateUser(writer modules.Writer, keyReader modules.KeyReader, db *database.DB, colorScheme menu.ColorScheme) bool {
	editor := user_editor.NewUserEditor(db, colorScheme)
	return editor.CreateUser(writer, keyReader)
}

// handleEditUser edits an existing user account
func handleEditUser(writer modules.Writer, keyReader modules.KeyReader, db *database.DB, colorScheme menu.ColorScheme) bool {
	editor := user_editor.NewUserEditor(db, colorScheme)
	return editor.EditUser(writer, keyReader)
}

// handleDeleteUser deletes a user account
func handleDeleteUser(writer modules.Writer, keyReader modules.KeyReader, db *database.DB, colorScheme menu.ColorScheme) bool {
	editor := user_editor.NewUserEditor(db, colorScheme)
	return editor.DeleteUser(writer, keyReader)
}

// handleViewUsers displays all users
func handleViewUsers(writer modules.Writer, keyReader modules.KeyReader, db *database.DB, colorScheme menu.ColorScheme) bool {
	editor := user_editor.NewUserEditor(db, colorScheme)
	return editor.ListUsers(writer, keyReader)
}

// handleChangePassword changes a user's password
func handleChangePassword(writer modules.Writer, keyReader modules.KeyReader, db *database.DB, colorScheme menu.ColorScheme) bool {
	editor := user_editor.NewUserEditor(db, colorScheme)
	return editor.ChangePassword(writer, keyReader)
}

// handleToggleUserStatus toggles a user's active status
func handleToggleUserStatus(writer modules.Writer, keyReader modules.KeyReader, db *database.DB, colorScheme menu.ColorScheme) bool {
	editor := user_editor.NewUserEditor(db, colorScheme)
	return editor.ToggleUserStatus(writer, keyReader)
}

// handleSystemStats displays system statistics
func handleSystemStats(writer modules.Writer, keyReader modules.KeyReader, db *database.DB, colorScheme menu.ColorScheme) bool {
	writer.Write([]byte(menu.ClearScreen))

	header := colorScheme.Colorize("--- System Statistics ---", "primary")
	centeredHeader := colorScheme.CenterText(header, 79)
	writer.Write([]byte(centeredHeader + "\n"))

	separator := colorScheme.DrawSeparator(len("System Statistics"), "═")
	centeredSeparator := colorScheme.CenterText(separator, 79)
	writer.Write([]byte(centeredSeparator + "\n\n"))

	// Get users count
	users, err := db.GetAllUsers(1000)
	if err != nil {
		showMessage(writer, keyReader, colorScheme, "Error retrieving user statistics: "+err.Error(), "error")
		return true
	}

	// Get bulletins count
	bulletins, err := db.GetBulletins(1000)
	if err != nil {
		showMessage(writer, keyReader, colorScheme, "Error retrieving bulletin statistics: "+err.Error(), "error")
		return true
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
		coloredStat := colorScheme.Colorize(stat, "text")
		centeredStat := colorScheme.CenterText(coloredStat, 79)
		writer.Write([]byte(centeredStat + "\n"))
	}

	writer.Write([]byte("\n"))
	prompt := colorScheme.Colorize("Press any key to continue...", "text")
	centeredPrompt := colorScheme.CenterText(prompt, 79)
	writer.Write([]byte(centeredPrompt))

	keyReader.ReadKey()
	return true
}

// handleBulletinManagement manages bulletins (placeholder for now)
func handleBulletinManagement(writer modules.Writer, keyReader modules.KeyReader, db *database.DB, colorScheme menu.ColorScheme) bool {
	showMessage(writer, keyReader, colorScheme, "Bulletin Management - Not yet implemented", "secondary")
	return true
}

// Helper function for showing messages
func showMessage(writer modules.Writer, keyReader modules.KeyReader, colorScheme menu.ColorScheme, message, messageType string) {
	writer.Write([]byte(menu.ClearScreen))

	coloredMessage := colorScheme.Colorize(message, messageType)
	centeredMessage := colorScheme.CenterText(coloredMessage, 79)
	writer.Write([]byte(centeredMessage + "\n\n"))

	prompt := colorScheme.Colorize("Press any key to continue...", "text")
	centeredPrompt := colorScheme.CenterText(prompt, 79)
	writer.Write([]byte(centeredPrompt))

	keyReader.ReadKey()
}
