package pager

// Writer interface for output operations
type Writer interface {
	Write([]byte) (int, error)
}

// KeyReader interface for reading user input
type KeyReader interface {
	ReadKey() (string, error)
}

// ColorScheme interface for colorizing text
type ColorScheme interface {
	Colorize(text string, style string) string
	CenterText(text string, width int) string
}

// TerminalSizer interface for getting terminal dimensions
type TerminalSizer interface {
	Size() (width, height int, err error)
}

// StatusBarManager interface for controlling status bar updates (optional)
type StatusBarManager interface {
	Pause()
	Resume()
}
