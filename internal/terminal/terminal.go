package terminal

import "io"

// Terminal represents either an SSH connection or local terminal
type Terminal interface {
	io.ReadWriter
	SetSize(width int, height int) error
	Size() (width int, height int, error error)
	MakeRaw() error
	Restore() error
	Close() error
	ReadLine() (string, error)
	SetPrompt(prompt string)
}
