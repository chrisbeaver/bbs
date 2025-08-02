package server

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	"bbs/internal/config"
	"bbs/internal/database"
	"bbs/internal/menu"
	"bbs/internal/terminal"
)

// Server represents a unified BBS server that can handle both SSH and local connections
type Server struct {
	config      *config.Config
	db          *database.DB
	colorScheme *ColorScheme
	sshConfig   *ssh.ServerConfig
}

// NewServer creates a new unified server
func NewServer(cfg *config.Config, db *database.DB) *Server {
	server := &Server{
		config:      cfg,
		db:          db,
		colorScheme: NewColorScheme(&cfg.BBS.Colors),
	}
	server.setupSSHConfig()
	return server
}

// setupSSHConfig configures SSH server settings
func (s *Server) setupSSHConfig() {
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

// passwordCallback handles SSH password authentication
func (s *Server) passwordCallback(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
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

// NewSession creates a new unified session
func (s *Server) NewSession(term terminal.Terminal, prefilledUsername string) *Session {
	session := &Session{
		terminal:          term,
		db:                s.db,
		config:            s.config,
		currentMenu:       "main",
		selectedIndex:     0,
		authenticated:     false,
		colorScheme:       s.colorScheme,
		prefilledUsername: prefilledUsername,
	}

	// Initialize the TerminalWriter for this session
	session.writer = &TerminalWriter{
		session:             session,
		lastStatusBarRedraw: time.Time{}, // Initialize to zero time
		pendingRedraw:       false,
	}

	// Initialize the MenuRenderer
	session.menuRenderer = menu.NewMenuRenderer(s.colorScheme, session.writer)

	return session
}

// NewLocalSession creates a session for local terminal access
func (s *Server) NewLocalSession(term terminal.Terminal) *Session {
	return s.NewSession(term, "") // No prefilled username for local
}

// HandleConnection handles SSH connections
func (s *Server) HandleConnection(netConn net.Conn) {
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

		// Get username from SSH authentication
		username := ""
		if user, ok := sshConn.Permissions.Extensions["username"]; ok {
			username = user
		}

		// Create SSH terminal interface
		sshTerm := terminal.NewSSHTerminal(channel)

		// Create unified session
		session := s.NewSession(sshTerm, username)

		go s.handleSSHSession(session, channel, requests)
	}
}

// handleSSHSession handles the SSH session setup and delegates to unified session
func (s *Server) handleSSHSession(session *Session, channel ssh.Channel, requests <-chan *ssh.Request) {
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

	// Run the unified session
	session.Run()
}

// TerminalWriter adapts session to Writer interface for modules
type TerminalWriter struct {
	session              *Session
	lastStatusBarRedraw  time.Time
	statusBarRedrawMutex sync.Mutex
	pendingRedraw        bool
}

func (w *TerminalWriter) Write(data []byte) (int, error) {
	// For SSH terminals, use the underlying term.Terminal for proper ANSI handling
	if sshTerm, ok := w.session.terminal.(*terminal.SSHTerminal); ok {
		terminalInstance := sshTerm.GetTerminal()
		n, err := terminalInstance.Write(data)
		// After any write, redraw status bar if screen was cleared
		w.handleStatusBarRedraw(data)
		return n, err
	}

	// For local terminals, also use term.Terminal for consistent ANSI processing
	if localTerm, ok := w.session.terminal.(*terminal.LocalTerminal); ok {
		terminalInstance := localTerm.GetTerminal()
		n, err := terminalInstance.Write(data)
		// After any write, redraw status bar if screen was cleared
		w.handleStatusBarRedraw(data)
		return n, err
	}

	// Fallback to direct write
	n, err := w.session.terminal.Write(data)
	// After any write, redraw status bar if screen was cleared
	w.handleStatusBarRedraw(data)
	return n, err
}

// handleStatusBarRedraw checks if screen was cleared and redraws status bar if needed
func (w *TerminalWriter) handleStatusBarRedraw(data []byte) {
	dataStr := string(data)

	// Check if screen was cleared (full screen clear or content area clear) and we have a status bar
	if (strings.Contains(dataStr, "\033[2J") || strings.Contains(dataStr, "\033[H\033[0J")) && w.session.statusBar != nil {
		w.statusBarRedrawMutex.Lock()
		defer w.statusBarRedrawMutex.Unlock()

		// Debounce rapid screen clears - only allow one redraw per 100ms
		now := time.Now()
		if now.Sub(w.lastStatusBarRedraw) < 100*time.Millisecond {
			// If there's already a pending redraw or recent redraw, skip this one
			if !w.pendingRedraw {
				w.pendingRedraw = true
				// Schedule a delayed redraw
				go func() {
					time.Sleep(100 * time.Millisecond)
					w.statusBarRedrawMutex.Lock()
					w.pendingRedraw = false
					w.lastStatusBarRedraw = time.Now()
					w.statusBarRedrawMutex.Unlock()
					w.doStatusBarRedraw()
				}()
			}
			return
		}

		// Immediate redraw if enough time has passed
		w.lastStatusBarRedraw = now
		w.doStatusBarRedraw()
	}
}

// doStatusBarRedraw performs the actual status bar redraw
func (w *TerminalWriter) doStatusBarRedraw() {
	// Get terminal height for proper positioning
	_, height, err := w.session.terminal.Size()
	if err != nil {
		height = 24 // Default height
	}

	// Save current cursor position
	saveCursor := "\033[s"

	// Position cursor at bottom line and clear the line
	positionCode := fmt.Sprintf("\033[%d;1H\033[2K", height)

	// Get status bar content without positioning
	statusBarContent := w.session.statusBar.RenderContent()

	// Restore cursor position
	restoreCursor := "\033[u"

	// Combine all the positioning and content
	statusBarOutput := saveCursor + positionCode + statusBarContent + restoreCursor

	// Write status bar directly to terminal (avoid recursion)
	if sshTerm, ok := w.session.terminal.(*terminal.SSHTerminal); ok {
		terminalInstance := sshTerm.GetTerminal()
		terminalInstance.Write([]byte(statusBarOutput))
	} else if localTerm, ok := w.session.terminal.(*terminal.LocalTerminal); ok {
		terminalInstance := localTerm.GetTerminal()
		terminalInstance.Write([]byte(statusBarOutput))
	} else {
		w.session.terminal.Write([]byte(statusBarOutput))
	}
}

// TerminalKeyReader adapts session to KeyReader interface for modules
type TerminalKeyReader struct {
	session *Session
}

func (r *TerminalKeyReader) ReadKey() (string, error) {
	return r.session.readKey()
}
