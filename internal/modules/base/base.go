package base

import (
	"strconv"

	"bbs/internal/database"
	"bbs/internal/menu"
	"bbs/internal/modules"
)

// MenuOption represents a generic menu option that can be executed
type MenuOption interface {
	GetID() string
	GetTitle() string
	GetDescription() string
	Execute(writer modules.Writer, keyReader modules.KeyReader, db *database.DB, colorScheme menu.ColorScheme) bool
}

// OptionProvider defines how modules provide their menu options
type OptionProvider interface {
	LoadOptions(db *database.DB) ([]MenuOption, error)
	GetMenuTitle() string
	GetInstructions() string
}

// Module provides common functionality for all menu-based modules
type Module struct {
	db            *database.DB
	colorScheme   menu.ColorScheme
	provider      OptionProvider
	options       []MenuOption
	selectedIndex int
	menuRenderer  *menu.MenuRenderer
}

// GetMenuTitle implements MenuProvider interface
func (m *Module) GetMenuTitle() string {
	return m.provider.GetMenuTitle()
}

// GetInstructions implements MenuProvider interface
func (m *Module) GetInstructions() string {
	return m.provider.GetInstructions()
}

// GetMenuItems implements MenuProvider interface
func (m *Module) GetMenuItems() []menu.MenuItem {
	var items []menu.MenuItem
	for _, option := range m.options {
		items = append(items, menu.MenuItem{
			ID:          option.GetID(),
			Description: option.GetDescription(),
			Data:        option,
		})
	}
	return items
}

// NewModule creates a new base module
func NewModule(db *database.DB, colorScheme menu.ColorScheme, provider OptionProvider) *Module {
	return &Module{
		db:           db,
		colorScheme:  colorScheme,
		provider:     provider,
		menuRenderer: menu.NewMenuRenderer(colorScheme, nil),
	}
}

// Execute runs the module using the unified menu system
func (m *Module) Execute(writer modules.Writer, keyReader modules.KeyReader) bool {
	m.menuRenderer = menu.NewMenuRenderer(m.colorScheme, writer)
	m.selectedIndex = 0

	// Load options from provider
	options, err := m.provider.LoadOptions(m.db)
	if err != nil {
		errorMsg := m.colorScheme.Colorize("Error loading menu options.", "error")
		centeredError := m.colorScheme.CenterText(errorMsg, 79)
		writer.Write([]byte(centeredError + "\n"))
		return true
	}

	m.options = options

	if len(m.options) == 0 {
		m.showEmptyMessage(writer, keyReader)
		return true
	}

	// Main menu loop
	for {
		m.renderMenu(writer)

		key, err := keyReader.ReadKey()
		if err != nil {
			break
		}

		if !m.handleKey(key, writer, keyReader) {
			break
		}
	}

	writer.Write([]byte(menu.ShowCursor))
	return true
}

// renderMenu renders the current menu state
func (m *Module) renderMenu(writer modules.Writer) {
	// Use the unified menu renderer instead of custom rendering
	m.menuRenderer.RenderModuleMenu(m, m.selectedIndex)
}

// handleKey processes keyboard input
func (m *Module) handleKey(key string, writer modules.Writer, keyReader modules.KeyReader) bool {
	switch key {
	case "up":
		m.selectedIndex--
		if m.selectedIndex < 0 {
			m.selectedIndex = len(m.options) - 1
		}
	case "down":
		m.selectedIndex++
		if m.selectedIndex >= len(m.options) {
			m.selectedIndex = 0
		}
	case "enter":
		option := m.options[m.selectedIndex]
		return option.Execute(writer, keyReader, m.db, m.colorScheme)
	case "q", "Q", "quit":
		return false
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		if num, err := strconv.Atoi(key); err == nil && num >= 1 && num <= len(m.options) {
			option := m.options[num-1]
			return option.Execute(writer, keyReader, m.db, m.colorScheme)
		}
	}
	return true
}

// showEmptyMessage displays a message when no items are available
func (m *Module) showEmptyMessage(writer modules.Writer, keyReader modules.KeyReader) {
	writer.Write([]byte(menu.ClearContentArea))
	header := m.colorScheme.Colorize("--- "+m.provider.GetMenuTitle()+" ---", "primary")
	centeredHeader := m.colorScheme.CenterText(header, 79)
	writer.Write([]byte(centeredHeader + "\n\n"))

	noMsg := m.colorScheme.Colorize("No items available.", "secondary")
	centeredNoMsg := m.colorScheme.CenterText(noMsg, 79)
	writer.Write([]byte(centeredNoMsg + "\n\n"))

	prompt := m.colorScheme.Colorize("Press any key to continue...", "text")
	centeredPrompt := m.colorScheme.CenterText(prompt, 79)
	writer.Write([]byte(centeredPrompt))

	keyReader.ReadKey()
}
