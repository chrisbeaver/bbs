package statusbar

import (
	"strings"
	"testing"
	"time"

	"bbs/internal/config"
)

func TestStatusBar_Render(t *testing.T) {
	cfg := &config.Config{
		BBS: config.BBSConfig{
			SystemName:    "Coastline BBS",
			MaxLineLength: 79,
		},
	}

	sb := New("testuser", cfg)

	rendered := sb.Render()

	// Check that it contains expected elements
	if !strings.Contains(rendered, "testuser") {
		t.Error("Status bar should contain username")
	}

	if !strings.Contains(rendered, "Coastline BBS") {
		t.Error("Status bar should contain system name")
	}

	if !strings.Contains(rendered, "00:00:") {
		t.Error("Status bar should contain duration timer")
	}

	// Check for ANSI color codes
	if !strings.Contains(rendered, "\033[44m") {
		t.Error("Status bar should contain blue background")
	}

	if !strings.Contains(rendered, "\033[92m") {
		t.Error("Status bar should contain bright green text")
	}

	if !strings.Contains(rendered, "\033[93m") {
		t.Error("Status bar should contain bright yellow text")
	}
}

func TestStatusBar_SetActive(t *testing.T) {
	cfg := &config.Config{
		BBS: config.BBSConfig{
			SystemName:    "Test BBS",
			MaxLineLength: 79,
		},
	}

	sb := New("testuser", cfg)

	// Should be active by default
	if sb.Render() == "" {
		t.Error("Status bar should render when active")
	}

	// Disable and test
	sb.SetActive(false)
	if sb.Render() != "" {
		t.Error("Status bar should not render when inactive")
	}

	// Re-enable and test
	sb.SetActive(true)
	if sb.Render() == "" {
		t.Error("Status bar should render when re-activated")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{0, "00:00:00"},
		{30 * time.Second, "00:00:30"},
		{90 * time.Second, "00:01:30"},
		{3661 * time.Second, "01:01:01"},
	}

	for _, test := range tests {
		result := formatDuration(test.duration)
		if result != test.expected {
			t.Errorf("formatDuration(%v) = %s, expected %s", test.duration, result, test.expected)
		}
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{"short", 10, "short"},
		{"this is a very long string", 10, "this is..."},
		{"exact", 5, "exact"},
		{"a", 1, "a"},
		{"ab", 1, "a"},
	}

	for _, test := range tests {
		result := truncateString(test.input, test.maxLen)
		if result != test.expected {
			t.Errorf("truncateString(%s, %d) = %s, expected %s", test.input, test.maxLen, result, test.expected)
		}
	}
}
