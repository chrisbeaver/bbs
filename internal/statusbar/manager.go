package statusbar

import (
	"fmt"
	"sync"
	"time"

	"bbs/internal/config"
)

// Manager handles status bar updates and rendering for terminal sessions
type Manager struct {
	statusBar      *StatusBar
	terminalHeight int
	mu             sync.RWMutex
	updateTicker   *time.Ticker
	stopChan       chan bool
	isInitialized  bool
}

// NewManager creates a new status bar manager
func NewManager(username string, cfg *config.Config, terminalHeight int) *Manager {
	return &Manager{
		statusBar:      New(username, cfg),
		terminalHeight: terminalHeight,
		stopChan:       make(chan bool),
		isInitialized:  false,
	}
}

// Start begins automatic status bar updates
func (m *Manager) Start(updateInterval time.Duration) <-chan string {
	m.mu.Lock()
	defer m.mu.Unlock()

	updateChan := make(chan string, 1)

	go func() {
		defer close(updateChan)

		// Send initial fixed setup ONLY
		if !m.isInitialized {
			statusBar := m.statusBar.InitializeFixed(m.terminalHeight)
			m.isInitialized = true
			updateChan <- statusBar
		}

		// Start timer updates for just the timer portion
		m.updateTicker = time.NewTicker(updateInterval)
		defer m.updateTicker.Stop()

		for {
			select {
			case <-m.updateTicker.C:
				// Only update the timer portion to avoid flicker
				timerUpdate := m.getTimerUpdate()
				if timerUpdate != "" {
					updateChan <- timerUpdate
				}
			case <-m.stopChan:
				return
			}
		}
	}()

	return updateChan
} // Stop stops automatic status bar updates
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.updateTicker != nil {
		m.updateTicker.Stop()
		m.updateTicker = nil
	}

	select {
	case m.stopChan <- true:
	default:
	}
}

// RenderNow returns the current status bar immediately
func (m *Manager) RenderNow() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.renderStatusBar()
}

// SetTerminalHeight updates the terminal height for positioning
func (m *Manager) SetTerminalHeight(height int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.terminalHeight = height
}

// SetActive enables or disables the status bar
func (m *Manager) SetActive(active bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statusBar.SetActive(active)
}

// GetContentHeight returns the available height for content (excluding status bar)
func (m *Manager) GetContentHeight() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.statusBar.GetContentHeight()
}

// Clear clears the status bar from the terminal
func (m *Manager) Clear() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.statusBar.Clear(m.terminalHeight)
}

// RenderAtPosition renders the status bar at the specified terminal height
func (m *Manager) RenderAtPosition(terminalHeight int) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Update terminal height in case it changed
	m.terminalHeight = terminalHeight

	// Position cursor at status bar line and render status bar
	positionCode := fmt.Sprintf("\033[%d;1H", terminalHeight)
	statusBarContent := m.statusBar.Render()

	return positionCode + statusBarContent
}

// RenderContent returns just the status bar content without positioning
func (m *Manager) RenderContent() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.statusBar.Render()
}

// getTimerUpdate returns just a timer update without repositioning the entire status bar
func (m *Manager) getTimerUpdate() string {
	if !m.isInitialized {
		return ""
	}

	// Get the current timer string
	durationStr := m.statusBar.GetTimerString()

	// Position cursor at the timer location (right side of status bar line)
	// The timer format is always "HH:MM:SS " (9 characters including space)
	timerStartCol := m.statusBar.GetWidth() - 9
	positionCode := fmt.Sprintf("\033[%d;%dH", m.terminalHeight, timerStartCol)

	// ANSI color codes for timer (bright yellow with blue background to match status bar)
	blue := "\033[44m"
	brightYellow := "\033[93m"
	reset := "\033[0m"

	// Clear the timer area by writing spaces, then reposition and write new timer
	clearSpaces := "         " // 9 spaces to clear the timer area
	repositionCode := fmt.Sprintf("\033[%d;%dH", m.terminalHeight, timerStartCol)

	return fmt.Sprintf("%s%s%s%s%s%s%s %s",
		positionCode, blue, clearSpaces,
		repositionCode, blue, brightYellow, durationStr, reset)
} // renderStatusBar generates the positioned status bar
func (m *Manager) renderStatusBar() string {
	positionCode := m.statusBar.GetPositionCode(m.terminalHeight)
	statusBarContent := m.statusBar.Render()
	// Return cursor to previous position after drawing status bar
	restoreCursor := "\033[u"
	saveCursor := "\033[s"

	return fmt.Sprintf("%s%s%s%s", saveCursor, positionCode, statusBarContent, restoreCursor)
}
