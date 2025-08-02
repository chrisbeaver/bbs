package user_editor

import (
	"fmt"
	"strconv"
	"strings"

	"bbs/internal/menu"
	"bbs/internal/modules"
)

// readLine reads a line of input from the user
func readLine(keyReader modules.KeyReader, writer modules.Writer) (string, error) {
	var line strings.Builder
	for {
		key, err := keyReader.ReadKey()
		if err != nil {
			return "", err
		}

		switch key {
		case "enter":
			writer.Write([]byte("\n"))
			return line.String(), nil
		case "backspace":
			if line.Len() > 0 {
				str := line.String()
				line.Reset()
				line.WriteString(str[:len(str)-1])
				writer.Write([]byte("\b \b")) // Backspace, space, backspace
			}
		case "escape", "ctrl+c":
			return "", fmt.Errorf("cancelled")
		default:
			if len(key) == 1 && key[0] >= 32 && key[0] <= 126 { // Printable ASCII
				line.WriteString(key)
				writer.Write([]byte(key)) // Echo the character
			}
		}
	}
}

// parseAccessLevel parses an access level string
func parseAccessLevel(s string) (int, error) {
	level, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0, fmt.Errorf("invalid access level")
	}
	if level < 0 || level > 255 {
		return 0, fmt.Errorf("access level must be 0-255")
	}
	return level, nil
}

// showMessage displays a message and waits for user input
func showMessage(writer modules.Writer, keyReader modules.KeyReader, colorScheme menu.ColorScheme, message, messageType string) {
	writer.Write([]byte(menu.ClearScreen))

	coloredMessage := colorScheme.Colorize(message, messageType)
	centeredMessage := colorScheme.CenterText(coloredMessage, 79)
	writer.Write([]byte(centeredMessage + "\n\n"))

	prompt := colorScheme.Colorize("Press any key to continue...", "text")
	centeredPrompt := colorScheme.CenterText(prompt, 79)
	writer.Write([]byte(centeredPrompt))

	keyReader.ReadKey()
}
