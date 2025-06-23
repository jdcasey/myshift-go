# myshift-go

A Go implementation of myshift - a command-line tool for managing PagerDuty on-call schedules.

This is a functional translation of the original Python [myshift](../myshift) tool, preserving all core functionality while leveraging Go's strengths for performance and deployment.

## Features

- **Next Shift**: View the next upcoming on-call shift for a user
- **Upcoming Shifts**: View all upcoming shifts for a user over a specified period
- **Plan Schedule**: Plan and visualize future schedule assignments  
- **Override Management**: Create schedule overrides for specific time periods
- **Interactive REPL**: Interactive shell for running multiple commands
- **Configuration Management**: YAML-based configuration with XDG compliance
- **Cross-platform**: Single binary deployment with no runtime dependencies

## Installation

### From Source

```bash
git clone https://github.com/jdcasey/myshift-go.git
cd myshift-go
go build -o myshift ./cmd/myshift
```

### Using Go Install

```bash
go install github.com/jdcasey/myshift-go/cmd/myshift@latest
```

## Configuration

Create a configuration file at one of these locations:

- **Linux**: `~/.config/myshift.yaml` or `$XDG_CONFIG_HOME/myshift.yaml`
- **macOS**: `~/Library/Application Support/myshift.yaml`

### Sample Configuration

```yaml
# PagerDuty API token (required)
pagerduty_token: "your-pagerduty-token"

# Default schedule ID (optional)
schedule_id: "your-default-schedule-id"

# Default user (optional)
my_user: "your-email@example.com"
```

Generate a sample configuration:

```bash
myshift config --print
```

## Usage

### Show Next Shift

```bash
# Show next shift for the configured user (uses my_user from config)
myshift next

# Show next shift for a specific user
myshift next --user user@example.com

# Look ahead 30 days instead of default 90
myshift next --user user@example.com --days 30
```

### Show Upcoming Shifts

```bash
# Show all upcoming shifts for the configured user (uses my_user from config)
myshift upcoming

# Show all upcoming shifts for a specific user (default: 28 days)
myshift upcoming --user user@example.com

# Look ahead 7 days
myshift upcoming --user user@example.com --days 7
```

### Plan Schedule

```bash
# Plan schedule for a date range
myshift plan --start 2025-01-01 --end 2025-01-31

# Plan schedule for next 14 days
myshift plan --days 14
```

### Create Override

```bash
# Create an override
myshift override --user substitute@example.com --target original@example.com --start "2025-01-15 09:00" --end "2025-01-15 17:00"
```

### Interactive REPL

```bash
# Start interactive shell
myshift repl

# Then use commands interactively:
# (myshift) next                    # Uses my_user from config
# (myshift) next user@example.com   # Check specific user
# (myshift) plan 7
# (myshift) upcoming                # Uses my_user from config  
# (myshift) upcoming user@example.com 14
# (myshift) quit
```

### Configuration Management

```bash
# Print sample configuration
myshift config --print

# Validate current configuration (basic info)
myshift config

# Detailed configuration validation with comprehensive report
myshift config --validate
```

## Architecture

The Go implementation follows idiomatic Go patterns while preserving the Python functionality:

```
myshift-go/
├── cmd/myshift/           # CLI entry point
├── internal/
│   ├── config/           # Configuration management
│   ├── pagerduty/        # PagerDuty API client
│   └── commands/         # Command implementations
├── pkg/myshift/          # Shared types and utilities
└── go.mod               # Go module definition
```

### Key Differences from Python Version

| Aspect | Python Version | Go Version |
|--------|----------------|------------|
| **Dependencies** | Multiple (pagerduty, pyyaml, dateutil) | Minimal (just gopkg.in/yaml.v3) |
| **CLI Framework** | argparse | flag (standard library) |
| **HTTP Client** | requests via pagerduty lib | net/http (standard library) |
| **Deployment** | Requires Python + deps | Single static binary |
| **Error Handling** | Exceptions with sys.exit() | Go error values with proper wrapping |
| **JSON/Time** | Manual parsing | Native Go time.Time with RFC3339 |

## Development

### Prerequisites

- Go 1.19 or later

### Building

```bash
go build -o myshift ./cmd/myshift
```

### Testing

```bash
go test ./...
```

### Cross-compilation

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o myshift-linux ./cmd/myshift

# macOS  
GOOS=darwin GOARCH=amd64 go build -o myshift-darwin ./cmd/myshift

# Windows
GOOS=windows GOARCH=amd64 go build -o myshift.exe ./cmd/myshift
```

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](../LICENSE) for details.

## Comparison with Python Version

This Go implementation provides:

✅ **Same functionality** - All core features preserved  
✅ **Better performance** - Native compilation vs interpreted Python  
✅ **Easier deployment** - Single binary vs Python + dependencies  
✅ **Better error handling** - Go's explicit error handling  
✅ **Memory safety** - Go's garbage collector and type safety  
✅ **Cross-platform** - Easy cross-compilation  

The API interactions and business logic remain identical to ensure consistent behavior across both implementations. 