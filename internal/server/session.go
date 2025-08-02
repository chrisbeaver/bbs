package server

import (
	"fmt"
	"strings"

	"bbs/internal/config"
	"bbs/internal/database"
	"bbs/internal/menu"
	"bbs/internal/modules/bulletins"
	"bbs/internal/modules/sysop/user_editor"
	"bbs/internal/terminal"
)

// Session represents a unified BBS session that can work with any terminal type
type Session struct {
	terminal          terminal.Terminal
	writer            *TerminalWriter // Use TerminalWriter for all output
	db                *database.DB
	config            *config.Config
	user              *database.User
	currentMenu       string
	menuHistory       []string
	selectedIndex     int
	authenticated     bool
	colorScheme       *ColorScheme
	prefilledUsername string // For SSH connections where username is already known
	menuRenderer      *menu.MenuRenderer
}

// Run is the unified entry point for all sessions (SSH and local)
func (s *Session) Run() {
	defer func() {
		if s.terminal != nil {
			s.terminal.Close()
		}
	}()

	// Display welcome message
	s.displayWelcome()

	// Handle authentication (username prefilled for SSH)
	if !s.handleLogin() {
		return
	}

	// Switch to raw mode for navigation (only for local terminals that support it)
	if s.terminal != nil {
		if err := s.terminal.MakeRaw(); err != nil {
			s.write([]byte("Warning: Could not set raw mode for navigation\n"))
		}
	}

	// Show bulletins after successful login
	bulletinsModule := bulletins.NewModule(s.db, s.colorScheme)
	writer := &TerminalWriter{session: s}
	keyReader := &TerminalKeyReader{session: s}
	bulletinsModule.Execute(writer, keyReader)

	// Set to main menu after bulletins
	s.currentMenu = "main"

	// Main menu loop
	s.menuLoop()
}

// handleLogin handles the login process for both SSH and local sessions
func (s *Session) handleLogin() bool {
	// For SSH sessions, user is already authenticated, just get user info
	if s.prefilledUsername != "" {
		user, err := s.db.GetUser(s.prefilledUsername)
		if err != nil {
			s.write([]byte(s.colorScheme.Colorize("Error retrieving user information.", "error") + "\n"))
			return false
		}
		s.user = user
		s.authenticated = true
		s.db.UpdateUserLastCall(s.prefilledUsername)

		s.write([]byte(s.colorScheme.Colorize(fmt.Sprintf("Welcome back, %s!", user.Username), "accent") + "\n"))
		if user.LastCall != nil {
			lastCallStr := fmt.Sprintf("Last call: %s", user.LastCall.Format("2006-01-02 15:04:05"))
			s.write([]byte(s.colorScheme.Colorize(lastCallStr, "text") + "\n"))
		} else {
			s.write([]byte(s.colorScheme.Colorize("Last call: First time login", "text") + "\n"))
		}
		totalCallsStr := fmt.Sprintf("Total calls: %d", user.TotalCalls)
		s.write([]byte(s.colorScheme.Colorize(totalCallsStr, "text") + "\n\n"))
		return true
	}

	// For local sessions, perform login process
	s.write([]byte(s.colorScheme.Colorize("=== Coastline BBS ===", "header") + "\n\n"))

	for attempts := 0; attempts < 3; attempts++ {
		// Get username
		s.write([]byte("Username: "))
		username, err := s.readInput(false)
		if err != nil {
			return false
		}

		if username == "" {
			continue
		}

		// Get password
		s.write([]byte("Password: "))
		password, err := s.readInput(true)
		if err != nil {
			return false
		}

		if password == "" {
			continue
		}

		// Validate credentials
		user, err := s.db.GetUser(username)
		if err != nil || user.Password != password {
			s.write([]byte(s.colorScheme.Colorize("Invalid username or password.", "error") + "\n"))
			continue
		}

		// Successful login
		s.user = user
		s.authenticated = true
		s.db.UpdateUserLastCall(username)

		s.write([]byte(s.colorScheme.Colorize(fmt.Sprintf("Welcome, %s!", user.Username), "accent") + "\n\n"))
		return true
	}

	s.write([]byte(s.colorScheme.Colorize("Too many failed attempts. Access denied.", "error") + "\n"))
	return false
}

// displayWelcome displays the welcome message
func (s *Session) displayWelcome() {
	banner := s.colorScheme.CreateWelcomeBanner(s.config.BBS.SystemName, s.config.BBS.WelcomeMsg)
	s.write([]byte(banner))
}

// readInput reads user input with optional masking (for passwords)
func (s *Session) readInput(maskInput bool) (string, error) {
	// Use character-by-character reading for both SSH and local terminals
	// to ensure consistent non-echoing behavior
	var input string
	buf := make([]byte, 1)

	for {
		n, err := s.terminal.Read(buf)
		if err != nil {
			return "", err
		}

		if n == 0 {
			continue
		}

		switch buf[0] {
		case 13, 10: // Enter or newline - finish input
			s.terminal.Write([]byte("\n"))
			return input, nil
		case 8, 127: // Backspace or DEL
			if len(input) > 0 {
				input = input[:len(input)-1]
				// Move cursor back, overwrite with space, move back again
				s.terminal.Write([]byte("\b \b"))
			}
		case 3: // Ctrl+C
			return "", fmt.Errorf("interrupted")
		case 27: // Escape - ignore
			continue
		default:
			// Add character to input
			if buf[0] >= 32 && buf[0] <= 126 { // Printable ASCII
				input += string(buf[0])
				// Don't echo input - let the user type without seeing characters
				if maskInput {
					s.terminal.Write([]byte("*"))
				}
				// For non-masked input, don't echo anything
			}
		}
	}
}

// write is a unified method to write to either terminal type
func (s *Session) write(data []byte) {
	// Use the same TerminalWriter that modules use for 100% consistency
	s.writer.Write(data)
}

// menuLoop handles the main menu interaction - unified for both SSH and local
func (s *Session) menuLoop() {
	for {
		// Find current menu
		var currentMenu *config.MenuItem
		for _, menu := range s.config.BBS.Menus {
			if menu.ID == s.currentMenu {
				currentMenu = &menu
				break
			}
		}

		if currentMenu == nil {
			s.write([]byte("Error: Menu not found\n"))
			return
		}

		// Build accessible menu items
		var accessibleItems []config.MenuItem
		for _, item := range currentMenu.Submenu {
			if s.user == nil || item.AccessLevel <= s.user.AccessLevel {
				accessibleItems = append(accessibleItems, item)
			}
		}

		if len(accessibleItems) == 0 {
			s.write([]byte("No menu items available\n"))
			return
		}

		// Display menu
		s.displayMenu(currentMenu)

		// Navigation loop
	NavigationLoop:
		for {
			key, err := s.readKey()
			if err != nil {
				return
			}

			switch key {
			case "up":
				s.selectedIndex--
				if s.selectedIndex < 0 {
					s.selectedIndex = len(accessibleItems) - 1
				}
				s.displayMenu(currentMenu)

			case "down":
				s.selectedIndex++
				if s.selectedIndex >= len(accessibleItems) {
					s.selectedIndex = 0
				}
				s.displayMenu(currentMenu)

			case "enter":
				// Execute selected item
				selectedItem := accessibleItems[s.selectedIndex]
				if !s.executeCommand(&selectedItem) {
					// Show cursor before exiting
					s.write([]byte(menu.ShowCursor))
					return
				}
				// Break out of navigation loop to redisplay menu
				break NavigationLoop

			case "quit", "q", "Q":
				// Handle Q key - return to previous menu (only works on submenus)
				if s.currentMenu == "main" {
					// Q does nothing on main menu
					continue
				} else {
					// Return to previous menu
					if len(s.menuHistory) > 0 {
						// Pop from history stack
						s.currentMenu = s.menuHistory[len(s.menuHistory)-1]
						s.menuHistory = s.menuHistory[:len(s.menuHistory)-1]
						s.selectedIndex = 0 // Reset selection
						break NavigationLoop
					} else {
						// Fallback to main menu if history is empty
						s.currentMenu = "main"
						s.selectedIndex = 0
						break NavigationLoop
					}
				}

			case "goodbye", "g", "G":
				// Handle G key - goodbye from any menu
				s.write([]byte(menu.ShowCursor))
				goodbyeMsg := s.colorScheme.Colorize("\nThank you for calling! Goodbye!\n", "success")
				s.write([]byte(goodbyeMsg))
				return

			default:
				// Check for hotkey matches
				if len(key) == 1 {
					keyLower := strings.ToLower(key)
					for _, item := range accessibleItems {
						if item.Hotkey != "" && strings.ToLower(item.Hotkey) == keyLower {
							// Found matching hotkey - execute the command
							if !s.executeCommand(&item) {
								// Show cursor before exiting
								s.write([]byte(menu.ShowCursor))
								return
							}
							// Break out of navigation loop to redisplay menu
							break NavigationLoop
						}
					}
				}
				// Ignore other keys
				continue
			}
		}
	}
}

// displayMenu displays the current menu - unified for both SSH and local
func (s *Session) displayMenu(menu *config.MenuItem) {
	// Get user access level (default to 0 if not authenticated)
	userAccessLevel := 0
	if s.user != nil {
		userAccessLevel = s.user.AccessLevel
	}

	// Use unified menu renderer with access level filtering
	s.menuRenderer.RenderConfigMenu(menu, s.selectedIndex, userAccessLevel)
}

// readKey reads a single key press - unified for both SSH and local
func (s *Session) readKey() (string, error) {
	// For SSH terminals, use the terminal interface
	if _, ok := s.terminal.(*terminal.SSHTerminal); ok {
		// SSH session - use existing SSH readKey logic
		return s.readKeySSH()
	} else {
		// Local session - use our terminal interface
		return s.readKeyLocal()
	}
}

// readKeyLocal handles key reading for local terminal
func (s *Session) readKeyLocal() (string, error) {
	buf := make([]byte, 1)
	n, err := s.terminal.Read(buf)
	if err != nil {
		return "", err
	}

	if n == 0 {
		return "", nil
	}

	// Handle single character
	switch buf[0] {
	case 13, 10: // Enter or newline
		return "enter", nil
	case 27: // Escape - check for arrow key sequence
		// Read next character to see if it's an arrow key
		buf2 := make([]byte, 1)
		n2, err := s.terminal.Read(buf2)
		if err != nil || n2 == 0 {
			return "escape", nil
		}

		if buf2[0] == 91 { // '['
			// Read the final character of the arrow key sequence
			buf3 := make([]byte, 1)
			n3, err := s.terminal.Read(buf3)
			if err != nil || n3 == 0 {
				return "escape", nil
			}

			switch buf3[0] {
			case 65: // 'A' - Up arrow
				return "up", nil
			case 66: // 'B' - Down arrow
				return "down", nil
			case 67: // 'C' - Right arrow
				return "right", nil
			case 68: // 'D' - Left arrow
				return "left", nil
			}
		}
		return "escape", nil
	case 'q', 'Q':
		return "quit", nil
	case 'g', 'G':
		return "goodbye", nil
	case 3: // Ctrl+C
		return "goodbye", nil
	default:
		return string(buf[0]), nil
	}
}

// readKeySSH handles key reading for SSH terminal
func (s *Session) readKeySSH() (string, error) {
	buf := make([]byte, 1)
	n, err := s.terminal.Read(buf)
	if err != nil {
		return "", err
	}

	if n == 0 {
		return "", nil
	}

	// Handle single character
	switch buf[0] {
	case 13, 10: // Enter or newline
		return "enter", nil
	case 27: // Escape - check for arrow key sequence
		// Read next character to see if it's an arrow key
		buf2 := make([]byte, 1)
		n2, err := s.terminal.Read(buf2)
		if err != nil || n2 == 0 {
			return "escape", nil
		}

		if buf2[0] == 91 { // '['
			// Read the final character of the arrow key sequence
			buf3 := make([]byte, 1)
			n3, err := s.terminal.Read(buf3)
			if err != nil || n3 == 0 {
				return "escape", nil
			}

			switch buf3[0] {
			case 65: // 'A' - Up arrow
				return "up", nil
			case 66: // 'B' - Down arrow
				return "down", nil
			case 67: // 'C' - Right arrow
				return "right", nil
			case 68: // 'D' - Left arrow
				return "left", nil
			}
		}
		return "escape", nil
	case 'q', 'Q':
		return "quit", nil
	case 'g', 'G':
		return "goodbye", nil
	case 3: // Ctrl+C
		return "goodbye", nil
	default:
		return string(buf[0]), nil
	}
}

// executeCommand executes the selected menu command - unified for both SSH and local
func (s *Session) executeCommand(item *config.MenuItem) bool {
	switch item.Command {
	case "bulletins":
		bulletinsModule := bulletins.NewModule(s.db, s.colorScheme)
		keyReader := &TerminalKeyReader{session: s}
		bulletinsModule.Execute(s.writer, keyReader)
		return true
	case "sysop_menu":
		// Check if user has sysop access
		if s.user == nil || s.user.AccessLevel < 255 {
			s.write([]byte(s.colorScheme.Colorize("Access denied. Sysop privileges required.", "error") + "\n"))
			s.waitForKey()
			return true
		}
		// Navigate to sysop_menu submenu
		s.menuHistory = append(s.menuHistory, s.currentMenu)
		s.currentMenu = "sysop_menu"
		s.selectedIndex = 0
		return true
	// Sysop command handlers
	case "create_user":
		if s.user == nil || s.user.AccessLevel < 255 {
			s.write([]byte(s.colorScheme.Colorize("Access denied. Sysop privileges required.", "error") + "\n"))
			s.waitForKey()
			return true
		}
		s.handleSysopCommand("create_user")
		return true
	case "edit_user":
		if s.user == nil || s.user.AccessLevel < 255 {
			s.write([]byte(s.colorScheme.Colorize("Access denied. Sysop privileges required.", "error") + "\n"))
			s.waitForKey()
			return true
		}
		s.handleSysopCommand("edit_user")
		return true
	case "delete_user":
		if s.user == nil || s.user.AccessLevel < 255 {
			s.write([]byte(s.colorScheme.Colorize("Access denied. Sysop privileges required.", "error") + "\n"))
			s.waitForKey()
			return true
		}
		s.handleSysopCommand("delete_user")
		return true
	case "view_users":
		if s.user == nil || s.user.AccessLevel < 255 {
			s.write([]byte(s.colorScheme.Colorize("Access denied. Sysop privileges required.", "error") + "\n"))
			s.waitForKey()
			return true
		}
		s.handleSysopCommand("view_users")
		return true
	case "change_password":
		if s.user == nil || s.user.AccessLevel < 255 {
			s.write([]byte(s.colorScheme.Colorize("Access denied. Sysop privileges required.", "error") + "\n"))
			s.waitForKey()
			return true
		}
		s.handleSysopCommand("change_password")
		return true
	case "toggle_user":
		if s.user == nil || s.user.AccessLevel < 255 {
			s.write([]byte(s.colorScheme.Colorize("Access denied. Sysop privileges required.", "error") + "\n"))
			s.waitForKey()
			return true
		}
		s.handleSysopCommand("toggle_user")
		return true
	case "system_stats":
		if s.user == nil || s.user.AccessLevel < 255 {
			s.write([]byte(s.colorScheme.Colorize("Access denied. Sysop privileges required.", "error") + "\n"))
			s.waitForKey()
			return true
		}
		s.handleSysopCommand("system_stats")
		return true
	case "bulletin_management":
		if s.user == nil || s.user.AccessLevel < 255 {
			s.write([]byte(s.colorScheme.Colorize("Access denied. Sysop privileges required.", "error") + "\n"))
			s.waitForKey()
			return true
		}
		s.handleSysopCommand("bulletin_management")
		return true
	case "messages":
		// TODO: Implement messages module
		s.write([]byte(s.colorScheme.Colorize("Messages feature coming soon...", "text") + "\n"))
		s.waitForKey()
		return true
	case "goodbye":
		return false
	case "logout":
		return false
	default:
		// Check if this item has submenus
		if len(item.Submenu) > 0 {
			// Navigate to submenu
			s.menuHistory = append(s.menuHistory, s.currentMenu)
			s.currentMenu = item.ID
			s.selectedIndex = 0
		} else {
			s.write([]byte(s.colorScheme.Colorize(fmt.Sprintf("Command '%s' not implemented yet.", item.Command), "text") + "\n"))
			s.waitForKey()
		}
		return true
	}
}

// waitForKey waits for any key press - unified for both SSH and local
func (s *Session) waitForKey() {
	s.write([]byte(s.colorScheme.Colorize("\nPress any key to continue...", "text")))

	// For both SSH and local, use the unified terminal interface
	if sshTerm, ok := s.terminal.(*terminal.SSHTerminal); ok {
		// SSH session - use ReadLine since it handles enter key nicely
		sshTerm.ReadLine()
	} else {
		// Local session - read single byte
		buf := make([]byte, 1)
		s.terminal.Read(buf)
	}
}

// handleSysopCommand executes sysop commands using the user_editor package
func (s *Session) handleSysopCommand(command string) {
	// Create user editor instance
	editor := user_editor.NewUserEditor(s.db, s.colorScheme)
	keyReader := &TerminalKeyReader{session: s}

	// Map commands to user_editor methods
	switch command {
	case "create_user":
		editor.CreateUser(s.writer, keyReader)
	case "edit_user":
		editor.EditUser(s.writer, keyReader)
	case "delete_user":
		editor.DeleteUser(s.writer, keyReader)
	case "view_users":
		editor.ListUsers(s.writer, keyReader)
	case "change_password":
		editor.ChangePassword(s.writer, keyReader)
	case "toggle_user":
		editor.ToggleUserStatus(s.writer, keyReader)
	case "system_stats":
		s.handleSystemStats()
	case "bulletin_management":
		s.write([]byte(s.colorScheme.Colorize("Bulletin Management - Not yet implemented", "secondary") + "\n"))
		s.waitForKey()
	default:
		s.write([]byte(s.colorScheme.Colorize(fmt.Sprintf("Unknown sysop command: %s", command), "error") + "\n"))
		s.waitForKey()
	}
}

// handleSystemStats displays system statistics
func (s *Session) handleSystemStats() {
	s.write([]byte(menu.ClearScreen))

	header := s.colorScheme.Colorize("--- System Statistics ---", "primary")
	centeredHeader := s.colorScheme.CenterText(header, 79)
	s.write([]byte(centeredHeader + "\n"))

	separator := s.colorScheme.DrawSeparator(len("System Statistics"), "═")
	centeredSeparator := s.colorScheme.CenterText(separator, 79)
	s.write([]byte(centeredSeparator + "\n\n"))

	// Get users count
	users, err := s.db.GetAllUsers(1000)
	if err != nil {
		s.write([]byte(s.colorScheme.Colorize("Error retrieving user statistics: "+err.Error(), "error") + "\n"))
		s.waitForKey()
		return
	}

	// Get bulletins count
	bulletins, err := s.db.GetBulletins(1000)
	if err != nil {
		s.write([]byte(s.colorScheme.Colorize("Error retrieving bulletin statistics: "+err.Error(), "error") + "\n"))
		s.waitForKey()
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
		coloredStat := s.colorScheme.Colorize(stat, "text")
		centeredStat := s.colorScheme.CenterText(coloredStat, 79)
		s.write([]byte(centeredStat + "\n"))
	}

	s.waitForKey()
}
