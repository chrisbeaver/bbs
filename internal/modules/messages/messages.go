package messages

import (
	"bbs/internal/database"
	"bbs/internal/menu"
	"bbs/internal/modules/base"
)

// Module implements the messages functionality
type Module struct {
	*base.Module
	db          *database.DB
	colorScheme menu.ColorScheme
}

// NewModule creates a new messages module
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
	topics, err := db.GetTopics()
	if err != nil {
		return nil, err
	}

	var options []base.MenuOption
	for i, topic := range topics {
		option := NewTopicOption(&topic, i, m.colorScheme)
		options = append(options, option)
	}

	return options, nil
}

// GetMenuTitle implements OptionProvider interface
func (m *Module) GetMenuTitle() string {
	return "MESSAGE BOARDS"
}

// GetInstructions implements OptionProvider interface
func (m *Module) GetInstructions() string {
	return "Navigate: ↑↓  View: Enter  Back: Q"
}