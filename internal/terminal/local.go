package terminal

import (
	"io"
	"os"

	"golang.org/x/term"
)

// LocalTerminal implements Terminal interface for local console access
type LocalTerminal struct {
	stdin    *os.File
	stdout   *os.File
	oldState *term.State
	terminal *term.Terminal
	rawMode  bool
}

// NewLocalTerminal creates a new local terminal
func NewLocalTerminal() *LocalTerminal {
	return &LocalTerminal{
		stdin:   os.Stdin,
		stdout:  os.Stdout,
		rawMode: false,
	}
}

func (t *LocalTerminal) Read(p []byte) (n int, err error) {
	return t.stdin.Read(p)
}

func (t *LocalTerminal) Write(p []byte) (n int, err error) {
	// Use term.Terminal.Write() for consistent formatting with SSH connections
	if t.terminal == nil {
		// Create a ReadWriter that combines stdin and stdout
		rw := struct {
			io.Reader
			io.Writer
		}{t.stdin, t.stdout}
		t.terminal = term.NewTerminal(rw, "")
	}
	return t.terminal.Write(p)
}

func (t *LocalTerminal) Size() (width int, height int, error error) {
	w, h, err := term.GetSize(int(t.stdin.Fd()))
	if err != nil {
		return 80, 24, err
	}
	return w, h, nil
}

func (t *LocalTerminal) MakeRaw() error {
	if !term.IsTerminal(int(t.stdin.Fd())) {
		return nil // Not a terminal, no raw mode needed
	}

	state, err := term.MakeRaw(int(t.stdin.Fd()))
	if err != nil {
		return err
	}
	t.oldState = state
	t.rawMode = true
	return nil
}

func (t *LocalTerminal) Restore() error {
	if t.oldState != nil && term.IsTerminal(int(t.stdin.Fd())) && t.rawMode {
		err := term.Restore(int(t.stdin.Fd()), t.oldState)
		t.rawMode = false
		return err
	}
	return nil
}

func (t *LocalTerminal) SetSize(width int, height int) error {
	// Local terminal size is controlled by the terminal emulator
	return nil
}

func (t *LocalTerminal) Close() error {
	return t.Restore()
}

func (t *LocalTerminal) ReadLine() (string, error) {
	// Ensure we're not in raw mode for line reading
	if t.rawMode {
		t.Restore()
	}

	// Create terminal for reading lines when not in raw mode
	if t.terminal == nil {
		t.terminal = term.NewTerminal(os.Stdin, "")
	}
	return t.terminal.ReadLine()
}

func (t *LocalTerminal) SetPrompt(prompt string) {
	// Ensure we're not in raw mode for prompts
	if t.rawMode {
		t.Restore()
	}

	// Create terminal for setting prompts when not in raw mode
	if t.terminal == nil {
		t.terminal = term.NewTerminal(os.Stdin, "")
	}
	t.terminal.SetPrompt(prompt)
}

// GetTerminal returns the underlying term.Terminal for compatibility
func (t *LocalTerminal) GetTerminal() *term.Terminal {
	if t.terminal == nil {
		t.terminal = term.NewTerminal(os.Stdin, "")
	}
	return t.terminal
}

// IsRaw returns whether the terminal is in raw mode
func (t *LocalTerminal) IsRaw() bool {
	return t.rawMode
}
