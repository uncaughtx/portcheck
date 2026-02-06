package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/uncaughtx/portcheck/internal/detector"
	"github.com/uncaughtx/portcheck/internal/ui"
)

var scanCmd = &cobra.Command{
	Use:   "scan <start>-<end>",
	Short: "Scan a range of ports",
	Long: `Scan a range of ports to find which ones are in use.

Examples:
  portcheck scan 3000-3010      Scan ports 3000 to 3010
  portcheck scan 8000-9000      Scan ports 8000 to 9000
  portcheck scan 80-443 --json  Output as JSON`,
	Args: cobra.ExactArgs(1),
	RunE: runScan,
}

func runScan(cmd *cobra.Command, args []string) error {
	parts := strings.Split(args[0], "-")
	if len(parts) != 2 {
		return fmt.Errorf("invalid range format: use start-end (e.g., 3000-3010)")
	}

	start, err := strconv.Atoi(parts[0])
	if err != nil || start < 1 || start > 65535 {
		return fmt.Errorf("invalid start port: %s", parts[0])
	}

	end, err := strconv.Atoi(parts[1])
	if err != nil || end < 1 || end > 65535 {
		return fmt.Errorf("invalid end port: %s", parts[1])
	}

	if start > end {
		return fmt.Errorf("start port must be less than or equal to end port")
	}

	ctx := context.Background()
	det := detector.New()

	ports, err := det.FindByRange(ctx, start, end)
	if err != nil {
		return fmt.Errorf("failed to scan ports: %w", err)
	}

	if jsonOut {
		return outputScanJSON(start, end, ports)
	}

	if len(ports) == 0 {
		fmt.Printf("No listening ports found in range %d-%d\n", start, end)
		return nil
	}

	fmt.Printf("Found %d listening port(s) in range %d-%d:\n\n", len(ports), start, end)
	for _, pi := range ports {
		fmt.Printf("  %s %d\n", ui.HighlightStyle.Render("Port"), pi.Port)
		fmt.Printf("    %s %s (PID %d)\n", ui.LabelStyle.Render("Process:"), pi.Process.Name, pi.Process.PID)
		if pi.Process.ProjectName != "" {
			fmt.Printf("    %s %s\n", ui.LabelStyle.Render("Project:"), pi.Process.ProjectName)
		}
		fmt.Printf("    %s %s\n", ui.LabelStyle.Render("Running:"), pi.Process.Uptime())
		fmt.Println()
	}

	return nil
}

func outputScanJSON(start, end int, ports interface{}) error {
	result := map[string]interface{}{
		"range": map[string]int{
			"start": start,
			"end":   end,
		},
		"ports": ports,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
