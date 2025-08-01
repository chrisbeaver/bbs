package server

import (
	"fmt"
	"strings"

	"bbs/internal/config"
	"bbs/internal/database"
	"bbs/internal/modules/bulletins"
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
	s.write([]byte(s.colorScheme.Colorize("=== Searchlight BBS ===", "header") + "\n\n"))

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
	// For SSH terminals, we can use ReadLine method for both cases
	if sshTerm, ok := s.terminal.(*terminal.SSHTerminal); ok {
		if maskInput {
			// For password input, we'll need to implement a ReadPassword method in SSHTerminal
			// For now, use ReadLine and note that SSH terminals often handle masking at the client level
			return sshTerm.ReadLine()
		} else {
			// For username input
			return sshTerm.ReadLine()
		}
	}

	// For local sessions using Terminal interface
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
				if maskInput {
					s.terminal.Write([]byte("*"))
				} else {
					s.terminal.Write([]byte(string(buf[0])))
				}
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
					s.write([]byte(ShowCursor))
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
				s.write([]byte(ShowCursor))
				goodbyeMsg := s.colorScheme.Colorize("\nThank you for calling! Goodbye!\n", "success")
				s.write([]byte(goodbyeMsg))
				return

			default:
				// Ignore other keys
				continue
			}
		}
	}
}

// displayMenu displays the current menu - unified for both SSH and local
func (s *Session) displayMenu(menu *config.MenuItem) {
	// Clear screen and hide cursor
	s.write([]byte(ClearScreen + HideCursor))

	// Terminal width for centering
	terminalWidth := 79

	// Menu title with color and centering
	title := s.colorScheme.Colorize(menu.Title, "primary")
	centeredTitle := s.colorScheme.CenterText(title, terminalWidth)
	s.write([]byte(fmt.Sprintf("%s\n", centeredTitle)))

	// Decorative separator (centered to match title width)
	// Calculate the actual title length (without ANSI codes) for proper separator width
	cleanTitle := s.colorScheme.StripAnsiCodes(title)
	separator := s.colorScheme.DrawSeparator(len(cleanTitle), "═")
	centeredSeparator := s.colorScheme.CenterText(separator, terminalWidth)
	s.write([]byte(centeredSeparator + "\n\n"))

	// Build accessible menu items (filter by access level)
	var accessibleItems []config.MenuItem
	for _, item := range menu.Submenu {
		if s.user == nil || item.AccessLevel <= s.user.AccessLevel {
			accessibleItems = append(accessibleItems, item)
		}
	}

	// Calculate maximum width needed for highlight bar
	maxWidth := 0
	for _, item := range accessibleItems {
		if len(item.Description) > maxWidth {
			maxWidth = len(item.Description)
		}
	}
	// Add some padding
	maxWidth += 4

	// Calculate centering offset for menu items
	centerOffset := (terminalWidth - maxWidth) / 2
	if centerOffset < 0 {
		centerOffset = 0
	}

	// Create decorative border pattern
	borderPattern := s.colorScheme.CreateBorderPattern(maxWidth, "-=")
	centerPadding := strings.Repeat(" ", centerOffset)

	// Top border
	s.write([]byte(centerPadding + borderPattern + "\n"))

	// Ensure selected index is valid
	if s.selectedIndex >= len(accessibleItems) {
		s.selectedIndex = 0
	}
	if s.selectedIndex < 0 {
		s.selectedIndex = len(accessibleItems) - 1
	}

	// Display menu items with highlighting and centering
	for i, item := range accessibleItems {
		selected := (i == s.selectedIndex)
		menuLine := s.colorScheme.HighlightSelection(item.Description, selected, maxWidth)
		s.write([]byte(centerPadding + menuLine + "\n"))
	}

	// Bottom border
	s.write([]byte(centerPadding + borderPattern + "\n"))

	// Instructions (centered) - different for main menu vs submenus
	var instructions string
	if s.currentMenu == "main" {
		instructions = s.colorScheme.Colorize("Use ", "text") +
			s.colorScheme.Colorize("↑↓", "accent") +
			s.colorScheme.Colorize(" arrow keys to navigate, ", "text") +
			s.colorScheme.Colorize("Enter", "accent") +
			s.colorScheme.Colorize(" to select, ", "text") +
			s.colorScheme.Colorize("G", "accent") +
			s.colorScheme.Colorize(" for goodbye", "text")
	} else {
		instructions = s.colorScheme.Colorize("Use ", "text") +
			s.colorScheme.Colorize("↑↓", "accent") +
			s.colorScheme.Colorize(" arrow keys to navigate, ", "text") +
			s.colorScheme.Colorize("Enter", "accent") +
			s.colorScheme.Colorize(" to select, ", "text") +
			s.colorScheme.Colorize("Q", "accent") +
			s.colorScheme.Colorize(" to return, ", "text") +
			s.colorScheme.Colorize("G", "accent") +
			s.colorScheme.Colorize(" for goodbye", "text")
	}

	centeredInstructions := s.colorScheme.CenterText(instructions, terminalWidth)
	s.write([]byte("\n" + centeredInstructions))
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
	buf := make([]byte, 3)
	n, err := s.terminal.Read(buf)
	if err != nil {
		return "", err
	}

	switch n {
	case 1:
		// Single character
		char := string(buf[0])
		switch buf[0] {
		case 13: // Enter
			return "enter", nil
		case 27: // Escape (might be start of arrow key sequence)
			return "escape", nil
		case 'q', 'Q':
			return "quit", nil
		default:
			return char, nil
		}
	case 3:
		// Arrow key sequence: ESC [ A/B/C/D
		if buf[0] == 27 && buf[1] == 91 {
			switch buf[2] {
			case 65: // Up arrow
				return "up", nil
			case 66: // Down arrow
				return "down", nil
			case 67: // Right arrow
				return "right", nil
			case 68: // Left arrow
				return "left", nil
			}
		}
	}

	return string(buf[:n]), nil
}

// executeCommand executes the selected menu command - unified for both SSH and local
func (s *Session) executeCommand(item *config.MenuItem) bool {
	switch item.Command {
	case "bulletins":
		bulletinsModule := bulletins.NewModule(s.db, s.colorScheme)
		keyReader := &TerminalKeyReader{session: s}
		bulletinsModule.Execute(s.writer, keyReader)
		return true
	case "messages":
		// TODO: Implement messages module
		s.write([]byte(s.colorScheme.Colorize("Messages feature coming soon...", "text") + "\n"))
		s.waitForKey()
		return true
	case "sysop_menu", "sysop_stats", "sysop_user_menu", "sysop_bulletin_menu", "sysop_config", "sysop_maintenance":
		// Check if user has sysop access
		if s.user == nil || s.user.AccessLevel < 255 {
			s.write([]byte(s.colorScheme.Colorize("Access denied. Sysop privileges required.", "error") + "\n"))
			s.waitForKey()
			return true
		}

		// For now, show a placeholder message for sysop functionality
		s.write([]byte(s.colorScheme.Colorize("Sysop functionality will be integrated soon...", "text") + "\n"))
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
