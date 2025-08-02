package components

import (
	"fmt"
	"strings"
)

// Form represents a collection of form components
type Form struct {
	title        string
	components   []FormComponent
	focusManager *FocusManager
	colorScheme  ColorScheme
	submitted    bool
	cancelled    bool
	width        int
}

// FormConfig holds configuration for a form
type FormConfig struct {
	Title string
	Width int
}

// NewForm creates a new form
func NewForm(config FormConfig, colorScheme ColorScheme) *Form {
	if config.Width <= 0 {
		config.Width = 79
	}

	return &Form{
		title:        config.Title,
		components:   make([]FormComponent, 0),
		focusManager: NewFocusManager(),
		colorScheme:  colorScheme,
		submitted:    false,
		cancelled:    false,
		width:        config.Width,
	}
}

// AddComponent adds a component to the form
func (f *Form) AddComponent(component FormComponent) {
	f.components = append(f.components, component)
	f.focusManager.AddComponent(component)
}

// HandleKey handles keyboard input for the form
func (f *Form) HandleKey(key rune) bool {
	switch key {
	case '\t': // Tab - next field
		f.focusManager.HandleTab()
		return true
	case '\r', '\n': // Enter - submit form
		f.submitted = true
		return true
	case 27: // Escape - cancel form
		f.cancelled = true
		return true
	default:
		return f.focusManager.HandleKey(key)
	}
}

// Render renders the entire form
func (f *Form) Render() string {
	var result strings.Builder

	// Clear screen
	result.WriteString("\033[2J\033[H")

	if f.title != "" {
		header := f.colorScheme.Colorize(f.title, "primary")
		centeredHeader := f.colorScheme.CenterText(header, f.width)
		result.WriteString(centeredHeader + "\n")

		separator := f.colorScheme.DrawSeparator(len(f.title), "═")
		centeredSeparator := f.colorScheme.CenterText(separator, f.width)
		result.WriteString(centeredSeparator + "\n\n")
	}

	// Render all components
	for i, component := range f.components {
		// Center each component
		componentRender := component.Render()
		lines := strings.Split(componentRender, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				centeredLine := f.colorScheme.CenterText(line, f.width)
				result.WriteString(centeredLine)
			}
			result.WriteString("\n")
		}

		if i < len(f.components)-1 {
			result.WriteString("\n")
		}
	}

	// Show instructions
	result.WriteString("\n")
	instructions := f.colorScheme.Colorize("Tab: Next Field  Enter: Submit  Esc: Cancel", "secondary")
	centeredInstructions := f.colorScheme.CenterText(instructions, f.width)
	result.WriteString(centeredInstructions)

	return result.String()
}

// Start activates the form for input
func (f *Form) Start() {
	f.focusManager.SetActive(true)
	f.submitted = false
	f.cancelled = false
}

// IsSubmitted returns whether the form was submitted
func (f *Form) IsSubmitted() bool {
	return f.submitted
}

// IsCancelled returns whether the form was cancelled
func (f *Form) IsCancelled() bool {
	return f.cancelled
}

// Validate validates all form components
func (f *Form) Validate() []error {
	var errors []error
	for _, component := range f.components {
		if err := component.Validate(); err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

// GetValues returns all form values as a map
func (f *Form) GetValues() map[string]interface{} {
	values := make(map[string]interface{})
	for _, component := range f.components {
		values[component.GetName()] = component.GetValue()
	}
	return values
}

// GetStringValues returns all form values as strings
func (f *Form) GetStringValues() map[string]string {
	values := make(map[string]string)
	for _, component := range f.components {
		if textInput, ok := component.(*TextInput); ok {
			values[component.GetName()] = textInput.GetStringValue()
		} else {
			values[component.GetName()] = fmt.Sprintf("%v", component.GetValue())
		}
	}
	return values
}

// Reset resets the form state
func (f *Form) Reset() {
	f.submitted = false
	f.cancelled = false
	f.focusManager.SetActive(false)
}
