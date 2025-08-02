# Status Bar Component

A terminal status bar component for the Coastline BBS system that displays user information, system name, and call duration at the bottom of the terminal.

## Features

-   **Fixed positioning**: Always appears at the bottom of the terminal
-   **Classic BBS styling**: Blue background with colored text sections
-   **Real-time updates**: Automatic duration timer updates
-   **Thread-safe**: Safe for concurrent access across goroutines
-   **Configurable**: Uses system configuration for display settings
-   **Responsive**: Handles terminal resizing and long text truncation

## Display Layout

```
 username        Coastline BBS        HH:MM:SS
[  white  ]   [  bright green  ]   [bright yellow]
```

-   **Left**: Username in white text
-   **Center**: System name from config in bright green
-   **Right**: Call duration timer in bright yellow
-   **Background**: Blue background across entire width

## Integration

The status bar is automatically integrated into all BBS sessions:

-   **SSH Sessions**: Status bar appears immediately after SSH authentication
-   **Local Sessions**: Status bar appears after login authentication
-   **Automatic cleanup**: Status bar is cleared when sessions end
-   **Real-time updates**: Duration timer updates every second

## Configuration

The status bar uses the following config values:

```yaml
bbs:
    system_name: "Coastline BBS" # Displayed in center (bright green)
    max_line_length: 79 # Maximum width of status bar
```

## ANSI Escape Codes

The status bar uses these ANSI escape codes:

-   `\033[44m` - Blue background
-   `\033[37m` - White text (username)
-   `\033[92m` - Bright green text (system name)
-   `\033[93m` - Bright yellow text (duration)
-   `\033[0m` - Reset formatting
-   `\033[2K` - Clear line
-   `\033[{row};1H` - Position cursor at bottom row
-   `\033[s` / `\033[u` - Save/restore cursor position

## Duration Formatting

Call duration is formatted as `HH:MM:SS`:

-   `00:00:30` - 30 seconds
-   `00:01:30` - 1 minute 30 seconds
-   `01:01:01` - 1 hour 1 minute 1 second

## Example Output

```
 sysop                    Coastline BBS                    00:05:42
```

This creates a professional, classic BBS look that maintains the authentic terminal experience while providing useful session information.
