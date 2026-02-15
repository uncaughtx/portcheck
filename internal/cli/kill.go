package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/spf13/cobra"
	"github.com/uncaughtx/portcheck/internal/detector"
	"github.com/uncaughtx/portcheck/internal/ui"
)

var forceKill bool

var killCmd = &cobra.Command{
	Use:   "kill [port]",
	Short: "Kill the process listening on a port",
	Long: `Kill the process listening on a port.
Asking for confirmation by default, use --force to skip.`,
	Args: cobra.ExactArgs(1),
	RunE: runKill,
}

func init() {
	killCmd.Flags().BoolVarP(&forceKill, "force", "f", false, "force kill without confirmation")
}

func runKill(cmd *cobra.Command, args []string) error {
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
		fmt.Printf("Port %d is not in use\n", port)
		return nil
	}

	// Display process info using UI styles
	fmt.Println(ui.BoxStyle.Render(fmt.Sprintf("  Process on port %d  ", port)))
	fmt.Printf("  %s %s\n", ui.LabelStyle.Render("Name:"), ui.ValueStyle.Render(info.Name))
	fmt.Printf("  %s %s\n", ui.LabelStyle.Render("PID:"), ui.HighlightStyle.Render(fmt.Sprintf("%d", info.PID)))
	fmt.Printf("  %s %s\n", ui.LabelStyle.Render("Command:"), ui.ValueStyle.Render(ui.Truncate(info.Cmdline, 100)))

	if info.ProjectName != "" {
		fmt.Printf("  %s %s\n", ui.LabelStyle.Render("Project:"), ui.HighlightStyle.Render(info.ProjectName))
	}
	fmt.Println()

	if !forceKill {
		fmt.Printf("%s Are you sure you want to kill this process? (y/N) ", ui.WarningStyle.Render("?"))
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read input: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	proc, err := process.NewProcess(info.PID)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	if err := proc.Terminate(); err != nil {
		// Fallback to Kill if Terminate fails
		if err := proc.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
	}

	fmt.Println(ui.SuccessStyle.Render(fmt.Sprintf("Process %d killed successfully", info.PID)))
	return nil
}
