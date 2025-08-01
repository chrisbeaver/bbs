package sysop

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"bbs/internal/database"
	"bbs/internal/menu"
)

// handleCreateUser creates a new user account
func handleCreateUser(writer Writer, keyReader KeyReader, db *database.DB, colorScheme menu.ColorScheme) bool {
	writer.Write([]byte(menu.ClearScreen))

	header := colorScheme.Colorize("--- Create New User ---", "primary")
	centeredHeader := colorScheme.CenterText(header, 79)
	writer.Write([]byte(centeredHeader + "\n\n"))

	// Get username
	writer.Write([]byte(colorScheme.Colorize("Enter username: ", "text")))
	username, err := readLine(keyReader, writer)
	if err != nil || strings.TrimSpace(username) == "" {
		showMessage(writer, keyReader, colorScheme, "Operation cancelled.", "error")
		return true
	}

	// Check if user exists
	if _, err := db.GetUser(username); err == nil {
		showMessage(writer, keyReader, colorScheme, "User already exists!", "error")
		return true
	}

	// Get password
	writer.Write([]byte(colorScheme.Colorize("Enter password: ", "text")))
	password, err := readLine(keyReader, writer)
	if err != nil || strings.TrimSpace(password) == "" {
		showMessage(writer, keyReader, colorScheme, "Operation cancelled.", "error")
		return true
	}

	// Get access level
	writer.Write([]byte(colorScheme.Colorize("Enter access level (0-255, default 10): ", "text")))
	accessLevelStr, err := readLine(keyReader, writer)
	if err != nil {
		showMessage(writer, keyReader, colorScheme, "Operation cancelled.", "error")
		return true
	}

	accessLevel := 10 // Default access level
	if strings.TrimSpace(accessLevelStr) != "" {
		if level, err := parseAccessLevel(accessLevelStr); err == nil {
			accessLevel = level
		} else {
			showMessage(writer, keyReader, colorScheme, "Invalid access level, using default (10).", "secondary")
		}
	}

	// Create user
	user := &database.User{
		Username:    strings.TrimSpace(username),
		Password:    strings.TrimSpace(password), // TODO: Hash password
		AccessLevel: accessLevel,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}

	if err := db.CreateUser(user); err != nil {
		showMessage(writer, keyReader, colorScheme, "Failed to create user: "+err.Error(), "error")
		return true
	}

	showMessage(writer, keyReader, colorScheme, "User created successfully!", "primary")
	return true
}

// handleEditUser edits an existing user account
func handleEditUser(writer Writer, keyReader KeyReader, db *database.DB, colorScheme menu.ColorScheme) bool {
	writer.Write([]byte(menu.ClearScreen))

	header := colorScheme.Colorize("--- Edit User Account ---", "primary")
	centeredHeader := colorScheme.CenterText(header, 79)
	writer.Write([]byte(centeredHeader + "\n\n"))

	// Get username to edit
	writer.Write([]byte(colorScheme.Colorize("Enter username to edit: ", "text")))
	username, err := readLine(keyReader, writer)
	if err != nil || strings.TrimSpace(username) == "" {
		showMessage(writer, keyReader, colorScheme, "Operation cancelled.", "error")
		return true
	}

	// Get user
	user, err := db.GetUser(strings.TrimSpace(username))
	if err != nil {
		showMessage(writer, keyReader, colorScheme, "User not found!", "error")
		return true
	}

	// Show current user info
	writer.Write([]byte(menu.ClearScreen))
	writer.Write([]byte(centeredHeader + "\n\n"))

	info := fmt.Sprintf("Current user: %s (Access Level: %d, Active: %v)",
		user.Username, user.AccessLevel, user.IsActive)
	centeredInfo := colorScheme.CenterText(colorScheme.Colorize(info, "secondary"), 79)
	writer.Write([]byte(centeredInfo + "\n\n"))

	// Get new password (optional)
	writer.Write([]byte(colorScheme.Colorize("New password (press Enter to keep current): ", "text")))
	newPassword, err := readLine(keyReader, writer)
	if err != nil {
		showMessage(writer, keyReader, colorScheme, "Operation cancelled.", "error")
		return true
	}

	// Get new access level (optional)
	currentLevelStr := fmt.Sprintf("New access level (current: %d, press Enter to keep): ", user.AccessLevel)
	writer.Write([]byte(colorScheme.Colorize(currentLevelStr, "text")))
	accessLevelStr, err := readLine(keyReader, writer)
	if err != nil {
		showMessage(writer, keyReader, colorScheme, "Operation cancelled.", "error")
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
			showMessage(writer, keyReader, colorScheme, "Invalid access level, keeping current value.", "secondary")
		}
	}

	if err := db.UpdateUser(user.ID, user.Username, user.Password, user.RealName, user.Email, user.AccessLevel, user.IsActive); err != nil {
		showMessage(writer, keyReader, colorScheme, "Failed to update user: "+err.Error(), "error")
		return true
	}

	showMessage(writer, keyReader, colorScheme, "User updated successfully!", "primary")
	return true
}

// handleDeleteUser deletes a user account
func handleDeleteUser(writer Writer, keyReader KeyReader, db *database.DB, colorScheme menu.ColorScheme) bool {
	writer.Write([]byte(menu.ClearScreen))

	header := colorScheme.Colorize("--- Delete User Account ---", "primary")
	centeredHeader := colorScheme.CenterText(header, 79)
	writer.Write([]byte(centeredHeader + "\n\n"))

	// Get username to delete
	writer.Write([]byte(colorScheme.Colorize("Enter username to delete: ", "text")))
	username, err := readLine(keyReader, writer)
	if err != nil || strings.TrimSpace(username) == "" {
		showMessage(writer, keyReader, colorScheme, "Operation cancelled.", "error")
		return true
	}

	// Get user to get ID
	user, err := db.GetUser(strings.TrimSpace(username))
	if err != nil {
		showMessage(writer, keyReader, colorScheme, "User not found!", "error")
		return true
	}

	// Confirm deletion
	confirmMsg := fmt.Sprintf("Are you sure you want to delete user '%s'? (y/N): ", user.Username)
	writer.Write([]byte(colorScheme.Colorize(confirmMsg, "text")))
	confirm, err := readLine(keyReader, writer)
	if err != nil || strings.ToLower(strings.TrimSpace(confirm)) != "y" {
		showMessage(writer, keyReader, colorScheme, "Operation cancelled.", "error")
		return true
	}

	// Delete user
	if err := db.DeleteUser(user.ID); err != nil {
		showMessage(writer, keyReader, colorScheme, "Failed to delete user: "+err.Error(), "error")
		return true
	}

	showMessage(writer, keyReader, colorScheme, "User deleted successfully!", "primary")
	return true
}

// handleViewUsers displays all users
func handleViewUsers(writer Writer, keyReader KeyReader, db *database.DB, colorScheme menu.ColorScheme) bool {
	writer.Write([]byte(menu.ClearScreen))

	header := colorScheme.Colorize("--- All Users ---", "primary")
	centeredHeader := colorScheme.CenterText(header, 79)
	writer.Write([]byte(centeredHeader + "\n\n"))

	users, err := db.GetAllUsers(1000)
	if err != nil {
		showMessage(writer, keyReader, colorScheme, "Failed to retrieve users: "+err.Error(), "error")
		return true
	}

	if len(users) == 0 {
		showMessage(writer, keyReader, colorScheme, "No users found.", "secondary")
		return true
	}

	// Display users
	for i, user := range users {
		status := "Active"
		if !user.IsActive {
			status = "Inactive"
		}
		userInfo := fmt.Sprintf("%d) %s (Level: %d, %s) - Last: %s",
			i+1, user.Username, user.AccessLevel, status,
			func() string {
				if user.LastCall != nil {
					return user.LastCall.Format("2006-01-02")
				}
				return "Never"
			}())
		coloredInfo := colorScheme.Colorize(userInfo, "text")
		centeredInfo := colorScheme.CenterText(coloredInfo, 79)
		writer.Write([]byte(centeredInfo + "\n"))
	}

	writer.Write([]byte("\n"))
	prompt := colorScheme.Colorize("Press any key to continue...", "text")
	centeredPrompt := colorScheme.CenterText(prompt, 79)
	writer.Write([]byte(centeredPrompt))

	keyReader.ReadKey()
	return true
}

// handleChangePassword changes a user's password
func handleChangePassword(writer Writer, keyReader KeyReader, db *database.DB, colorScheme menu.ColorScheme) bool {
	writer.Write([]byte(menu.ClearScreen))

	header := colorScheme.Colorize("--- Change User Password ---", "primary")
	centeredHeader := colorScheme.CenterText(header, 79)
	writer.Write([]byte(centeredHeader + "\n\n"))

	// Get username
	writer.Write([]byte(colorScheme.Colorize("Enter username: ", "text")))
	username, err := readLine(keyReader, writer)
	if err != nil || strings.TrimSpace(username) == "" {
		showMessage(writer, keyReader, colorScheme, "Operation cancelled.", "error")
		return true
	}

	// Get user
	user, err := db.GetUser(strings.TrimSpace(username))
	if err != nil {
		showMessage(writer, keyReader, colorScheme, "User not found!", "error")
		return true
	}

	// Get new password
	writer.Write([]byte(colorScheme.Colorize("Enter new password: ", "text")))
	newPassword, err := readLine(keyReader, writer)
	if err != nil || strings.TrimSpace(newPassword) == "" {
		showMessage(writer, keyReader, colorScheme, "Operation cancelled.", "error")
		return true
	}

	// Update password
	user.Password = strings.TrimSpace(newPassword) // TODO: Hash password
	if err := db.UpdateUser(user.ID, user.Username, user.Password, user.RealName, user.Email, user.AccessLevel, user.IsActive); err != nil {
		showMessage(writer, keyReader, colorScheme, "Failed to update password: "+err.Error(), "error")
		return true
	}

	showMessage(writer, keyReader, colorScheme, "Password updated successfully!", "primary")
	return true
}

// handleToggleUserStatus toggles a user's active status
func handleToggleUserStatus(writer Writer, keyReader KeyReader, db *database.DB, colorScheme menu.ColorScheme) bool {
	writer.Write([]byte(menu.ClearScreen))

	header := colorScheme.Colorize("--- Toggle User Status ---", "primary")
	centeredHeader := colorScheme.CenterText(header, 79)
	writer.Write([]byte(centeredHeader + "\n\n"))

	// Get username
	writer.Write([]byte(colorScheme.Colorize("Enter username: ", "text")))
	username, err := readLine(keyReader, writer)
	if err != nil || strings.TrimSpace(username) == "" {
		showMessage(writer, keyReader, colorScheme, "Operation cancelled.", "error")
		return true
	}

	// Get user
	user, err := db.GetUser(strings.TrimSpace(username))
	if err != nil {
		showMessage(writer, keyReader, colorScheme, "User not found!", "error")
		return true
	}

	// Toggle status
	user.IsActive = !user.IsActive
	if err := db.UpdateUser(user.ID, user.Username, user.Password, user.RealName, user.Email, user.AccessLevel, user.IsActive); err != nil {
		showMessage(writer, keyReader, colorScheme, "Failed to update user status: "+err.Error(), "error")
		return true
	}

	status := "activated"
	if !user.IsActive {
		status = "deactivated"
	}

	message := fmt.Sprintf("User %s %s successfully!", user.Username, status)
	showMessage(writer, keyReader, colorScheme, message, "primary")
	return true
}

// handleSystemStats displays system statistics
func handleSystemStats(writer Writer, keyReader KeyReader, db *database.DB, colorScheme menu.ColorScheme) bool {
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
func handleBulletinManagement(writer Writer, keyReader KeyReader, db *database.DB, colorScheme menu.ColorScheme) bool {
	showMessage(writer, keyReader, colorScheme, "Bulletin Management - Not yet implemented", "secondary")
	return true
}

// Helper functions

// readLine reads a line of input from the user
func readLine(keyReader KeyReader, writer Writer) (string, error) {
	var line strings.Builder
	for {
		key, err := keyReader.ReadKey()
		if err != nil {
			return "", err
		}

		switch key {
		case "enter":
			writer.Write([]byte("\n"))
			return line.String(), nil
		case "backspace":
			if line.Len() > 0 {
				str := line.String()
				line.Reset()
				line.WriteString(str[:len(str)-1])
				writer.Write([]byte("\b \b")) // Backspace, space, backspace
			}
		case "escape", "ctrl+c":
			return "", fmt.Errorf("cancelled")
		default:
			if len(key) == 1 && key[0] >= 32 && key[0] <= 126 { // Printable ASCII
				line.WriteString(key)
				writer.Write([]byte(key)) // Echo the character
			}
		}
	}
}

// parseAccessLevel parses an access level string
func parseAccessLevel(s string) (int, error) {
	level, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0, fmt.Errorf("invalid access level")
	}
	if level < 0 || level > 255 {
		return 0, fmt.Errorf("access level must be 0-255")
	}
	return level, nil
}

// showMessage displays a message and waits for user input
func showMessage(writer Writer, keyReader KeyReader, colorScheme menu.ColorScheme, message, messageType string) {
	writer.Write([]byte(menu.ClearScreen))

	coloredMessage := colorScheme.Colorize(message, messageType)
	centeredMessage := colorScheme.CenterText(coloredMessage, 79)
	writer.Write([]byte(centeredMessage + "\n\n"))

	prompt := colorScheme.Colorize("Press any key to continue...", "text")
	centeredPrompt := colorScheme.CenterText(prompt, 79)
	writer.Write([]byte(centeredPrompt))

	keyReader.ReadKey()
}
