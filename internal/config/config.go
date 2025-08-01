package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// SysopMenuOption represents a sysop menu option from config
type SysopMenuOption struct {
	ID          string `yaml:"id"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Command     string `yaml:"command"`
}

// SysopConfig represents the sysop menu configuration
type SysopConfig struct {
	Title        string            `yaml:"title"`
	Instructions string            `yaml:"instructions"`
	Options      []SysopMenuOption `yaml:"options"`
}

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	BBS      BBSConfig      `yaml:"bbs"`
	Sysop    SysopConfig    `yaml:"sysop"`
}

type ServerConfig struct {
	Port        int    `yaml:"port"`
	HostKeyPath string `yaml:"host_key_path"`
	MaxUsers    int    `yaml:"max_users"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type BBSConfig struct {
	SystemName    string      `yaml:"system_name"`
	SysopName     string      `yaml:"sysop_name"`
	WelcomeMsg    string      `yaml:"welcome_message"`
	MaxLineLength int         `yaml:"max_line_length"`
	Colors        ColorConfig `yaml:"colors"`
	Menus         []MenuItem  `yaml:"menus"`
}

type ColorConfig struct {
	Primary    string `yaml:"primary"`    // Main color (default: cyan)
	Secondary  string `yaml:"secondary"`  // Secondary color (default: red)
	Accent     string `yaml:"accent"`     // Accent color (default: yellow)
	Text       string `yaml:"text"`       // Normal text (default: white)
	Background string `yaml:"background"` // Background (default: black)
	Border     string `yaml:"border"`     // Borders and frames (default: blue)
	Success    string `yaml:"success"`    // Success messages (default: green)
	Error      string `yaml:"error"`      // Error messages (default: red)
	Highlight  string `yaml:"highlight"`  // Highlighted text (default: bright_white)
}

type MenuItem struct {
	ID          string     `yaml:"id"`
	Title       string     `yaml:"title"`
	Description string     `yaml:"description"`
	Command     string     `yaml:"command"`
	AccessLevel int        `yaml:"access_level"`
	Hotkey      string     `yaml:"hotkey,omitempty"`
	Submenu     []MenuItem `yaml:"submenu,omitempty"`
}

func Load(filename string) (*Config, error) {
	// Set default config
	config := &Config{
		Server: ServerConfig{
			Port:        2323,
			HostKeyPath: "host_key",
			MaxUsers:    100,
		},
		Database: DatabaseConfig{
			Path: "bbs.db",
		},
		BBS: BBSConfig{
			SystemName:    "Coastline BBS",
			SysopName:     "Sysop",
			WelcomeMsg:    "Welcome to Coastline BBS!",
			MaxLineLength: 79,
			Colors: ColorConfig{
				Primary:    "cyan",
				Secondary:  "red",
				Accent:     "yellow",
				Text:       "white",
				Background: "black",
				Border:     "blue",
				Success:    "green",
				Error:      "red",
				Highlight:  "bright_white",
			},
			Menus: []MenuItem{
				{
					ID:          "main",
					Title:       "Main Menu",
					Description: "Main BBS Menu",
					Command:     "main_menu",
					AccessLevel: 0,
					Submenu: []MenuItem{
						{ID: "bulletins", Title: "Bulletins", Description: "Read system bulletins", Command: "bulletins", AccessLevel: 0},
						{ID: "messages", Title: "Messages", Description: "Message areas", Command: "messages", AccessLevel: 0},
						{ID: "files", Title: "Files", Description: "File areas", Command: "files", AccessLevel: 0},
						{ID: "games", Title: "Games", Description: "Online games", Command: "games", AccessLevel: 0},
						{ID: "users", Title: "Users", Description: "User listings", Command: "users", AccessLevel: 0},
						{ID: "sysop", Title: "Sysop", Description: "System operator menu", Command: "sysop", AccessLevel: 255},
						{ID: "goodbye", Title: "Goodbye", Description: "Logoff system", Command: "goodbye", AccessLevel: 0},
					},
				},
			},
		},
		Sysop: SysopConfig{
			Title:        "System Administration",
			Instructions: "Use ↑↓ arrow keys to navigate, Enter to select, Q to quit",
			Options: []SysopMenuOption{
				{ID: "create_user", Title: "Create New User", Description: "1) Create New User Account", Command: "create_user"},
				{ID: "edit_user", Title: "Edit User Account", Description: "2) Edit User Account", Command: "edit_user"},
				{ID: "delete_user", Title: "Delete User Account", Description: "3) Delete User Account", Command: "delete_user"},
				{ID: "view_users", Title: "View All Users", Description: "4) View All Users", Command: "view_users"},
				{ID: "change_password", Title: "Change User Password", Description: "5) Change User Password", Command: "change_password"},
				{ID: "toggle_user", Title: "Toggle User Status", Description: "6) Toggle User Active Status", Command: "toggle_user"},
				{ID: "system_stats", Title: "System Statistics", Description: "7) System Statistics", Command: "system_stats"},
				{ID: "bulletin_management", Title: "Bulletin Management", Description: "8) Bulletin Management", Command: "bulletin_management"},
			},
		},
	}

	// Try to load config file if it exists
	if _, err := os.Stat(filename); err == nil {
		data, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, err
		}
	}

	return config, nil
}

func (c *Config) Save(filename string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
