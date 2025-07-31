# BBS

A modern classic BBS software project, implemented in Go with SSH connectivity.

## Features

-   **SSH Terminal Interface**: Connect using any SSH client
-   **Multi-user Support**: Concurrent connections handled by goroutines
-   **SQLite Database**: Lightweight, embedded database for user data, messages, and bulletins
-   **Flexible Menu System**: Easily configurable menus via YAML configuration
-   **Classic BBS Experience**: Authentic bulletin board system feel
-   **User Management**: User accounts with access levels
-   **Message Areas**: Private messaging between users
-   **Bulletin System**: System announcements and information
-   **Configurable**: Easy customization through config.yaml

## Quick Start

1. **Initialize the database and load seed data:**

    ```bash
    go run cmd/setup/main.go
    ```

2. **Start the BBS server:**

    ```bash
    go run main.go
    ```

3. **Connect via SSH:**

    ```bash
    ssh -p 2323 sysop@localhost
    # Default password: password

    # Or connect as test user:
    ssh -p 2323 test@localhost
    # Password: test
    ```

## Configuration and Data

The BBS is configured through `config.yaml` for system settings:

-   **Server settings**: Port, host key path, max users
-   **Database**: SQLite database file path
-   **BBS settings**: System name, sysop name, welcome message
-   **Menu system**: Fully configurable menu structure with access levels
-   **Seed data**: Initial users and bulletins are built into the code and loaded during setup

### Default Users

The setup creates two default users:

-   **sysop** (password: password) - Full system access (level 255)
-   **test** (password: test) - Regular user access (level 10)

### Menu System

The menu system is highly flexible and defined in the configuration file. Each menu item can have:

-   `id`: Unique identifier
-   `title`: Display title
-   `description`: User-friendly description
-   `command`: Command to execute
-   `access_level`: Minimum user access level required (0-255)
-   `submenu`: Nested menu items

## Project Structure

```
bbs/
├── main.go                 # Main server entry point
├── config.yaml            # Configuration file
├── cmd/
│   └── setup/
│       └── main.go        # Database setup utility
└── internal/
    ├── config/
    │   └── config.go      # Configuration management
    ├── database/
    │   └── database.go    # SQLite database operations
    └── server/
        └── ssh.go         # SSH server and BBS logic
```

## Database Schema

The system uses SQLite with the following tables:

-   **users**: User accounts and authentication
-   **messages**: Private messages between users
-   **bulletins**: System bulletins and announcements
-   **sessions**: Active user sessions

## Access Levels

-   `0`: Regular user (default)
-   `10-254`: Various privilege levels
-   `255`: Sysop (full access)

## Development

### Building

```bash
go build -o bbs main.go
```

### Dependencies

-   `golang.org/x/crypto/ssh`: SSH server implementation
-   `golang.org/x/term`: Terminal handling
-   `github.com/mattn/go-sqlite3`: SQLite driver
-   `gopkg.in/yaml.v2`: YAML configuration parsing

## Extending the BBS

### Adding New Menu Commands

1. Add the command to your menu configuration in `config.yaml`
2. Implement the command handler in `internal/server/ssh.go`
3. Add the case to the `executeCommand` function

### Adding New Database Tables

1. Add the table structure to `createTables()` in `internal/database/database.go`
2. Create corresponding Go structs
3. Implement CRUD operations as needed

## Security Notes

-   **Default passwords**: Change default passwords in production
-   **Password hashing**: Implement proper password hashing (currently plain text)
-   **Host keys**: Generate and securely store SSH host keys
-   **Access control**: Review and configure access levels appropriately

## License

This project is inspired by the original Searchlight BBS software by Frank LaRosa. This implementation is created for educational and nostalgic purposes.

## Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues for bugs and feature requests.
