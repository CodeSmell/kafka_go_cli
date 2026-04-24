package cli

import (
	"kafka_go_cli/internal/processor"

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

	// Cobra requires flags to be pre-defined to be used in config loading and validation
	// for the most part we are not setting defaults here
	// that is done in config package so they are done in one place
	rootCmd.PersistentFlags().String("config", "", "Path to config file (optional)")
	rootCmd.PersistentFlags().Bool("check", false, "Validate resolved configuration and exit")
	rootCmd.PersistentFlags().String("log-level", "", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("message-location", "", "Directory path to scan for message files")
	rootCmd.PersistentFlags().Bool("run-once", false, "Run a single scan cycle then exits")
	rootCmd.PersistentFlags().Bool("no-delete-files", true, "Do not delete files after successful processing")
	rootCmd.PersistentFlags().Int("delay", 0, "Number of ms to wait between polling cycles (e.g., 5000 for 5 seconds)")
	rootCmd.PersistentFlags().Int("max-cycles", 0, "Max number of times to poll (default -1 polls indefinitely)")
	rootCmd.PersistentFlags().String("processor", "noop", "Processor type (kafka, pulsar, noop)")

	// Dynamically register processor-specific configuration flags
	// Each processor declares config parameters via RegisterConfigParams() in its init()
	registerProcessorFlagsWithCobra(rootCmd)

	return rootCmd
}

// registerProcessorFlagsWithCobra dynamically adds processor-specific flags to the command
// This allows adding new processors without modifying this file
func registerProcessorFlagsWithCobra(cmd *cobra.Command) {
	allProcessors := processor.GetAllRegisteredProcessors()
	for _, processorConfigParams := range allProcessors {
		for _, param := range processorConfigParams {
			cmd.PersistentFlags().String(param.Flag, "", param.Description)
		}
	}
}

// Execute runs the CLI and returns a conventional process exit code.
func Execute() int {
	cmd := NewRootCommand()
	if err := cmd.Execute(); err != nil {
		return 1
	}
	return 0
}
