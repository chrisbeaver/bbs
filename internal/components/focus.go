package components

// FocusManager handles focus management for form components
type FocusManager struct {
	components   []Focusable
	currentIndex int
	isActive     bool
}

// NewFocusManager creates a new focus manager
func NewFocusManager() *FocusManager {
	return &FocusManager{
		components:   make([]Focusable, 0),
		currentIndex: 0,
		isActive:     false,
	}
}

// AddComponent adds a component to the focus manager
func (fm *FocusManager) AddComponent(component Focusable) {
	fm.components = append(fm.components, component)
}

// SetActive activates/deactivates the focus manager
func (fm *FocusManager) SetActive(active bool) {
	fm.isActive = active
	if active && len(fm.components) > 0 {
		fm.components[fm.currentIndex].SetFocus(true)
	} else {
		fm.clearAllFocus()
	}
}

// HandleTab moves focus to the next component
func (fm *FocusManager) HandleTab() {
	if !fm.isActive || len(fm.components) == 0 {
		return
	}

	fm.components[fm.currentIndex].SetFocus(false)
	fm.currentIndex = (fm.currentIndex + 1) % len(fm.components)
	fm.components[fm.currentIndex].SetFocus(true)
}

// HandleShiftTab moves focus to the previous component
func (fm *FocusManager) HandleShiftTab() {
	if !fm.isActive || len(fm.components) == 0 {
		return
	}

	fm.components[fm.currentIndex].SetFocus(false)
	fm.currentIndex = (fm.currentIndex - 1 + len(fm.components)) % len(fm.components)
	fm.components[fm.currentIndex].SetFocus(true)
}

// HandleKey passes key input to the focused component
func (fm *FocusManager) HandleKey(key rune) bool {
	if !fm.isActive || len(fm.components) == 0 {
		return false
	}

	return fm.components[fm.currentIndex].HandleKey(key)
}

// GetFocusedComponent returns the currently focused component
func (fm *FocusManager) GetFocusedComponent() Focusable {
	if !fm.isActive || len(fm.components) == 0 {
		return nil
	}
	return fm.components[fm.currentIndex]
}

// clearAllFocus removes focus from all components
func (fm *FocusManager) clearAllFocus() {
	for _, component := range fm.components {
		component.SetFocus(false)
	}
}
