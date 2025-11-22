# Coastline BBS Architecture Reference

This document provides a technical overview of the system architecture, key interfaces, and data flow patterns.

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    SSH Client                               │
│                (Terminal Emulator)                         │
└─────────────────────┬───────────────────────────────────────┘
                      │ SSH Protocol (Port 2323)
┌─────────────────────▼───────────────────────────────────────┐
│                SSH Server                                   │
│  - Authentication (username/password)                      │
│  - Connection handling                                      │  
│  - SSH channel management                                   │
└─────────────────────┬───────────────────────────────────────┘
                      │ Go Goroutines
┌─────────────────────▼───────────────────────────────────────┐
│               Session Manager                              │
│  - User session state                                      │
│  - Menu navigation                                          │
│  - Command routing                                          │
│  - Terminal I/O handling                                    │
└─────────────────────┬───────────────────────────────────────┘
                      │
        ┌─────────────┼─────────────┐
        │             │             │
┌───────▼──────┐ ┌────▼────┐ ┌─────▼─────┐
│ Menu System  │ │ Modules │ │ Terminal  │
│ - Config-    │ │ - Base  │ │ - SSH     │
│   driven     │ │ - Bulls │ │ - Local   │
│ - Access     │ │ - Sysop │ │ - Colors  │
│   control    │ │ - Forms │ │           │
└───────┬──────┘ └────┬────┘ └─────┬─────┘
        │             │            │
        └─────────────┼────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                Database Layer                               │
│  - SQLite operations                                        │
│  - User management                                          │
│  - Content storage (bulletins, messages)                   │
│  - Session tracking                                         │
└─────────────────────────────────────────────────────────────┘
```

## Core Interfaces

### Terminal Interface
```go
// internal/terminal/terminal.go
type Terminal interface {
    io.ReadWriter                                    // Basic I/O
    SetSize(width int, height int) error            // Terminal sizing  
    Size() (width int, height int, error error)     // Get terminal size
    MakeRaw() error                                 // Enter raw mode
    Restore() error                                 // Restore normal mode
    Close() error                                   // Cleanup
    ReadLine() (string, error)                      // Read line input
    SetPrompt(prompt string)                        // Set input prompt
}
```

**Implementations:**
- `SSHTerminal`: Wraps SSH channel for remote connections
- `LocalTerminal`: Direct terminal access for local connections

### Module System Interfaces

```go
// internal/modules/base/base.go
type OptionProvider interface {
    LoadOptions(db *database.DB) ([]MenuOption, error)  // Load menu options
    GetMenuTitle() string                               // Menu title
    GetInstructions() string                           // User instructions  
}

type MenuOption interface {
    GetTitle() string                                  // Option display title
    GetDescription() string                           // Option description  
    Execute(writer Writer, reader KeyReader) error   // Execute option
}

type Writer interface {
    Write([]byte) (int, error)                       // Terminal output
}

type KeyReader interface {
    ReadKey() (string, error)                        // Read single keypress
}
```

### Database Interface

```go
// internal/database/database.go
type DB struct {
    conn *sql.DB                                     // SQLite connection
}

// Core entities
type User struct {
    ID          int        `json:"id"`
    Username    string     `json:"username"`
    Password    string     `json:"password"`          // TODO: Hash
    AccessLevel int        `json:"access_level"`      // 0-255
    LastCall    *time.Time `json:"last_call"`
    TotalCalls  int        `json:"total_calls"`
    // ... more fields
}

type Bulletin struct {
    ID        int        `json:"id"`
    Title     string     `json:"title"`
    Body      string     `json:"body"`
    Author    string     `json:"author"`
    CreatedAt time.Time  `json:"created_at"`
    ExpiresAt *time.Time `json:"expires_at"`
}

type Message struct {
    ID        int       `json:"id"`
    FromUser  string    `json:"from_user"`
    ToUser    string    `json:"to_user"`
    Subject   string    `json:"subject"`
    Body      string    `json:"body"`
    Area      string    `json:"area"`              // Message board area
    CreatedAt time.Time `json:"created_at"`
    IsRead    bool      `json:"is_read"`
}
```

### Configuration System

```go
// internal/config/config.go
type Config struct {
    Server   ServerConfig          `yaml:"server"`
    Database DatabaseConfig        `yaml:"database"`
    BBS      BBSConfig             `yaml:"bbs"`
    Modules  map[string]MenuConfig `yaml:",inline"`
}

type BBSConfig struct {
    SystemName    string      `yaml:"system_name"`
    SysopName     string      `yaml:"sysop_name"`
    WelcomeMsg    string      `yaml:"welcome_message"`
    MaxLineLength int         `yaml:"max_line_length"`
    Colors        ColorConfig `yaml:"colors"`
    Menus         []MenuItem  `yaml:"menus"`
}

type MenuItem struct {
    ID          string     `yaml:"id"`
    Title       string     `yaml:"title"`
    Description string     `yaml:"description"`
    Command     string     `yaml:"command"`
    AccessLevel int        `yaml:"access_level"`
    Submenu     []MenuItem `yaml:"submenu"`
}
```

## Data Flow Patterns

### Connection Lifecycle

```
1. SSH Client connects to port 2323
   ├─ SSH handshake and authentication
   └─ User credentials validated against database

2. Session Creation  
   ├─ New goroutine spawned for connection
   ├─ Session struct initialized with user data
   ├─ Terminal wrapper created (SSH or Local)
   └─ Color scheme loaded from config

3. Welcome Sequence
   ├─ Welcome message displayed
   ├─ Bulletins module auto-executed
   └─ Menu system initialized

4. Menu Loop
   ├─ Current menu loaded from config
   ├─ Menu items filtered by user access level
   ├─ Menu displayed with navigation
   └─ User input processed
       ├─ Arrow keys: Navigation
       ├─ Enter: Execute selected command  
       ├─ Q: Return to previous menu
       └─ G: Goodbye (exit)

5. Command Execution
   ├─ Command routed to appropriate handler
   ├─ Module or function executed
   ├─ Database operations performed as needed
   └─ Return to menu loop

6. Session Cleanup
   ├─ User activity updated in database
   ├─ Terminal restored to normal mode
   └─ Connection closed
```

### Menu System Flow

```
Configuration Loading (config.yaml)
    ↓
Menu Tree Construction
    ↓  
Access Level Filtering (per user)
    ↓
Menu Rendering (ANSI colors + layout)
    ↓
User Navigation (arrow keys)
    ↓
Command Execution
    ↓
Module/Handler Execution
    ↓
Return to Menu Loop
```

### Module Execution Pattern

```go
// Pattern used by modules like bulletins, sysop tools
type Module struct {
    *base.Module
    db          *database.DB
    colorScheme menu.ColorScheme
}

// Execution flow:
1. Module.LoadOptions(db) -> []MenuOption
2. Base module displays menu with options
3. User navigates and selects option
4. Option.Execute(writer, reader) called
5. Option performs its function
6. Returns to module menu or exits
```

### Database Operation Patterns

**Read Operations:**
```go
func (db *DB) GetBulletins(limit int) ([]Bulletin, error) {
    rows, err := db.conn.Query(`
        SELECT id, title, body, author, created_at, expires_at 
        FROM bulletins 
        WHERE expires_at IS NULL OR expires_at > datetime('now')
        ORDER BY created_at DESC 
        LIMIT ?`, limit)
    // ... scan results into structs
    return bulletins, nil
}
```

**Write Operations:**
```go
func (db *DB) CreateBulletin(bulletin *Bulletin) error {
    _, err := db.conn.Exec(`
        INSERT INTO bulletins (title, body, author, expires_at) 
        VALUES (?, ?, ?, ?)`,
        bulletin.Title, bulletin.Body, bulletin.Author, bulletin.ExpiresAt)
    return err
}
```

**Transaction Pattern:**
```go
func (db *DB) UpdateUserWithHistory(user *User) error {
    tx, err := db.conn.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Multiple operations in transaction
    _, err = tx.Exec("UPDATE users SET last_call = ? WHERE id = ?", time.Now(), user.ID)
    if err != nil {
        return err
    }
    
    _, err = tx.Exec("INSERT INTO login_history (...) VALUES (...)")
    if err != nil {
        return err
    }
    
    return tx.Commit()
}
```

## Key Design Patterns

### 1. Terminal Abstraction Pattern
```go
// Unified interface for SSH and local terminals
type Terminal interface { ... }

// Used by session without knowing connection type
session := server.NewSession(terminal, username)
session.Run()
```

### 2. Configuration-Driven Menus
```yaml
# Menu structure defined in config, not code
menus:
  - id: "main"
    submenu:
      - id: "bulletins"
        command: "bulletins"
        access_level: 0
```

### 3. Module Plugin Pattern
```go
// Modules implement standard interfaces
type BulletinModule struct {
    *base.Module
}

func (m *BulletinModule) LoadOptions(db) ([]MenuOption, error)
func (m *BulletinModule) Execute(writer, reader)
```

### 4. Access Control Pattern
```go
// Consistent access level checking throughout system
if s.user.AccessLevel < item.AccessLevel {
    // Hide menu item or deny access
    continue
}
```

### 5. Color Scheme Pattern
```go
// Centralized color management
type ColorScheme struct {
    colors map[string]string
}

func (cs *ColorScheme) Colorize(text, colorName string) string
func (cs *ColorScheme) CenterText(text string, width int) string
func (cs *ColorScheme) DrawSeparator(length int, char string) string
```

## Concurrency Model

### Goroutine Usage
```go
// Main server accepts connections
for {
    conn, err := listener.Accept()
    if err != nil {
        continue
    }
    go bbsServer.HandleConnection(conn)  // New goroutine per connection
}

// Each connection gets isolated session
func (s *Server) HandleConnection(netConn net.Conn) {
    sshConn, _, _, err := ssh.NewServerConn(netConn, s.sshConfig)
    // ... setup session
    session.Run()  // Blocks until user disconnects
}
```

### Thread Safety
- Database operations are thread-safe (SQLite handles locking)
- Sessions are isolated (no shared state between connections)
- Configuration loaded once at startup (read-only after init)
- Each user gets independent menu state and navigation history

### Resource Management
```go
// Session cleanup pattern
defer func() {
    if s.user != nil {
        s.db.UpdateUserActivity(s.user.Username)
    }
    s.terminal.Restore()
    s.terminal.Close()
}()
```

## Security Architecture

### Authentication Flow
```
1. SSH password authentication
2. Lookup user in database
3. Plain text password comparison (TODO: bcrypt)
4. Set session user context
5. Access levels enforced per menu/command
```

### Access Control
- Menu items filtered by user access level (0-255)
- Commands check access level before execution  
- Database operations don't enforce access control (done at session level)
- No role-based permissions (just numeric levels)

### Data Validation
```go
// Input validation patterns
func validateUsername(username string) error {
    if len(username) == 0 || len(username) > 50 {
        return fmt.Errorf("invalid username length")
    }
    // Additional validation...
    return nil
}
```

## Extension Points

### Adding New Commands
1. Add menu item to `config.yaml`
2. Add case to `session.executeCommand()` 
3. Implement handler method

### Adding New Modules  
1. Implement `OptionProvider` interface
2. Create option implementations
3. Add to command routing
4. Follow base module patterns

### Adding Database Tables
1. Add CREATE TABLE to `createTables()`
2. Define struct type  
3. Implement CRUD methods
4. Update seed data if needed

### Adding Configuration
1. Add fields to config structs
2. Set defaults in `config.Load()`
3. Use throughout application
4. Document in config.yaml comments

---

This architecture supports the classic BBS experience while providing modern extensibility and maintainability. The modular design allows new features to be added without modifying core systems.