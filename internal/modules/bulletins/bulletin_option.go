package bulletins

import (
	"fmt"
	"strings"

	"bbs/internal/database"
	"bbs/internal/menu"
	"bbs/internal/modules"
)

// BulletinOption represents a bulletin menu option
type BulletinOption struct {
	bulletin    *database.Bulletin
	index       int
	colorScheme menu.ColorScheme
}

// NewBulletinOption creates a new bulletin option
func NewBulletinOption(bulletin *database.Bulletin, index int, colorScheme menu.ColorScheme) *BulletinOption {
	return &BulletinOption{
		bulletin:    bulletin,
		index:       index,
		colorScheme: colorScheme,
	}
}

// GetID implements MenuOption interface
func (b *BulletinOption) GetID() string {
	return fmt.Sprintf("bulletin_%d", b.index)
}

// GetTitle implements MenuOption interface
func (b *BulletinOption) GetTitle() string {
	return b.bulletin.Title
}

// GetDescription implements MenuOption interface
func (b *BulletinOption) GetDescription() string {
	return fmt.Sprintf("%d) %s", b.index+1, b.bulletin.Title)
}

// Execute implements MenuOption interface
func (b *BulletinOption) Execute(writer modules.Writer, keyReader modules.KeyReader, db *database.DB, colorScheme menu.ColorScheme) bool {
	writer.Write([]byte(menu.ClearScreen))

	// Header with bulletin title
	header := fmt.Sprintf("--- %s ---", b.bulletin.Title)
	headerColored := colorScheme.Colorize(header, "primary")
	centeredHeader := colorScheme.CenterText(headerColored, 79)
	writer.Write([]byte(centeredHeader + "\n"))

	separator := colorScheme.DrawSeparator(len(header), "═")
	centeredSeparator := colorScheme.CenterText(separator, 79)
	writer.Write([]byte(centeredSeparator + "\n\n"))

	// Author and date info
	info := fmt.Sprintf("By: %s | Date: %s", b.bulletin.Author, b.bulletin.CreatedAt.Format("January 2, 2006"))
	infoColored := colorScheme.Colorize(info, "secondary")
	centeredInfo := colorScheme.CenterText(infoColored, 79)
	writer.Write([]byte(centeredInfo + "\n\n"))

	// Bulletin body - word wrap at 75 characters and center
	bodyLines := wrapText(b.bulletin.Body, 75)
	for _, line := range bodyLines {
		if strings.TrimSpace(line) == "" {
			writer.Write([]byte("\n"))
		} else {
			lineColored := colorScheme.Colorize(line, "text")
			centeredLine := colorScheme.CenterText(lineColored, 79)
			writer.Write([]byte(centeredLine + "\n"))
		}
	}

	// Return prompt
	writer.Write([]byte("\n"))
	prompt := colorScheme.Colorize("Press ", "text") +
		colorScheme.Colorize("any key", "accent") +
		colorScheme.Colorize(" to return to bulletin list...", "text")
	centeredPrompt := colorScheme.CenterText(prompt, 79)
	writer.Write([]byte(centeredPrompt))

	keyReader.ReadKey()
	return true
}

// wrapText wraps text to specified width
func wrapText(text string, width int) []string {
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
