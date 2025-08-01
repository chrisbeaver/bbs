package menu

import (
	"fmt"
	"strings"

	"bbs/internal/config"
)

// ColorScheme interface for menu rendering
type ColorScheme interface {
	Colorize(text, colorName string) string
	ColorizeWithBg(text, fgColor, bgColor string) string
	CenterText(text string, terminalWidth int) string
	DrawSeparator(width int, char string) string
	CreateBorderPattern(width int, pattern string) string
	HighlightSelection(text string, selected bool, maxWidth int) string
	StripAnsiCodes(text string) string
}

// Writer interface for output
type Writer interface {
	Write([]byte) (int, error)
}

// MenuProvider interface for modules that provide menu items
type MenuProvider interface {
	GetMenuTitle() string
	GetMenuItems() []MenuItem
	GetInstructions() string
}

// MenuItem represents a single menu item
type MenuItem struct {
	ID          string
	Description string
	Data        interface{} // Module-specific data
}

// MenuRenderer handles display logic for all menu types
type MenuRenderer struct {
	colorScheme   ColorScheme
	writer        Writer
	terminalWidth int
}

// Screen control constants
const (
	ClearScreen = "\033[2J\033[H"
	HideCursor  = "\033[?25l"
	ShowCursor  = "\033[?25h"
)

// NewMenuRenderer creates a new menu renderer
func NewMenuRenderer(colorScheme ColorScheme, writer Writer) *MenuRenderer {
	return &MenuRenderer{
		colorScheme:   colorScheme,
		writer:        writer,
		terminalWidth: 79, // Classic BBS width
	}
}

// RenderConfigMenu displays a config-based menu with access level filtering
func (r *MenuRenderer) RenderConfigMenu(menuItem *config.MenuItem, selectedIndex int, userAccessLevel int) {
	// Create menu items from config, filtering by access level
	var items []MenuItem
	for _, item := range menuItem.Submenu {
		if item.AccessLevel <= userAccessLevel {
			// Process the description to highlight hotkeys
			description := r.highlightHotkey(item.Description, item.Hotkey)
			items = append(items, MenuItem{
				ID:          item.ID,
				Description: description,
				Data:        item,
			})
		}
	}

	// Default instructions for config menus with hotkey info
	instructions := "Use ↑↓ arrow keys to navigate, Enter to select, hotkeys to execute, Q for goodbye"

	r.renderMenu(menuItem.Title, items, selectedIndex, instructions)
}

// RenderModuleMenu displays a module-provided menu
func (r *MenuRenderer) RenderModuleMenu(provider MenuProvider, selectedIndex int) {
	title := provider.GetMenuTitle()
	items := provider.GetMenuItems()
	instructions := provider.GetInstructions()

	r.renderMenu(title, items, selectedIndex, instructions)
}

// renderMenu is the unified rendering method
func (r *MenuRenderer) renderMenu(title string, items []MenuItem, selectedIndex int, instructions string) {
	// Clear screen and hide cursor
	r.writer.Write([]byte(ClearScreen + HideCursor))

	// Menu title with color and centering
	coloredTitle := r.colorScheme.Colorize(title, "primary")
	centeredTitle := r.colorScheme.CenterText(coloredTitle, r.terminalWidth)
	r.writer.Write([]byte(fmt.Sprintf("%s\n\n", centeredTitle)))

	// Calculate maximum width needed for highlight bar
	maxWidth := r.calculateMaxWidth(items)

	// Calculate centering offset for menu items
	centerOffset := (r.terminalWidth - maxWidth) / 2
	if centerOffset < 0 {
		centerOffset = 0
	}

	// Create decorative border pattern longer than menu options
	borderPattern := r.colorScheme.CreateBorderPattern(maxWidth+8, "-=")
	borderCenterPadding := strings.Repeat(" ", (r.terminalWidth-(maxWidth+8))/2)
	menuCenterPadding := strings.Repeat(" ", centerOffset)

	// Top border (centered under title)
	r.writer.Write([]byte(borderCenterPadding + borderPattern + "\n"))

	// Ensure selected index is valid
	if selectedIndex >= len(items) {
		selectedIndex = 0
	}
	if selectedIndex < 0 {
		selectedIndex = len(items) - 1
	}

	// Display menu items with highlighting and centering
	for i, item := range items {
		selected := (i == selectedIndex)
		menuLine := r.colorScheme.HighlightSelection(item.Description, selected, maxWidth)
		r.writer.Write([]byte(menuCenterPadding + menuLine + "\n"))
	}

	// Bottom border (centered under title)
	r.writer.Write([]byte(borderCenterPadding + borderPattern + "\n"))

	// Instructions with proper color styling
	r.renderInstructions(instructions)
}

// renderInstructions displays formatted instructions
func (r *MenuRenderer) renderInstructions(instructionText string) {
	// Parse instruction text to colorize special parts
	instructions := r.colorScheme.Colorize("Use ", "text") +
		r.colorScheme.Colorize("↑↓", "accent") +
		r.colorScheme.Colorize(" arrow keys to navigate, ", "text") +
		r.colorScheme.Colorize("Enter", "accent")

	// Add custom instruction ending
	if strings.Contains(instructionText, "read") {
		instructions += r.colorScheme.Colorize(" to read, ", "text")
	} else {
		instructions += r.colorScheme.Colorize(" to select, ", "text")
	}

	// Add hotkey information if mentioned in instructions
	if strings.Contains(instructionText, "hotkey") {
		instructions += r.colorScheme.Colorize("hotkeys", "accent") +
			r.colorScheme.Colorize(" to execute, ", "text")
	}

	instructions += r.colorScheme.Colorize("Q", "accent")

	if strings.Contains(instructionText, "quit") {
		instructions += r.colorScheme.Colorize(" to quit", "text")
	} else {
		instructions += r.colorScheme.Colorize(" for goodbye", "text")
	}

	centeredInstructions := r.colorScheme.CenterText(instructions, r.terminalWidth)
	r.writer.Write([]byte("\n" + centeredInstructions))
}

// calculateMaxWidth determines the maximum width needed for menu items
func (r *MenuRenderer) calculateMaxWidth(items []MenuItem) int {
	maxWidth := 0
	for _, item := range items {
		cleanDesc := r.colorScheme.StripAnsiCodes(item.Description)
		if len(cleanDesc) > maxWidth {
			maxWidth = len(cleanDesc)
		}
	}
	// Add some padding but keep it reasonable
	maxWidth += 4

	// Cap the max width to keep it manageable
	if maxWidth > 60 {
		maxWidth = 60
	}

	return maxWidth
}

// highlightHotkey highlights the hotkey character in the description
func (r *MenuRenderer) highlightHotkey(description, hotkey string) string {
	if hotkey == "" {
		return description
	}

	// Convert hotkey to both upper and lower case for matching
	hotkeyLower := strings.ToLower(hotkey)
	hotkeyUpper := strings.ToUpper(hotkey)

	// Find the first occurrence of the hotkey character (case insensitive)
	for i, char := range description {
		charStr := string(char)
		if charStr == hotkeyLower || charStr == hotkeyUpper {
			// Split the description and highlight the hotkey character
			before := description[:i]
			hotkeyChr := description[i : i+1]
			after := description[i+1:]

			// Highlight the hotkey character
			highlightedHotkey := r.colorScheme.Colorize(hotkeyChr, "accent")
			return before + highlightedHotkey + after
		}
	}

	// If hotkey character not found in description, return as-is
	return description
}
