# InfraSpec TUI - Claude Code Inspired Interface

A beautiful, interactive Terminal User Interface (TUI) for running InfraSpec tests, inspired by Claude Code's design patterns and user experience.

## Features

### 🎯 Core Functionality
- **Real-time test execution** with live streaming output
- **Interactive file selection** with path validation
- **Progress indicators** and status updates
- **Keyboard-driven navigation** similar to Claude Code
- **Responsive design** that adapts to terminal size

### 🎨 UI Components
- **Header Bar**: Shows application title, current status, and execution timing
- **Input Section**: Feature file path entry with validation
- **Output Viewport**: Scrollable, timestamped test output with syntax highlighting
- **Footer Bar**: Context-sensitive keyboard shortcuts
- **Help Screen**: Comprehensive usage documentation

### ⌨️ Keyboard Shortcuts
| Key | Action | Context |
|-----|--------|---------|
| `Enter` | Run test | When idle |
| `Esc` | Cancel test | When running |
| `R` | Re-run last test | When complete/failed |
| `?` | Toggle help screen | Any time |
| `Q` / `Ctrl+C` | Quit application | Any time |

### 🎨 Visual Design
- **Color-coded output**: Success (green), errors (red), warnings (yellow), info (blue)
- **Timestamped logs**: Each output line includes precise timing
- **Status indicators**: Visual symbols (✓, ✗, ⚠, •) for different message types
- **Responsive layout**: Automatically adjusts to terminal dimensions
- **Smooth animations**: Spinner and progress indicators

## Usage

### Basic Usage
```bash
# Launch the interactive TUI
infraspec ui

# View demo without requiring interactive terminal
infraspec ui --demo
```

### Example Workflow
1. **Launch**: Run `infraspec ui` in your terminal
2. **Enter Path**: Type the path to your feature file (e.g., `features/aws/s3/s3_bucket.feature`)
3. **Execute**: Press `Enter` to start test execution
4. **Monitor**: Watch real-time output with timestamps and status indicators
5. **Re-run**: Press `R` to run the same test again, or enter a new path

### Sample Feature Files
```
features/aws/s3/s3_bucket.feature
features/aws/dynamodb/dynamodb_table.feature
features/terraform/hello_world.feature
examples/aws/rds/mysql/main.tf
```

## Technical Implementation

### Architecture
- **Built with Bubble Tea**: Leverages the Charm.sh ecosystem for TUI components
- **Streaming I/O**: Real-time command output capture and display
- **Concurrent Processing**: Non-blocking test execution with cancellation support
- **Memory Efficient**: Pointer receivers to handle large model structures

### Components
- `internal/tui/model.go`: Core TUI model and state management
- `internal/tui/app.go`: Application lifecycle and initialization
- `internal/tui/runner.go`: Test execution and output streaming
- `internal/tui/demo.go`: Demo mode for non-interactive environments
- `cmd/ui.go`: CLI command integration

### Dependencies
- `github.com/charmbracelet/bubbletea`: TUI framework
- `github.com/charmbracelet/bubbles`: Pre-built TUI components
- `github.com/charmbracelet/lipgloss`: Styling and layout

## Integration

The TUI integrates seamlessly with InfraSpec's existing infrastructure:
- Uses the same configuration system
- Leverages existing test runner logic  
- Supports all current CLI flags (`--verbose`)
- Maintains compatibility with existing workflows

## Development

### Building
```bash
go build ./cmd/infraspec
```

### Testing
```bash
# Test the TUI (requires interactive terminal)
./infraspec ui

# Test demo mode (works in any environment)
./infraspec ui --demo

# Test with existing feature files
./infraspec ui
# Then enter: features/terraform/hello_world.feature
```

### Code Quality
```bash
# Format code
make fmt

# Run linting
make lint

# Run tests
make go-test-cover
```

## Inspiration

This TUI draws inspiration from Claude Code's excellent user experience patterns:
- **Keyboard-first navigation**: Efficient shortcuts for power users
- **Real-time feedback**: Immediate visual response to user actions
- **Contextual help**: Always-available assistance without leaving the interface
- **Clean aesthetics**: Minimal, focused design that prioritizes functionality
- **Responsive interaction**: Smooth, predictable behavior

## Future Enhancements

Potential improvements for future versions:
- **File browser**: Interactive feature file selection
- **Test history**: Recent test runs and results
- **Auto-completion**: Smart path completion for feature files
- **Syntax highlighting**: Enhanced output formatting
- **Parallel execution**: Multiple test runners
- **Configuration UI**: Interactive settings management

---

*This TUI represents InfraSpec's commitment to developer experience, making infrastructure testing more accessible and enjoyable.*