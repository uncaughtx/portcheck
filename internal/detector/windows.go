//go:build windows

package detector

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/process"
	procinfo "github.com/uncaughtx/portcheck/internal/process"
)

type windowsDetector struct{}

func newPlatformDetector() Detector {
	return &windowsDetector{}
}

func (d *windowsDetector) FindByPort(ctx context.Context, port int) (*procinfo.Info, error) {
	// Use netstat to find the process using this port
	cmd := exec.CommandContext(ctx, "netstat", "-ano", "-p", "TCP")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("netstat failed: %w", err)
	}

	portStr := fmt.Sprintf(":%d ", port)
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	for scanner.Scan() {
		line := scanner.Text()
		// Look for LISTENING state on our port
		if strings.Contains(line, portStr) && strings.Contains(line, "LISTENING") {
			fields := strings.Fields(line)
			if len(fields) >= 5 {
				pid, err := strconv.ParseInt(fields[4], 10, 32)
				if err != nil {
					continue
				}
				return d.enrichProcessInfo(ctx, int32(pid))
			}
		}
	}

	return nil, nil // Port not in use
}

func (d *windowsDetector) FindByRange(ctx context.Context, start, end int) ([]*procinfo.PortInfo, error) {
	// Get all listening ports first
	all, err := d.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	// Filter by range
	var result []*procinfo.PortInfo
	for _, pi := range all {
		if pi.Port >= start && pi.Port <= end {
			result = append(result, pi)
		}
	}

	return result, nil
}

func (d *windowsDetector) ListAll(ctx context.Context) ([]*procinfo.PortInfo, error) {
	cmd := exec.CommandContext(ctx, "netstat", "-ano", "-p", "TCP")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("netstat failed: %w", err)
	}

	var result []*procinfo.PortInfo
	seenPorts := make(map[int]bool)
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, "LISTENING") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		// Parse local address (e.g., "0.0.0.0:3000" or "[::]:3000")
		localAddr := fields[1]
		lastColon := strings.LastIndex(localAddr, ":")
		if lastColon == -1 {
			continue
		}

		portStr := localAddr[lastColon+1:]
		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}

		// Skip if we've already seen this port
		if seenPorts[port] {
			continue
		}
		seenPorts[port] = true

		pid, err := strconv.ParseInt(fields[4], 10, 32)
		if err != nil {
			continue
		}

		info, err := d.enrichProcessInfo(ctx, int32(pid))
		if err != nil {
			continue
		}

		result = append(result, &procinfo.PortInfo{
			Port:     port,
			Protocol: "tcp",
			State:    "LISTEN",
			Process:  info,
		})
	}

	return result, nil
}

func (d *windowsDetector) enrichProcessInfo(ctx context.Context, pid int32) (*procinfo.Info, error) {
	proc, err := process.NewProcessWithContext(ctx, pid)
	if err != nil {
		return nil, err
	}

	name, _ := proc.NameWithContext(ctx)
	cmdline, _ := proc.CmdlineWithContext(ctx)
	cwd, _ := proc.CwdWithContext(ctx)
	username, _ := proc.UsernameWithContext(ctx)
	createTime, _ := proc.CreateTimeWithContext(ctx)
	ppid, _ := proc.PpidWithContext(ctx)
	memInfo, _ := proc.MemoryInfoWithContext(ctx)
	cpuPercent, _ := proc.CPUPercentWithContext(ctx)
	children, _ := proc.ChildrenWithContext(ctx)

	info := &procinfo.Info{
		PID:        pid,
		Name:       name,
		Cmdline:    cmdline,
		Cwd:        cwd,
		User:       username,
		StartTime:  time.UnixMilli(createTime),
		ParentPID:  ppid,
		ChildCount: len(children),
		CPUPercent: cpuPercent,
	}

	if memInfo != nil {
		info.MemoryMB = float64(memInfo.RSS) / 1024 / 1024
	}

	// Try to detect project name
	if cwd != "" {
		info.ProjectName = detectProjectName(cwd)
	}

	// Get parent name
	if ppid > 0 {
		parent, err := process.NewProcessWithContext(ctx, ppid)
		if err == nil {
			info.ParentName, _ = parent.NameWithContext(ctx)
		}
	}

	return info, nil
}

func detectProjectName(cwd string) string {
	// Check package.json
	if data, err := os.ReadFile(filepath.Join(cwd, "package.json")); err == nil {
		var pkg struct{ Name string }
		if json.Unmarshal(data, &pkg) == nil && pkg.Name != "" {
			return pkg.Name
		}
	}

	// Check go.mod
	if data, err := os.ReadFile(filepath.Join(cwd, "go.mod")); err == nil {
		lines := strings.Split(string(data), "\n")
		if len(lines) > 0 && strings.HasPrefix(lines[0], "module ") {
			return strings.TrimPrefix(lines[0], "module ")
		}
	}

	// Check Cargo.toml (Rust)
	if _, err := os.Stat(filepath.Join(cwd, "Cargo.toml")); err == nil {
		return filepath.Base(cwd) + " (Rust)"
	}

	// Check pom.xml (Java/Maven)
	if _, err := os.Stat(filepath.Join(cwd, "pom.xml")); err == nil {
		return filepath.Base(cwd) + " (Maven)"
	}

	// Fall back to directory name
	return filepath.Base(cwd)
}
