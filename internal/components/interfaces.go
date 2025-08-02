package components

// ColorScheme interface to match your existing pattern
type ColorScheme interface {
	Colorize(text, colorName string) string
	ColorizeWithBg(text, fgColor, bgColor string) string
	CenterText(text string, terminalWidth int) string
	DrawSeparator(width int, char string) string
}

// Focusable represents a component that can receive focus
type Focusable interface {
	SetFocus(focused bool)
	IsFocused() bool
	HandleKey(key rune) bool
	Render() string
	GetValue() interface{}
	SetValue(value interface{})
	Validate() error
}

// FormComponent represents a form component with validation
type FormComponent interface {
	Focusable
	GetName() string
	IsRequired() bool
	GetLabel() string
}
