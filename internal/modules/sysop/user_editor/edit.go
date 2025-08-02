package user_editor

import (
	"fmt"
	"strings"

	"bbs/internal/menu"
	"bbs/internal/modules"
)

// EditUser edits an existing user account
func (ue *UserEditor) EditUser(writer modules.Writer, keyReader modules.KeyReader) bool {
	writer.Write([]byte(menu.ClearScreen))

	header := ue.colorScheme.Colorize("--- Edit User Account ---", "primary")
	centeredHeader := ue.colorScheme.CenterText(header, 79)
	writer.Write([]byte(centeredHeader + "\n\n"))

	// Get username to edit
	writer.Write([]byte(ue.colorScheme.Colorize("Enter username to edit: ", "text")))
	username, err := readLine(keyReader, writer)
	if err != nil || strings.TrimSpace(username) == "" {
		showMessage(writer, keyReader, ue.colorScheme, "Operation cancelled.", "error")
		return true
	}

	// Get user
	user, err := ue.db.GetUser(strings.TrimSpace(username))
	if err != nil {
		showMessage(writer, keyReader, ue.colorScheme, "User not found!", "error")
		return true
	}

	// Show current user info
	writer.Write([]byte(menu.ClearScreen))
	writer.Write([]byte(centeredHeader + "\n\n"))

	info := fmt.Sprintf("Current user: %s (Access Level: %d, Active: %v)",
		user.Username, user.AccessLevel, user.IsActive)
	centeredInfo := ue.colorScheme.CenterText(ue.colorScheme.Colorize(info, "secondary"), 79)
	writer.Write([]byte(centeredInfo + "\n\n"))

	// Get new password (optional)
	writer.Write([]byte(ue.colorScheme.Colorize("New password (press Enter to keep current): ", "text")))
	newPassword, err := readLine(keyReader, writer)
	if err != nil {
		showMessage(writer, keyReader, ue.colorScheme, "Operation cancelled.", "error")
		return true
	}

	// Get new access level (optional)
	currentLevelStr := fmt.Sprintf("New access level (current: %d, press Enter to keep): ", user.AccessLevel)
	writer.Write([]byte(ue.colorScheme.Colorize(currentLevelStr, "text")))
	accessLevelStr, err := readLine(keyReader, writer)
	if err != nil {
		showMessage(writer, keyReader, ue.colorScheme, "Operation cancelled.", "error")
		return true
	}

	// Update user
	if strings.TrimSpace(newPassword) != "" {
		user.Password = strings.TrimSpace(newPassword) // TODO: Hash password
	}

	if strings.TrimSpace(accessLevelStr) != "" {
		if level, err := parseAccessLevel(accessLevelStr); err == nil {
			user.AccessLevel = level
		} else {
			showMessage(writer, keyReader, ue.colorScheme, "Invalid access level, keeping current value.", "secondary")
		}
	}

	if err := ue.db.UpdateUser(user.ID, user.Username, user.Password, user.RealName, user.Email, user.AccessLevel, user.IsActive); err != nil {
		showMessage(writer, keyReader, ue.colorScheme, "Failed to update user: "+err.Error(), "error")
		return true
	}

	showMessage(writer, keyReader, ue.colorScheme, "User updated successfully!", "primary")
	return true
}
