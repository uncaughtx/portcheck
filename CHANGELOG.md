# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Initial release of portcheck
- Interactive TUI for port investigation
- Cross-platform support (Linux, macOS, Windows)
- `portcheck <port>` - Check a specific port
- `portcheck scan <start>-<end>` - Scan port range
- `portcheck list` - List all listening ports
- `portcheck version` - Show version info
- `portcheck completion` - Generate shell completions
- `--json` flag for scriptable output
- `--no-color` flag for plain output
- Process information: PID, name, command, working directory, uptime
- Parent process detection
- Project name detection (package.json, go.mod)
- Kill process with confirmation

### Security

- Safe process termination with explicit confirmation
- No blind process killing

## [0.1.0] - 2026-02-05

### Added

- Initial public release
