package cli

import (
	"github.com/spf13/cobra"
)

// NewRootCommand builds the CLI command tree.
// This project intentionally keeps the command structure flat (no subcommands)
func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "kafka-go-cli",
		Short: "Kafka publishing util",
		Long:  "Directory polling Kafka publishing utility",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          runE,
	}

	rootCmd.PersistentFlags().String("config", "", "Path to config file (optional)")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("message-location", "", "Directory path to scan for message files")
	rootCmd.PersistentFlags().Bool("run-once", false, "Run a single scan cycle then exits")
	rootCmd.PersistentFlags().Bool("no-delete-files", true, "Do not delete files after successful processing")
	rootCmd.PersistentFlags().Bool("check", false, "Validate resolved configuration and exit")

	return rootCmd
}

// Execute runs the CLI and returns a conventional process exit code.
func Execute() int {
	cmd := NewRootCommand()
	if err := cmd.Execute(); err != nil {
		return 1
	}
	return 0
}
