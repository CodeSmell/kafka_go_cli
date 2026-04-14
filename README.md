# Kafka Publish Util (in Go)

This is a Go learning project that will become a CLI utility for polling a directory and publishing each file to a Kafka topic.

The primary goal is to experiment with the Go language on a practical but small application.

Related Java project for comparison:
- [KafkaPubCLI](https://github.com/CodeSmell/KafkaPubCLI)


## Project Layout

```
kafka_go_cli/
  ├── go.mod
  ├── README.md
  ├── cmd/
  │   └── kafka-go-cli/
  │       └── main.go
  ├── internal/
  │   ├── cli/
  │   │   ├── root.go
  │   │   └── app.go
  │   ├── file/
  │   │    ├── poller.go              # DirectoryPoller 
  |   │    └── poller_test.go
  │   ├── processor/
  │   │    ├── processor.go           # the interface/type (Processor)
  │   │    └── processor_factory.go
  │   └── processors/
  │        ├── processor_noop.go
  │        ├── kafka/
  │        │     └── processor_kafka.go     # KafkaFileProcessor implementation
  │        └── pulsar/
  |              └── processor_pulsar.go    # PulsarFileProcessor implementation
  ├── configs/
  │   └── sample config file
  └── docs/
      └── additional doc files
```

Common Go layout:
- `cmd/` contains executable entrypoints
- `internal/` contains private application packages
- `configs/` stores config files
- `docs/` stores documents and learning notes 

## Basic commands for developer workflow
Unlike Java, building Go produces a native executable artifact. During development, using `make` keeps the workflow fast and repeatable. The underlying Go commands can be found in the `Makefile`.

### Basic commands during development
Run all quality checks as you develop

```bash
make check
```

Cleanup and verify everything before creating a PR

```bash
make ci
```
### Running the application

Run the app (with Make). This will create an executable file and run it

```bash
make run ARGS="--log-level debug --no-delete-files --run-once --message-location ~/dev/messages"
```

Run the app (with Go). This will run the app without creating an executable file.

```bash
go run ./cmd/kafka-go-cli --log-level debug --no-delete-files --message-location ~/dev/messages
```


## Current CLI and config behavior
- Cobra command tree with a single command (no subcommands)
- Viper-based config resolution using this precedence:
  - defaults
  - config file (optional)
  - environment variables
  - CLI flags (highest priority)

### CLI parameters
Running the application

```bash
go run ./cmd/kafka-go-cli [flags]
```

| Parameter | Description |
|---|---|
| `--config` | Optional path to a config file (for example YAML). |
| `--check` | Validate resolved config and exit. |
| `--log-level` | Sets logging level. Use `debug`, `info`, `warn`, or `error`. |
| `--message-location` | Directory to scan for message files (can come from flag, config file, or env). |
| `--run-once` | Runs a single scan cycle then exits. Default is `false`. |
| `--no-delete-files` | Keeps files after processing. Default is `true`. |
| `--delay` | Number of ms to wait between polling cycles |
| `--max-cycles` | Max number times to poll. If < 0 then use run-once or keep polling indefinitely. Default is -1 |
| `--processor` | specify which processor will handle the files (noop, kafka, pulsar etc). Default is noop |

Logs are structured text logs using Go's standard `log/slog` package.

## Sample commands 

Show command tree:

```bash
go run ./cmd/kafka-go-cli --help
```

Run the scan and start publishing

```bash
go run ./cmd/kafka-go-cli --message-location ~/dev/messages
```

To run the application while only checking the configuration add the `--check` flag

```bash
go run ./cmd/kafka-go-cli \
--config ./configs/config.sample.yaml \
--message-location ~/dev/messages \
--check
```
