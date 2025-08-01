package base

import (
	"fmt"

	"bbs/internal/database"
	"bbs/internal/menu"
	"bbs/internal/modules"
)

// CommandOption represents a menu option that executes a handler function
type CommandOption struct {
	ID          string
	Title       string
	Description string
	Handler     func(writer modules.Writer, keyReader modules.KeyReader, db *database.DB, colorScheme menu.ColorScheme) bool
}

// GetID implements MenuOption interface
func (c *CommandOption) GetID() string {
	return c.ID
}

// GetTitle implements MenuOption interface
func (c *CommandOption) GetTitle() string {
	return c.Title
}

// GetDescription implements MenuOption interface
func (c *CommandOption) GetDescription() string {
	return c.Description
}

// Execute implements MenuOption interface
func (c *CommandOption) Execute(writer modules.Writer, keyReader modules.KeyReader, db *database.DB, colorScheme menu.ColorScheme) bool {
	if c.Handler == nil {
		errorMsg := colorScheme.Colorize(fmt.Sprintf("No handler defined for command: %s", c.ID), "error")
		centeredError := colorScheme.CenterText(errorMsg, 79)
		writer.Write([]byte(centeredError + "\n"))
		return true
	}
	return c.Handler(writer, keyReader, db, colorScheme)
}
