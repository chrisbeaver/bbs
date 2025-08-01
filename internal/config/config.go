package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// MenuOption represents a generic menu option from config
type MenuOption struct {
	ID          string `yaml:"id"`
	Title       string `yaml:"title"`
	Description string `yaml:"description"`
	Command     string `yaml:"command"`
}

// MenuConfig represents a generic menu configuration
type MenuConfig struct {
	Title        string       `yaml:"title"`
	Instructions string       `yaml:"instructions"`
	Options      []MenuOption `yaml:"options"`
}

type Config struct {
	Server   ServerConfig          `yaml:"server"`
	Database DatabaseConfig        `yaml:"database"`
	BBS      BBSConfig             `yaml:"bbs"`
	Modules  map[string]MenuConfig `yaml:",inline"`
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
	// Set minimal default config
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
		},
		Modules: make(map[string]MenuConfig),
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

// GetMenuConfig returns a menu configuration by name for generic access
func (c *Config) GetMenuConfig(name string) *MenuConfig {
	if config, exists := c.Modules[name]; exists {
		return &config
	}
	return nil
}
