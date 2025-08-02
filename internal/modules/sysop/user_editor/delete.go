package user_editor

import (
	"fmt"
	"strings"

	"bbs/internal/menu"
	"bbs/internal/modules"
)

// DeleteUser deletes a user account
func (ue *UserEditor) DeleteUser(writer modules.Writer, keyReader modules.KeyReader) bool {
	writer.Write([]byte(menu.ClearScreen))

	header := ue.colorScheme.Colorize("--- Delete User Account ---", "primary")
	centeredHeader := ue.colorScheme.CenterText(header, 79)
	writer.Write([]byte(centeredHeader + "\n\n"))

	// Get username to delete
	writer.Write([]byte(ue.colorScheme.Colorize("Enter username to delete: ", "text")))
	username, err := readLine(keyReader, writer)
	if err != nil || strings.TrimSpace(username) == "" {
		showMessage(writer, keyReader, ue.colorScheme, "Operation cancelled.", "error")
		return true
	}

	// Get user to get ID
	user, err := ue.db.GetUser(strings.TrimSpace(username))
	if err != nil {
		showMessage(writer, keyReader, ue.colorScheme, "User not found!", "error")
		return true
	}

	// Confirm deletion
	confirmMsg := fmt.Sprintf("Are you sure you want to delete user '%s'? (y/N): ", user.Username)
	writer.Write([]byte(ue.colorScheme.Colorize(confirmMsg, "text")))
	confirm, err := readLine(keyReader, writer)
	if err != nil || strings.ToLower(strings.TrimSpace(confirm)) != "y" {
		showMessage(writer, keyReader, ue.colorScheme, "Operation cancelled.", "error")
		return true
	}

	// Delete user
	if err := ue.db.DeleteUser(user.ID); err != nil {
		showMessage(writer, keyReader, ue.colorScheme, "Failed to delete user: "+err.Error(), "error")
		return true
	}

	showMessage(writer, keyReader, ue.colorScheme, "User deleted successfully!", "primary")
	return true
}
