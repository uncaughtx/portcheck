package cli

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	jsonOut bool
	noColor bool
)

var rootCmd = &cobra.Command{
	Use:   "portcheck [port]",
	Short: "Investigate what's using a port",
	Long: `portcheck helps you understand what process is using a port,
with context like working directory, uptime, and parent process.

Examples:
  portcheck 3000            Check port 3000 interactively
  portcheck 3000 --json     Output JSON for scripting
  portcheck scan 3000-3010  Scan a range of ports
  portcheck list            List all listening ports`,
	Args: cobra.MaximumNArgs(1),
	RunE: runCheck,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default $HOME/.portcheck.yaml)")
	rootCmd.PersistentFlags().BoolVar(&jsonOut, "json", false, "output as JSON")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")

	// Subcommands
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(completionCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".portcheck")
	}

	viper.AutomaticEnv()
	// Ignore errors if config file doesn't exist
	_ = viper.ReadInConfig()
}
