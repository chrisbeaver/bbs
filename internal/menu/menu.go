package menu

import (
	"fmt"
	"strings"

	"bbs/internal/config"
	"bbs/internal/modules"
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
type Writer = modules.Writer

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
	ClearScreen      = "\033[2J\033[H"
	ClearContentArea = "\033[H\033[0J" // Home cursor and clear from cursor to end of screen
	HideCursor       = "\033[?25l"
	ShowCursor       = "\033[?25h"
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
	instructions := "Navigate: ↑↓  Select: Enter  Hotkeys: Execute  Quit: Q"

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
	// Clear content area only (respects scroll region) and hide cursor
	r.writer.Write([]byte(ClearContentArea + HideCursor))

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
	// Build the plain text version first to calculate proper centering
	plainInstructions := "Navigate: ↑↓  Select: Enter"

	// Add hotkeys section if mentioned in instructions
	if strings.Contains(instructionText, "hotkey") || strings.Contains(instructionText, "Hotkeys") {
		plainInstructions += "  Hotkeys: Execute"
	}

	// Add read section if mentioned in instructions
	if strings.Contains(instructionText, "read") || strings.Contains(instructionText, "Read") {
		plainInstructions += "  Read: Enter"
	}

	plainInstructions += "  Quit: Q"

	// Calculate centering based on plain text
	textLen := len(plainInstructions)
	padding := (r.terminalWidth - textLen) / 2
	if padding < 0 {
		padding = 0
	}

	// Now build the colored version
	coloredInstructions := r.colorScheme.Colorize("Navigate: ", "text") +
		r.colorScheme.Colorize("↑↓", "accent") +
		r.colorScheme.Colorize("  Select: ", "text") +
		r.colorScheme.Colorize("Enter", "accent")

	// Add hotkeys section if mentioned in instructions
	if strings.Contains(instructionText, "hotkey") || strings.Contains(instructionText, "Hotkeys") {
		coloredInstructions += r.colorScheme.Colorize("  Hotkeys: ", "text") +
			r.colorScheme.Colorize("Execute", "accent")
	}

	// Add read section if mentioned in instructions
	if strings.Contains(instructionText, "read") || strings.Contains(instructionText, "Read") {
		coloredInstructions += r.colorScheme.Colorize("  Read: ", "text") +
			r.colorScheme.Colorize("Enter", "accent")
	}

	coloredInstructions += r.colorScheme.Colorize("  Quit: ", "text") +
		r.colorScheme.Colorize("Q", "accent")

	// Apply the padding calculated from plain text to the colored version
	centeredInstructions := strings.Repeat(" ", padding) + coloredInstructions
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
