package sysop

import (
	"fmt"

	"bbs/internal/config"
	"bbs/internal/database"
	"bbs/internal/menu"
	"bbs/internal/modules"
	"bbs/internal/modules/base"
)

// Module implements the sysop functionality using config-driven options
type Module struct {
	*base.Module
	config      *config.Config
	db          *database.DB
	colorScheme menu.ColorScheme
}

// NewModule creates a new sysop module
func NewModule(db *database.DB, colorScheme menu.ColorScheme, cfg *config.Config) *Module {
	m := &Module{
		config:      cfg,
		db:          db,
		colorScheme: colorScheme,
	}
	m.Module = base.NewModule(db, colorScheme, m)
	return m
}

// LoadOptions implements OptionProvider interface
func (m *Module) LoadOptions(db *database.DB) ([]base.MenuOption, error) {
	var options []base.MenuOption

	sysopConfig := m.config.GetMenuConfig("sysop")
	if sysopConfig == nil {
		return options, nil
	}

	for _, configOption := range sysopConfig.Options {
		// Map command names to handler functions
		handler := m.getHandlerForCommand(configOption.Command)

		option := &base.CommandOption{
			ID:          configOption.ID,
			Title:       configOption.Title,
			Description: configOption.Description,
			Handler:     handler,
		}
		options = append(options, option)
	}

	return options, nil
}

// GetMenuTitle implements OptionProvider interface
func (m *Module) GetMenuTitle() string {
	sysopConfig := m.config.GetMenuConfig("sysop")
	if sysopConfig == nil {
		return "System Administration"
	}
	return sysopConfig.Title
}

// GetInstructions implements OptionProvider interface
func (m *Module) GetInstructions() string {
	sysopConfig := m.config.GetMenuConfig("sysop")
	if sysopConfig == nil {
		return "Navigate: ↑↓  Select: Enter  Quit: Q"
	}
	return sysopConfig.Instructions
}

// getHandlerForCommand maps command names to handler functions
func (m *Module) getHandlerForCommand(command string) func(modules.Writer, modules.KeyReader, *database.DB, menu.ColorScheme) bool {
	switch command {
	case "create_user":
		return handleCreateUser
	case "edit_user":
		return handleEditUser
	case "delete_user":
		return handleDeleteUser
	case "view_users":
		return handleViewUsers
	case "change_password":
		return handleChangePassword
	case "toggle_user":
		return handleToggleUserStatus
	case "system_stats":
		return handleSystemStats
	case "bulletin_management":
		return handleBulletinManagement
	default:
		return func(writer modules.Writer, keyReader modules.KeyReader, db *database.DB, colorScheme menu.ColorScheme) bool {
			errorMsg := colorScheme.Colorize(fmt.Sprintf("Unknown command: %s", command), "error")
			centeredError := colorScheme.CenterText(errorMsg, 79)
			writer.Write([]byte(centeredError + "\n"))
			keyReader.ReadKey()
			return true
		}
	}
}
