package sysop

import (
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/term"

	"bbs/internal/components"
	"bbs/internal/database"
)

// UserEditor implements the sysop user management functionality
type UserEditor struct {
	db          *database.DB
	colorScheme ColorScheme
}

// NewUserEditor creates a new sysop user editor
func NewUserEditor(db *database.DB, colorScheme ColorScheme) *UserEditor {
	return &UserEditor{
		db:          db,
		colorScheme: colorScheme,
	}
}

// Execute runs the user editor
func (ue *UserEditor) Execute(term *term.Terminal) bool {
	for {
		// Clear screen and show menu
		term.Write([]byte(ClearScreen + HideCursor))

		// Header
		header := ue.colorScheme.Colorize("Sysop User Editor", "primary")
		centeredHeader := ue.colorScheme.CenterText(header, 79)
		term.Write([]byte(centeredHeader + "\n"))

		separator := ue.colorScheme.DrawSeparator(len("Sysop User Editor"), "═")
		centeredSeparator := ue.colorScheme.CenterText(separator, 79)
		term.Write([]byte(centeredSeparator + "\n\n"))

		// Menu options
		options := []string{
			"1) List all users",
			"2) Create new user",
			"3) Edit existing user",
			"4) Delete user",
			"5) Change user password",
			"6) Toggle user active status",
			"Q) Return to sysop menu",
		}

		for _, option := range options {
			coloredOption := ue.colorScheme.Colorize(option, "text")
			centeredOption := ue.colorScheme.CenterText(coloredOption, 79)
			term.Write([]byte(centeredOption + "\n"))
		}

		// Prompt
		prompt := ue.colorScheme.Colorize("\nEnter your choice: ", "accent")
		centeredPrompt := ue.colorScheme.CenterText(prompt, 79)
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
			ue.ListUsers(term)
		case "2":
			ue.CreateUser(term)
		case "3":
			ue.EditUser(term)
		case "4":
			ue.DeleteUser(term)
		case "5":
			ue.ChangePassword(term)
		case "6":
			ue.ToggleUserStatus(term)
		case "q", "quit":
			term.Write([]byte(ShowCursor))
			return true
		default:
			ue.showMessage(term, "Invalid choice. Press any key to continue...", "error")
		}
	}
}

// ListUsers displays a list of all users
func (ue *UserEditor) ListUsers(term *term.Terminal) {
	users, err := ue.db.GetAllUsers(100)
	if err != nil {
		ue.showMessage(term, "Error retrieving users: "+err.Error(), "error")
		return
	}

	term.Write([]byte(ClearScreen))

	header := ue.colorScheme.Colorize("All Users", "primary")
	centeredHeader := ue.colorScheme.CenterText(header, 79)
	term.Write([]byte(centeredHeader + "\n"))

	separator := ue.colorScheme.DrawSeparator(len("All Users"), "═")
	centeredSeparator := ue.colorScheme.CenterText(separator, 79)
	term.Write([]byte(centeredSeparator + "\n\n"))

	if len(users) == 0 {
		msg := ue.colorScheme.Colorize("No users found.", "secondary")
		centeredMsg := ue.colorScheme.CenterText(msg, 79)
		term.Write([]byte(centeredMsg + "\n"))
	} else {
		// Header line
		headerLine := fmt.Sprintf("%-4s %-15s %-20s %-5s %-8s %-6s", "ID", "Username", "Real Name", "Level", "Calls", "Active")
		coloredHeaderLine := ue.colorScheme.Colorize(headerLine, "accent")
		centeredHeaderLine := ue.colorScheme.CenterText(coloredHeaderLine, 79)
		term.Write([]byte(centeredHeaderLine + "\n"))

		// Separator line
		sepLine := strings.Repeat("-", 60)
		coloredSepLine := ue.colorScheme.Colorize(sepLine, "secondary")
		centeredSepLine := ue.colorScheme.CenterText(coloredSepLine, 79)
		term.Write([]byte(centeredSepLine + "\n"))

		for _, user := range users {
			status := "Yes"
			if !user.IsActive {
				status = "No"
			}

			// Truncate real name if too long
			realName := user.RealName
			if len(realName) > 20 {
				realName = realName[:17] + "..."
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
			term.Write([]byte(centeredLine + "\n"))
		}
	}

	ue.showMessage(term, "\nPress any key to continue...", "text")
}

// CreateUser creates a new user using form components
func (ue *UserEditor) CreateUser(term *term.Terminal) {
	// Create the form
	form := components.NewForm(components.FormConfig{
		Title: "Create New User",
		Width: 79,
	}, ue.colorScheme)

	// Add username field
	usernameField := components.NewTextInput(components.TextInputConfig{
		Name:        "username",
		Label:       "Username",
		Placeholder: "Enter username...",
		MaxLength:   32,
		Required:    true,
		Width:       40,
		Validator: func(value string) error {
			trimmed := strings.TrimSpace(value)
			if len(trimmed) < 3 {
				return fmt.Errorf("username must be at least 3 characters")
			}
			// Check if user already exists
			existingUser, _ := ue.db.GetUser(trimmed)
			if existingUser != nil {
				return fmt.Errorf("user already exists")
			}
			return nil
		},
	}, ue.colorScheme)

	// Add password field
	passwordField := components.NewTextInput(components.TextInputConfig{
		Name:        "password",
		Label:       "Password",
		Placeholder: "Enter password...",
		MaxLength:   64,
		Required:    true,
		Width:       40,
		Validator: func(value string) error {
			if len(strings.TrimSpace(value)) < 6 {
				return fmt.Errorf("password must be at least 6 characters")
			}
			return nil
		},
	}, ue.colorScheme)

	// Add real name field
	realNameField := components.NewTextInput(components.TextInputConfig{
		Name:        "real_name",
		Label:       "Real Name",
		Placeholder: "Enter real name (optional)...",
		MaxLength:   64,
		Required:    false,
		Width:       40,
	}, ue.colorScheme)

	// Add email field
	emailField := components.NewTextInput(components.TextInputConfig{
		Name:        "email",
		Label:       "Email",
		Placeholder: "Enter email (optional)...",
		MaxLength:   128,
		Required:    false,
		Width:       40,
	}, ue.colorScheme)

	// Add access level field
	accessLevelField := components.NewTextInput(components.TextInputConfig{
		Name:        "access_level",
		Label:       "Access Level",
		Placeholder: "0-255",
		Value:       "0", // Default to 0
		MaxLength:   3,
		Required:    true,
		Width:       40,
		Validator: func(value string) error {
			accessLevel, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil || accessLevel < 0 || accessLevel > 255 {
				return fmt.Errorf("access level must be 0-255")
			}
			return nil
		},
	}, ue.colorScheme)

	// Add components to form
	form.AddComponent(usernameField)
	form.AddComponent(passwordField)
	form.AddComponent(realNameField)
	form.AddComponent(emailField)
	form.AddComponent(accessLevelField)

	// Start form interaction
	form.Start()

	// Simple character input loop
	for {
		// Render form
		term.Write([]byte(form.Render()))

		// Read single character input using a simple approach
		term.Write([]byte("\n" + ue.colorScheme.Colorize("Controls: [t]ab, [c]lear, [b]ackspace, [s]ubmit, [q]uit, or type text: ", "secondary")))
		input, err := term.ReadLine()
		if err != nil {
			break
		}

		input = strings.TrimSpace(input)

		// Handle special commands first
		if len(input) == 1 {
			switch strings.ToLower(input) {
			case "t":
				form.HandleKey('\t')
				continue
			case "b":
				form.HandleKey('\b')
				continue
			case "s":
				form.HandleKey('\r')
			case "q":
				form.HandleKey(27)
			default:
				// Single character input to focused field
				form.HandleKey(rune(input[0]))
				continue
			}
		} else if input == "" {
			// Empty input = submit
			form.HandleKey('\r')
		} else {
			// Multi-character input - add each character to focused field
			for _, char := range input {
				form.HandleKey(char)
			}
			continue
		}

		// Check form state
		if form.IsSubmitted() {
			errors := form.Validate()
			if len(errors) == 0 {
				// Form is valid, create user
				values := form.GetStringValues()
				accessLevel, _ := strconv.Atoi(values["access_level"])

				user := &database.User{
					Username:    strings.TrimSpace(values["username"]),
					Password:    strings.TrimSpace(values["password"]), // TODO: Hash password
					RealName:    strings.TrimSpace(values["real_name"]),
					Email:       strings.TrimSpace(values["email"]),
					AccessLevel: accessLevel,
					IsActive:    true,
				}

				if err := ue.db.CreateUser(user); err != nil {
					ue.showMessage(term, "Error creating user: "+err.Error(), "error")
				} else {
					ue.showMessage(term, "User created successfully!", "success")
				}
				break
			} else {
				// Show validation errors
				errorMsg := "Validation errors:\n"
				for _, err := range errors {
					errorMsg += "• " + err.Error() + "\n"
				}
				ue.showMessage(term, errorMsg, "error")
				form.Reset()
				form.Start()
			}
		}

		if form.IsCancelled() {
			break
		}
	}

	form.Reset()
}

// EditUser edits an existing user
func (ue *UserEditor) EditUser(term *term.Terminal) {
	term.Write([]byte(ClearScreen + ShowCursor))

	term.Write([]byte(ue.colorScheme.Colorize("Edit User\n\n", "primary")))
	term.Write([]byte(ue.colorScheme.Colorize("Enter user ID to edit: ", "text")))

	idStr, err := term.ReadLine()
	if err != nil {
		return
	}

	id, err := strconv.Atoi(strings.TrimSpace(idStr))
	if err != nil {
		ue.showMessage(term, "Invalid ID format.", "error")
		return
	}

	// Get existing user
	user, err := ue.db.GetUserByID(id)
	if err != nil {
		ue.showMessage(term, "User not found.", "error")
		return
	}

	// Show current values and get new ones
	term.Write([]byte(ue.colorScheme.Colorize(fmt.Sprintf("Current username: %s\n", user.Username), "secondary")))
	term.Write([]byte(ue.colorScheme.Colorize("New username (or press Enter to keep current): ", "text")))
	newUsername, err := term.ReadLine()
	if err != nil {
		return
	}
	if strings.TrimSpace(newUsername) == "" {
		newUsername = user.Username
	}

	term.Write([]byte(ue.colorScheme.Colorize(fmt.Sprintf("Current real name: %s\n", user.RealName), "secondary")))
	term.Write([]byte(ue.colorScheme.Colorize("New real name (or press Enter to keep current): ", "text")))
	newRealName, err := term.ReadLine()
	if err != nil {
		return
	}
	if strings.TrimSpace(newRealName) == "" {
		newRealName = user.RealName
	}

	term.Write([]byte(ue.colorScheme.Colorize(fmt.Sprintf("Current email: %s\n", user.Email), "secondary")))
	term.Write([]byte(ue.colorScheme.Colorize("New email (or press Enter to keep current): ", "text")))
	newEmail, err := term.ReadLine()
	if err != nil {
		return
	}
	if strings.TrimSpace(newEmail) == "" {
		newEmail = user.Email
	}

	term.Write([]byte(ue.colorScheme.Colorize(fmt.Sprintf("Current access level: %d\n", user.AccessLevel), "secondary")))
	term.Write([]byte(ue.colorScheme.Colorize("New access level (or press Enter to keep current): ", "text")))
	newAccessLevelStr, err := term.ReadLine()
	if err != nil {
		return
	}

	newAccessLevel := user.AccessLevel
	if strings.TrimSpace(newAccessLevelStr) != "" {
		newAccessLevel, err = strconv.Atoi(strings.TrimSpace(newAccessLevelStr))
		if err != nil || newAccessLevel < 0 || newAccessLevel > 255 {
			ue.showMessage(term, "Invalid access level. Must be 0-255.", "error")
			return
		}
	}

	// Update user
	err = ue.db.UpdateUser(id, strings.TrimSpace(newUsername), user.Password,
		strings.TrimSpace(newRealName), strings.TrimSpace(newEmail), newAccessLevel, user.IsActive)
	if err != nil {
		ue.showMessage(term, "Error updating user: "+err.Error(), "error")
	} else {
		ue.showMessage(term, "User updated successfully!", "success")
	}
}

// DeleteUser deletes a user account
func (ue *UserEditor) DeleteUser(term *term.Terminal) {
	term.Write([]byte(ClearScreen + ShowCursor))

	term.Write([]byte(ue.colorScheme.Colorize("Delete User\n\n", "primary")))
	term.Write([]byte(ue.colorScheme.Colorize("Enter user ID to delete: ", "text")))

	idStr, err := term.ReadLine()
	if err != nil {
		return
	}

	id, err := strconv.Atoi(strings.TrimSpace(idStr))
	if err != nil {
		ue.showMessage(term, "Invalid ID format.", "error")
		return
	}

	// Get user to show what will be deleted
	user, err := ue.db.GetUserByID(id)
	if err != nil {
		ue.showMessage(term, "User not found.", "error")
		return
	}

	// Prevent deletion of sysop account
	if user.AccessLevel == 255 {
		ue.showMessage(term, "Cannot delete sysop account.", "error")
		return
	}

	// Confirm deletion
	term.Write([]byte(ue.colorScheme.Colorize(fmt.Sprintf("Delete user: %s (%s)\n", user.Username, user.RealName), "secondary")))
	term.Write([]byte(ue.colorScheme.Colorize("Are you sure? (y/N): ", "text")))

	confirm, err := term.ReadLine()
	if err != nil {
		return
	}

	if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
		err = ue.db.DeleteUser(id)
		if err != nil {
			ue.showMessage(term, "Error deleting user: "+err.Error(), "error")
		} else {
			ue.showMessage(term, "User deleted successfully!", "success")
		}
	} else {
		ue.showMessage(term, "Deletion cancelled.", "text")
	}
}

// ChangePassword changes a user's password
func (ue *UserEditor) ChangePassword(term *term.Terminal) {
	term.Write([]byte(ClearScreen + ShowCursor))

	term.Write([]byte(ue.colorScheme.Colorize("Change User Password\n\n", "primary")))
	term.Write([]byte(ue.colorScheme.Colorize("Enter user ID: ", "text")))

	idStr, err := term.ReadLine()
	if err != nil {
		return
	}

	id, err := strconv.Atoi(strings.TrimSpace(idStr))
	if err != nil {
		ue.showMessage(term, "Invalid ID format.", "error")
		return
	}

	// Get existing user
	user, err := ue.db.GetUserByID(id)
	if err != nil {
		ue.showMessage(term, "User not found.", "error")
		return
	}

	term.Write([]byte(ue.colorScheme.Colorize(fmt.Sprintf("Changing password for: %s\n", user.Username), "secondary")))
	term.Write([]byte(ue.colorScheme.Colorize("Enter new password: ", "text")))

	newPassword, err := term.ReadLine()
	if err != nil {
		return
	}

	if strings.TrimSpace(newPassword) == "" {
		ue.showMessage(term, "Password cannot be empty.", "error")
		return
	}

	// Update user with new password
	err = ue.db.UpdateUser(id, user.Username, strings.TrimSpace(newPassword),
		user.RealName, user.Email, user.AccessLevel, user.IsActive)
	if err != nil {
		ue.showMessage(term, "Error updating password: "+err.Error(), "error")
	} else {
		ue.showMessage(term, "Password updated successfully!", "success")
	}
}

// ToggleUserStatus toggles user active status
func (ue *UserEditor) ToggleUserStatus(term *term.Terminal) {
	term.Write([]byte(ClearScreen + ShowCursor))

	term.Write([]byte(ue.colorScheme.Colorize("Toggle User Status\n\n", "primary")))
	term.Write([]byte(ue.colorScheme.Colorize("Enter user ID: ", "text")))

	idStr, err := term.ReadLine()
	if err != nil {
		return
	}

	id, err := strconv.Atoi(strings.TrimSpace(idStr))
	if err != nil {
		ue.showMessage(term, "Invalid ID format.", "error")
		return
	}

	// Get existing user
	user, err := ue.db.GetUserByID(id)
	if err != nil {
		ue.showMessage(term, "User not found.", "error")
		return
	}

	// Prevent disabling sysop account
	if user.AccessLevel == 255 && user.IsActive {
		ue.showMessage(term, "Cannot disable sysop account.", "error")
		return
	}

	newStatus := !user.IsActive
	statusText := "active"
	if !newStatus {
		statusText = "inactive"
	}

	term.Write([]byte(ue.colorScheme.Colorize(fmt.Sprintf("User %s will be set to: %s\n", user.Username, statusText), "secondary")))
	term.Write([]byte(ue.colorScheme.Colorize("Confirm? (y/N): ", "text")))

	confirm, err := term.ReadLine()
	if err != nil {
		return
	}

	if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
		err = ue.db.UpdateUser(id, user.Username, user.Password,
			user.RealName, user.Email, user.AccessLevel, newStatus)
		if err != nil {
			ue.showMessage(term, "Error updating user status: "+err.Error(), "error")
		} else {
			ue.showMessage(term, "User status updated successfully!", "success")
		}
	} else {
		ue.showMessage(term, "Operation cancelled.", "text")
	}
}

// showMessage displays a message and waits for key press
func (ue *UserEditor) showMessage(term *term.Terminal, message, colorType string) {
	term.Write([]byte(HideCursor))

	coloredMsg := ue.colorScheme.Colorize(message, colorType)
	centeredMsg := ue.colorScheme.CenterText(coloredMsg, 79)
	term.Write([]byte("\n" + centeredMsg))

	term.Write([]byte(ShowCursor))
	term.ReadLine()
	term.Write([]byte(HideCursor))
}
