package main

import (
	"os"

	"kafka_go_cli/internal/cli"

	// blank imports to register processor constructors with the factory
	// via their init() functions. This allows dynamic processor selection
	// based on configuration without hardcoding dependencies.
	_ "kafka_go_cli/internal/processors/kafka"
	_ "kafka_go_cli/internal/processors/noop"
)

func main() {
	os.Exit(cli.Execute())
}
