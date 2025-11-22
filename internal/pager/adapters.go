package pager

// TerminalSizerFromWriter attempts to get terminal dimensions from a writer
// by using type assertions to find a Size() method
type TerminalSizerFromWriter struct {
	writer Writer
}

// NewTerminalSizerFromWriter creates a terminal sizer from a writer
func NewTerminalSizerFromWriter(writer Writer) *TerminalSizerFromWriter {
	return &TerminalSizerFromWriter{writer: writer}
}

// Size attempts to get terminal dimensions from the writer
func (t *TerminalSizerFromWriter) Size() (width, height int, err error) {
	// Try type assertion to get terminal size
	type terminalSizer interface {
		Size() (int, int, error)
	}

	// Check if writer implements a terminal sizer interface
	if sizer, ok := t.writer.(terminalSizer); ok {
		return sizer.Size()
	}

	// Try to get session and terminal from writer
	type sessionHolder interface {
		GetSession() interface{}
	}

	if sh, ok := t.writer.(sessionHolder); ok {
		session := sh.GetSession()
		if termGetter, ok := session.(interface{ GetTerminal() interface{} }); ok {
			term := termGetter.GetTerminal()
			if sizer, ok := term.(terminalSizer); ok {
				return sizer.Size()
			}
		}
	}

	// Fallback to standard dimensions
	return 80, 24, nil
}

// WriterAdapter wraps a Writer and provides additional functionality
type WriterAdapter struct {
	Writer
	underlyingTerminalSizer TerminalSizer // Reference to get real terminal dimensions
	StatusBarMgr            StatusBarManager // Exported for external access
}

// NewWriterAdapter creates a new WriterAdapter
func NewWriterAdapter(writer Writer, terminalSizer TerminalSizer) *WriterAdapter {
	return &WriterAdapter{
		Writer:                  writer,
		underlyingTerminalSizer: terminalSizer,
		StatusBarMgr:            nil,
	}
}

// WithStatusBarManager adds status bar management to the adapter
func (w *WriterAdapter) WithStatusBarManager(mgr StatusBarManager) *WriterAdapter {
	w.StatusBarMgr = mgr
	return w
}

// Size returns the terminal dimensions from the underlying terminal sizer
func (w *WriterAdapter) Size() (width, height int, err error) {
	if w.underlyingTerminalSizer != nil {
		return w.underlyingTerminalSizer.Size()
	}
	// Fallback to standard dimensions
	return 80, 24, nil
}

// Pause pauses status bar updates (implements StatusBarManager)
func (w *WriterAdapter) Pause() {
	if w.StatusBarMgr != nil {
		w.StatusBarMgr.Pause()
	}
}

// Resume resumes status bar updates (implements StatusBarManager)
func (w *WriterAdapter) Resume() {
	if w.StatusBarMgr != nil {
		w.StatusBarMgr.Resume()
	}
}
