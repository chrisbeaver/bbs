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
	if selected {
		// Create a full-width highlight bar
		padding := width - len(text)
		if padding < 0 {
			padding = 0
		}
		highlightText := " " + text + strings.Repeat(" ", padding-1)
		return cs.ColorizeWithBg(highlightText, "highlight", "primary")
	}
	// Non-selected items get normal padding
	padding := width - len(text)
	if padding < 0 {
		padding = 0
	}
	normalText := " " + text + strings.Repeat(" ", padding-1)
	return cs.Colorize(normalText, "text")
}

// Create decorative border pattern
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
	// Simple ANSI code removal - finds \033[...m sequences
	result := ""
	inEscape := false

	for i := 0; i < len(text); i++ {
		if text[i] == '\033' && i+1 < len(text) && text[i+1] == '[' {
			inEscape = true
			i++ // skip the '['
			continue
		}

		if inEscape {
			if text[i] == 'm' {
				inEscape = false
			}
			continue
		}

		result += string(text[i])
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
