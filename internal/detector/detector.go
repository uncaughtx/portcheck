package detector

import (
	"context"

	"github.com/uncaughtx/portcheck/internal/process"
)

// Detector finds processes listening on network ports
type Detector interface {
	// FindByPort returns the process listening on the given port
	FindByPort(ctx context.Context, port int) (*process.Info, error)

	// FindByRange returns all processes listening on ports in range
	FindByRange(ctx context.Context, start, end int) ([]*process.PortInfo, error)

	// ListAll returns all listening ports and their processes
	ListAll(ctx context.Context) ([]*process.PortInfo, error)
}

// New returns the appropriate detector for the current platform
func New() Detector {
	// Build tags ensure correct implementation is compiled
	return newPlatformDetector()
}
