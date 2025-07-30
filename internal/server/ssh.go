package server

import (
	"fmt"
	"net"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"

	"bbs/internal/config"
	"bbs/internal/database"
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
		for {
			key, err := s.readKey(session)
			if err != nil {
				return
			}

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
				// Don't reset selection - keep it at the current position
				break

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

			// Break out of navigation loop to redisplay menu
			if key == "enter" {
				break
			}
		}
	}
}

func (s *SSHServer) displayMenu(session *Session, menu *config.MenuItem) {
	// Clear screen and hide cursor
	session.term.Write([]byte(ClearScreen + HideCursor))

	// Menu title with color
	title := s.colorScheme.Colorize(menu.Title, "primary")
	session.term.Write([]byte(fmt.Sprintf("%s\n", title)))

	// Decorative separator
	separator := s.colorScheme.DrawSeparator(len(menu.Title), "═")
	session.term.Write([]byte(separator + "\n\n"))

	// Build accessible menu items (filter by access level)
	var accessibleItems []config.MenuItem
	for _, item := range menu.Submenu {
		if session.user == nil || item.AccessLevel <= session.user.AccessLevel {
			accessibleItems = append(accessibleItems, item)
		}
	}

	// Ensure selected index is valid
	if session.selectedIndex >= len(accessibleItems) {
		session.selectedIndex = 0
	}
	if session.selectedIndex < 0 {
		session.selectedIndex = len(accessibleItems) - 1
	}

	// Display menu items with highlighting
	for i, item := range accessibleItems {
		selected := (i == session.selectedIndex)
		menuLine := s.colorScheme.HighlightSelection(item.Description, selected)
		session.term.Write([]byte(menuLine + "\n"))
	}

	// Instructions
	instructions := "\n" +
		s.colorScheme.Colorize("Use ", "text") +
		s.colorScheme.Colorize("↑↓", "accent") +
		s.colorScheme.Colorize(" arrow keys to navigate, ", "text") +
		s.colorScheme.Colorize("Enter", "accent") +
		s.colorScheme.Colorize(" to select, ", "text") +
		s.colorScheme.Colorize("Q", "accent") +
		s.colorScheme.Colorize(" to quit", "text")

	session.term.Write([]byte(instructions))
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

func (s *SSHServer) executeCommand(session *Session, item *config.MenuItem) bool {
	switch item.Command {
	case "bulletins":
		return s.showBulletins(session)
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
	case "sysop":
		msg := s.colorScheme.Colorize("Sysop menu not yet implemented.", "secondary")
		session.term.Write([]byte(msg + "\n"))
		return true
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

func (s *SSHServer) showBulletins(session *Session) bool {
	bulletins, err := s.db.GetBulletins(10)
	if err != nil {
		errorMsg := s.colorScheme.Colorize("Error retrieving bulletins.", "error")
		session.term.Write([]byte(errorMsg + "\n"))
		return true
	}

	// Colorized header
	header := s.colorScheme.Colorize("\n--- System Bulletins ---\n\n", "primary")
	session.term.Write([]byte(header))

	if len(bulletins) == 0 {
		noMsg := s.colorScheme.Colorize("No bulletins available.", "secondary")
		session.term.Write([]byte(noMsg + "\n"))
	} else {
		for i, bulletin := range bulletins {
			number := s.colorScheme.Colorize(fmt.Sprintf("%d)", i+1), "accent")
			title := s.colorScheme.Colorize(bulletin.Title, "highlight")
			author := s.colorScheme.Colorize(fmt.Sprintf("by %s", bulletin.Author), "text")
			date := s.colorScheme.Colorize(bulletin.CreatedAt.Format("2006-01-02"), "secondary")

			session.term.Write([]byte(fmt.Sprintf("%s %s (%s, %s)\n",
				number, title, author, date)))
		}
	}

	// Colorized prompt
	prompt := s.colorScheme.Colorize("\nPress Enter to continue...", "text")
	session.term.Write([]byte(prompt))
	session.term.ReadLine()
	return true
}

func (s *SSHServer) showMessages(session *Session) bool {
	if session.user == nil {
		errorMsg := s.colorScheme.Colorize("You must be logged in to read messages.", "error")
		session.term.Write([]byte(errorMsg + "\n"))
		return true
	}

	messages, err := s.db.GetMessages(session.user.Username, 10)
	if err != nil {
		errorMsg := s.colorScheme.Colorize("Error retrieving messages.", "error")
		session.term.Write([]byte(errorMsg + "\n"))
		return true
	}

	// Colorized header
	header := s.colorScheme.Colorize("\n--- Your Messages ---\n\n", "primary")
	session.term.Write([]byte(header))

	if len(messages) == 0 {
		noMsg := s.colorScheme.Colorize("No messages for you.", "secondary")
		session.term.Write([]byte(noMsg + "\n"))
	} else {
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

			session.term.Write([]byte(fmt.Sprintf("%s %s %s %s (%s)\n",
				number, subject, fromUser, statusColored, date)))
		}
	}

	// Colorized prompt
	prompt := s.colorScheme.Colorize("\nPress Enter to continue...", "text")
	session.term.Write([]byte(prompt))
	session.term.ReadLine()
	return true
}
