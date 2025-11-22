package bulletins

import (
	"fmt"
	"strings"

	"bbs/internal/database"
	"bbs/internal/menu"
	"bbs/internal/modules"
	"bbs/internal/pager"
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
	// Prepare bulletin content lines
	bodyLines := wrapText(b.bulletin.Body, 75)

	// Build complete content with header and body
	var contentLines []string

	// Author and date info
	info := fmt.Sprintf("By: %s | Date: %s", b.bulletin.Author, b.bulletin.CreatedAt.Format("January 2, 2006"))
	infoColored := colorScheme.Colorize(info, "secondary")
	centeredInfo := colorScheme.CenterText(infoColored, 79)
	contentLines = append(contentLines, centeredInfo, "")

	// Add body lines with proper formatting
	for _, line := range bodyLines {
		if strings.TrimSpace(line) == "" {
			contentLines = append(contentLines, "")
		} else {
			lineColored := colorScheme.Colorize(line, "text")
			centeredLine := colorScheme.CenterText(lineColored, 79)
			contentLines = append(contentLines, centeredLine)
		}
	}

	// Create terminal sizer from writer (will use real terminal dimensions)
	termSizer := pager.NewTerminalSizerFromWriter(writer)

	// Create pager adapter that uses real terminal dimensions
	writerAdapter := pager.NewWriterAdapter(writer, termSizer)

	// Check if writer implements StatusBarManager (e.g., TerminalWriter does)
	// This uses type assertion to provide status bar control without polluting interfaces
	type StatusBarController interface {
		Pause()
		Resume()
	}
	if sbCtrl, ok := writer.(StatusBarController); ok {
		writerAdapter.WithStatusBarManager(sbCtrl)
	}

	// Create pager instance
	p := pager.NewPager(writerAdapter, keyReader, writerAdapter, colorScheme)

	// Pass status bar control to pager if available
	if writerAdapter.StatusBarMgr != nil {
		p.WithStatusBar(writerAdapter)
	}

	// Display bulletin using pager
	title := fmt.Sprintf("--- %s ---", b.bulletin.Title)
	p.Display(contentLines, title)

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
