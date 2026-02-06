package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/uncaughtx/portcheck/internal/detector"
	"github.com/uncaughtx/portcheck/internal/ui"
)

func runCheck(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	port, err := strconv.Atoi(args[0])
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid port: %s (must be 1-65535)", args[0])
	}

	ctx := context.Background()
	det := detector.New()

	info, err := det.FindByPort(ctx, port)
	if err != nil {
		return fmt.Errorf("failed to check port: %w", err)
	}

	if info == nil {
		if jsonOut {
			fmt.Println(`{"status":"available","port":` + args[0] + `}`)
		} else {
			fmt.Printf("✓ Port %d is available\n", port)
		}
		return nil
	}

	if jsonOut {
		return outputJSON(map[string]interface{}{
			"status":  "in_use",
			"port":    port,
			"process": info,
		})
	}

	// Interactive mode
	p := tea.NewProgram(ui.NewModel(port, info))
	_, err = p.Run()
	return err
}

func outputJSON(v interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
