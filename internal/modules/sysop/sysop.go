package sysop

import (
	"strconv"

	"bbs/internal/database"
	"bbs/internal/menu"
)

// KeyReader interface for reading keys
type KeyReader interface {
	ReadKey() (string, error)
}

// Writer interface for writing output
type Writer interface {
	Write([]byte) (int, error)
}

// SysopOption represents a sysop menu option
type SysopOption struct {
	ID          string
	Title       string
	Description string
	Handler     func(writer Writer, keyReader KeyReader, db *database.DB, colorScheme menu.ColorScheme) bool
}

// Module implements the sysop functionality using the unified menu system
type Module struct {
	db            *database.DB
	colorScheme   menu.ColorScheme
	options       []SysopOption
	selectedIndex int
}

// NewModule creates a new sysop module
func NewModule(db *database.DB, colorScheme menu.ColorScheme) *Module {
	options := []SysopOption{
		{
			ID:          "create_user",
			Title:       "Create New User",
			Description: "1) Create New User Account",
			Handler:     handleCreateUser,
		},
		{
			ID:          "edit_user",
			Title:       "Edit User Account",
			Description: "2) Edit User Account",
			Handler:     handleEditUser,
		},
		{
			ID:          "delete_user",
			Title:       "Delete User Account",
			Description: "3) Delete User Account",
			Handler:     handleDeleteUser,
		},
		{
			ID:          "view_users",
			Title:       "View All Users",
			Description: "4) View All Users",
			Handler:     handleViewUsers,
		},
		{
			ID:          "user_password",
			Title:       "Change User Password",
			Description: "5) Change User Password",
			Handler:     handleChangePassword,
		},
		{
			ID:          "toggle_user",
			Title:       "Toggle User Status",
			Description: "6) Toggle User Active Status",
			Handler:     handleToggleUserStatus,
		},
		{
			ID:          "system_stats",
			Title:       "System Statistics",
			Description: "7) System Statistics",
			Handler:     handleSystemStats,
		},
		{
			ID:          "bulletin_editor",
			Title:       "Bulletin Management",
			Description: "8) Bulletin Management",
			Handler:     handleBulletinManagement,
		},
	}

	return &Module{
		db:          db,
		colorScheme: colorScheme,
		options:     options,
	}
}

// GetMenuTitle implements MenuProvider interface
func (m *Module) GetMenuTitle() string {
	return "System Administration"
}

// GetMenuItems implements MenuProvider interface
func (m *Module) GetMenuItems() []menu.MenuItem {
	var items []menu.MenuItem
	for i, option := range m.options {
		items = append(items, menu.MenuItem{
			ID:          option.ID,
			Description: option.Description,
			Data:        &m.options[i], // Store option reference
		})
	}
	return items
}

// GetInstructions implements MenuProvider interface
func (m *Module) GetInstructions() string {
	return "Use ↑↓ arrow keys to navigate, Enter to select, Q to quit"
}

// Execute runs the sysop module using the unified menu system
func (m *Module) Execute(writer Writer, keyReader KeyReader) bool {
	// Initialize menu renderer
	menuRenderer := menu.NewMenuRenderer(m.colorScheme, writer)
	m.selectedIndex = 0

	// Main menu loop using unified menu system
	for {
		// Render menu using unified renderer
		menuRenderer.RenderModuleMenu(m, m.selectedIndex)

		// Get key input
		key, err := keyReader.ReadKey()
		if err != nil {
			break
		}

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
			// Execute selected option
			option := &m.options[m.selectedIndex]
			if !option.Handler(writer, keyReader, m.db, m.colorScheme) {
				return false // User wants to exit completely
			}
		case "q", "Q", "quit":
			writer.Write([]byte(menu.ShowCursor))
			return true
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			// Direct number selection
			if num, err := strconv.Atoi(key); err == nil && num >= 1 && num <= len(m.options) {
				option := &m.options[num-1]
				if !option.Handler(writer, keyReader, m.db, m.colorScheme) {
					return false
				}
			}
		}
	}

	writer.Write([]byte(menu.ShowCursor))
	return true
}
