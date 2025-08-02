package user_editor

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"bbs/internal/components"
	"bbs/internal/database"
	"bbs/internal/modules"
)

// CreateUser creates a new user using form components
func (ue *UserEditor) CreateUser(writer modules.Writer, keyReader modules.KeyReader) bool {
	// Create the form
	form := components.NewForm(components.FormConfig{
		Title: "Create New User",
		Width: 79,
	}, ue.getComponentAdapter())

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
			if _, err := ue.db.GetUser(trimmed); err == nil {
				return fmt.Errorf("user already exists")
			}
			return nil
		},
	}, ue.getComponentAdapter())

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
	}, ue.getComponentAdapter())

	// Add real name field
	realNameField := components.NewTextInput(components.TextInputConfig{
		Name:        "real_name",
		Label:       "Real Name",
		Placeholder: "Enter real name (optional)...",
		MaxLength:   64,
		Required:    false,
		Width:       40,
	}, ue.getComponentAdapter())

	// Add email field
	emailField := components.NewTextInput(components.TextInputConfig{
		Name:        "email",
		Label:       "Email",
		Placeholder: "Enter email (optional)...",
		MaxLength:   128,
		Required:    false,
		Width:       40,
	}, ue.getComponentAdapter())

	// Add access level field
	accessLevelField := components.NewTextInput(components.TextInputConfig{
		Name:        "access_level",
		Label:       "Access Level",
		Placeholder: "0-255",
		Value:       "10", // Default access level
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
	}, ue.getComponentAdapter())

	// Add components to form
	form.AddComponent(usernameField)
	form.AddComponent(passwordField)
	form.AddComponent(realNameField)
	form.AddComponent(emailField)
	form.AddComponent(accessLevelField)

	// Start form interaction
	form.Start()

	// Interactive form loop
	for {
		// Render form
		writer.Write([]byte(form.Render()))

		// Read single character input
		keyStr, err := keyReader.ReadKey()
		if err != nil {
			break
		}

		// Convert special key names to runes
		var char rune
		switch keyStr {
		case "enter":
			char = '\r'
		case "escape":
			char = 27
		case "quit", "goodbye":
			char = 27
		default:
			// Regular character
			if len(keyStr) > 0 {
				char = rune(keyStr[0])
			} else {
				continue
			}
		}

		// Handle the key
		form.HandleKey(char)

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
					CreatedAt:   time.Now(),
				}

				if err := ue.db.CreateUser(user); err != nil {
					showMessage(writer, keyReader, ue.colorScheme, "Error creating user: "+err.Error(), "error")
				} else {
					showMessage(writer, keyReader, ue.colorScheme, "User created successfully!", "success")
				}
				return true
			} else {
				// Show validation errors
				errorMsg := "Validation errors:\n"
				for _, err := range errors {
					errorMsg += "• " + err.Error() + "\n"
				}
				showMessage(writer, keyReader, ue.colorScheme, errorMsg, "error")
				form.Reset()
				form.Start()
			}
		}

		if form.IsCancelled() {
			return true
		}
	}

	return true
}
