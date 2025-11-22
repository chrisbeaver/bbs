# Coastline BBS Quick Reference

This is a quick reference for common development tasks and patterns in the Coastline BBS codebase.

## Quick Start Commands

```bash
# Setup development environment
make dev-setup          # Install deps and setup database
make run                # Start SSH server on port 2323
go run main.go -l       # Start in local terminal mode

# Testing connections
ssh -p 2323 sysop@localhost    # Sysop access (password: password)
ssh -p 2323 test@localhost     # Regular user (password: test)

# Build and deploy
make build              # Build binary
make clean              # Clean artifacts
make setup              # Reset database
```

## Project Navigation

### Key Entry Points
- `main.go` → `cmd/root.go` - Application entry and CLI setup
- `cmd/root.go:runServerMode()` - SSH server startup
- `cmd/root.go:runLocalMode()` - Local terminal mode  
- `internal/server/session.go:Run()` - Session main loop
- `internal/server/session.go:menuLoop()` - Menu navigation
- `internal/server/session.go:executeCommand()` - Command dispatch

### Core Components Location
- **Server**: `internal/server/server.go` - SSH config, connection handling
- **Sessions**: `internal/server/session.go` - User session management
- **Database**: `internal/database/database.go` - SQLite operations
- **Config**: `internal/config/config.go` - YAML configuration loading
- **Menus**: Menu system spread across config + session.go
- **Modules**: `internal/modules/` - Feature implementations

## Common Code Patterns

### Adding a New Menu Command

1. **Add to config.yaml:**
```yaml
bbs:
  menus:
    - id: "main"
      submenu:
        - id: "my_feature"
          title: "My Feature" 
          description: "Description of my feature"
          command: "my_feature"
          access_level: 10
```

2. **Add command handler in session.go:**
```go
// In executeCommand() switch statement
case "my_feature":
    s.handleMyFeature()
    return true
```

3. **Implement handler method:**
```go
func (s *Session) handleMyFeature() {
    s.write([]byte(menu.ClearScreen))
    
    header := s.colorScheme.Colorize("My Feature", "primary")
    centeredHeader := s.colorScheme.CenterText(header, 79)
    s.write([]byte(centeredHeader + "\n\n"))
    
    // Your feature logic here
    s.write([]byte("Feature content...\n"))
    
    s.waitForKey() // Wait for user input before returning to menu
}
```

### Database Operations

**Add new table:**
```go
// In database.go createTables()
`CREATE TABLE IF NOT EXISTS my_table (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    data TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
)`
```

**Create data struct:**
```go
type MyData struct {
    ID        int       `json:"id"`
    Name      string    `json:"name"`
    Data      string    `json:"data"`
    CreatedAt time.Time `json:"created_at"`
}
```

**Add database methods:**
```go
func (db *DB) GetMyData(limit int) ([]MyData, error) {
    rows, err := db.conn.Query("SELECT id, name, data, created_at FROM my_table ORDER BY created_at DESC LIMIT ?", limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var items []MyData
    for rows.Next() {
        var item MyData
        err := rows.Scan(&item.ID, &item.Name, &item.Data, &item.CreatedAt)
        if err != nil {
            return nil, err
        }
        items = append(items, item)
    }
    return items, nil
}

func (db *DB) CreateMyData(data *MyData) error {
    _, err := db.conn.Exec("INSERT INTO my_table (name, data) VALUES (?, ?)", 
        data.Name, data.Data)
    return err
}
```

### Creating a Module

1. **Create module struct:**
```go
// internal/modules/mymodule/mymodule.go
package mymodule

import (
    "bbs/internal/database"
    "bbs/internal/menu"
    "bbs/internal/modules/base"
)

type Module struct {
    *base.Module
    db          *database.DB
    colorScheme menu.ColorScheme
}

func NewModule(db *database.DB, colorScheme menu.ColorScheme) *Module {
    m := &Module{
        db:          db,
        colorScheme: colorScheme,
    }
    m.Module = base.NewModule(db, colorScheme, m)
    return m
}
```

2. **Implement OptionProvider interface:**
```go
func (m *Module) LoadOptions(db *database.DB) ([]base.MenuOption, error) {
    // Load your options from database
    data, err := db.GetMyData(50)
    if err != nil {
        return nil, err
    }
    
    var options []base.MenuOption
    for i, item := range data {
        option := NewMyOption(&item, i, m.colorScheme)
        options = append(options, option)
    }
    return options, nil
}

func (m *Module) GetMenuTitle() string {
    return "My Feature Menu"
}

func (m *Module) GetInstructions() string {
    return "Navigate: ↑↓  Select: Enter  Quit: Q"
}
```

3. **Create option implementation:**
```go
// internal/modules/mymodule/my_option.go
type MyOption struct {
    data        *database.MyData
    index       int
    colorScheme menu.ColorScheme
}

func NewMyOption(data *database.MyData, index int, colorScheme menu.ColorScheme) *MyOption {
    return &MyOption{
        data:        data,
        index:       index,
        colorScheme: colorScheme,
    }
}

func (o *MyOption) GetTitle() string {
    return fmt.Sprintf("%d. %s", o.index+1, o.data.Name)
}

func (o *MyOption) GetDescription() string {
    return o.data.Data
}

func (o *MyOption) Execute(writer base.Writer, reader base.KeyReader) error {
    // Implement what happens when user selects this option
    writer.Write([]byte(menu.ClearScreen))
    
    header := o.colorScheme.Colorize(o.data.Name, "primary")
    writer.Write([]byte(header + "\n\n"))
    
    content := o.colorScheme.Colorize(o.data.Data, "text")
    writer.Write([]byte(content + "\n\n"))
    
    prompt := o.colorScheme.Colorize("Press any key to continue...", "highlight")
    writer.Write([]byte(prompt))
    
    _, err := reader.ReadKey()
    return err
}
```

4. **Integrate with session:**
```go
// In session.go executeCommand()
case "my_feature":
    myModule := mymodule.NewModule(s.db, s.colorScheme)
    writer := &TerminalWriter{session: s}
    keyReader := &TerminalKeyReader{session: s}
    myModule.Execute(writer, keyReader)
    return true
```

### Common Terminal Operations

```go
// Clear screen and position cursor
s.write([]byte(menu.ClearScreen))

// Write colored text
coloredText := s.colorScheme.Colorize("Hello", "primary")
s.write([]byte(coloredText + "\n"))

// Center text on screen (79 chars)
centered := s.colorScheme.CenterText("Title", 79)
s.write([]byte(centered + "\n"))

// Draw separator line
separator := s.colorScheme.DrawSeparator(len("Title"), "═")
s.write([]byte(separator + "\n"))

// Wait for any key
s.waitForKey()

// Read specific key
key, err := s.readKey()
if err != nil {
    return err
}
switch key {
case "enter":
    // Handle enter
case "q", "Q":
    // Handle quit
}
```

### Color Usage

Available colors (defined in config.yaml):
- `primary` - Main accent color (cyan)
- `secondary` - Secondary accent (red) 
- `accent` - Highlight color (yellow)
- `text` - Normal text (white)
- `background` - Background (black)
- `border` - Borders/lines (blue)
- `success` - Success messages (green)
- `error` - Error messages (red)
- `highlight` - Emphasis (bright_white)

```go
// Colorize text
text := s.colorScheme.Colorize("Success!", "success")
error := s.colorScheme.Colorize("Error occurred", "error")
title := s.colorScheme.Colorize("Menu Title", "primary")
```

### Access Level Checking

```go
// Check if user has required access level
if s.user.AccessLevel < requiredLevel {
    s.write([]byte(s.colorScheme.Colorize("Access denied", "error") + "\n"))
    s.waitForKey()
    return
}

// Access levels:
// 0 = Public
// 10 = Regular user  
// 255 = Sysop
```

## File Organization Patterns

### Module Structure
```
internal/modules/mymodule/
├── mymodule.go         # Main module struct and interfaces
├── my_option.go        # Option implementations  
├── handlers.go         # Command handlers
└── forms.go            # Form/input handling (if needed)
```

### Component Structure  
```
internal/components/
├── interfaces.go       # Interface definitions
├── form.go            # Form component
├── textinput.go       # Text input component
└── focus.go           # Focus management
```

## Testing Patterns

### Manual Testing Checklist
- [ ] SSH connection works
- [ ] Local mode works  
- [ ] All menu navigation functions
- [ ] Access levels are enforced
- [ ] Color scheme displays correctly
- [ ] Error handling works
- [ ] Graceful shutdown (Ctrl+C)

### Common Test Scenarios
```bash
# Test different user levels
ssh -p 2323 sysop@localhost  # Should see sysop menu
ssh -p 2323 test@localhost   # Should NOT see sysop menu

# Test invalid credentials
ssh -p 2323 invalid@localhost # Should be rejected

# Test concurrent connections
# Open multiple SSH sessions simultaneously

# Test local mode
go run main.go -l            # Should work without SSH
```

## Debugging Tips

### Common Issues and Solutions

**"SSH connection refused"**
- Check if server is running on port 2323
- Verify host key exists: `ls -la host_key`
- Check firewall settings

**"Menu not found"**
- Check config.yaml menu structure
- Verify menu ID matches in configuration
- Check access level requirements

**"Database errors"**
- Run `make setup` to reset database
- Check file permissions on bbs.db
- Verify SQLite3 is installed

### Debug Output
```go
// Add temporary debug output
fmt.Printf("DEBUG: User %s accessing menu %s\n", s.user.Username, s.currentMenu)
```

### VS Code Debugging
Use `.vscode/launch.json` configurations:
- "Launch BBS Server" - Debug main server
- "Debug Database Setup" - Debug setup process

## Performance Considerations

### Memory Usage
- Sessions are created per connection
- Database connections are persistent  
- Menu configurations loaded once at startup

### Concurrency
- Each SSH connection runs in separate goroutine
- Database operations are thread-safe
- No shared session state

### Resource Limits
- `max_users` in config.yaml controls concurrent connections
- No automatic session cleanup (TODO)
- SSH idle timeout handled by SSH client

---

For detailed information, see the full [DEVELOPER_GUIDE.md](DEVELOPER_GUIDE.md).