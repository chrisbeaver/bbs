package user_editor

import (
	"strings"

	"bbs/internal/menu"
	"bbs/internal/modules"
)

// ChangePassword changes a user's password
func (ue *UserEditor) ChangePassword(writer modules.Writer, keyReader modules.KeyReader) bool {
	writer.Write([]byte(menu.ClearScreen))

	header := ue.colorScheme.Colorize("--- Change User Password ---", "primary")
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

	// Get new password
	writer.Write([]byte(ue.colorScheme.Colorize("Enter new password: ", "text")))
	newPassword, err := readLine(keyReader, writer)
	if err != nil || strings.TrimSpace(newPassword) == "" {
		showMessage(writer, keyReader, ue.colorScheme, "Operation cancelled.", "error")
		return true
	}

	// Update password
	user.Password = strings.TrimSpace(newPassword) // TODO: Hash password
	if err := ue.db.UpdateUser(user.ID, user.Username, user.Password, user.RealName, user.Email, user.AccessLevel, user.IsActive); err != nil {
		showMessage(writer, keyReader, ue.colorScheme, "Failed to update password: "+err.Error(), "error")
		return true
	}

	showMessage(writer, keyReader, ue.colorScheme, "Password updated successfully!", "primary")
	return true
}
