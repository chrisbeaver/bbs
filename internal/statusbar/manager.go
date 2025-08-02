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
}

// NewManager creates a new status bar manager
func NewManager(username string, cfg *config.Config, terminalHeight int) *Manager {
	return &Manager{
		statusBar:      New(username, cfg),
		terminalHeight: terminalHeight,
		stopChan:       make(chan bool),
	}
}

// Start begins automatic status bar updates
func (m *Manager) Start(updateInterval time.Duration) <-chan string {
	m.mu.Lock()
	defer m.mu.Unlock()

	updateChan := make(chan string, 1)

	if m.updateTicker != nil {
		m.updateTicker.Stop()
	}

	m.updateTicker = time.NewTicker(updateInterval)

	go func() {
		defer close(updateChan)

		// Send initial render
		updateChan <- m.renderStatusBar()

		for {
			select {
			case <-m.updateTicker.C:
				select {
				case updateChan <- m.renderStatusBar():
				default:
					// Channel is full, skip this update
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

// Clear clears the status bar from the terminal
func (m *Manager) Clear() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.statusBar.Clear(m.terminalHeight)
}

// renderStatusBar generates the positioned status bar
func (m *Manager) renderStatusBar() string {
	positionCode := m.statusBar.GetPositionCode(m.terminalHeight)
	statusBarContent := m.statusBar.Render()
	// Return cursor to previous position after drawing status bar
	restoreCursor := "\033[u"
	saveCursor := "\033[s"

	return fmt.Sprintf("%s%s%s%s", saveCursor, positionCode, statusBarContent, restoreCursor)
}
