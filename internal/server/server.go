package server

import (
	"fmt"
	"net"

	"golang.org/x/crypto/ssh"

	"bbs/internal/config"
	"bbs/internal/database"
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
	session.writer = &TerminalWriter{session: session}

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
	session *Session
}

func (w *TerminalWriter) Write(data []byte) (int, error) {
	// For SSH terminals, use the underlying term.Terminal for proper ANSI handling
	if sshTerm, ok := w.session.terminal.(*terminal.SSHTerminal); ok {
		terminalInstance := sshTerm.GetTerminal()
		return terminalInstance.Write(data)
	}

	// For local terminals, also use term.Terminal for consistent ANSI processing
	if localTerm, ok := w.session.terminal.(*terminal.LocalTerminal); ok {
		terminalInstance := localTerm.GetTerminal()
		return terminalInstance.Write(data)
	}

	// Fallback to direct write
	return w.session.terminal.Write(data)
}

// TerminalKeyReader adapts session to KeyReader interface for modules
type TerminalKeyReader struct {
	session *Session
}

func (r *TerminalKeyReader) ReadKey() (string, error) {
	return r.session.readKey()
}
