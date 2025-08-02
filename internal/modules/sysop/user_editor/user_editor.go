package user_editor

import (
	"bbs/internal/components"
	"bbs/internal/database"
	"bbs/internal/menu"
)

// UserEditor implements the sysop user management functionality
type UserEditor struct {
	db          *database.DB
	colorScheme menu.ColorScheme
}

// NewUserEditor creates a new sysop user editor
func NewUserEditor(db *database.DB, colorScheme menu.ColorScheme) *UserEditor {
	return &UserEditor{
		db:          db,
		colorScheme: colorScheme,
	}
}

// ComponentColorSchemeAdapter adapts menu.ColorScheme to components.ColorScheme
type ComponentColorSchemeAdapter struct {
	colorScheme menu.ColorScheme
}

func (a *ComponentColorSchemeAdapter) Colorize(text, colorName string) string {
	return a.colorScheme.Colorize(text, colorName)
}

func (a *ComponentColorSchemeAdapter) ColorizeWithBg(text, fgColor, bgColor string) string {
	return a.colorScheme.ColorizeWithBg(text, fgColor, bgColor)
}

func (a *ComponentColorSchemeAdapter) CenterText(text string, terminalWidth int) string {
	return a.colorScheme.CenterText(text, terminalWidth)
}

func (a *ComponentColorSchemeAdapter) DrawSeparator(width int, char string) string {
	return a.colorScheme.DrawSeparator(width, char)
}

// getComponentAdapter returns a component-compatible color scheme
func (ue *UserEditor) getComponentAdapter() components.ColorScheme {
	return &ComponentColorSchemeAdapter{colorScheme: ue.colorScheme}
}
