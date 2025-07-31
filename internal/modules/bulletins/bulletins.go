package bulletins

import (
	"fmt"
	"strconv"
	"strings"

	"bbs/internal/config"
	"bbs/internal/database"
)

// ColorScheme interface to avoid import cycle
type ColorScheme interface {
	Colorize(text, colorName string) string
	ColorizeWithBg(text, fgColor, bgColor string) string
	CenterText(text string, terminalWidth int) string
	CenterContainerLeftAlign(text string, containerWidth, terminalWidth int) string
	DrawSeparator(width int, char string) string
	CreateBorderPattern(width int, pattern string) string
	HighlightSelection(text string, selected bool, maxWidth int) string
	StripAnsiCodes(text string) string // Made this exported
}

// KeyReader interface for reading keys
type KeyReader interface {
	ReadKey() (string, error)
}

// Writer interface for writing output
type Writer interface {
	Write([]byte) (int, error)
}

// Screen control constants
const (
	ClearScreen = "\033[2J\033[H"
	HideCursor  = "\033[?25l"
	ShowCursor  = "\033[?25h"
)

// Module implements the bulletins functionality using the same menu system
type Module struct {
	db            *database.DB
	colorScheme   ColorScheme
	bulletins     []database.Bulletin
	selectedIndex int
}

// NewModule creates a new bulletins module
func NewModule(db *database.DB, colorScheme ColorScheme) *Module {
	return &Module{
		db:          db,
		colorScheme: colorScheme,
	}
}

// Execute runs the bulletins module using the same menu navigation system
func (m *Module) Execute(writer Writer, keyReader KeyReader) bool {
	// Get bulletins from database
	bulletins, err := m.db.GetBulletins(50)
	if err != nil {
		errorMsg := m.colorScheme.Colorize("Error retrieving bulletins.", "error")
		centeredError := m.colorScheme.CenterText(errorMsg, 79)
		writer.Write([]byte(centeredError + "\n"))
		return true
	}

	if len(bulletins) == 0 {
		writer.Write([]byte(ClearScreen))
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

	// Create dynamic menu structure for bulletins
	bulletinMenu := &config.MenuItem{
		ID:          "bulletins",
		Title:       "System Bulletins",
		Description: "System Bulletins",
	}

	// Convert bulletins to menu items
	for i, bulletin := range bulletins {
		// Keep descriptions shorter and cleaner like regular menu items
		description := fmt.Sprintf("%d) %s", i+1, bulletin.Title)

		bulletinMenu.Submenu = append(bulletinMenu.Submenu, config.MenuItem{
			ID:          fmt.Sprintf("bulletin_%d", i),
			Title:       bulletin.Title,
			Description: description,
			Command:     fmt.Sprintf("read_bulletin_%d", i),
		})
	}

	// Use the same menu display logic as regular menus
	for {
		m.displayBulletinMenu(writer, bulletinMenu)

		// Get key input using the same navigation as menus
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
			writer.Write([]byte(ShowCursor))
			return true
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			// Direct number selection
			if num, err := strconv.Atoi(key); err == nil && num >= 1 && num <= len(bulletins) {
				m.showBulletin(writer, keyReader, &bulletins[num-1])
			}
		}
	}

	writer.Write([]byte(ShowCursor))
	return true
}

// displayBulletinMenu uses the same display logic as the regular menu system
func (m *Module) displayBulletinMenu(writer Writer, menu *config.MenuItem) {
	// Clear screen and hide cursor (same as menu system)
	writer.Write([]byte(ClearScreen + HideCursor))

	// Terminal width for centering (same as menu system)
	terminalWidth := 79

	// Menu title with color and centering (same as menu system)
	title := m.colorScheme.Colorize(menu.Title, "primary")
	centeredTitle := m.colorScheme.CenterText(title, terminalWidth)
	writer.Write([]byte(fmt.Sprintf("%s\n", centeredTitle)))

	// Decorative separator (same as menu system)
	cleanTitle := m.colorScheme.StripAnsiCodes(title)
	separator := m.colorScheme.DrawSeparator(len(cleanTitle), "═")
	centeredSeparator := m.colorScheme.CenterText(separator, terminalWidth)
	writer.Write([]byte(centeredSeparator + "\n\n"))

	// Calculate maximum width needed for highlight bar (same as menu system)
	maxWidth := 0
	for _, item := range menu.Submenu {
		cleanDesc := m.colorScheme.StripAnsiCodes(item.Description)
		if len(cleanDesc) > maxWidth {
			maxWidth = len(cleanDesc)
		}
	}
	// Add some padding but keep it reasonable
	maxWidth += 4

	// Cap the max width to keep it manageable like regular menus
	if maxWidth > 60 {
		maxWidth = 60
	}

	// Calculate centering offset for menu items (same as menu system)
	centerOffset := (terminalWidth - maxWidth) / 2
	if centerOffset < 0 {
		centerOffset = 0
	}

	// Create decorative border pattern (same as menu system)
	borderPattern := m.colorScheme.CreateBorderPattern(maxWidth, "-=")
	centerPadding := strings.Repeat(" ", centerOffset)

	// Top border (same as menu system)
	writer.Write([]byte(centerPadding + borderPattern + "\n"))

	// Ensure selected index is valid (same as menu system)
	if m.selectedIndex >= len(menu.Submenu) {
		m.selectedIndex = 0
	}
	if m.selectedIndex < 0 {
		m.selectedIndex = len(menu.Submenu) - 1
	}

	// Display menu items with highlighting and centering (same as menu system)
	for i, item := range menu.Submenu {
		selected := (i == m.selectedIndex)
		menuLine := m.colorScheme.HighlightSelection(item.Description, selected, maxWidth)
		writer.Write([]byte(centerPadding + menuLine + "\n"))
	}

	// Bottom border (same as menu system)
	writer.Write([]byte(centerPadding + borderPattern + "\n"))

	// Instructions with proper color styling (same as menu system)
	instructions := m.colorScheme.Colorize("Use ", "text") +
		m.colorScheme.Colorize("↑↓", "accent") +
		m.colorScheme.Colorize(" arrow keys to navigate, ", "text") +
		m.colorScheme.Colorize("Enter", "accent") +
		m.colorScheme.Colorize(" to read, ", "text") +
		m.colorScheme.Colorize("Q", "accent") +
		m.colorScheme.Colorize(" to quit", "text")

	centeredInstructions := m.colorScheme.CenterText(instructions, terminalWidth)
	writer.Write([]byte("\n" + centeredInstructions))
}

// showBulletin displays the full content of a bulletin (same formatting as before)
func (m *Module) showBulletin(writer Writer, keyReader KeyReader, bulletin *database.Bulletin) {
	writer.Write([]byte(ClearScreen))

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
