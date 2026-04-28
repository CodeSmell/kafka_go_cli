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

Processor-specific flags are registered dynamically based on the selected processor. For Kafka, see `--kafka-brokers`, `--kafka-topic` etc

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

## Testing

This project uses a 2-layer testing strategy:

**Layer 1: Fast Unit Tests** — These tests have no external dependencies and are intended to be fast and portable. They provide fast feedback during development. They often take advantage of mocks and doubles. They can be run with `make test` or `go test ./...`. 

**Layer 2: Integration Tests** — These tests rely on a real Kafka broker via testcontainers-go (requires Docker, Podman, or Colima). They can be run with `make test-integration` or `go test -tags=integration ./...`. These tests automatically provision a Kafka container, publish messages, and verify behavior end-to-end. Tests gracefully skip if no supported container runtime is available.

Note: There was an issue running integration tests with Podman on a Mac and using Docker Desktop can involve licensing issues. [Colima](https://ariel-ibarra.medium.com/goodbye-docker-rancher-desktop-switching-to-colima-for-lightweight-containers-f6a4215486d1) is a CLI only viable alternative on the Mac.

Installing Colima (for integration tests)

```bash
brew update
brew upgrade
brew cleanup
brew install colima docker
colima version
colima version 0.10.1
```

Running Colima

```bash
colima start
docker --context colima version
```

Run the Kafka integration test with the repo-local target:

```bash
make test-integration-colima
```

This target sets `DOCKER_HOST` to the Colima socket and disables Ryuk only for this test command. That avoids changing your global Podman setup or relying on `/var/run/docker.sock`.

### Integration test variants

- Override the broker container image used by the integration test (defaults to Redpanda):

```bash
KAFKA_IT_IMAGE="redpandadata/redpanda:v23.3.3" make test-integration-colima
```


### Writing tests

Unit tests use the constructor seam pattern, setting `newKafkaProducer` to a stub that returns a mock producer:

```go
oldProducer := newKafkaProducer
newKafkaProducer = func(/* args */) (*kafka.Producer, error) {
    return mockProducer, nil
}
defer func() { newKafkaProducer = oldProducer }()
```

Integration tests use `//go:build integration` to conditionally compile. See `processor_kafka_integration_test.go` for an example.

## Adding new processors (in addition to Kafka)
The application uses a factory and registry pattern. 
The `main.go` relies on blank imports to load each Go file with a `Processor`
That delegates to the `init()` functions which will register the `Processor` constructor with the factory. 
This allows dynamic processor selection based on configuration without hardcoding dependencies.

Adding new processors is trivial: 
- create the package `foo` under `processors`
- create the Go file `processor_foo.go`
- add `New()` method as constructor
- add `init()` with `processor.Register()` passing in the constructor
- add the blank-import in `main.go`
- implement the `Processor` methods (`Process` and `Close`)
