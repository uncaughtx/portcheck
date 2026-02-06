//go:build linux

package detector

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/process"
	procinfo "github.com/uncaughtx/portcheck/internal/process"
)

type linuxDetector struct{}

func newPlatformDetector() Detector {
	return &linuxDetector{}
}

func (d *linuxDetector) FindByPort(ctx context.Context, port int) (*procinfo.Info, error) {
	// Try using ss command first (more reliable, shows PIDs directly)
	pid, err := d.findPIDByPortSS(ctx, port)
	if err == nil && pid > 0 {
		return d.enrichProcessInfo(ctx, pid)
	}

	// Fallback to /proc/net approach
	inode, err := d.findInodeByPort(port)
	if err != nil {
		return nil, err
	}

	if inode == 0 {
		return nil, nil // Port not in use
	}

	// Find process with this socket inode
	pid, err = d.findPIDByInode(inode)
	if err != nil {
		return nil, err
	}

	return d.enrichProcessInfo(ctx, pid)
}

// findPIDByPortSS uses the ss command to find the PID listening on a port
func (d *linuxDetector) findPIDByPortSS(ctx context.Context, port int) (int32, error) {
	// ss -tlnp shows listening TCP sockets with process info
	cmd := exec.CommandContext(ctx, "ss", "-tlnp", fmt.Sprintf("sport = :%d", port))
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	// Parse output to find PID
	// Format: LISTEN 0 128 *:3000 *:* users:(("node",pid=1234,fd=3))
	pidRegex := regexp.MustCompile(`pid=(\d+)`)
	matches := pidRegex.FindStringSubmatch(string(output))
	if len(matches) >= 2 {
		pid, err := strconv.ParseInt(matches[1], 10, 32)
		if err == nil {
			return int32(pid), nil
		}
	}

	return 0, fmt.Errorf("no process found on port %d", port)
}

func (d *linuxDetector) FindByRange(ctx context.Context, start, end int) ([]*procinfo.PortInfo, error) {
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

func (d *linuxDetector) ListAll(ctx context.Context) ([]*procinfo.PortInfo, error) {
	// Try ss command first (more reliable)
	result, err := d.listAllSS(ctx)
	if err == nil && len(result) > 0 {
		return result, nil
	}

	// Fallback to /proc/net parsing
	return d.listAllProc(ctx)
}

// listAllSS uses ss command to list all listening ports with PIDs
func (d *linuxDetector) listAllSS(ctx context.Context) ([]*procinfo.PortInfo, error) {
	// ss -tlnp shows all listening TCP sockets with process info
	cmd := exec.CommandContext(ctx, "ss", "-tlnp")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var result []*procinfo.PortInfo
	seenPorts := make(map[int]bool)

	// Parse output line by line
	// Format: State Recv-Q Send-Q Local Address:Port Peer Address:Port Process
	// LISTEN 0 128 *:22 *:* users:(("sshd",pid=1234,fd=3))
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	scanner.Scan() // Skip header

	portRegex := regexp.MustCompile(`:(\d+)\s`)
	pidRegex := regexp.MustCompile(`pid=(\d+)`)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "LISTEN") {
			continue
		}

		// Extract port
		portMatches := portRegex.FindStringSubmatch(line)
		if len(portMatches) < 2 {
			continue
		}
		port, err := strconv.Atoi(portMatches[1])
		if err != nil {
			continue
		}

		if seenPorts[port] {
			continue
		}
		seenPorts[port] = true

		// Extract PID
		pidMatches := pidRegex.FindStringSubmatch(line)
		if len(pidMatches) < 2 {
			continue
		}
		pid, err := strconv.ParseInt(pidMatches[1], 10, 32)
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

// listAllProc uses /proc/net/tcp parsing (fallback)
func (d *linuxDetector) listAllProc(ctx context.Context) ([]*procinfo.PortInfo, error) {
	var result []*procinfo.PortInfo
	seenPorts := make(map[int]bool)

	for _, proto := range []string{"tcp", "tcp6"} {
		file, err := os.Open(filepath.Join("/proc/net", proto))
		if err != nil {
			continue
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanner.Scan() // Skip header

		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) < 10 {
				continue
			}

			// State field[3] == "0A" means LISTEN
			if fields[3] != "0A" {
				continue
			}

			// Parse port from local_address (fields[1])
			localParts := strings.Split(fields[1], ":")
			if len(localParts) != 2 {
				continue
			}

			portHex := localParts[1]
			port64, err := strconv.ParseInt(portHex, 16, 32)
			if err != nil {
				continue
			}
			port := int(port64)

			if seenPorts[port] {
				continue
			}
			seenPorts[port] = true

			inode, _ := strconv.ParseUint(fields[9], 10, 64) //nolint:errcheck // We check for zero value below
			if inode == 0 {
				continue
			}

			pid, err := d.findPIDByInode(inode)
			if err != nil {
				continue
			}

			info, err := d.enrichProcessInfo(ctx, pid)
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
	}

	return result, nil
}

func (d *linuxDetector) findInodeByPort(port int) (uint64, error) {
	portHex := fmt.Sprintf("%04X", port)

	for _, proto := range []string{"tcp", "tcp6"} {
		file, err := os.Open(filepath.Join("/proc/net", proto))
		if err != nil {
			continue
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanner.Scan() // Skip header

		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) < 10 {
				continue
			}

			// local_address is fields[1], format: IP:PORT (hex)
			localParts := strings.Split(fields[1], ":")
			if len(localParts) == 2 && localParts[1] == portHex {
				// State field[3] == "0A" means LISTEN
				if fields[3] == "0A" {
					inode, _ := strconv.ParseUint(fields[9], 10, 64) //nolint:errcheck // Error returns 0 which is handled
					return inode, nil
				}
			}
		}
	}

	return 0, nil
}

func (d *linuxDetector) findPIDByInode(inode uint64) (int32, error) {
	inodeStr := fmt.Sprintf("socket:[%d]", inode)

	procDir, err := os.Open("/proc")
	if err != nil {
		return 0, err
	}
	defer procDir.Close()

	entries, err := procDir.Readdirnames(-1)
	if err != nil {
		return 0, err
	}

	for _, entry := range entries {
		pid, err := strconv.ParseInt(entry, 10, 32)
		if err != nil {
			continue // Not a PID directory
		}

		fdDir := filepath.Join("/proc", entry, "fd")
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			continue
		}

		for _, fd := range fds {
			link, err := os.Readlink(filepath.Join(fdDir, fd.Name()))
			if err == nil && link == inodeStr {
				return int32(pid), nil
			}
		}
	}

	return 0, fmt.Errorf("process not found for inode %d", inode)
}

func (d *linuxDetector) enrichProcessInfo(ctx context.Context, pid int32) (*procinfo.Info, error) {
	proc, err := process.NewProcessWithContext(ctx, pid)
	if err != nil {
		return nil, err
	}

	// Process info fields may fail due to permissions or process state - gracefully ignore
	name, _ := proc.NameWithContext(ctx)             //nolint:errcheck // Optional field
	cmdline, _ := proc.CmdlineWithContext(ctx)       //nolint:errcheck // Optional field
	cwd, _ := proc.CwdWithContext(ctx)               //nolint:errcheck // Optional field
	username, _ := proc.UsernameWithContext(ctx)     //nolint:errcheck // Optional field
	createTime, _ := proc.CreateTimeWithContext(ctx) //nolint:errcheck // Optional field
	ppid, _ := proc.PpidWithContext(ctx)             //nolint:errcheck // Optional field
	memInfo, _ := proc.MemoryInfoWithContext(ctx)    //nolint:errcheck // Optional field
	cpuPercent, _ := proc.CPUPercentWithContext(ctx) //nolint:errcheck // Optional field
	children, _ := proc.ChildrenWithContext(ctx)     //nolint:errcheck // Optional field

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
			info.ParentName, _ = parent.NameWithContext(ctx) //nolint:errcheck // Optional field
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
