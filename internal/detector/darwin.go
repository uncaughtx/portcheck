//go:build darwin

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

type darwinDetector struct{}

func newPlatformDetector() Detector {
	return &darwinDetector{}
}

func (d *darwinDetector) FindByPort(ctx context.Context, port int) (*procinfo.Info, error) {
	// Use lsof to find the process using this port
	cmd := exec.CommandContext(ctx, "lsof", "-i", fmt.Sprintf(":%d", port), "-sTCP:LISTEN", "-n", "-P")
	output, err := cmd.Output()
	if err != nil {
		// lsof returns exit code 1 if no process found
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil // Port not in use
		}
		return nil, fmt.Errorf("lsof failed: %w", err)
	}

	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	scanner.Scan() // Skip header

	if scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 2 {
			pid, err := strconv.ParseInt(fields[1], 10, 32)
			if err != nil {
				return nil, err
			}
			return d.enrichProcessInfo(ctx, int32(pid))
		}
	}

	return nil, nil
}

func (d *darwinDetector) FindByRange(ctx context.Context, start, end int) ([]*procinfo.PortInfo, error) {
	all, err := d.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	var result []*procinfo.PortInfo
	for _, pi := range all {
		if pi.Port >= start && pi.Port <= end {
			result = append(result, pi)
		}
	}

	return result, nil
}

func (d *darwinDetector) ListAll(ctx context.Context) ([]*procinfo.PortInfo, error) {
	cmd := exec.CommandContext(ctx, "lsof", "-i", "-sTCP:LISTEN", "-n", "-P")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil, nil // No listening ports
		}
		return nil, fmt.Errorf("lsof failed: %w", err)
	}

	var result []*procinfo.PortInfo
	seenPorts := make(map[int]bool)
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	scanner.Scan() // Skip header

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 9 {
			continue
		}

		pid, err := strconv.ParseInt(fields[1], 10, 32)
		if err != nil {
			continue
		}

		// Parse port from NAME field (e.g., "*:3000" or "127.0.0.1:3000")
		nameField := fields[8]
		lastColon := strings.LastIndex(nameField, ":")
		if lastColon == -1 {
			continue
		}

		portStr := nameField[lastColon+1:]
		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}

		if seenPorts[port] {
			continue
		}
		seenPorts[port] = true

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

func (d *darwinDetector) enrichProcessInfo(ctx context.Context, pid int32) (*procinfo.Info, error) {
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

	// Fall back to directory name
	return filepath.Base(cwd)
}
