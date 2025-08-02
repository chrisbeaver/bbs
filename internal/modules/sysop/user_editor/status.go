package user_editor

import (
	"fmt"
	"strings"

	"bbs/internal/menu"
	"bbs/internal/modules"
)

// ToggleUserStatus toggles a user's active status
func (ue *UserEditor) ToggleUserStatus(writer modules.Writer, keyReader modules.KeyReader) bool {
	writer.Write([]byte(menu.ClearScreen))

	header := ue.colorScheme.Colorize("--- Toggle User Status ---", "primary")
	centeredHeader := ue.colorScheme.CenterText(header, 79)
	writer.Write([]byte(centeredHeader + "\n\n"))

	// Get username
	writer.Write([]byte(ue.colorScheme.Colorize("Enter username: ", "text")))
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

	// Toggle status
	user.IsActive = !user.IsActive
	if err := ue.db.UpdateUser(user.ID, user.Username, user.Password, user.RealName, user.Email, user.AccessLevel, user.IsActive); err != nil {
		showMessage(writer, keyReader, ue.colorScheme, "Failed to update user status: "+err.Error(), "error")
		return true
	}

	status := "activated"
	if !user.IsActive {
		status = "deactivated"
	}

	message := fmt.Sprintf("User %s %s successfully!", user.Username, status)
	showMessage(writer, keyReader, ue.colorScheme, message, "primary")
	return true
}
