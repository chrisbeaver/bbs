package bulletins

import (
	"bbs/internal/database"
	"bbs/internal/menu"
	"bbs/internal/modules/base"
)

// Module implements the bulletins functionality using database-driven options
type Module struct {
	*base.Module
	db          *database.DB
	colorScheme menu.ColorScheme
}

// NewModule creates a new bulletins module
func NewModule(db *database.DB, colorScheme menu.ColorScheme) *Module {
	m := &Module{
		db:          db,
		colorScheme: colorScheme,
	}
	m.Module = base.NewModule(db, colorScheme, m)
	return m
}

// LoadOptions implements OptionProvider interface
func (m *Module) LoadOptions(db *database.DB) ([]base.MenuOption, error) {
	bulletins, err := db.GetBulletins(50)
	if err != nil {
		return nil, err
	}

	var options []base.MenuOption
	for i, bulletin := range bulletins {
		option := NewBulletinOption(&bulletin, i, m.colorScheme)
		options = append(options, option)
	}

	return options, nil
}

// GetMenuTitle implements OptionProvider interface
func (m *Module) GetMenuTitle() string {
	return "System Bulletins"
}

// GetInstructions implements OptionProvider interface
func (m *Module) GetInstructions() string {
	return "Navigate: ↑↓  Read: Enter  Quit: Q"
}
