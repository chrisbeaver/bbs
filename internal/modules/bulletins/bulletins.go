package bulletins

import (
	"fmt"
	"strconv"
	"strings"

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

// Module implements the bulletins functionality using the unified menu system
type Module struct {
	db            *database.DB
	colorScheme   menu.ColorScheme
	bulletins     []database.Bulletin
	selectedIndex int
}

// NewModule creates a new bulletins module
func NewModule(db *database.DB, colorScheme menu.ColorScheme) *Module {
	return &Module{
		db:          db,
		colorScheme: colorScheme,
	}
}

// GetMenuTitle implements MenuProvider interface
func (m *Module) GetMenuTitle() string {
	return "System Bulletins"
}

// GetMenuItems implements MenuProvider interface
func (m *Module) GetMenuItems() []menu.MenuItem {
	var items []menu.MenuItem
	for i, bulletin := range m.bulletins {
		description := fmt.Sprintf("%d) %s", i+1, bulletin.Title)
		items = append(items, menu.MenuItem{
			ID:          fmt.Sprintf("bulletin_%d", i),
			Description: description,
			Data:        &bulletin, // Store bulletin reference
		})
	}
	return items
}

// GetInstructions implements MenuProvider interface
func (m *Module) GetInstructions() string {
	return "Use ↑↓ arrow keys to navigate, Enter to read, Q to quit"
}

// Execute runs the bulletins module using the unified menu system
func (m *Module) Execute(writer Writer, keyReader KeyReader) bool {
	// Initialize menu renderer
	menuRenderer := menu.NewMenuRenderer(m.colorScheme, writer)

	// Get bulletins from database
	bulletins, err := m.db.GetBulletins(50)
	if err != nil {
		errorMsg := m.colorScheme.Colorize("Error retrieving bulletins.", "error")
		centeredError := m.colorScheme.CenterText(errorMsg, 79)
		writer.Write([]byte(centeredError + "\n"))
		return true
	}

	if len(bulletins) == 0 {
		writer.Write([]byte(menu.ClearScreen))
		header := m.colorScheme.Colorize("--- System Bulletins ---", "primary")
		centeredHeader := m.colorScheme.CenterText(header, 79)
		writer.Write([]byte(centeredHeader + "\n\n"))

		noMsg := m.colorScheme.Colorize("No bulletins available.", "secondary")
		centeredNoMsg := m.colorScheme.CenterText(noMsg, 79)
		writer.Write([]byte(centeredNoMsg + "\n\n"))

		prompt := m.colorScheme.Colorize("Press any key to continue...", "text")
		centeredPrompt := m.colorScheme.CenterText(prompt, 79)
		writer.Write([]byte(centeredPrompt))

		keyReader.ReadKey()
		return true
	}

	m.bulletins = bulletins
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
				m.selectedIndex = len(bulletins) - 1
			}
		case "down":
			m.selectedIndex++
			if m.selectedIndex >= len(bulletins) {
				m.selectedIndex = 0
			}
		case "enter":
			// Show selected bulletin content
			m.showBulletin(writer, keyReader, &bulletins[m.selectedIndex])
		case "q", "Q", "quit":
			writer.Write([]byte(menu.ShowCursor))
			return true
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			// Direct number selection
			if num, err := strconv.Atoi(key); err == nil && num >= 1 && num <= len(bulletins) {
				m.showBulletin(writer, keyReader, &bulletins[num-1])
			}
		}
	}

	writer.Write([]byte(menu.ShowCursor))
	return true
}

// showBulletin displays the full content of a bulletin
func (m *Module) showBulletin(writer Writer, keyReader KeyReader, bulletin *database.Bulletin) {
	writer.Write([]byte(menu.ClearScreen))

	// Header with bulletin title
	header := fmt.Sprintf("--- %s ---", bulletin.Title)
	headerColored := m.colorScheme.Colorize(header, "primary")
	centeredHeader := m.colorScheme.CenterText(headerColored, 79)
	writer.Write([]byte(centeredHeader + "\n"))

	separator := m.colorScheme.DrawSeparator(len(header), "═")
	centeredSeparator := m.colorScheme.CenterText(separator, 79)
	writer.Write([]byte(centeredSeparator + "\n\n"))

	// Author and date info
	info := fmt.Sprintf("By: %s | Date: %s", bulletin.Author, bulletin.CreatedAt.Format("January 2, 2006"))
	infoColored := m.colorScheme.Colorize(info, "secondary")
	centeredInfo := m.colorScheme.CenterText(infoColored, 79)
	writer.Write([]byte(centeredInfo + "\n\n"))

	// Bulletin body - word wrap at 75 characters and center
	bodyLines := m.wrapText(bulletin.Body, 75)
	for _, line := range bodyLines {
		if strings.TrimSpace(line) == "" {
			writer.Write([]byte("\n"))
		} else {
			lineColored := m.colorScheme.Colorize(line, "text")
			centeredLine := m.colorScheme.CenterText(lineColored, 79)
			writer.Write([]byte(centeredLine + "\n"))
		}
	}

	// Return prompt with proper color styling
	writer.Write([]byte("\n"))
	prompt := m.colorScheme.Colorize("Press ", "text") +
		m.colorScheme.Colorize("any key", "accent") +
		m.colorScheme.Colorize(" to return to bulletin list...", "text")
	centeredPrompt := m.colorScheme.CenterText(prompt, 79)
	writer.Write([]byte(centeredPrompt))

	keyReader.ReadKey()
}

// wrapText wraps text to specified width (same as before)
func (m *Module) wrapText(text string, width int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{}
	}

	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		if currentLine.Len() == 0 {
			currentLine.WriteString(word)
		} else if currentLine.Len()+1+len(word) <= width {
			currentLine.WriteString(" " + word)
		} else {
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLine.WriteString(word)
		}
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return lines
}
