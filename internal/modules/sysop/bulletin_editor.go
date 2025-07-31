package sysop

import (
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/term"

	"bbs/internal/database"
)

// ColorScheme interface to avoid import cycle
type ColorScheme interface {
	Colorize(text, colorName string) string
	ColorizeWithBg(text, fgColor, bgColor string) string
	CenterText(text string, terminalWidth int) string
	DrawSeparator(width int, char string) string
}

// Screen control constants
const (
	ClearScreen = "\033[2J\033[H"
	HideCursor  = "\033[?25l"
	ShowCursor  = "\033[?25h"
)

// Module implements the sysop bulletin management functionality
type BulletinEditor struct {
	db          *database.DB
	colorScheme ColorScheme
}

// NewBulletinEditor creates a new sysop bulletin editor
func NewBulletinEditor(db *database.DB, colorScheme ColorScheme) *BulletinEditor {
	return &BulletinEditor{
		db:          db,
		colorScheme: colorScheme,
	}
}

// Execute runs the bulletin editor
func (be *BulletinEditor) Execute(term *term.Terminal) bool {
	for {
		// Clear screen and show menu
		term.Write([]byte(ClearScreen + HideCursor))

		// Header
		header := be.colorScheme.Colorize("Sysop Bulletin Editor", "primary")
		centeredHeader := be.colorScheme.CenterText(header, 79)
		term.Write([]byte(centeredHeader + "\n"))

		separator := be.colorScheme.DrawSeparator(len("Sysop Bulletin Editor"), "═")
		centeredSeparator := be.colorScheme.CenterText(separator, 79)
		term.Write([]byte(centeredSeparator + "\n\n"))

		// Menu options
		options := []string{
			"1) List all bulletins",
			"2) Create new bulletin",
			"3) Edit existing bulletin",
			"4) Delete bulletin",
			"Q) Return to main menu",
		}

		for _, option := range options {
			coloredOption := be.colorScheme.Colorize(option, "text")
			centeredOption := be.colorScheme.CenterText(coloredOption, 79)
			term.Write([]byte(centeredOption + "\n"))
		}

		// Prompt
		prompt := be.colorScheme.Colorize("\nEnter your choice: ", "accent")
		centeredPrompt := be.colorScheme.CenterText(prompt, 79)
		term.Write([]byte(centeredPrompt))
		term.Write([]byte(ShowCursor))

		// Get input
		input, err := term.ReadLine()
		if err != nil {
			return true
		}

		term.Write([]byte(HideCursor))

		switch strings.ToLower(strings.TrimSpace(input)) {
		case "1":
			be.ListBulletins(term)
		case "2":
			be.CreateBulletin(term)
		case "3":
			be.EditBulletin(term)
		case "4":
			be.DeleteBulletin(term)
		case "q", "quit":
			term.Write([]byte(ShowCursor))
			return true
		default:
			be.showMessage(term, "Invalid choice. Press any key to continue...", "error")
		}
	}
}

// ListBulletins displays a list of all bulletins
func (be *BulletinEditor) ListBulletins(term *term.Terminal) {
	bulletins, err := be.db.GetBulletins(50)
	if err != nil {
		be.showMessage(term, "Error retrieving bulletins: "+err.Error(), "error")
		return
	}

	term.Write([]byte(ClearScreen))

	header := be.colorScheme.Colorize("All Bulletins", "primary")
	centeredHeader := be.colorScheme.CenterText(header, 79)
	term.Write([]byte(centeredHeader + "\n"))

	separator := be.colorScheme.DrawSeparator(len("All Bulletins"), "═")
	centeredSeparator := be.colorScheme.CenterText(separator, 79)
	term.Write([]byte(centeredSeparator + "\n\n"))

	if len(bulletins) == 0 {
		msg := be.colorScheme.Colorize("No bulletins found.", "secondary")
		centeredMsg := be.colorScheme.CenterText(msg, 79)
		term.Write([]byte(centeredMsg + "\n"))
	} else {
		for _, bulletin := range bulletins {
			line := fmt.Sprintf("ID: %d | %s | by %s | %s",
				bulletin.ID,
				bulletin.Title,
				bulletin.Author,
				bulletin.CreatedAt.Format("2006-01-02"))

			coloredLine := be.colorScheme.Colorize(line, "text")
			centeredLine := be.colorScheme.CenterText(coloredLine, 79)
			term.Write([]byte(centeredLine + "\n"))
		}
	}

	be.showMessage(term, "\nPress any key to continue...", "text")
}

// CreateBulletin creates a new bulletin
func (be *BulletinEditor) CreateBulletin(term *term.Terminal) {
	term.Write([]byte(ClearScreen + ShowCursor))

	// Get title
	term.Write([]byte(be.colorScheme.Colorize("Create New Bulletin\n\n", "primary")))
	term.Write([]byte(be.colorScheme.Colorize("Enter bulletin title: ", "text")))
	title, err := term.ReadLine()
	if err != nil {
		return
	}

	if strings.TrimSpace(title) == "" {
		be.showMessage(term, "Title cannot be empty.", "error")
		return
	}

	// Get body (simple single-line for now)
	term.Write([]byte(be.colorScheme.Colorize("Enter bulletin body: ", "text")))
	body, err := term.ReadLine()
	if err != nil {
		return
	}

	if strings.TrimSpace(body) == "" {
		be.showMessage(term, "Body cannot be empty.", "error")
		return
	}

	// Create bulletin
	bulletin := &database.Bulletin{
		Title:  strings.TrimSpace(title),
		Body:   strings.TrimSpace(body),
		Author: "Sysop",
	}

	err = be.db.CreateBulletin(bulletin)
	if err != nil {
		be.showMessage(term, "Error creating bulletin: "+err.Error(), "error")
	} else {
		be.showMessage(term, "Bulletin created successfully!", "success")
	}
}

// EditBulletin edits an existing bulletin
func (be *BulletinEditor) EditBulletin(term *term.Terminal) {
	term.Write([]byte(ClearScreen + ShowCursor))

	term.Write([]byte(be.colorScheme.Colorize("Edit Bulletin\n\n", "primary")))
	term.Write([]byte(be.colorScheme.Colorize("Enter bulletin ID to edit: ", "text")))

	idStr, err := term.ReadLine()
	if err != nil {
		return
	}

	id, err := strconv.Atoi(strings.TrimSpace(idStr))
	if err != nil {
		be.showMessage(term, "Invalid ID format.", "error")
		return
	}

	// Get existing bulletin
	bulletin, err := be.db.GetBulletinByID(id)
	if err != nil {
		be.showMessage(term, "Bulletin not found.", "error")
		return
	}

	// Show current values and get new ones
	term.Write([]byte(be.colorScheme.Colorize(fmt.Sprintf("Current title: %s\n", bulletin.Title), "secondary")))
	term.Write([]byte(be.colorScheme.Colorize("New title (or press Enter to keep current): ", "text")))
	newTitle, err := term.ReadLine()
	if err != nil {
		return
	}

	if strings.TrimSpace(newTitle) == "" {
		newTitle = bulletin.Title
	}

	term.Write([]byte(be.colorScheme.Colorize(fmt.Sprintf("Current body: %s\n", bulletin.Body), "secondary")))
	term.Write([]byte(be.colorScheme.Colorize("New body (or press Enter to keep current): ", "text")))
	newBody, err := term.ReadLine()
	if err != nil {
		return
	}

	if strings.TrimSpace(newBody) == "" {
		newBody = bulletin.Body
	}

	// Update bulletin
	err = be.db.UpdateBulletin(id, strings.TrimSpace(newTitle), strings.TrimSpace(newBody))
	if err != nil {
		be.showMessage(term, "Error updating bulletin: "+err.Error(), "error")
	} else {
		be.showMessage(term, "Bulletin updated successfully!", "success")
	}
}

// DeleteBulletin deletes a bulletin
func (be *BulletinEditor) DeleteBulletin(term *term.Terminal) {
	term.Write([]byte(ClearScreen + ShowCursor))

	term.Write([]byte(be.colorScheme.Colorize("Delete Bulletin\n\n", "primary")))
	term.Write([]byte(be.colorScheme.Colorize("Enter bulletin ID to delete: ", "text")))

	idStr, err := term.ReadLine()
	if err != nil {
		return
	}

	id, err := strconv.Atoi(strings.TrimSpace(idStr))
	if err != nil {
		be.showMessage(term, "Invalid ID format.", "error")
		return
	}

	// Get bulletin to show what will be deleted
	bulletin, err := be.db.GetBulletinByID(id)
	if err != nil {
		be.showMessage(term, "Bulletin not found.", "error")
		return
	}

	// Confirm deletion
	term.Write([]byte(be.colorScheme.Colorize(fmt.Sprintf("Delete bulletin: %s\n", bulletin.Title), "secondary")))
	term.Write([]byte(be.colorScheme.Colorize("Are you sure? (y/N): ", "text")))

	confirm, err := term.ReadLine()
	if err != nil {
		return
	}

	if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
		err = be.db.DeleteBulletin(id)
		if err != nil {
			be.showMessage(term, "Error deleting bulletin: "+err.Error(), "error")
		} else {
			be.showMessage(term, "Bulletin deleted successfully!", "success")
		}
	} else {
		be.showMessage(term, "Deletion cancelled.", "text")
	}
}

// showMessage displays a message and waits for key press
func (be *BulletinEditor) showMessage(term *term.Terminal, message, colorType string) {
	term.Write([]byte(HideCursor))

	coloredMsg := be.colorScheme.Colorize(message, colorType)
	centeredMsg := be.colorScheme.CenterText(coloredMsg, 79)
	term.Write([]byte("\n" + centeredMsg))

	term.Write([]byte(ShowCursor))
	term.ReadLine()
	term.Write([]byte(HideCursor))
}
