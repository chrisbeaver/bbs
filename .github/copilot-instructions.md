<!-- Use this file to provide workspace-specific custom instructions to Copilot. For more details, visit https://code.visualstudio.com/docs/copilot/copilot-customization#_use-a-githubcopilotinstructionsmd-file -->

# Searchlight BBS Project Instructions

This is a Go project that recreates the classic Searchlight BBS software as an SSH terminal server.

## Project Context

- **Language**: Go 1.21+
- **Architecture**: SSH terminal server with goroutine-based concurrent connections
- **Database**: SQLite for data persistence
- **Configuration**: YAML-based flexible menu system
- **Terminal Interface**: Classic BBS experience over SSH

## Key Components

1. **SSH Server** (`internal/server/ssh.go`): Handles SSH connections and terminal sessions
2. **Database Layer** (`internal/database/database.go`): SQLite operations for users, messages, bulletins
3. **Configuration** (`internal/config/config.go`): YAML-based configuration management
4. **Main Server** (`main.go`): Entry point and server lifecycle management

## Coding Standards

- Use standard Go conventions and formatting
- Implement proper error handling with meaningful error messages
- Use goroutines for concurrent connection handling
- Keep database operations in the database package
- Configuration should be flexible and easily modifiable by sysops
- Maintain the classic BBS aesthetic in terminal output

## Key Dependencies

- `golang.org/x/crypto/ssh`: SSH server functionality
- `golang.org/x/term`: Terminal handling (replaces deprecated ssh/terminal)
- `github.com/mattn/go-sqlite3`: SQLite database driver
- `gopkg.in/yaml.v2`: YAML configuration parsing

## BBS-Specific Guidelines

- Menu systems should be configurable via YAML
- Access levels (0-255) control feature availability
- Terminal output should maintain 79-character line limits
- Use ANSI escape codes for screen clearing and formatting
- Implement classic BBS features: bulletins, messages, file areas, games
- Each user connection should be handled in its own goroutine
- Database operations should be safe for concurrent access

## Security Considerations

- Implement proper password hashing (currently using plain text for demo)
- Generate and manage SSH host keys securely
- Validate user input to prevent SQL injection
- Implement proper session management
- Use access levels to restrict sensitive operations

When working on this project, prioritize the authentic BBS experience while leveraging modern Go practices for reliability and performance.
