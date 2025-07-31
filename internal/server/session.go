package server

import (
	"fmt"
	"strings"

	"golang.org/x/term"

	"bbs/internal/config"
	"bbs/internal/database"
	"bbs/internal/modules/bulletins"
	"bbs/internal/terminal"
)

// Session represents a unified BBS session that can work with any terminal type
type Session struct {
	terminal      terminal.Terminal
	termTerminal  *term.Terminal // For compatibility with existing SSH code
	db            *database.DB
	config        *config.Config
	user          *database.User
	currentMenu   string
	menuHistory   []string
	selectedIndex int
	authenticated bool
	colorScheme   *ColorScheme
}

// NewSession creates a new unified session
func NewSession(term terminal.Terminal, termTerm *term.Terminal, db *database.DB, cfg *config.Config) *Session {
	return &Session{
		terminal:      term,
		termTerminal:  termTerm,
		db:            db,
		config:        cfg,
		currentMenu:   "main",
		selectedIndex: 0,
		authenticated: false,
		colorScheme:   NewColorScheme(&cfg.BBS.Colors),
	}
}

// LocalSession wraps the unified session for local terminal access
type LocalSession struct {
	*Session
}

// NewLocalSession creates a new local session
func NewLocalSession(term terminal.Terminal, db *database.DB, cfg *config.Config) *LocalSession {
	// For local sessions, we don't have a term.Terminal, so we pass nil
	session := NewSession(term, nil, db, cfg)
	return &LocalSession{Session: session}
}

// Run starts the local BBS session
func (s *LocalSession) Run() {
	defer s.terminal.Close()

	// Display welcome message
	s.displayWelcome()

	// Login process for local sessions
	if !s.handleLogin() {
		return
	}

	// Switch to raw mode before showing bulletins for proper navigation
	if err := s.terminal.MakeRaw(); err != nil {
		s.terminal.Write([]byte("Warning: Could not set raw mode for navigation\n"))
	}

	// Show bulletins after successful login
	bulletinsModule := bulletins.NewModule(s.Session.db, s.Session.colorScheme)
	writer := &LocalWriter{session: s.Session}
	keyReader := &LocalKeyReader{session: s.Session}
	bulletinsModule.Execute(writer, keyReader)

	// Set to main menu after bulletins
	s.currentMenu = "main"

	// Main menu loop - now using unified session logic
	s.menuLoop()
}

// handleLogin handles the login process for local sessions
func (s *LocalSession) handleLogin() bool {
	s.write([]byte(s.colorScheme.Colorize("=== Searchlight BBS Local Access ===", "header") + "\n\n"))

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
func (s *LocalSession) displayWelcome() {
	banner := s.colorScheme.CreateWelcomeBanner(s.config.BBS.SystemName, s.config.BBS.WelcomeMsg)
	s.write([]byte(banner))
}

// write is a unified method to write to either terminal type
func (s *Session) write(data []byte) {
	if s.termTerminal != nil {
		// SSH session - use term.Terminal
		s.termTerminal.Write(data)
	} else {
		// Local session - use our Terminal interface
		s.terminal.Write(data)
	}
}

// Write implements io.Writer interface for bulletins module
func (s *Session) Write(data []byte) (int, error) {
	if s.termTerminal != nil {
		// SSH session - use term.Terminal
		return s.termTerminal.Write(data)
	} else {
		// Local session - use our Terminal interface
		s.terminal.Write(data)
		return len(data), nil
	}
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
	cleanTitle := s.colorScheme.stripAnsiCodes(title)
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
	if s.termTerminal != nil {
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

// readKeySSH handles key reading for SSH terminal (placeholder - implement with existing SSH logic)
func (s *Session) readKeySSH() (string, error) {
	// This should be implemented with the existing SSH readKey logic
	// For now, just return enter to avoid compilation errors
	return "enter", nil
}

// executeCommand executes the selected menu command - unified for both SSH and local
func (s *Session) executeCommand(item *config.MenuItem) bool {
	switch item.Command {
	case "bulletins":
		bulletinsModule := bulletins.NewModule(s.db, s.colorScheme)
		writer := &LocalWriter{session: s}
		keyReader := &LocalKeyReader{session: s}
		bulletinsModule.Execute(writer, keyReader)
		return true
	case "messages":
		// TODO: Implement messages module
		s.write([]byte(s.colorScheme.Colorize("Messages feature coming soon...", "text") + "\n"))
		s.waitForKey()
		return true
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

	// Read a single key
	if s.termTerminal != nil {
		// SSH session
		s.termTerminal.ReadLine()
	} else {
		// Local session
		buf := make([]byte, 1)
		s.terminal.Read(buf)
	}
}

// readInput reads user input with optional masking (for passwords) - for local sessions only
func (s *LocalSession) readInput(maskInput bool) (string, error) {
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
				}
				// For username, don't echo anything (no output)
			}
		}
	}
}

// LocalWriter implements the Writer interface for local sessions
type LocalWriter struct {
	session *Session
}

func (w *LocalWriter) Write(data []byte) (int, error) {
	w.session.write(data)
	return len(data), nil
}

// LocalKeyReader implements the KeyReader interface for local sessions
type LocalKeyReader struct {
	session *Session
}

func (l *LocalKeyReader) ReadKey() (string, error) {
	return l.session.readKey()
}
