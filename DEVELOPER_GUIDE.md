# Coastline BBS Developer Guide

Welcome to the Coastline BBS project! This guide will help you understand the codebase architecture, key components, and how to extend the system.

## Table of Contents

1. [Project Overview](#project-overview)
2. [Architecture](#architecture)
3. [Project Structure](#project-structure)
4. [Key Components](#key-components)
5. [Development Setup](#development-setup)
6. [Database Layer](#database-layer)
7. [Terminal Interface](#terminal-interface)
8. [Menu System](#menu-system)
9. [Module System](#module-system)
10. [Adding New Features](#adding-new-features)
11. [Configuration](#configuration)
12. [Testing](#testing)
13. [Deployment](#deployment)

## Project Overview

Coastline BBS is a modern recreation of classic bulletin board systems (BBS), accessible via SSH. It's written in Go and provides an authentic terminal-based experience while leveraging modern technology for reliability and security.

### Key Technologies
- **Language**: Go 1.24+
- **Database**: SQLite for data persistence
- **Networking**: SSH server for secure connections
- **Configuration**: YAML-based flexible menu system
- **Concurrency**: Goroutines for handling multiple connections

### Design Philosophy
- Authentic BBS experience with modern reliability
- Modular architecture for easy extension
- Configuration-driven menu system
- Secure by default with SSH encryption

## Architecture

The system follows a layered architecture:

```
┌─────────────────┐
│   SSH Client    │
└─────────────────┘
         │
┌─────────────────┐
│  SSH Server     │  ← Entry point, handles connections
└─────────────────┘
         │
┌─────────────────┐
│   Session       │  ← Manages user session state
└─────────────────┘
         │
┌─────────────────┐
│  Menu System    │  ← Configuration-driven navigation
└─────────────────┘
         │
┌─────────────────┐
│   Modules       │  ← Feature implementations (bulletins, messages, etc.)
└─────────────────┘
         │
┌─────────────────┐
│   Database      │  ← SQLite persistence layer
└─────────────────┘
```

### Concurrency Model
- Each SSH connection spawns a new goroutine
- Database operations are thread-safe
- Sessions are isolated from each other
- Graceful shutdown handling with signal management

## Project Structure

```
bbs/
├── main.go                           # Entry point (delegates to cmd/)
├── config.yaml                       # Main configuration file
├── go.mod                           # Go module definition
├── Makefile                         # Build automation
├── README.md                        # Project README
│
├── cmd/                             # Command-line interface
│   ├── root.go                      # Cobra CLI setup, server/local mode
│   └── setup/
│       └── main.go                  # Database initialization utility
│
└── internal/                        # Private packages
    ├── components/                   # UI components (forms, inputs)
    │   ├── focus.go
    │   ├── form.go
    │   ├── interfaces.go
    │   └── textinput.go
    │
    ├── config/                      # Configuration management
    │   └── config.go                # YAML config parsing
    │
    ├── database/                    # Data persistence layer
    │   ├── database.go              # SQLite operations, schemas
    │   └── seed.go                  # Default data seeding
    │
    ├── menu/                        # Menu rendering system
    │   └── menu.go                  # Menu display and navigation
    │
    ├── modules/                     # Feature modules
    │   ├── module.go                # Module interfaces
    │   ├── base/                    # Base module functionality
    │   │   ├── base.go
    │   │   └── command.go
    │   ├── bulletins/               # Bulletin system
    │   │   ├── bulletins.go
    │   │   └── bulletin_option.go
    │   └── sysop/                   # System operator tools
    │       ├── sysop.go
    │       ├── handlers.go
    │       ├── bulletin_editor.go
    │       └── user_editor/         # User management
    │           ├── user_editor.go
    │           ├── create.go
    │           ├── edit.go
    │           ├── delete.go
    │           ├── list.go
    │           ├── password.go
    │           ├── status.go
    │           └── utils.go
    │
    ├── pager/                       # Text pagination system
    │   ├── pager.go
    │   ├── interfaces.go
    │   └── adapters.go
    │
    ├── server/                      # Server core
    │   ├── server.go                # Main server, SSH config
    │   ├── session.go               # User session management
    │   ├── colors.go                # ANSI color schemes
    │   └── hostkey.go               # SSH key management
    │
    ├── statusbar/                   # Status bar system
    │   ├── statusbar.go
    │   └── manager.go
    │
    └── terminal/                    # Terminal abstraction
        ├── terminal.go              # Interface definition
        ├── ssh.go                   # SSH terminal implementation
        └── local.go                 # Local terminal implementation
```

## Key Components

### 1. Main Entry Point (`main.go`)
The entry point simply delegates to the Cobra CLI system in `cmd/`.

### 2. Command System (`cmd/`)
- **`root.go`**: Implements Cobra CLI with two modes:
  - Server mode (default): Starts SSH server
  - Local mode (`-l`): Direct terminal connection
- **`setup/main.go`**: Database initialization and seeding

### 3. Server (`internal/server/`)
- **`server.go`**: Core server struct, SSH configuration, connection handling
- **`session.go`**: User session management, menu navigation, command execution
- **`colors.go`**: ANSI color scheme management
- **`hostkey.go`**: SSH host key generation and management

### 4. Configuration (`internal/config/`)
YAML-based configuration system supporting:
- Server settings (port, host key, max users)
- Database configuration
- BBS settings (system name, colors, etc.)
- Dynamic menu structure with access levels

### 5. Database (`internal/database/`)
SQLite-based persistence with:
- **Users**: Authentication and profile data
- **Messages**: Private messaging system
- **Bulletins**: System announcements
- **Sessions**: Active connection tracking

### 6. Terminal Abstraction (`internal/terminal/`)
Unified interface for both SSH and local terminal connections:
- Input/output operations
- Terminal size management
- Raw mode handling

### 7. Module System (`internal/modules/`)
Pluggable feature system:
- **Base module**: Common functionality for all modules
- **Bulletins**: System bulletin management
- **Sysop**: Administrative tools

## Development Setup

### Prerequisites
- Go 1.24 or later
- SQLite3
- SSH client for testing

### Initial Setup

1. **Clone and build:**
   ```bash
   git clone <repository>
   cd bbs
   make deps          # Install Go dependencies
   ```

2. **Initialize database:**
   ```bash
   make setup         # Creates bbs.db with seed data
   ```

3. **Start development server:**
   ```bash
   make run           # Starts SSH server on port 2323
   ```

4. **Connect for testing:**
   ```bash
   ssh -p 2323 sysop@localhost    # Password: password
   ssh -p 2323 test@localhost     # Password: test
   ```

### Development Tools

**VS Code Integration:**
- `.vscode/launch.json`: Debug configurations
- `.vscode/tasks.json`: Build tasks

**Make Targets:**
- `make build`: Compile binary
- `make setup`: Initialize database
- `make run`: Start server
- `make start`: Setup + run
- `make clean`: Remove build artifacts
- `make fmt`: Format code
- `make lint`: Run linter

**Testing Local Mode:**
```bash
go run main.go -l    # Direct terminal connection (no SSH)
```

## Database Layer

### Schema Design

```sql
-- Users table
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,           -- TODO: Implement proper hashing
    real_name TEXT,
    email TEXT,
    access_level INTEGER DEFAULT 0,   -- 0-255 access control
    last_call DATETIME,
    total_calls INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT 1
);

-- Messages table  
CREATE TABLE messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    from_user TEXT NOT NULL,
    to_user TEXT NOT NULL,
    subject TEXT NOT NULL,
    body TEXT NOT NULL,
    area TEXT DEFAULT 'general',      -- Message areas/forums
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    is_read BOOLEAN DEFAULT 0
);

-- Bulletins table
CREATE TABLE bulletins (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    author TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME               -- Optional expiration
);

-- Sessions table (active connections)
CREATE TABLE sessions (
    id TEXT PRIMARY KEY,              -- Session UUID
    username TEXT,
    start_time DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_activity DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Access Patterns

**User Management:**
```go
// Get user by credentials
user, err := db.GetUserByUsername(username)
if err != nil || user.Password != password {
    return nil, fmt.Errorf("invalid credentials")
}

// Update user activity
err := db.UpdateUserActivity(userID)

// Create new user (sysop only)
user := &database.User{
    Username:    "newuser",
    Password:    "password",  // TODO: Hash this
    AccessLevel: 10,
}
err := db.CreateUser(user)
```

**Bulletin Operations:**
```go
// Get recent bulletins
bulletins, err := db.GetBulletins(50)

// Create new bulletin (sysop only) 
bulletin := &database.Bulletin{
    Title:  "System Maintenance",
    Body:   "System will be down...",
    Author: "Sysop",
}
err := db.CreateBulletin(bulletin)
```

### Migration Strategy
Currently, the system recreates tables on startup. For production:
1. Implement proper migration system
2. Version the database schema
3. Add rollback capabilities

## Terminal Interface

The terminal system provides a unified interface for both SSH and local connections.

### Terminal Interface
```go
type Terminal interface {
    io.ReadWriter
    SetSize(width int, height int) error
    Size() (width int, height int, error error)
    MakeRaw() error
    Restore() error
    Close() error
    ReadLine() (string, error)
    SetPrompt(prompt string)
}
```

### SSH Terminal (`internal/terminal/ssh.go`)
Wraps SSH channel for terminal operations:
- Handles SSH-specific terminal modes
- Manages window size changes
- Implements raw mode for character-by-character input

### Local Terminal (`internal/terminal/local.go`)
Direct terminal interface for local connections:
- Uses `golang.org/x/term` for terminal control
- Handles Unix terminal modes
- Supports direct testing without SSH

### Color System
ANSI color support with configurable schemes:
```go
type ColorScheme struct {
    colors map[string]string
}

func (cs *ColorScheme) Colorize(text, colorName string) string {
    if color, exists := cs.colors[colorName]; exists {
        return color + text + "\033[0m"  // Reset
    }
    return text
}
```

**Supported Colors:**
- `primary`, `secondary`, `accent`
- `text`, `background`, `border`  
- `success`, `error`, `highlight`

## Menu System

The menu system is entirely configuration-driven, allowing dynamic menu structures without code changes.

### Menu Configuration
```yaml
bbs:
  menus:
    - id: "main"
      title: "Main Menu"
      description: "Main BBS Menu"
      command: "main_menu"
      access_level: 0
      submenu:
        - id: "bulletins"
          title: "Bulletins"
          description: "Read system bulletins"
          command: "bulletins"
          access_level: 0
        - id: "sysop"
          title: "Sysop Menu"
          description: "System administration"
          command: "sysop_menu"
          access_level: 255
          submenu:
            - id: "user_management"
              title: "User Management"
              description: "Manage user accounts"
              command: "user_management"
              access_level: 255
```

### Menu Navigation
- **Arrow keys**: Navigate up/down
- **Enter**: Select menu item
- **Q**: Return to previous menu (submenus only)
- **G**: Goodbye (exit from any menu)

### Access Control
Each menu item has an `access_level` (0-255):
- `0`: Public access
- `10`: Regular user features
- `255`: Sysop-only features

Menu items are filtered based on user's access level.

### Menu Rendering
The `MenuRenderer` handles:
- ANSI screen clearing
- Centered title display
- Color-coded menu items
- Status bar integration
- Navigation hints

## Module System

The module system provides a pluggable architecture for BBS features.

### Base Module (`internal/modules/base/`)
Common functionality for all modules:
- Menu option loading from database
- Standard navigation handling
- Color scheme management
- Input validation

### Module Interface
```go
type OptionProvider interface {
    LoadOptions(db *database.DB) ([]MenuOption, error)
    GetMenuTitle() string
    GetInstructions() string
}

type MenuOption interface {
    GetTitle() string
    GetDescription() string
    Execute(writer Writer, reader KeyReader) error
}
```

### Creating New Modules

1. **Create module struct:**
   ```go
   type MyModule struct {
       *base.Module
       db          *database.DB
       colorScheme menu.ColorScheme
   }
   ```

2. **Implement OptionProvider:**
   ```go
   func (m *MyModule) LoadOptions(db *database.DB) ([]base.MenuOption, error) {
       // Load options from database or configuration
   }
   
   func (m *MyModule) GetMenuTitle() string {
       return "My Feature"
   }
   
   func (m *MyModule) GetInstructions() string {
       return "Navigate: ↑↓  Select: Enter  Quit: Q"
   }
   ```

3. **Create option implementations:**
   ```go
   type MyOption struct {
       title       string
       description string
       data        interface{}
       colorScheme menu.ColorScheme
   }
   
   func (o *MyOption) Execute(writer Writer, reader KeyReader) error {
       // Implement the feature logic
   }
   ```

4. **Integrate with session:**
   ```go
   // In session.go executeCommand()
   case "my_feature":
       myModule := myfeature.NewModule(s.db, s.colorScheme)
       writer := &TerminalWriter{session: s}
       keyReader := &TerminalKeyReader{session: s}
       myModule.Execute(writer, keyReader)
   ```

### Existing Modules

**Bulletins Module:**
- Loads bulletins from database
- Provides paginated display
- Supports bulletin reading and navigation

**Sysop Module:**
- User management (create, edit, delete users)
- Bulletin management
- System statistics
- Access level management

## Adding New Features

### 1. Database-Driven Features

For features that require data persistence:

1. **Add database tables:**
   ```go
   // In database.go createTables()
   `CREATE TABLE IF NOT EXISTS my_table (
       id INTEGER PRIMARY KEY AUTOINCREMENT,
       name TEXT NOT NULL,
       data TEXT,
       created_at DATETIME DEFAULT CURRENT_TIMESTAMP
   )`
   ```

2. **Create data structures:**
   ```go
   type MyData struct {
       ID        int       `json:"id"`
       Name      string    `json:"name"`
       Data      string    `json:"data"`
       CreatedAt time.Time `json:"created_at"`
   }
   ```

3. **Add database methods:**
   ```go
   func (db *DB) GetMyData(limit int) ([]MyData, error) {
       // Implementation
   }
   
   func (db *DB) CreateMyData(data *MyData) error {
       // Implementation
   }
   ```

### 2. Menu-Driven Features

For configuration-driven menu features:

1. **Add menu configuration:**
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

2. **Implement command handler:**
   ```go
   // In session.go executeCommand()
   case "my_feature":
       s.handleMyFeature()
       return true
   ```

3. **Create handler method:**
   ```go
   func (s *Session) handleMyFeature() {
       s.write([]byte(menu.ClearScreen))
       
       header := s.colorScheme.Colorize("My Feature", "primary")
       s.write([]byte(header + "\n\n"))
       
       // Feature implementation
       
       s.waitForKey()
   }
   ```

### 3. Interactive Features

For complex interactive features:

1. **Create module structure**
2. **Implement form components** (see `internal/components/`)
3. **Handle user input validation**
4. **Provide status feedback**

Example form integration:
```go
form := components.NewForm()
form.AddField("name", "Name:", "", true)
form.AddField("email", "Email:", "", false)

result, err := form.Execute(writer, keyReader)
if err != nil {
    return err
}
```

## Configuration

### Configuration Structure

The system uses a hierarchical YAML configuration:

```yaml
server:                    # SSH server settings
  port: 2323
  host_key_path: "host_key"
  max_users: 100

database:                  # Database settings
  path: "bbs.db"

bbs:                       # BBS-specific settings
  system_name: "Coastline BBS"
  sysop_name: "Sysop"
  welcome_message: |
    Welcome to Coastline BBS!
    A classic bulletin board system experience over SSH.
  max_line_length: 79
  
  colors:                  # Color scheme
    primary: "cyan"
    secondary: "red"
    # ... more colors
    
  menus:                   # Menu structure
    - id: "main"
      title: "Main Menu"
      # ... menu definition
```

### Loading Configuration

```go
// Default configuration with fallbacks
config := &Config{
    Server: ServerConfig{
        Port:        2323,
        HostKeyPath: "host_key",
        MaxUsers:    100,
    },
    // ... other defaults
}

// Load from file if it exists
if _, err := os.Stat(filename); err == nil {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }
    
    if err := yaml.Unmarshal(data, config); err != nil {
        return nil, fmt.Errorf("failed to parse config file: %w", err)
    }
}
```

### Environment Integration

Configuration supports environment variable override through Viper:
```bash
COASTLINE_SERVER_PORT=2323 go run main.go
```

## Testing

### Manual Testing

**SSH Testing:**
```bash
# Test sysop access
ssh -p 2323 sysop@localhost

# Test regular user
ssh -p 2323 test@localhost

# Test local mode
go run main.go -l
```

**Feature Testing:**
1. Test all menu navigation
2. Verify access level restrictions
3. Test bulletin system
4. Test user management (sysop only)
5. Verify graceful shutdown

### Unit Testing (TODO)

Framework for unit testing:
```go
// Example test structure
func TestUserAuthentication(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    // Test valid credentials
    user, err := db.GetUserByUsername("test")
    assert.NoError(t, err)
    assert.Equal(t, "test", user.Username)
    
    // Test invalid credentials
    user, err = db.GetUserByUsername("nonexistent")
    assert.Error(t, err)
}
```

### Integration Testing

Test complete workflows:
1. User login flow
2. Menu navigation
3. Module execution
4. Database operations
5. Session management

## Deployment

### Production Considerations

**Security:**
1. Implement proper password hashing
2. Generate secure SSH host keys
3. Configure firewall rules
4. Set up log rotation
5. Implement rate limiting

**Performance:**
1. Database connection pooling
2. Session cleanup
3. Memory management
4. Connection limits

**Monitoring:**
1. Connection logging
2. Error tracking
3. Performance metrics
4. User activity monitoring

### Deployment Options

**Systemd Service:**
```ini
[Unit]
Description=Coastline BBS Server
After=network.target

[Service]
Type=simple
User=bbs
WorkingDirectory=/opt/bbs
ExecStart=/opt/bbs/bbs
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

**Docker Deployment:**
```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o bbs main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/bbs .
COPY config.yaml .
CMD ["./bbs"]
```

### Configuration Management

**Production config.yaml:**
- Change default passwords
- Configure proper database path
- Set appropriate max users
- Customize welcome message
- Configure SSL/TLS if needed

## Contributing

### Code Style
- Follow standard Go conventions
- Use `go fmt` for formatting
- Implement proper error handling
- Add meaningful comments for exported functions
- Use meaningful variable and function names

### Git Workflow
1. Create feature branches
2. Write descriptive commit messages
3. Test thoroughly before committing
4. Keep commits atomic and focused

### Documentation
- Update this guide for architectural changes
- Document new modules and features
- Include configuration examples
- Provide upgrade instructions

## Future Enhancements

### Planned Features
- [ ] Private messaging system
- [ ] File transfer areas
- [ ] Door games integration
- [ ] Multi-line chat
- [ ] Email gateway
- [ ] Web interface
- [ ] API endpoints

### Technical Improvements
- [ ] Password hashing (bcrypt)
- [ ] Database migrations
- [ ] Comprehensive test suite
- [ ] Performance optimization
- [ ] Logging framework
- [ ] Configuration validation

### Security Enhancements
- [ ] Rate limiting
- [ ] Session timeout
- [ ] Input sanitization
- [ ] Audit logging
- [ ] Two-factor authentication
- [ ] Public key authentication

---

This guide should provide a solid foundation for understanding and extending the Coastline BBS codebase. For questions or clarifications, please reach out to the development team or create an issue in the project repository.