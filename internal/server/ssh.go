package server

import (
	"fmt"
	"net"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"

	"bbs/internal/config"
	"bbs/internal/database"
	"bbs/internal/modules/sysop"
)

type SSHServer struct {
	config      *config.Config
	db          *database.DB
	sshConfig   *ssh.ServerConfig
	colorScheme *ColorScheme
}

type Session struct {
	conn          ssh.Conn
	channel       ssh.Channel
	term          *term.Terminal
	user          *database.User
	currentMenu   string
	selectedIndex int
	authenticated bool
}

func NewSSHServer(cfg *config.Config, db *database.DB) *SSHServer {
	server := &SSHServer{
		config:      cfg,
		db:          db,
		colorScheme: NewColorScheme(&cfg.BBS.Colors),
	}

	server.setupSSHConfig()
	return server
}

func (s *SSHServer) setupSSHConfig() {
	s.sshConfig = &ssh.ServerConfig{
		PasswordCallback: s.passwordCallback,
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			return nil, fmt.Errorf("public key authentication not supported")
		},
	}

	// Generate or load host key
	hostKey, err := GenerateHostKey(s.config.Server.HostKeyPath)
	if err != nil {
		panic(fmt.Sprintf("Failed to load host key: %v", err))
	}
	s.sshConfig.AddHostKey(hostKey)
}

func (s *SSHServer) passwordCallback(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	username := conn.User()

	// Try to authenticate user
	user, err := s.db.GetUser(username)
	if err != nil {
		return nil, fmt.Errorf("authentication failed")
	}

	// Simple password check (in production, use proper hashing)
	if user.Password != string(password) {
		return nil, fmt.Errorf("authentication failed")
	}

	return &ssh.Permissions{
		Extensions: map[string]string{
			"username": username,
		},
	}, nil
}

func (s *SSHServer) HandleConnection(netConn net.Conn) {
	defer netConn.Close()

	// Perform SSH handshake
	sshConn, chans, reqs, err := ssh.NewServerConn(netConn, s.sshConfig)
	if err != nil {
		return
	}
	defer sshConn.Close()

	// Handle out-of-band requests
	go ssh.DiscardRequests(reqs)

	// Handle channels
	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unsupported channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			continue
		}

		// Create session
		session := &Session{
			conn:          sshConn,
			channel:       channel,
			currentMenu:   "main",
			selectedIndex: 0,
			authenticated: true,
		}

		// Get user info from permissions
		if username, ok := sshConn.Permissions.Extensions["username"]; ok {
			user, err := s.db.GetUser(username)
			if err == nil {
				session.user = user
				s.db.UpdateUserLastCall(username)
			}
		}

		go s.handleSession(session, channel, requests)
	}
}

func (s *SSHServer) handleSession(session *Session, channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()

	// Handle session requests
	go func() {
		for req := range requests {
			switch req.Type {
			case "shell", "exec":
				if req.WantReply {
					req.Reply(true, nil)
				}
			case "pty-req":
				if req.WantReply {
					req.Reply(true, nil)
				}
			default:
				if req.WantReply {
					req.Reply(false, nil)
				}
			}
		}
	}()

	// Create terminal
	terminal := term.NewTerminal(channel, "")
	session.term = terminal

	// Display welcome message
	s.displayWelcome(session)

	// Main menu loop
	s.menuLoop(session)
}

func (s *SSHServer) displayWelcome(session *Session) {
	// Create colorized welcome banner
	banner := s.colorScheme.CreateWelcomeBanner(s.config.BBS.SystemName, s.config.BBS.WelcomeMsg)
	session.term.Write([]byte(banner))

	if session.user != nil {
		session.term.Write([]byte(s.colorScheme.Colorize(fmt.Sprintf("Welcome back, %s!", session.user.Username), "accent") + "\n"))

		if session.user.LastCall != nil {
			lastCallStr := fmt.Sprintf("Last call: %s", session.user.LastCall.Format("2006-01-02 15:04:05"))
			session.term.Write([]byte(s.colorScheme.Colorize(lastCallStr, "text") + "\n"))
		} else {
			session.term.Write([]byte(s.colorScheme.Colorize("Last call: First time login", "text") + "\n"))
		}

		totalCallsStr := fmt.Sprintf("Total calls: %d", session.user.TotalCalls)
		session.term.Write([]byte(s.colorScheme.Colorize(totalCallsStr, "text") + "\n\n"))
	}
}

func (s *SSHServer) menuLoop(session *Session) {
	for {
		// Find current menu
		var currentMenu *config.MenuItem
		for _, menu := range s.config.BBS.Menus {
			if menu.ID == session.currentMenu {
				currentMenu = &menu
				break
			}
		}

		if currentMenu == nil {
			session.term.Write([]byte("Error: Menu not found\n"))
			return
		}

		// Build accessible menu items
		var accessibleItems []config.MenuItem
		for _, item := range currentMenu.Submenu {
			if session.user == nil || item.AccessLevel <= session.user.AccessLevel {
				accessibleItems = append(accessibleItems, item)
			}
		}

		if len(accessibleItems) == 0 {
			session.term.Write([]byte("No menu items available\n"))
			return
		}

		// Display menu
		s.displayMenu(session, currentMenu)

		// Navigation loop
	NavigationLoop:
		for {
			key, err := s.readKey(session)
			if err != nil {
				return
			}

			// Debug: Show what key was pressed at main menu level
			//session.term.Write([]byte(fmt.Sprintf("\nMAIN MENU DEBUG: Key pressed: '%s'\n", key)))

			switch key {
			case "up":
				session.selectedIndex--
				if session.selectedIndex < 0 {
					session.selectedIndex = len(accessibleItems) - 1
				}
				s.displayMenu(session, currentMenu)

			case "down":
				session.selectedIndex++
				if session.selectedIndex >= len(accessibleItems) {
					session.selectedIndex = 0
				}
				s.displayMenu(session, currentMenu)

			case "enter":
				// Execute selected item
				selectedItem := accessibleItems[session.selectedIndex]
				if !s.executeCommand(session, &selectedItem) {
					// Show cursor before exiting
					session.term.Write([]byte(ShowCursor))
					return
				}
				// Break out of navigation loop to redisplay menu
				break NavigationLoop

			case "quit", "q", "Q":
				// Show cursor and exit
				session.term.Write([]byte(ShowCursor))
				goodbyeMsg := s.colorScheme.Colorize("\nThank you for calling! Goodbye!\n", "success")
				session.term.Write([]byte(goodbyeMsg))
				return

			default:
				// Ignore other keys
				continue
			}
		}
	}
}

func (s *SSHServer) displayMenu(session *Session, menu *config.MenuItem) {
	// Clear screen and hide cursor
	session.term.Write([]byte(ClearScreen + HideCursor))

	// Terminal width for centering
	terminalWidth := 79

	// Menu title with color and centering
	title := s.colorScheme.Colorize(menu.Title, "primary")
	centeredTitle := s.colorScheme.CenterText(title, terminalWidth)
	session.term.Write([]byte(fmt.Sprintf("%s\n", centeredTitle)))

	// Decorative separator (centered to match title width)
	// Calculate the actual title length (without ANSI codes) for proper separator width
	cleanTitle := s.colorScheme.stripAnsiCodes(title)
	separator := s.colorScheme.DrawSeparator(len(cleanTitle), "═")
	centeredSeparator := s.colorScheme.CenterText(separator, terminalWidth)
	session.term.Write([]byte(centeredSeparator + "\n\n"))

	// Build accessible menu items (filter by access level)
	var accessibleItems []config.MenuItem
	for _, item := range menu.Submenu {
		if session.user == nil || item.AccessLevel <= session.user.AccessLevel {
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
	session.term.Write([]byte(centerPadding + borderPattern + "\n"))

	// Ensure selected index is valid
	if session.selectedIndex >= len(accessibleItems) {
		session.selectedIndex = 0
	}
	if session.selectedIndex < 0 {
		session.selectedIndex = len(accessibleItems) - 1
	}

	// Display menu items with highlighting and centering
	for i, item := range accessibleItems {
		selected := (i == session.selectedIndex)
		menuLine := s.colorScheme.HighlightSelection(item.Description, selected, maxWidth)
		session.term.Write([]byte(centerPadding + menuLine + "\n"))
	}

	// Bottom border
	session.term.Write([]byte(centerPadding + borderPattern + "\n"))

	// Instructions (centered)
	instructions := s.colorScheme.Colorize("Use ", "text") +
		s.colorScheme.Colorize("↑↓", "accent") +
		s.colorScheme.Colorize(" arrow keys to navigate, ", "text") +
		s.colorScheme.Colorize("Enter", "accent") +
		s.colorScheme.Colorize(" to select, ", "text") +
		s.colorScheme.Colorize("Q", "accent") +
		s.colorScheme.Colorize(" to quit", "text")

	centeredInstructions := s.colorScheme.CenterText(instructions, terminalWidth)
	session.term.Write([]byte("\n" + centeredInstructions))
}

// readKey reads a single key press, handling special keys like arrows
func (s *SSHServer) readKey(session *Session) (string, error) {
	buf := make([]byte, 3)
	n, err := session.channel.Read(buf)
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

func (s *SSHServer) processCommand(session *Session, menu *config.MenuItem, input string) bool {
	if input == "quit" || input == "q" || input == "goodbye" {
		goodbyeMsg := s.colorScheme.Colorize("\nThank you for calling! Goodbye!\n", "success")
		session.term.Write([]byte(goodbyeMsg))
		return false
	}

	// Try to parse as number
	if len(input) == 1 && input[0] >= '1' && input[0] <= '9' {
		choice := int(input[0] - '1')
		if choice < len(menu.Submenu) {
			item := menu.Submenu[choice]

			// Check access level
			if session.user != nil && item.AccessLevel > session.user.AccessLevel {
				accessDenied := s.colorScheme.Colorize("Access denied.", "error")
				session.term.Write([]byte(accessDenied + "\n"))
				return true
			}

			return s.executeCommand(session, &item)
		}
	}

	invalidChoice := s.colorScheme.Colorize("Invalid choice. Please try again.", "error")
	session.term.Write([]byte(invalidChoice + "\n"))
	return true
}

// isMenuCommand checks if a command refers to a menu
func (s *SSHServer) isMenuCommand(command string) bool {
	for _, menu := range s.config.BBS.Menus {
		if menu.ID == command {
			return true
		}
	}
	return false
}

func (s *SSHServer) executeCommand(session *Session, item *config.MenuItem) bool {
	// First check if this command refers to a menu
	if s.isMenuCommand(item.Command) {
		session.currentMenu = item.Command
		return true
	}

	// Otherwise execute as a command
	switch item.Command {
	case "bulletins":
		return s.executeBulletinsModule(session)
	case "messages":
		return s.showMessages(session)
	case "files":
		msg := s.colorScheme.Colorize("File areas not yet implemented.", "secondary")
		session.term.Write([]byte(msg + "\n"))
		return true
	case "games":
		msg := s.colorScheme.Colorize("Games not yet implemented.", "secondary")
		session.term.Write([]byte(msg + "\n"))
		return true
	case "users":
		msg := s.colorScheme.Colorize("User listings not yet implemented.", "secondary")
		session.term.Write([]byte(msg + "\n"))
		return true
	case "sysop_stats":
		return s.executeSysopStats(session)
	case "sysop_config":
		msg := s.colorScheme.Colorize("System configuration not yet implemented.", "secondary")
		session.term.Write([]byte(msg + "\n"))
		return true
	case "sysop_maintenance":
		msg := s.colorScheme.Colorize("Database maintenance not yet implemented.", "secondary")
		session.term.Write([]byte(msg + "\n"))
		return true
	// Sysop User Management Commands
	case "sysop_user_list":
		return s.executeSysopUserList(session)
	case "sysop_user_create":
		return s.executeSysopUserCreate(session)
	case "sysop_user_edit":
		return s.executeSysopUserEdit(session)
	case "sysop_user_delete":
		return s.executeSysopUserDelete(session)
	case "sysop_user_password":
		return s.executeSysopUserPassword(session)
	case "sysop_user_toggle":
		return s.executeSysopUserToggle(session)
	// Sysop Bulletin Management Commands
	case "sysop_bulletin_list":
		return s.executeSysopBulletinList(session)
	case "sysop_bulletin_create":
		return s.executeSysopBulletinCreate(session)
	case "sysop_bulletin_edit":
		return s.executeSysopBulletinEdit(session)
	case "sysop_bulletin_delete":
		return s.executeSysopBulletinDelete(session)
	case "goodbye":
		goodbyeMsg := s.colorScheme.Colorize("\nThank you for calling! Goodbye!\n", "success")
		session.term.Write([]byte(goodbyeMsg))
		return false
	default:
		msg := s.colorScheme.Colorize(fmt.Sprintf("Command '%s' not implemented.", item.Command), "error")
		session.term.Write([]byte(msg + "\n"))
		return true
	}
}

// executeBulletinsModule creates and runs the bulletins module with proper key handling
func (s *SSHServer) executeBulletinsModule(session *Session) bool {
	// Get bulletins from database
	bulletins, err := s.db.GetBulletins(50)
	if err != nil {
		errorMsg := s.colorScheme.Colorize("Error retrieving bulletins.", "error")
		centeredError := s.colorScheme.CenterText(errorMsg, 79)
		session.term.Write([]byte(centeredError + "\n"))
		session.term.ReadLine() // Wait for key
		return true
	}

	if len(bulletins) == 0 {
		// Clear screen and show no bulletins message
		session.term.Write([]byte(ClearScreen + HideCursor))

		header := s.colorScheme.Colorize("System Bulletins", "primary")
		centeredHeader := s.colorScheme.CenterText(header, 79)
		session.term.Write([]byte(centeredHeader + "\n"))

		separator := s.colorScheme.DrawSeparator(len("System Bulletins"), "═")
		centeredSeparator := s.colorScheme.CenterText(separator, 79)
		session.term.Write([]byte(centeredSeparator + "\n\n"))

		noMsg := s.colorScheme.Colorize("No bulletins available.", "secondary")
		centeredNoMsg := s.colorScheme.CenterText(noMsg, 79)
		session.term.Write([]byte(centeredNoMsg + "\n\n"))

		prompt := s.colorScheme.Colorize("Press any key to continue...", "text")
		centeredPrompt := s.colorScheme.CenterText(prompt, 79)
		session.term.Write([]byte(centeredPrompt))

		session.term.ReadLine()
		session.term.Write([]byte(ShowCursor))
		return true
	}

	// Show navigable bulletin list using the same key handling as menus
	return s.showNavigableBulletinList(session, bulletins)
}

// Helper function to check sysop access
func (s *SSHServer) checkSysopAccess(session *Session) bool {
	if session.user == nil || session.user.AccessLevel < 255 {
		errorMsg := s.colorScheme.Colorize("Access denied. Sysop access required.", "error")
		centeredError := s.colorScheme.CenterText(errorMsg, 79)
		session.term.Write([]byte(centeredError + "\n"))
		session.term.ReadLine() // Wait for key
		return false
	}
	return true
}

// Sysop Statistics
func (s *SSHServer) executeSysopStats(session *Session) bool {
	if !s.checkSysopAccess(session) {
		return true
	}

	session.term.Write([]byte(ClearScreen))

	header := s.colorScheme.Colorize("System Statistics", "primary")
	centeredHeader := s.colorScheme.CenterText(header, 79)
	session.term.Write([]byte(centeredHeader + "\n"))

	separator := s.colorScheme.DrawSeparator(len("System Statistics"), "═")
	centeredSeparator := s.colorScheme.CenterText(separator, 79)
	session.term.Write([]byte(centeredSeparator + "\n\n"))

	// Get users count
	users, err := s.db.GetAllUsers(1000)
	if err != nil {
		s.showSysopMessage(session, "Error retrieving user statistics: "+err.Error(), "error")
		return true
	}

	// Get bulletins count
	bulletins, err := s.db.GetBulletins(1000)
	if err != nil {
		s.showSysopMessage(session, "Error retrieving bulletin statistics: "+err.Error(), "error")
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
		fmt.Sprintf("Total Users: %d", len(users)),
		fmt.Sprintf("Active Users: %d", activeUsers),
		fmt.Sprintf("Inactive Users: %d", len(users)-activeUsers),
		fmt.Sprintf("Total Bulletins: %d", len(bulletins)),
		fmt.Sprintf("Total System Calls: %d", totalCalls),
	}

	for _, stat := range stats {
		coloredStat := s.colorScheme.Colorize(stat, "text")
		centeredStat := s.colorScheme.CenterText(coloredStat, 79)
		session.term.Write([]byte(centeredStat + "\n"))
	}

	s.showSysopMessage(session, "\nPress any key to continue...", "text")
	return true
}

// Sysop User Management Commands
func (s *SSHServer) executeSysopUserList(session *Session) bool {
	if !s.checkSysopAccess(session) {
		return true
	}

	userEditor := sysop.NewUserEditor(s.db, s.colorScheme)
	userEditor.ListUsers(session.term)
	return true
}

func (s *SSHServer) executeSysopUserCreate(session *Session) bool {
	if !s.checkSysopAccess(session) {
		return true
	}

	userEditor := sysop.NewUserEditor(s.db, s.colorScheme)
	userEditor.CreateUser(session.term)
	return true
}

func (s *SSHServer) executeSysopUserEdit(session *Session) bool {
	if !s.checkSysopAccess(session) {
		return true
	}

	userEditor := sysop.NewUserEditor(s.db, s.colorScheme)
	userEditor.EditUser(session.term)
	return true
}

func (s *SSHServer) executeSysopUserDelete(session *Session) bool {
	if !s.checkSysopAccess(session) {
		return true
	}

	userEditor := sysop.NewUserEditor(s.db, s.colorScheme)
	userEditor.DeleteUser(session.term)
	return true
}

func (s *SSHServer) executeSysopUserPassword(session *Session) bool {
	if !s.checkSysopAccess(session) {
		return true
	}

	userEditor := sysop.NewUserEditor(s.db, s.colorScheme)
	userEditor.ChangePassword(session.term)
	return true
}

func (s *SSHServer) executeSysopUserToggle(session *Session) bool {
	if !s.checkSysopAccess(session) {
		return true
	}

	userEditor := sysop.NewUserEditor(s.db, s.colorScheme)
	userEditor.ToggleUserStatus(session.term)
	return true
}

// Sysop Bulletin Management Commands
func (s *SSHServer) executeSysopBulletinList(session *Session) bool {
	if !s.checkSysopAccess(session) {
		return true
	}

	bulletinEditor := sysop.NewBulletinEditor(s.db, s.colorScheme)
	bulletinEditor.ListBulletins(session.term)
	return true
}

func (s *SSHServer) executeSysopBulletinCreate(session *Session) bool {
	if !s.checkSysopAccess(session) {
		return true
	}

	bulletinEditor := sysop.NewBulletinEditor(s.db, s.colorScheme)
	bulletinEditor.CreateBulletin(session.term)
	return true
}

func (s *SSHServer) executeSysopBulletinEdit(session *Session) bool {
	if !s.checkSysopAccess(session) {
		return true
	}

	bulletinEditor := sysop.NewBulletinEditor(s.db, s.colorScheme)
	bulletinEditor.EditBulletin(session.term)
	return true
}

func (s *SSHServer) executeSysopBulletinDelete(session *Session) bool {
	if !s.checkSysopAccess(session) {
		return true
	}

	bulletinEditor := sysop.NewBulletinEditor(s.db, s.colorScheme)
	bulletinEditor.DeleteBulletin(session.term)
	return true
}

// Helper method for sysop messages
func (s *SSHServer) showSysopMessage(session *Session, message, colorType string) {
	session.term.Write([]byte(HideCursor))

	coloredMsg := s.colorScheme.Colorize(message, colorType)
	centeredMsg := s.colorScheme.CenterText(coloredMsg, 79)
	session.term.Write([]byte("\n" + centeredMsg))

	session.term.Write([]byte(ShowCursor))
	session.term.ReadLine()
	session.term.Write([]byte(HideCursor))
}

// showNavigableBulletinList displays a navigable list of bulletins
func (s *SSHServer) showNavigableBulletinList(session *Session, bulletins []database.Bulletin) bool {
	selectedIndex := 0

	for {
		// Clear screen and hide cursor
		session.term.Write([]byte(ClearScreen + HideCursor))

		// Display header
		header := s.colorScheme.Colorize("System Bulletins", "primary")
		centeredHeader := s.colorScheme.CenterText(header, 79)
		session.term.Write([]byte(centeredHeader + "\n"))

		separator := s.colorScheme.DrawSeparator(len("System Bulletins"), "═")
		centeredSeparator := s.colorScheme.CenterText(separator, 79)
		session.term.Write([]byte(centeredSeparator + "\n\n"))

		// Calculate display area
		terminalWidth := 79
		contentWidth := 70
		centerOffset := (terminalWidth - contentWidth) / 2
		centerPadding := strings.Repeat(" ", centerOffset)

		// Display bulletin list with navigation
		for i, bulletin := range bulletins {
			isSelected := (i == selectedIndex)

			// Format bulletin line
			number := fmt.Sprintf("%2d)", i+1)
			title := bulletin.Title
			author := fmt.Sprintf("by %s", bulletin.Author)
			date := bulletin.CreatedAt.Format("2006-01-02")

			// Truncate title if too long
			maxTitleLength := contentWidth - len(number) - len(author) - len(date) - 6 // spaces and parentheses
			if len(title) > maxTitleLength {
				title = title[:maxTitleLength-3] + "..."
			}

			bulletinLine := fmt.Sprintf("%s %s (%s, %s)", number, title, author, date)

			// Pad to content width
			if len(bulletinLine) < contentWidth {
				bulletinLine += strings.Repeat(" ", contentWidth-len(bulletinLine))
			}

			if isSelected {
				// Highlight selected item
				coloredLine := s.colorScheme.ColorizeWithBg(bulletinLine, "highlight", "primary")
				session.term.Write([]byte(centerPadding + coloredLine + "\n"))
			} else {
				// Normal item
				numberColored := s.colorScheme.Colorize(number, "accent")
				titleColored := s.colorScheme.Colorize(title, "text")
				authorColored := s.colorScheme.Colorize(fmt.Sprintf("(%s, %s)", author, date), "secondary")

				normalLine := fmt.Sprintf("%s %s %s", numberColored, titleColored, authorColored)
				// Pad the line to maintain consistent spacing
				paddedLine := normalLine + strings.Repeat(" ", contentWidth-len(fmt.Sprintf("%s %s (%s, %s)", number, title, author, date)))
				session.term.Write([]byte(centerPadding + paddedLine + "\n"))
			}
		}

		// Instructions
		instructions := s.colorScheme.Colorize("\nUse ", "text") +
			s.colorScheme.Colorize("↑↓", "accent") +
			s.colorScheme.Colorize(" to navigate, ", "text") +
			s.colorScheme.Colorize("Enter", "accent") +
			s.colorScheme.Colorize(" to read, ", "text") +
			s.colorScheme.Colorize("Q", "accent") +
			s.colorScheme.Colorize(" to return", "text")

		centeredInstructions := s.colorScheme.CenterText(instructions, 79)
		session.term.Write([]byte("\n" + centeredInstructions))

		// Handle input using the same key reading as menus
		key, err := s.readKey(session)
		if err != nil {
			session.term.Write([]byte(ShowCursor))
			return true
		}

		switch key {
		case "up":
			selectedIndex--
			if selectedIndex < 0 {
				selectedIndex = len(bulletins) - 1
			}
		case "down":
			selectedIndex++
			if selectedIndex >= len(bulletins) {
				selectedIndex = 0
			}
		case "enter":
			// Show selected bulletin
			if selectedIndex >= 0 && selectedIndex < len(bulletins) {
				s.showSingleBulletin(session, &bulletins[selectedIndex])
			}
		case "quit", "q", "Q":
			session.term.Write([]byte(ShowCursor))
			return true
		}
	}
}

// showSingleBulletin displays a single bulletin
func (s *SSHServer) showSingleBulletin(session *Session, bulletin *database.Bulletin) {
	// Clear screen
	session.term.Write([]byte(ClearScreen + HideCursor))

	terminalWidth := 79
	contentWidth := 70
	centerOffset := (terminalWidth - contentWidth) / 2
	centerPadding := strings.Repeat(" ", centerOffset)

	// Header with bulletin title
	title := s.colorScheme.Colorize(bulletin.Title, "primary")
	centeredTitle := s.colorScheme.CenterText(title, terminalWidth)
	session.term.Write([]byte(centeredTitle + "\n"))

	// Separator
	separator := s.colorScheme.DrawSeparator(len(bulletin.Title), "═")
	centeredSeparator := s.colorScheme.CenterText(separator, terminalWidth)
	session.term.Write([]byte(centeredSeparator + "\n\n"))

	// Metadata
	author := s.colorScheme.Colorize(fmt.Sprintf("Author: %s", bulletin.Author), "accent")
	date := s.colorScheme.Colorize(fmt.Sprintf("Date: %s", bulletin.CreatedAt.Format("2006-01-02 15:04:05")), "secondary")

	session.term.Write([]byte(centerPadding + author + "\n"))
	session.term.Write([]byte(centerPadding + date + "\n\n"))

	// Bulletin body - word wrap to content width
	body := bulletin.Body
	lines := s.wrapText(body, contentWidth)

	for _, line := range lines {
		coloredLine := s.colorScheme.Colorize(line, "text")
		session.term.Write([]byte(centerPadding + coloredLine + "\n"))
	}

	// Footer prompt
	prompt := s.colorScheme.Colorize("\nPress any key to return to bulletin list...", "text")
	centeredPrompt := s.colorScheme.CenterText(prompt, terminalWidth)
	session.term.Write([]byte(centeredPrompt))

	// Wait for key press
	session.term.ReadLine()
	session.term.Write([]byte(ShowCursor))
}

// wrapText wraps text to specified width
func (s *SSHServer) wrapText(text string, width int) []string {
	var lines []string
	words := strings.Fields(text)

	if len(words) == 0 {
		return lines
	}

	currentLine := ""
	for _, word := range words {
		// Check if adding this word would exceed the width
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) <= width {
			currentLine = testLine
		} else {
			// Start a new line
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}

	// Add the last line
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

func (s *SSHServer) showBulletins(session *Session) bool {
	bulletins, err := s.db.GetBulletins(10)
	if err != nil {
		errorMsg := s.colorScheme.Colorize("Error retrieving bulletins.", "error")
		centeredError := s.colorScheme.CenterText(errorMsg, 79)
		session.term.Write([]byte(centeredError + "\n"))
		return true
	}

	// Colorized header (centered)
	header := s.colorScheme.Colorize("\n--- System Bulletins ---\n\n", "primary")
	centeredHeader := s.colorScheme.CenterText(header, 79)
	session.term.Write([]byte(centeredHeader))

	if len(bulletins) == 0 {
		noMsg := s.colorScheme.Colorize("No bulletins available.", "secondary")
		centeredNoMsg := s.colorScheme.CenterContainerLeftAlign(noMsg, 60, 79)
		session.term.Write([]byte(centeredNoMsg + "\n"))
	} else {
		// Define content container width (about 60 characters)
		contentWidth := 60

		for i, bulletin := range bulletins {
			number := s.colorScheme.Colorize(fmt.Sprintf("%d)", i+1), "accent")
			title := s.colorScheme.Colorize(bulletin.Title, "highlight")
			author := s.colorScheme.Colorize(fmt.Sprintf("by %s", bulletin.Author), "text")
			date := s.colorScheme.Colorize(bulletin.CreatedAt.Format("2006-01-02"), "secondary")

			bulletinLine := fmt.Sprintf("%s %s (%s, %s)",
				number, title, author, date)
			centeredBulletin := s.colorScheme.CenterContainerLeftAlign(bulletinLine, contentWidth, 79)
			session.term.Write([]byte(centeredBulletin + "\n"))
		}
	}

	// Colorized prompt (centered)
	prompt := s.colorScheme.Colorize("\nPress Enter to continue...", "text")
	centeredPrompt := s.colorScheme.CenterText(prompt, 79)
	session.term.Write([]byte(centeredPrompt))
	session.term.ReadLine()
	return true
}

func (s *SSHServer) showMessages(session *Session) bool {
	if session.user == nil {
		errorMsg := s.colorScheme.Colorize("You must be logged in to read messages.", "error")
		centeredError := s.colorScheme.CenterText(errorMsg, 79)
		session.term.Write([]byte(centeredError + "\n"))
		return true
	}

	messages, err := s.db.GetMessages(session.user.Username, 10)
	if err != nil {
		errorMsg := s.colorScheme.Colorize("Error retrieving messages.", "error")
		centeredError := s.colorScheme.CenterText(errorMsg, 79)
		session.term.Write([]byte(centeredError + "\n"))
		return true
	}

	// Colorized header (centered)
	header := s.colorScheme.Colorize("\n--- Your Messages ---\n\n", "primary")
	centeredHeader := s.colorScheme.CenterText(header, 79)
	session.term.Write([]byte(centeredHeader))

	if len(messages) == 0 {
		noMsg := s.colorScheme.Colorize("No messages for you.", "secondary")
		centeredNoMsg := s.colorScheme.CenterContainerLeftAlign(noMsg, 60, 79)
		session.term.Write([]byte(centeredNoMsg + "\n"))
	} else {
		// Define content container width
		contentWidth := 60

		for i, msg := range messages {
			number := s.colorScheme.Colorize(fmt.Sprintf("%d)", i+1), "accent")
			subject := s.colorScheme.Colorize(msg.Subject, "highlight")
			fromUser := s.colorScheme.Colorize(fmt.Sprintf("from %s", msg.FromUser), "text")

			var status string
			var statusColor string
			if msg.IsRead {
				status = "READ"
				statusColor = "secondary"
			} else {
				status = "NEW"
				statusColor = "success"
			}
			statusColored := s.colorScheme.Colorize(fmt.Sprintf("[%s]", status), statusColor)
			date := s.colorScheme.Colorize(msg.CreatedAt.Format("2006-01-02"), "secondary")

			messageLine := fmt.Sprintf("%s %s %s %s (%s)",
				number, subject, fromUser, statusColored, date)
			centeredMessage := s.colorScheme.CenterContainerLeftAlign(messageLine, contentWidth, 79)
			session.term.Write([]byte(centeredMessage + "\n"))
		}
	}

	// Colorized prompt (centered)
	prompt := s.colorScheme.Colorize("\nPress Enter to continue...", "text")
	centeredPrompt := s.colorScheme.CenterText(prompt, 79)
	session.term.Write([]byte(centeredPrompt))
	session.term.ReadLine()
	return true
}
