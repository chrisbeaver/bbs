package terminal

import (
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// SSHTerminal wraps an SSH channel to implement the Terminal interface
type SSHTerminal struct {
	channel  ssh.Channel
	terminal *term.Terminal
}

// NewSSHTerminal creates a new SSH terminal wrapper
func NewSSHTerminal(channel ssh.Channel) *SSHTerminal {
	terminal := term.NewTerminal(channel, "")
	return &SSHTerminal{
		channel:  channel,
		terminal: terminal,
	}
}

func (t *SSHTerminal) Read(p []byte) (n int, err error) {
	return t.channel.Read(p)
}

func (t *SSHTerminal) Write(p []byte) (n int, err error) {
	return t.channel.Write(p)
}

func (t *SSHTerminal) SetSize(width int, height int) error {
	// SSH terminal size is managed by the SSH protocol
	return nil
}

func (t *SSHTerminal) Size() (width int, height int, error error) {
	// For SSH terminals, we'll use a default size
	// Terminal size should be handled via SSH window-change requests
	return 80, 24, nil
}

func (t *SSHTerminal) MakeRaw() error {
	// SSH channels are already in raw mode
	return nil
}

func (t *SSHTerminal) Restore() error {
	// No restoration needed for SSH
	return nil
}

func (t *SSHTerminal) Close() error {
	return t.channel.Close()
}

func (t *SSHTerminal) ReadLine() (string, error) {
	return t.terminal.ReadLine()
}

func (t *SSHTerminal) SetPrompt(prompt string) {
	t.terminal.SetPrompt(prompt)
}

// GetTerminal returns the underlying term.Terminal for compatibility
func (t *SSHTerminal) GetTerminal() *term.Terminal {
	return t.terminal
}
