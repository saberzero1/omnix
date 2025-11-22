# TUI - Terminal User Interface

The `om tui` command launches an interactive terminal user interface for omnix. This provides a more visual and interactive way to explore your Nix system compared to individual CLI commands.

## Usage

```bash
om tui
```

## Features

The TUI provides four main views accessible via keyboard shortcuts:

### 1. Dashboard (Press `1`)

The main landing page providing an overview and quick navigation guide.

### 2. Health Checks (Press `2`)

Displays all Nix system health checks with color-coded status indicators:
- ✓ Green: Check passed
- ✗ Red: Check failed

Health checks include:
- Nix version compatibility
- Flakes enablement
- Trusted users configuration
- Binary cache configuration
- Max jobs setting
- Shell configuration
- Direnv installation (if applicable)
- Homebrew installation (on macOS)
- Rosetta 2 (on Apple Silicon)

### 3. System Info (Press `3`)

Shows detailed information about your Nix installation:
- **Nix Version**: Currently installed version
- **Configuration**: System settings including:
  - System architecture
  - Substituters (binary caches)
  - Max jobs and cores
  - Experimental features enabled
- **Environment**: Operating system details and user information

### 4. Flake Browser (Press `4`)

Browse and explore Nix flake outputs (under development).

## Keyboard Shortcuts

### Navigation
- `1` - Go to Dashboard
- `2` - Go to Health Checks
- `3` - Go to System Info
- `4` - Go to Flake Browser
- `↑/k` - Move up (in scrollable views)
- `↓/j` - Move down (in scrollable views)
- `←/h` - Move left
- `→/l` - Move right

### Actions
- `r` - Refresh current view
- `?` - Toggle help screen
- `q` or `Ctrl+C` - Quit the TUI

## Implementation

The TUI is built using the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework from Charm, which provides:
- Elegant terminal rendering with [Lipgloss](https://github.com/charmbracelet/lipgloss) for styling
- Reusable components from [Bubbles](https://github.com/charmbracelet/bubbles)
- An Elm-inspired architecture for reliable state management

## Examples

### Quick Health Check
Launch the TUI and press `2` to immediately view health checks:
```bash
om tui
# Press 2 when the TUI loads
```

### View System Configuration
Check your Nix configuration details:
```bash
om tui
# Press 3 to view system info
```

### Interactive Exploration
Navigate between views to get a complete picture of your Nix setup:
```bash
om tui
# Use 1-4 to switch between views
# Press r to refresh data
# Press ? for help
```

## Comparison with CLI Commands

The TUI provides a unified interface for features available via separate CLI commands:

| TUI View | Equivalent CLI Command |
|----------|----------------------|
| Health Checks | `om health` |
| System Info | `nix show-config` + `nix --version` |
| Flake Browser | `om show [FLAKE]` |

The TUI advantage is the ability to quickly switch between views and see all information in one session.

## Troubleshooting

### Terminal Compatibility
The TUI works best with modern terminal emulators that support:
- 256 colors
- Unicode characters
- ANSI escape codes

Recommended terminals:
- **macOS**: iTerm2, Terminal.app
- **Linux**: GNOME Terminal, Konsole, Alacritty
- **Windows**: Windows Terminal, WSL2 with any Linux terminal

### Rendering Issues
If you experience rendering issues:
1. Ensure your terminal supports Unicode (UTF-8)
2. Try resizing the terminal window
3. Check that your `TERM` environment variable is set correctly
4. Use a different terminal emulator if problems persist

### Performance
The TUI loads data asynchronously. If health checks or system info seem slow:
- Check that `nix` commands run normally outside the TUI
- Ensure you have network connectivity for cache checks
- Try refreshing the view with `r`

## Future Enhancements

Planned features for the TUI:
- Interactive flake browser with navigation tree
- Log viewer for nix build outputs
- Configuration editor
- Package search and installation
- Rollback management for NixOS/nix-darwin
