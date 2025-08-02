package components

import (
	"fmt"
	"strings"
	"unicode"
)

// stripAnsiCodes removes ANSI escape codes from text for length calculation
func stripAnsiCodes(text string) string {
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

// TextInput represents a text input field
type TextInput struct {
	name        string
	label       string
	value       string
	placeholder string
	maxLength   int
	required    bool
	focused     bool
	colorScheme ColorScheme
	validator   func(string) error
	width       int
}

// TextInputConfig holds configuration for text input
type TextInputConfig struct {
	Name        string
	Label       string
	Value       string
	Placeholder string
	MaxLength   int
	Required    bool
	Width       int
	Validator   func(string) error
}

// NewTextInput creates a new text input component
func NewTextInput(config TextInputConfig, colorScheme ColorScheme) *TextInput {
	if config.Width <= 0 {
		config.Width = 40
	}
	if config.MaxLength <= 0 {
		config.MaxLength = 255
	}

	return &TextInput{
		name:        config.Name,
		label:       config.Label,
		value:       config.Value,
		placeholder: config.Placeholder,
		maxLength:   config.MaxLength,
		required:    config.Required,
		focused:     false,
		colorScheme: colorScheme,
		validator:   config.Validator,
		width:       config.Width,
	}
}

// SetFocus sets the focus state
func (t *TextInput) SetFocus(focused bool) {
	t.focused = focused
}

// IsFocused returns the focus state
func (t *TextInput) IsFocused() bool {
	return t.focused
}

// HandleKey handles keyboard input
func (t *TextInput) HandleKey(key rune) bool {
	switch key {
	case '\b', 127: // Backspace
		if len(t.value) > 0 {
			t.value = t.value[:len(t.value)-1]
		}
		return true
	case '\r', '\n': // Enter
		return false // Let form handle this
	case '\t': // Tab
		return false // Let focus manager handle this
	default:
		if unicode.IsPrint(key) && len(t.value) < t.maxLength {
			t.value += string(key)
			return true
		}
	}
	return false
}

// Render renders the text input
func (t *TextInput) Render() string {
	var result strings.Builder

	// Render label if provided
	if t.label != "" {
		labelText := t.label
		if t.required {
			labelText += " *"
		}
		result.WriteString(t.colorScheme.Colorize(labelText+":", "text"))
		result.WriteString("\n")
	}

	// Create the input content
	displayContent := t.value
	showPlaceholder := false
	if displayContent == "" && t.placeholder != "" {
		displayContent = t.placeholder
		showPlaceholder = true
	}

	// Pad content to full width
	if len(displayContent) > t.width {
		displayContent = displayContent[:t.width]
	}

	// Create the input field with blue background when focused
	if t.focused {
		// Focused state - blue background with cursor
		paddedContent := displayContent + strings.Repeat(" ", t.width-len(displayContent))
		if showPlaceholder {
			// Blue background with placeholder text in lighter color
			result.WriteString("\033[44m") // Blue background
			result.WriteString(t.colorScheme.Colorize(paddedContent, "secondary"))
			result.WriteString("\033[0m") // Reset
		} else {
			// Blue background with white text and cursor
			cursorPos := len(t.value)
			if cursorPos >= t.width {
				cursorPos = t.width - 1
			}

			beforeCursor := displayContent[:cursorPos]
			afterCursor := strings.Repeat(" ", t.width-cursorPos)

			result.WriteString("\033[44m") // Blue background
			result.WriteString("\033[37m") // White text
			result.WriteString(beforeCursor)
			result.WriteString("\033[47m\033[30m█\033[44m\033[37m") // White cursor on blue
			if len(afterCursor) > 1 {
				result.WriteString(afterCursor[1:]) // Skip one space for cursor
			}
			result.WriteString("\033[0m") // Reset
		}
	} else {
		// Unfocused state - no background, just content
		paddedContent := displayContent + strings.Repeat(" ", t.width-len(displayContent))
		if showPlaceholder {
			result.WriteString(t.colorScheme.Colorize(paddedContent, "secondary"))
		} else {
			result.WriteString(t.colorScheme.Colorize(paddedContent, "text"))
		}
	}

	return result.String()
}

// GetValue returns the current value
func (t *TextInput) GetValue() interface{} {
	return t.value
}

// SetValue sets the current value
func (t *TextInput) SetValue(value interface{}) {
	if str, ok := value.(string); ok {
		if len(str) <= t.maxLength {
			t.value = str
		}
	}
}

// Validate validates the input value
func (t *TextInput) Validate() error {
	if t.required && strings.TrimSpace(t.value) == "" {
		return fmt.Errorf("%s is required", t.label)
	}

	if t.validator != nil {
		return t.validator(t.value)
	}

	return nil
}

// GetName returns the field name
func (t *TextInput) GetName() string {
	return t.name
}

// IsRequired returns whether the field is required
func (t *TextInput) IsRequired() bool {
	return t.required
}

// GetLabel returns the field label
func (t *TextInput) GetLabel() string {
	return t.label
}

// GetStringValue returns the value as a string
func (t *TextInput) GetStringValue() string {
	return t.value
}
