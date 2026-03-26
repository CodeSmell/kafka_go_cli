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
  │   └── README.md
  ├── configs/
  │   └── README.md
  └── docs/
      └── README.md
```

Common Go layout:
- `cmd/` contains executable entrypoints
- `internal/` contains private application packages
- `configs/` stores config files
- `docs/` stores documents and learning notes 

## Basic commands for developer workflow 
Unlike Java, building the Go project produces an executable artifact.

Run the scaffold:

```bash
go run ./cmd/kafka-go-cli
```

Format code:

```bash
go fmt ./...
```

Run tests:

```bash
go test ./...
```

Vet code:

```bash
go vet ./...
```
