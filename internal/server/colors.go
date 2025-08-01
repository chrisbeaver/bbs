package server

import (
	"fmt"
	"strings"

	"bbs/internal/config"
)

// ANSI color codes
var colorCodes = map[string]string{
	"black":          "\033[30m",
	"red":            "\033[31m",
	"green":          "\033[32m",
	"yellow":         "\033[33m",
	"blue":           "\033[34m",
	"magenta":        "\033[35m",
	"cyan":           "\033[36m",
	"white":          "\033[37m",
	"bright_black":   "\033[90m",
	"bright_red":     "\033[91m",
	"bright_green":   "\033[92m",
	"bright_yellow":  "\033[93m",
	"bright_blue":    "\033[94m",
	"bright_magenta": "\033[95m",
	"bright_cyan":    "\033[96m",
	"bright_white":   "\033[97m",
	"reset":          "\033[0m",
	"bold":           "\033[1m",
	"underline":      "\033[4m",
	"reverse":        "\033[7m",
}

// Background color codes
var bgColorCodes = map[string]string{
	"black":   "\033[40m",
	"red":     "\033[41m",
	"green":   "\033[42m",
	"yellow":  "\033[43m",
	"blue":    "\033[44m",
	"magenta": "\033[45m",
	"cyan":    "\033[46m",
	"white":   "\033[47m",
}

type ColorScheme struct {
	config *config.ColorConfig
}

func NewColorScheme(cfg *config.ColorConfig) *ColorScheme {
	return &ColorScheme{config: cfg}
}

func (cs *ColorScheme) GetColor(colorName string) string {
	var configColor string
	switch colorName {
	case "primary":
		configColor = cs.config.Primary
	case "secondary":
		configColor = cs.config.Secondary
	case "accent":
		configColor = cs.config.Accent
	case "text":
		configColor = cs.config.Text
	case "background":
		configColor = cs.config.Background
	case "border":
		configColor = cs.config.Border
	case "success":
		configColor = cs.config.Success
	case "error":
		configColor = cs.config.Error
	case "highlight":
		configColor = cs.config.Highlight
	default:
		configColor = colorName
	}

	if code, exists := colorCodes[configColor]; exists {
		return code
	}
	return ""
}

func (cs *ColorScheme) GetBgColor(colorName string) string {
	var configColor string
	switch colorName {
	case "background":
		configColor = cs.config.Background
	default:
		configColor = colorName
	}

	if code, exists := bgColorCodes[configColor]; exists {
		return code
	}
	return ""
}

func (cs *ColorScheme) Colorize(text, colorName string) string {
	return cs.GetColor(colorName) + text + colorCodes["reset"]
}

func (cs *ColorScheme) ColorizeWithBg(text, fgColor, bgColor string) string {
	return cs.GetColor(fgColor) + cs.GetBgColor(bgColor) + text + colorCodes["reset"]
}

// Special ANSI sequences for screen control
const (
	ClearScreen    = "\033[2J\033[H"
	ClearLine      = "\033[2K"
	CursorHome     = "\033[H"
	CursorUp       = "\033[A"
	CursorDown     = "\033[B"
	CursorForward  = "\033[C"
	CursorBackward = "\033[D"
	SaveCursor     = "\033[s"
	RestoreCursor  = "\033[u"
	HideCursor     = "\033[?25l"
	ShowCursor     = "\033[?25h"
)

// Movement and positioning functions
func MoveCursorTo(row, col int) string {
	return fmt.Sprintf("\033[%d;%dH", row, col)
}

func MoveCursorUp(lines int) string {
	return fmt.Sprintf("\033[%dA", lines)
}

func MoveCursorDown(lines int) string {
	return fmt.Sprintf("\033[%dB", lines)
}

// Selection highlighting
func (cs *ColorScheme) HighlightSelection(text string, selected bool, width int) string {
	// Calculate padding based on clean text length (without ANSI codes)
	cleanText := cs.stripAnsiCodes(text)
	textLen := len(cleanText)

	if selected {
		// Create a full-width highlight bar - NO BACKGROUND COLOR, just bright white text
		padding := width - textLen - 2 // Account for leading and trailing spaces
		if padding < 0 {
			padding = 0
		}

		// For selected items, change text to bright white while keeping hotkey bright yellow
		adjustedText := cs.adjustColorsForSelection(text)
		highlightText := " " + adjustedText + strings.Repeat(" ", padding) + " "
		return highlightText // No background color applied
	}
	// Non-selected items get normal padding
	padding := width - textLen - 2 // Account for leading and trailing spaces
	if padding < 0 {
		padding = 0
	}
	normalText := " " + text + strings.Repeat(" ", padding) + " "
	return cs.Colorize(normalText, "text")
} // Create decorative border pattern
func (cs *ColorScheme) CreateBorderPattern(width int, pattern string) string {
	if len(pattern) == 0 {
		pattern = "-"
	}

	// Repeat the pattern to fill the width
	repeats := width / len(pattern)
	remainder := width % len(pattern)

	borderText := strings.Repeat(pattern, repeats)
	if remainder > 0 {
		borderText += pattern[:remainder]
	}

	return cs.Colorize(borderText, "border")
}

// Center text within a given terminal width
func (cs *ColorScheme) CenterText(text string, terminalWidth int) string {
	// Remove ANSI codes to calculate actual text length
	cleanText := cs.stripAnsiCodes(text)
	textLen := len(cleanText)

	if textLen >= terminalWidth {
		return text
	}

	padding := (terminalWidth - textLen) / 2
	return strings.Repeat(" ", padding) + text
}

// Helper function to strip ANSI codes for length calculation
func (cs *ColorScheme) stripAnsiCodes(text string) string {
	// More robust ANSI stripping using a simple state machine
	result := ""
	i := 0
	
	for i < len(text) {
		if i <= len(text)-4 && text[i] == '\033' && text[i+1] == '[' {
			// Found ESC[, skip until we find a letter (usually 'm')
			i += 2 // skip \033[
			for i < len(text) {
				char := text[i]
				i++
				// ANSI sequences typically end with a letter
				if (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') {
					break
				}
			}
		} else {
			result += string(text[i])
			i++
		}
	}
	
	return result
}

// StripAnsiCodes removes ANSI escape codes from text (public version for interface compatibility)
func (cs *ColorScheme) StripAnsiCodes(text string) string {
	return cs.stripAnsiCodes(text)
}

// Center container but left-align content within it
func (cs *ColorScheme) CenterContainerLeftAlign(text string, containerWidth, terminalWidth int) string {
	// Calculate padding to center the container
	centerOffset := (terminalWidth - containerWidth) / 2
	if centerOffset < 0 {
		centerOffset = 0
	}

	// Left-align text within the container (with small left margin)
	leftPadding := strings.Repeat(" ", centerOffset)
	textWithMargin := " " + text

	return leftPadding + textWithMargin
}

// ASCII Art and Box Drawing
func (cs *ColorScheme) DrawBox(title string, width int) string {
	if width < len(title)+4 {
		width = len(title) + 4
	}

	topLine := "╔" + strings.Repeat("═", width-2) + "╗"
	titleLine := fmt.Sprintf("║ %s%s ║",
		cs.Colorize(title, "primary"),
		strings.Repeat(" ", width-len(title)-4))
	bottomLine := "╚" + strings.Repeat("═", width-2) + "╝"

	return cs.Colorize(topLine, "border") + "\n" +
		cs.Colorize("║", "border") + titleLine + cs.Colorize("║", "border") + "\n" +
		cs.Colorize(bottomLine, "border")
}

func (cs *ColorScheme) DrawSeparator(width int, char string) string {
	if char == "" {
		char = "═"
	}
	return cs.Colorize(strings.Repeat(char, width), "border")
}

// BBS-style welcome banner
func (cs *ColorScheme) CreateWelcomeBanner(systemName, welcomeMsg string) string {
	width := 78

	banner := cs.Colorize(ClearScreen, "")

	// Top border
	banner += cs.Colorize("╔"+strings.Repeat("═", width-2)+"╗", "border") + "\n"

	// Title with padding
	titlePadding := (width - len(systemName) - 2) / 2
	leftPad := strings.Repeat(" ", titlePadding)
	rightPad := strings.Repeat(" ", width-len(systemName)-titlePadding-2)

	banner += cs.Colorize("║", "border") +
		cs.Colorize(leftPad+systemName+rightPad, "primary") +
		cs.Colorize("║", "border") + "\n"

	// Bottom border
	banner += cs.Colorize("╚"+strings.Repeat("═", width-2)+"╝", "border") + "\n\n"

	// Welcome message
	banner += cs.Colorize(welcomeMsg, "text") + "\n\n"

	return banner
}

// replaceTextColorInSelection replaces text color codes in a string while preserving accent (hotkey) colors
func (cs *ColorScheme) replaceTextColorInSelection(text, newColor string) string {
	// Get the color codes
	textColor := cs.GetColor("text")
	accentColor := cs.GetColor("accent")
	newColorCode := cs.GetColor(newColor)

	result := text

	// Replace text color with new color, but preserve accent colors
	if textColor != "" {
		// Replace text color followed by reset with new color followed by reset
		result = strings.ReplaceAll(result, textColor, newColorCode)
	}

	// Ensure accent colors (hotkeys) remain bright yellow by replacing any occurrence
	// of the new color that should be accent with accent color
	if accentColor != "" && newColorCode != "" {
		// This is a simple approach - in a real implementation you might need
		// more sophisticated ANSI sequence parsing
		result = strings.ReplaceAll(result, accentColor, accentColor)
	}

	return result
}

// adjustColorsForSelection changes text color to bright white while preserving bright yellow hotkey colors
func (cs *ColorScheme) adjustColorsForSelection(text string) string {
	// Get the color codes
	textColor := cs.GetColor("text")
	accentColor := cs.GetColor("accent")
	brightWhite := cs.GetColor("bright_white")
	brightYellow := cs.GetColor("bright_yellow")
	reset := colorCodes["reset"]

	// If there are no ANSI codes, just make the whole thing bright white
	if !strings.Contains(text, "\033[") {
		return brightWhite + text + reset
	}

	result := text

	// Replace text color with bright white
	if textColor != "" && brightWhite != "" {
		result = strings.ReplaceAll(result, textColor, brightWhite)
	}

	// Ensure accent colors become bright yellow
	if accentColor != "" && brightYellow != "" {
		result = strings.ReplaceAll(result, accentColor, brightYellow)
	}

	// Handle the case where we have mixed content - need to ensure
	// everything that's not explicitly colored as hotkey becomes bright white

	// Simple approach: if the text starts without a color code, prepend bright white
	if !strings.HasPrefix(result, "\033[") {
		result = brightWhite + result
	}

	// Ensure any reset sequences are followed by bright white (except before accent colors)
	resetPattern := reset
	if resetPattern != "" {
		// Replace reset codes with reset + bright white, but be careful around accent colors
		parts := strings.Split(result, resetPattern)
		if len(parts) > 1 {
			for i := 0; i < len(parts)-1; i++ {
				// Check if the next part starts with accent color
				nextPart := parts[i+1]
				if !strings.HasPrefix(nextPart, accentColor) && !strings.HasPrefix(nextPart, brightYellow) {
					parts[i] = parts[i] + resetPattern + brightWhite
				} else {
					parts[i] = parts[i] + resetPattern
				}
			}
			result = strings.Join(parts, "")
		}
	}

	return result
}
