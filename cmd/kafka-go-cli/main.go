package main

import (
	"os"

	"kafka_go_cli/internal/cli"
)

func main() {
	os.Exit(cli.Execute())
}
