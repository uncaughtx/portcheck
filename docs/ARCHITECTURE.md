# portcheck — Architecture

## Directory Structure

```
portcheck/
├── cmd/portcheck/          # Entry point
│   └── main.go
├── internal/
│   ├── cli/                # CLI commands (cobra)
│   │   ├── root.go         # Root command, flags
│   │   ├── check.go        # portcheck <port>
│   │   ├── scan.go         # portcheck scan
│   │   ├── list.go         # portcheck list
│   │   ├── version.go      # portcheck version
│   │   └── completion.go   # Shell completions
│   ├── detector/           # Platform-specific port detection
│   │   ├── detector.go     # Interface
│   │   ├── linux.go        # Linux: ss + /proc/net
│   │   ├── darwin.go       # macOS: lsof
│   │   └── windows.go      # Windows: netstat
│   ├── process/            # Process information
│   │   └── info.go         # Types and helpers
│   └── ui/                 # Interactive TUI
│       ├── model.go        # Bubbletea model
│       └── styles.go       # Lipgloss styles
└── pkg/                    # (Future) Public API
```

## Key Design Decisions

### 1. Platform Detection via Build Tags

Each platform has its own detector implementation selected at compile time:

- `linux.go` — Uses `ss` command + `/proc/net/tcp` parsing
- `darwin.go` — Uses `lsof` command
- `windows.go` — Uses `netstat -ano`

### 2. Process Enrichment

After finding the PID, we use `gopsutil` to get rich process information:

- Command line
- Working directory
- Memory/CPU usage
- Parent process
- Child count

### 3. Project Detection

Attempts to detect the project name from:

- `package.json` (Node.js)
- `go.mod` (Go)
- `Cargo.toml` (Rust)
- Directory name (fallback)

### 4. Interactive vs Scripting

- Default: Interactive TUI with bubbletea
- `--json`: Machine-readable output for scripting/CI
