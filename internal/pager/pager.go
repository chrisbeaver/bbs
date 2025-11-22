package pager

import (
	"fmt"
	"strings"
)

const (
	// ANSI escape codes
	ClearContentArea = "\033[H\033[0J" // Home cursor and clear from cursor to end (respects scroll region)
	HideCursor       = "\033[?25l"
	ShowCursor       = "\033[?25h"
)

// Pager handles paginated display of long content
type Pager struct {
	writer        Writer
	keyReader     KeyReader
	terminalSizer TerminalSizer
	colorScheme   ColorScheme
	statusBarMgr  StatusBarManager // Optional: for pausing timer updates
}

// NewPager creates a new pager instance
func NewPager(writer Writer, keyReader KeyReader, terminalSizer TerminalSizer, colorScheme ColorScheme) *Pager {
	return &Pager{
		writer:        writer,
		keyReader:     keyReader,
		terminalSizer: terminalSizer,
		colorScheme:   colorScheme,
		statusBarMgr:  nil, // No status bar manager by default
	}
}

// WithStatusBar adds status bar control to the pager
func (p *Pager) WithStatusBar(mgr StatusBarManager) *Pager {
	p.statusBarMgr = mgr
	return p
}

// Display shows content with pagination
func (p *Pager) Display(lines []string, title string) error {
	// Get terminal dimensions
	_, height, err := p.terminalSizer.Size()
	if err != nil {
		height = 24 // Default height
	}

	// Calculate available content height with very conservative margins
	// Simple approach: just avoid the bottom few lines entirely
	// Never write to the last 3 lines to ensure status bar is safe
	availableLines := height - 9 // Very conservative: header(3) + footer(1) + buffer(5)

	// If content fits on one screen, just display it without pagination
	if len(lines) <= availableLines {
		return p.displaySinglePage(lines, title)
	}

	// Calculate total pages
	totalPages := (len(lines) + availableLines - 1) / availableLines
	currentPage := 0

	// Hide cursor during navigation
	p.writer.Write([]byte(HideCursor))
	defer p.writer.Write([]byte(ShowCursor))

	for {
		// Calculate start and end indices for current page
		startIdx := currentPage * availableLines
		endIdx := startIdx + availableLines
		if endIdx > len(lines) {
			endIdx = len(lines)
		}

		// Get lines for current page
		pageLines := lines[startIdx:endIdx]

		// Display current page
		p.displayPage(pageLines, title, currentPage+1, totalPages)

		// Get user input for navigation
		key, err := p.keyReader.ReadKey()
		if err != nil {
			return err
		}

		// Handle navigation
		switch key {
		case "q", "Q":
			// Quit
			return nil
		case " ", "enter", "down":
			// Next page (or quit if on last page)
			if currentPage < totalPages-1 {
				currentPage++
			} else {
				// On last page, quit
				return nil
			}
		case "b", "B", "up":
			// Previous page
			if currentPage > 0 {
				currentPage--
			}
		}
	}
}

// displaySinglePage displays content that fits on one screen without pagination
func (p *Pager) displaySinglePage(lines []string, title string) error {
	// Get terminal height first
	_, height, err := p.terminalSizer.Size()
	if err != nil {
		height = 24 // Default
	}

	// Clear content area (respects existing scroll region protecting status bar)
	p.writer.Write([]byte(ClearContentArea))

	// Use absolute positioning for all elements to prevent scrolling
	currentLine := 1

	// Title at line 1
	coloredTitle := p.colorScheme.Colorize(title, "primary")
	centeredTitle := p.colorScheme.CenterText(coloredTitle, 79)
	position := fmt.Sprintf("\033[%d;1H", currentLine)
	p.writer.Write([]byte(position + centeredTitle))
	currentLine++

	// Separator at line 2
	separator := strings.Repeat("─", 79)
	coloredSeparator := p.colorScheme.Colorize(separator, "secondary")
	position = fmt.Sprintf("\033[%d;1H", currentLine)
	p.writer.Write([]byte(position + coloredSeparator))
	currentLine += 2 // Skip line 3 (blank)

	// Display content lines starting at line 4, using absolute positioning
	// Leave extra space to ensure status bar is never overwritten
	maxContentLine := height - 6 // Very conservative: never write to bottom 6 lines
	for _, line := range lines {
		if currentLine > maxContentLine {
			break // Stop if we would exceed the scroll region
		}
		position = fmt.Sprintf("\033[%d;1H", currentLine)
		p.writer.Write([]byte(position + line))
		currentLine++
	}

	// Footer at line height-5 (extra conservative to avoid status bar)
	footerLine := height - 5
	footerPosition := fmt.Sprintf("\033[%d;1H", footerLine)
	footer := p.colorScheme.Colorize("Press any key to return...", "text")
	centeredFooter := p.colorScheme.CenterText(footer, 79)
	p.writer.Write([]byte(footerPosition + centeredFooter))

	// Status bar is protected by scroll region and managed by timer updates
	// No manual redraw needed as it would interfere with cursor positioning

	// Wait for key press
	_, err = p.keyReader.ReadKey()
	return err
}

// displayPage displays a single page of content
func (p *Pager) displayPage(lines []string, title string, currentPage, totalPages int) {
	// Get terminal height first
	_, height, err := p.terminalSizer.Size()
	if err != nil {
		height = 24 // Default
	}

	// Clear content area (respects existing scroll region protecting status bar)
	p.writer.Write([]byte(ClearContentArea))

	// Use absolute positioning for all elements to prevent scrolling
	currentLine := 1

	// Title at line 1
	coloredTitle := p.colorScheme.Colorize(title, "primary")
	centeredTitle := p.colorScheme.CenterText(coloredTitle, 79)
	position := fmt.Sprintf("\033[%d;1H", currentLine)
	p.writer.Write([]byte(position + centeredTitle))
	currentLine++

	// Page indicator at line 2 (for multi-page)
	pageIndicator := fmt.Sprintf("Page %d of %d", currentPage, totalPages)
	coloredIndicator := p.colorScheme.Colorize(pageIndicator, "secondary")
	centeredIndicator := p.colorScheme.CenterText(coloredIndicator, 79)
	position = fmt.Sprintf("\033[%d;1H", currentLine)
	p.writer.Write([]byte(position + centeredIndicator))
	currentLine++

	// Separator at line 3
	separator := strings.Repeat("─", 79)
	coloredSeparator := p.colorScheme.Colorize(separator, "secondary")
	position = fmt.Sprintf("\033[%d;1H", currentLine)
	p.writer.Write([]byte(position + coloredSeparator))
	currentLine += 2 // Skip line 4 (blank)

	// Display content lines starting at line 5, using absolute positioning
	// Leave extra space to ensure status bar is never overwritten
	maxContentLine := height - 6 // Very conservative: never write to bottom 6 lines
	for _, line := range lines {
		if currentLine > maxContentLine {
			break // Stop if we would exceed the scroll region
		}
		position = fmt.Sprintf("\033[%d;1H", currentLine)
		p.writer.Write([]byte(position + line))
		currentLine++
	}

	// Footer with navigation instructions
	p.displayFooter(currentPage, totalPages, height)

	// Status bar is protected by scroll region and managed by timer updates
	// No manual redraw needed as it would interfere with cursor positioning
}

// displayHeader displays the page header with title and page indicator
func (p *Pager) displayHeader(title string, currentPage, totalPages int) {
	// Title line
	coloredTitle := p.colorScheme.Colorize(title, "primary")
	centeredTitle := p.colorScheme.CenterText(coloredTitle, 79)
	p.writer.Write([]byte(centeredTitle + "\n"))

	// Page indicator (only show if multiple pages)
	if totalPages > 1 {
		pageIndicator := fmt.Sprintf("Page %d of %d", currentPage, totalPages)
		coloredIndicator := p.colorScheme.Colorize(pageIndicator, "secondary")
		centeredIndicator := p.colorScheme.CenterText(coloredIndicator, 79)
		p.writer.Write([]byte(centeredIndicator + "\n"))
	}

	// Separator line
	separator := strings.Repeat("─", 79)
	coloredSeparator := p.colorScheme.Colorize(separator, "secondary")
	p.writer.Write([]byte(coloredSeparator + "\n\n"))
}

// displayFooter displays navigation instructions using absolute positioning
func (p *Pager) displayFooter(currentPage, totalPages, terminalHeight int) {
	// Build navigation instructions based on current page
	var instructions string
	if currentPage < totalPages {
		// Not on last page - show next/previous/quit
		if currentPage == 1 {
			// First page - only show next and quit
			instructions = "Space: Next | Q: Quit"
		} else {
			// Middle page - show all options
			instructions = "Space: Next | B: Back | Q: Quit"
		}
	} else {
		// Last page - show back and quit
		if totalPages > 1 {
			instructions = "Space: Return | B: Back"
		} else {
			instructions = "Press any key to return..."
		}
	}

	// Position footer very conservatively to avoid status bar
	// Use terminalHeight-5 to leave plenty of buffer
	footerLine := terminalHeight - 5
	footerPosition := fmt.Sprintf("\033[%d;1H", footerLine)

	coloredInstructions := p.colorScheme.Colorize(instructions, "text")
	centeredInstructions := p.colorScheme.CenterText(coloredInstructions, 79)
	p.writer.Write([]byte(footerPosition + centeredInstructions))
}
