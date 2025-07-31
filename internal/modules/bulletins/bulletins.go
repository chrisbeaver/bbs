package bulletins

import (
	"fmt"
	"strings"

	"golang.org/x/term"

	"bbs/internal/database"
)

// ColorScheme interface to avoid import cycle
type ColorScheme interface {
	Colorize(text, colorName string) string
	ColorizeWithBg(text, fgColor, bgColor string) string
	CenterText(text string, terminalWidth int) string
	DrawSeparator(width int, char string) string
}

// Screen control constants
const (
	ClearScreen    = "\033[2J\033[H"
	HideCursor     = "\033[?25l"
	ShowCursor     = "\033[?25h"
)

// Module implements the bulletins functionality
type Module struct {
	db          *database.DB
	colorScheme ColorScheme
}

// NewModule creates a new bulletins module
func NewModule(db *database.DB, colorScheme ColorScheme) *Module {
	return &Module{
		db:          db,
		colorScheme: colorScheme,
	}
}

// Execute runs the bulletins module
func (m *Module) Execute(term *term.Terminal) bool {
	// Get bulletins from database
	bulletins, err := m.db.GetBulletins(50) // Get more bulletins for navigation
	if err != nil {
		errorMsg := m.colorScheme.Colorize("Error retrieving bulletins.", "error")
		centeredError := m.colorScheme.CenterText(errorMsg, 79)
		term.Write([]byte(centeredError + "\n"))
		m.waitForKey(term)
		return true
	}

	if len(bulletins) == 0 {
		// Clear screen and show no bulletins message
		term.Write([]byte(ClearScreen + HideCursor))
		
		header := m.colorScheme.Colorize("System Bulletins", "primary")
		centeredHeader := m.colorScheme.CenterText(header, 79)
		term.Write([]byte(centeredHeader + "\n"))
		
		separator := m.colorScheme.DrawSeparator(len("System Bulletins"), "═")
		centeredSeparator := m.colorScheme.CenterText(separator, 79)
		term.Write([]byte(centeredSeparator + "\n\n"))
		
		noMsg := m.colorScheme.Colorize("No bulletins available.", "secondary")
		centeredNoMsg := m.colorScheme.CenterText(noMsg, 79)
		term.Write([]byte(centeredNoMsg + "\n\n"))
		
		prompt := m.colorScheme.Colorize("Press any key to continue...", "text")
		centeredPrompt := m.colorScheme.CenterText(prompt, 79)
		term.Write([]byte(centeredPrompt))
		
		m.waitForKey(term)
		term.Write([]byte(ShowCursor))
		return true
	}

	// Show navigable bulletin list
	return m.showBulletinList(term, bulletins)
}

// showBulletinList displays a navigable list of bulletins
func (m *Module) showBulletinList(term *term.Terminal, bulletins []database.Bulletin) bool {
	selectedIndex := 0

	for {
		// Clear screen and hide cursor
		term.Write([]byte(ClearScreen + HideCursor))

		// Display header
		header := m.colorScheme.Colorize("System Bulletins", "primary")
		centeredHeader := m.colorScheme.CenterText(header, 79)
		term.Write([]byte(centeredHeader + "\n"))
		
		separator := m.colorScheme.DrawSeparator(len("System Bulletins"), "═")
		centeredSeparator := m.colorScheme.CenterText(separator, 79)
		term.Write([]byte(centeredSeparator + "\n\n"))

		// Calculate display area
		terminalWidth := 79
		contentWidth := 70
		centerOffset := (terminalWidth - contentWidth) / 2
		centerPadding := strings.Repeat(" ", centerOffset)

		// Display bulletin list with navigation
		for i, bulletin := range bulletins {
			isSelected := (i == selectedIndex)
			
			// Format bulletin line
			number := fmt.Sprintf("%2d)", i+1)
			title := bulletin.Title
			author := fmt.Sprintf("by %s", bulletin.Author)
			date := bulletin.CreatedAt.Format("2006-01-02")
			
			// Truncate title if too long
			maxTitleLength := contentWidth - len(number) - len(author) - len(date) - 6 // spaces and parentheses
			if len(title) > maxTitleLength {
				title = title[:maxTitleLength-3] + "..."
			}
			
			bulletinLine := fmt.Sprintf("%s %s (%s, %s)", number, title, author, date)
			
			// Pad to content width
			if len(bulletinLine) < contentWidth {
				bulletinLine += strings.Repeat(" ", contentWidth-len(bulletinLine))
			}
			
			if isSelected {
				// Highlight selected item
				coloredLine := m.colorScheme.ColorizeWithBg(bulletinLine, "highlight", "primary")
				term.Write([]byte(centerPadding + coloredLine + "\n"))
			} else {
				// Normal item
				numberColored := m.colorScheme.Colorize(number, "accent")
				titleColored := m.colorScheme.Colorize(title, "text")
				authorColored := m.colorScheme.Colorize(fmt.Sprintf("(%s, %s)", author, date), "secondary")
				
				normalLine := fmt.Sprintf("%s %s %s", numberColored, titleColored, authorColored)
				// Pad the line to maintain consistent spacing
				paddedLine := normalLine + strings.Repeat(" ", contentWidth-len(fmt.Sprintf("%s %s (%s, %s)", number, title, author, date)))
				term.Write([]byte(centerPadding + paddedLine + "\n"))
			}
		}

		// Instructions
		instructions := m.colorScheme.Colorize("\nUse ", "text") +
			m.colorScheme.Colorize("↑↓", "accent") +
			m.colorScheme.Colorize(" to navigate, ", "text") +
			m.colorScheme.Colorize("Enter", "accent") +
			m.colorScheme.Colorize(" to read, ", "text") +
			m.colorScheme.Colorize("Q", "accent") +
			m.colorScheme.Colorize(" to return", "text")

		centeredInstructions := m.colorScheme.CenterText(instructions, 79)
		term.Write([]byte("\n" + centeredInstructions))

		// Handle input
		key, err := m.readKey(term)
		if err != nil {
			term.Write([]byte(ShowCursor))
			return true
		}

		switch key {
		case "up":
			selectedIndex--
			if selectedIndex < 0 {
				selectedIndex = len(bulletins) - 1
			}
		case "down":
			selectedIndex++
			if selectedIndex >= len(bulletins) {
				selectedIndex = 0
			}
		case "enter":
			// Show selected bulletin
			if selectedIndex >= 0 && selectedIndex < len(bulletins) {
				m.showBulletin(term, &bulletins[selectedIndex])
			}
		case "quit", "q", "Q":
			term.Write([]byte(ShowCursor))
			return true
		}
	}
}

// showBulletin displays a single bulletin
func (m *Module) showBulletin(term *term.Terminal, bulletin *database.Bulletin) {
	// Clear screen
	term.Write([]byte(ClearScreen + HideCursor))

	terminalWidth := 79
	contentWidth := 70
	centerOffset := (terminalWidth - contentWidth) / 2
	centerPadding := strings.Repeat(" ", centerOffset)

	// Header with bulletin title
	title := m.colorScheme.Colorize(bulletin.Title, "primary")
	centeredTitle := m.colorScheme.CenterText(title, terminalWidth)
	term.Write([]byte(centeredTitle + "\n"))

	// Separator
	separator := m.colorScheme.DrawSeparator(len(bulletin.Title), "═")
	centeredSeparator := m.colorScheme.CenterText(separator, terminalWidth)
	term.Write([]byte(centeredSeparator + "\n\n"))

	// Metadata
	author := m.colorScheme.Colorize(fmt.Sprintf("Author: %s", bulletin.Author), "accent")
	date := m.colorScheme.Colorize(fmt.Sprintf("Date: %s", bulletin.CreatedAt.Format("2006-01-02 15:04:05")), "secondary")
	
	term.Write([]byte(centerPadding + author + "\n"))
	term.Write([]byte(centerPadding + date + "\n\n"))

	// Bulletin body - word wrap to content width
	body := bulletin.Body
	lines := m.wrapText(body, contentWidth)
	
	for _, line := range lines {
		coloredLine := m.colorScheme.Colorize(line, "text")
		term.Write([]byte(centerPadding + coloredLine + "\n"))
	}

	// Footer prompt
	prompt := m.colorScheme.Colorize("\nPress any key to return to bulletin list...", "text")
	centeredPrompt := m.colorScheme.CenterText(prompt, terminalWidth)
	term.Write([]byte(centeredPrompt))

	// Wait for key press
	m.waitForKey(term)
	term.Write([]byte(ShowCursor))
}

// wrapText wraps text to specified width
func (m *Module) wrapText(text string, width int) []string {
	var lines []string
	words := strings.Fields(text)
	
	if len(words) == 0 {
		return lines
	}

	currentLine := ""
	for _, word := range words {
		// Check if adding this word would exceed the width
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) <= width {
			currentLine = testLine
		} else {
			// Start a new line
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}

	// Add the last line
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return lines
}

// readKey reads a single key press, handling arrow keys
// Note: This is a simplified implementation that uses letter keys instead of arrow keys
// In a full implementation, you'd need direct access to the SSH channel
func (m *Module) readKey(term *term.Terminal) (string, error) {
	input, err := term.ReadLine()
	if err != nil {
		return "", err
	}
	
	if input == "" {
		return "enter", nil
	}
	
	switch strings.ToLower(input) {
	case "q", "quit":
		return "quit", nil
	case "u", "up":
		return "up", nil
	case "d", "down":
		return "down", nil
	default:
		return "enter", nil
	}
}

// waitForKey waits for any key press
func (m *Module) waitForKey(term *term.Terminal) {
	term.ReadLine()
}
