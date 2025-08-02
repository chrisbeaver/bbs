package user_editor

import (
	"fmt"

	"bbs/internal/menu"
	"bbs/internal/modules"
)

// ListUsers displays all users
func (ue *UserEditor) ListUsers(writer modules.Writer, keyReader modules.KeyReader) bool {
	writer.Write([]byte(menu.ClearScreen))

	header := ue.colorScheme.Colorize("--- All Users ---", "primary")
	centeredHeader := ue.colorScheme.CenterText(header, 79)
	writer.Write([]byte(centeredHeader + "\n\n"))

	users, err := ue.db.GetAllUsers(1000)
	if err != nil {
		showMessage(writer, keyReader, ue.colorScheme, "Failed to retrieve users: "+err.Error(), "error")
		return true
	}

	if len(users) == 0 {
		showMessage(writer, keyReader, ue.colorScheme, "No users found.", "secondary")
		return true
	}

	// Header line
	headerLine := "ID   Username        Real Name            Level Calls  Status"
	coloredHeader := ue.colorScheme.Colorize(headerLine, "accent")
	centeredHeaderLine := ue.colorScheme.CenterText(coloredHeader, 79)
	writer.Write([]byte(centeredHeaderLine + "\n"))

	// Separator line
	separator := ue.colorScheme.DrawSeparator(len(headerLine), "─")
	centeredSeparator := ue.colorScheme.CenterText(separator, 79)
	writer.Write([]byte(centeredSeparator + "\n"))

	// Display users
	for _, user := range users {
		// Truncate real name if too long
		realName := user.RealName
		if len(realName) > 20 {
			realName = realName[:17] + "..."
		}

		status := "Active"
		if !user.IsActive {
			status = "Inactive"
		}

		line := fmt.Sprintf("%-4d %-15s %-20s %-5d %-8d %-6s",
			user.ID,
			user.Username,
			realName,
			user.AccessLevel,
			user.TotalCalls,
			status)

		coloredLine := ue.colorScheme.Colorize(line, "text")
		centeredLine := ue.colorScheme.CenterText(coloredLine, 79)
		writer.Write([]byte(centeredLine + "\n"))
	}

	writer.Write([]byte("\n"))
	prompt := ue.colorScheme.Colorize("Press any key to continue...", "text")
	centeredPrompt := ue.colorScheme.CenterText(prompt, 79)
	writer.Write([]byte(centeredPrompt))

	keyReader.ReadKey()
	return true
}
