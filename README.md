# portcheck

<div align="center">

**Understand what's using your port, then decide what to do about it.**

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/uncaughtx/portcheck?include_prereleases)](https://github.com/uncaughtx/portcheck/releases)

A cross-platform CLI tool that transforms the cryptic output of `lsof`/`netstat` into actionable, human-readable information with interactive decision-making.

</div>

---

## The Problem

I've hit "Port already in use" **5+ times daily**. Current solutions either:

- Give cryptic output (`lsof -i :3000`)
- Blind-kill processes (`npx kill-port 3000`)

Nothing helps you **understand** what's happening before taking action.

## The Solution

`portcheck` provides:

- **Human-readable process information** — name, PID, command, working directory, uptime
- **Interactive decision-making** — investigate before killing
- **Cross-platform** — single binary, works on Linux, macOS, Windows
- **Scriptable output** — JSON mode for automation

---

## Installation

### Using Go

```bash
go install github.com/uncaughtx/portcheck/cmd/portcheck@latest
```

### Binary Releases

Download from [GitHub Releases](https://github.com/uncaughtx/portcheck/releases):

```bash
# Linux (amd64)
curl -L https://github.com/uncaughtx/portcheck/releases/latest/download/portcheck-linux-amd64 -o portcheck
chmod +x portcheck
sudo mv portcheck /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/uncaughtx/portcheck/releases/latest/download/portcheck-darwin-arm64 -o portcheck
chmod +x portcheck
sudo mv portcheck /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri https://github.com/uncaughtx/portcheck/releases/latest/download/portcheck-windows-amd64.exe -OutFile portcheck.exe
```

### Build from Source

```bash
git clone https://github.com/uncaughtx/portcheck.git
cd portcheck
make build
```

---

## Usage

### Check a Port (Interactive)

```bash
portcheck 3000
```

![Interactive TUI](docs/assets/demo.gif)

**Keybindings:**

- `i` — Show detailed info
- `k` — Kill process (with confirmation)
- `q` — Quit

### Check a Port (JSON)

```bash
portcheck 3000 --json
```

```json
{
  "status": "in_use",
  "port": 3000,
  "process": {
    "pid": 12345,
    "name": "node",
    "cmdline": "node server.js",
    "cwd": "/home/user/myproject",
    "user": "user",
    "uptime": "2 hours"
  }
}
```

### Scan Port Range

```bash
portcheck scan 3000-3010
```

### List All Listening Ports

```bash
portcheck list
```

### Shell Completions

```bash
# Bash
portcheck completion bash > /etc/bash_completion.d/portcheck

# Zsh
portcheck completion zsh > "${fpath[1]}/_portcheck"

# Fish
portcheck completion fish > ~/.config/fish/completions/portcheck.fish

# PowerShell
portcheck completion powershell > portcheck.ps1
```

---

## Examples

### CI/CD Integration

```yaml
# .github/workflows/test.yml
- name: Ensure port 3000 is available
  run: |
    if portcheck 3000 --json | jq -e '.status == "in_use"'; then
      echo "Port 3000 is in use!"
      portcheck 3000 --json
      exit 1
    fi
```

### Docker Development

```bash
# Check if something is already using your dev ports
portcheck scan 3000-3010 --json | jq '.ports[] | "\(.port): \(.process.name)"'
```

### Scripts

```bash
#!/bin/bash
# wait-for-port.sh - Wait for a port to become available

PORT=${1:-3000}
TIMEOUT=${2:-30}

for i in $(seq 1 $TIMEOUT); do
  if portcheck $PORT --json | jq -e '.status == "available"' > /dev/null; then
    echo "Port $PORT is available"
    exit 0
  fi
  sleep 1
done

echo "Timeout waiting for port $PORT"
exit 1
```

---

## Comparison

| Tool | Info | Kill | Interactive | Cross-Platform | JSON |
|------|:----:|:----:|:-----------:|:--------------:|:----:|
| **portcheck** | Rich | Safe | Yes | Yes | Yes |
| `lsof` | Cryptic | No | No | macOS/Linux | No |
| `netstat` | Minimal | No | No | Partial | No |
| `kill-port` | None | Blind | No | Yes | No |
| `fuser` | Minimal | Yes | No | Linux only | No |

---

## Configuration

Create `~/.portcheck.yaml`:

```yaml
# Default output format (json, table, minimal)
output: table

# Disable colors
no-color: false

# Common development ports to highlight
highlight-ports:
  - 3000
  - 8080
  - 5432
  - 6379
```

---

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](docs/CONTRIBUTING.md) for guidelines.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## License

MIT License — see [LICENSE](LICENSE) for details.

---

<div align="center">

Made by [@uncaughtx](https://github.com/uncaughtx)

**Star this repo if you find it useful!**

</div>
