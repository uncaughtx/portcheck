package process

import (
	"fmt"
	"time"
)

// Info contains detailed process information
type Info struct {
	PID         int32     `json:"pid"`
	Name        string    `json:"name"`
	Cmdline     string    `json:"cmdline"`
	Cwd         string    `json:"cwd"`
	User        string    `json:"user"`
	StartTime   time.Time `json:"start_time"`
	ParentPID   int32     `json:"parent_pid"`
	ParentName  string    `json:"parent_name,omitempty"`
	ChildCount  int       `json:"child_count"`
	MemoryMB    float64   `json:"memory_mb"`
	CPUPercent  float64   `json:"cpu_percent"`
	ProjectName string    `json:"project_name,omitempty"` // from package.json, go.mod, etc.
}

// PortInfo associates a port with its process
type PortInfo struct {
	Port     int    `json:"port"`
	Protocol string `json:"protocol"` // tcp, udp
	State    string `json:"state"`    // LISTEN, ESTABLISHED, etc.
	Process  *Info  `json:"process"`
}

// Uptime returns human-readable duration since process started
func (i *Info) Uptime() string {
	if i.StartTime.IsZero() {
		return "unknown"
	}
	
	d := time.Since(i.StartTime)
	switch {
	case d < time.Minute:
		return "just started"
	case d < time.Hour:
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	case d < 24*time.Hour:
		hours := int(d.Hours())
		if hours == 1 {
			return "1 hour"
		}
		return fmt.Sprintf("%d hours", hours)
	default:
		days := int(d.Hours() / 24)
		if days == 1 {
			return "1 day"
		}
		return fmt.Sprintf("%d days", days)
	}
}

// Summary returns a one-line summary of the process
func (i *Info) Summary() string {
	if i.ProjectName != "" {
		return fmt.Sprintf("%s (%s)", i.Name, i.ProjectName)
	}
	return i.Name
}
