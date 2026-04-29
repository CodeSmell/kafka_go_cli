# Kafka Publish Util (in Go)
This utility will monitor a directory. It will publish each file in the specified directory to a Kafka topic. This CLI utility was built to enable easy testing of products that consume from Kafka topics. 

This repo is a rewrite of the utility originally written as a Java project. This was undertaken with the primary goal being to learn and experiment with the Go language on a practical but small application.

The original Java version:
- [KafkaPubCLI](https://github.com/CodeSmell/KafkaPubCLI)

## Using the Utility
The utility is meant to be run as a simple CLI relying on configuration to tell it where to find files and where to publish them

```bash
go run ./cmd/kafka-go-cli [flags]
```

The cofiguration flags/parameters values from lowest to highest precedence
  - defaults
  - config file (optional)
  - environment variables
  - CLI flags (highest priority)

| Flag/Parameter | Description |
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

Processor-specific flags are registered dynamically based on the selected processor. 

The Kafka based values
| Flag/Parameter | Description |
|---|---|
| `--kafka-brokers` | comma separated list of brokers (host:port). Default is localhost:9092  |
| `--kafka-broker-address-family` | Kafka broker address family (for example v4). Default is blank |
| `--kafka-topic` | Kafka topic where messages will be published |
| `--kafka-acks` | Kafka acks setting (0, 1, or all). Default is all |
| `--kafka-retries` | Number of retries for failed sends |
| `--kafka-retry-delay` | Milliseconds between retries |
| `--kafka-max-in-flight` | Max messages in-flight |
| `--kafka-batch-size` | Max batch size in bytes |
| `--kafka-batch-delay` | Milliseconds to wait before sending batch |
| `--kafka-producer-id` | Producer ID for batching |

### Format for the Files
The utility does not expect any format when publishing only a body on a Kafka message. 
The entire contents of the file will be published as the body of the message.
It will not treat the EOL in any special way so this allows the content (like JSON) to be pretty printed and still be sent in its entirety. 

For example a JSON that is pretty printed. 

```
{
	"foo": "bar",
	"code": "smell"
}
```

This allows easy scanning and editing the file when testing.

### Adding a Key and Headers to the Kafka message
However, if a key and headers are needed in the Kafka message the following format can be used:

At the top of the file add the `--key` delimiter.
Every line above that delimiter will be expected to be the value of the key. 
It is also expected that it will be one line and always appear first if it is needed.

At the top of the file (but below the `--key` delimiter if there is one) add the `--header` delimiter.
Every line above that delimiter will be expected to be a key value pair separated by a colon (:).
All of the content below that delimiter will be considered the body of the Kafka message.

```
foo
--key
hello:world
ghost:buster
--header
foobar
```

The contents above will result in a single Kafka message.
The body will be `foobar` and there will be two headers. 
The key for one of the headers will be `hello` and the value will be `world`. 
The key for the other header will be `ghost` and the value will be `buster`. 
The key for the Kafka payload will be `foo`.

### Quick Kafka commands
#### List the topics
Verify the topics on the broker

```bash
kafka-topics --list --bootstrap-server localhost:9092
```

#### Create a topics
Topics (in dev) are often created automatically. However if you want more control to set the number of partitions, replicas:

```bash
kafka-topics --create --topic alertsTopic --bootstrap-server localhost:9092 --partitions 10
```

To verify the topic:

```bash
kafka-topics --describe --topic alertsTopic --bootstrap-server localhost:9092 
```

#### Consuming messages 
An easy way to test the application is to consume them with the Kafka CLI consumer to verify things are working

```bash
kafka-console-consumer --bootstrap-server localhost:9092 \
  --topic alertsTopic \
  --property print.offset=true \
  --property print.partition=true \
  --property print.key=true \
  --property print.headers=true 
```

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
  │   ├── config/
  │   ├── model/
  │   ├── file/
  │   │    ├── poller.go              # DirectoryPoller 
  |   │    └── poller_test.go
  │   ├── processor/
  │   │    ├── processor.go           # the interface/type (Processor)
  │   │    └── processor_factory.go
  │   └── processors/
  │        ├── noop
  │        │     └── processor_noop.go
  │        ├── kafka/
  │        │     └── processor_kafka.go
  │        └── pulsar/
  |              └── processor_pulsar.go
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

Major packages & highlights
- Cobra command tree with a single command (no subcommands)
- Viper-based config resolution
- Logs are structured text logs using Go's standard `log/slog` package
- Testcontainers-go for integration tests

The basic design pattern allows each processor to register a constructor and set of parameters with a factory. The specified processor will be created and those parameters will be used to connect and publish the files that are in the specified directory (`message-location`)

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

## Sample commands 
Running the application

```bash
go run ./cmd/kafka-go-cli [flags]
```

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
