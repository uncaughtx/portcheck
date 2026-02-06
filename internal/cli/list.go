package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/spf13/cobra"
	"github.com/uncaughtx/portcheck/internal/detector"
	"github.com/uncaughtx/portcheck/internal/process"
	"github.com/uncaughtx/portcheck/internal/ui"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all listening ports",
	Long: `List all listening ports on the system with their associated processes.

Examples:
  portcheck list         List all listening ports
  portcheck list --json  Output as JSON`,
	RunE: runList,
}

func runList(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	det := detector.New()

	ports, err := det.ListAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to list ports: %w", err)
	}

	if jsonOut {
		return outputListJSON(ports)
	}

	if len(ports) == 0 {
		fmt.Println("No listening ports found")
		return nil
	}

	// Sort by port number
	sort.Slice(ports, func(i, j int) bool {
		return ports[i].Port < ports[j].Port
	})

	fmt.Printf("Found %d listening port(s):\n\n", len(ports))

	// Group by process for cleaner output
	byProcess := make(map[int32][]*process.PortInfo)
	for _, pi := range ports {
		byProcess[pi.Process.PID] = append(byProcess[pi.Process.PID], pi)
	}

	// Print grouped
	for _, pi := range ports {
		// Check if we already printed this process
		if byProcess[pi.Process.PID] == nil {
			continue
		}

		processPorts := byProcess[pi.Process.PID]
		byProcess[pi.Process.PID] = nil // Mark as printed

		fmt.Printf("  %s %s (PID %d)\n",
			ui.LabelStyle.Render("Process:"),
			ui.HighlightStyle.Render(pi.Process.Name),
			pi.Process.PID)

		if pi.Process.User != "" {
			fmt.Printf("  %s %s\n", ui.LabelStyle.Render("User:"), pi.Process.User)
		}

		if pi.Process.ProjectName != "" {
			fmt.Printf("  %s %s\n", ui.LabelStyle.Render("Project:"), pi.Process.ProjectName)
		}

		fmt.Printf("  %s %s\n", ui.LabelStyle.Render("Running:"), pi.Process.Uptime())

		fmt.Printf("  %s ", ui.LabelStyle.Render("Ports:"))
		for i, p := range processPorts {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(ui.HighlightStyle.Render(fmt.Sprintf("%d", p.Port)))
		}
		fmt.Println()
		fmt.Println()
	}

	return nil
}

func outputListJSON(ports interface{}) error {
	result := map[string]interface{}{
		"ports": ports,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
