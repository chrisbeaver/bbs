package main

import (
	"fmt"

	"bbs/internal/components"
)

// Simple ColorScheme implementation for testing
type TestColorScheme struct{}

func (cs *TestColorScheme) Colorize(text, colorName string) string {
	return text // No coloring for test
}

func (cs *TestColorScheme) ColorizeWithBg(text, fgColor, bgColor string) string {
	return text
}

func (cs *TestColorScheme) CenterText(text string, terminalWidth int) string {
	if len(text) >= terminalWidth {
		return text
	}
	padding := (terminalWidth - len(text)) / 2
	return fmt.Sprintf("%*s%s", padding, "", text)
}

func (cs *TestColorScheme) DrawSeparator(width int, char string) string {
	result := ""
	for i := 0; i < width; i++ {
		result += char
	}
	return result
}

func testForm() {
	colorScheme := &TestColorScheme{}

	// Create the form
	form := components.NewForm(components.FormConfig{
		Title: "Test Form",
		Width: 79,
	}, colorScheme)

	// Add a test field
	testField := components.NewTextInput(components.TextInputConfig{
		Name:        "test",
		Label:       "Test Field",
		Placeholder: "Enter something...",
		MaxLength:   32,
		Required:    true,
		Width:       40,
	}, colorScheme)

	form.AddComponent(testField)
	form.Start()

	// Render the form
	fmt.Print(form.Render())
}
