// Package model defines shared data types across layers.
// This package has no external dependencies and can be safely imported everywhere.
package model

// KeyValue represents a single key/value pair.
// Used for both message keys and headers.
type KeyValue struct {
	Key   string
	Value string
}

// Message is the structured representation of a polled file.
// Multiple keys and headers (including duplicates) are allowed.
type Message struct {
	Keys    []KeyValue
	Headers []KeyValue
	Body    string
}

// ConfigParam describes a single configuration parameter for a processor.
// Used for dynamic CLI flag registration and documentation.
type ConfigParam struct {
	Name        string // "topic" - cache key in ProcessorConfig map
	Flag        string // "kafka-topic" - CLI flag name
	Description string // Help text for --help
}

// Settings holds the complete runtime configuration for the application.
// It aggregates values from all configuration sources (defaults, config file, env, CLI flags).
type Settings struct {
	ConfigFile      string                 `json:"config_file"`
	LogLevel        string                 `mapstructure:"log_level" json:"log_level"`
	MessageLocation string                 `mapstructure:"message_location" json:"message_location"`
	RunOnce         bool                   `mapstructure:"run_once" json:"run_once"`
	NoDeleteFiles   bool                   `mapstructure:"no_delete_files" json:"no_delete_files"`
	Delay           int                    `mapstructure:"delay" json:"delay"`
	MaxCycles       int                    `mapstructure:"max_cycles" json:"max_cycles"`
	ProcessorType   string                 `mapstructure:"processor" json:"processor"`
	ProcessorConfig map[string]interface{} `mapstructure:"-" json:"-"` // Raw processor-specific config
}
