// Package processor implements a registry-based factory for processor implementations.
//
// This follows the database/sql driver registration pattern:
//   - Implementations call processor.Register(name, factory) in init()
//   - The application blank-imports processor packages to enable them
//   - NewProcessor(settings) looks up and invokes the registered constructor in the factory
//
// This allows the processor package to avoid importing concrete implementations,
// enabling cleaner dependency graphs and dynamic processor selection.
package processor

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"kafka_go_cli/internal/model"
)

// FactoryFunc creates a Processor instance for a given processor type.
// Implementations register a FactoryFunc function (constructor) that NewProcessor invokes.
type FactoryFunc func(ctx context.Context, logger *slog.Logger, settings model.Settings) (Processor, error)

var (
	factoriesMu      sync.RWMutex
	factoryFunctions = map[string]FactoryFunc{}
	configParamsMu   sync.RWMutex
	configParamsMap  = map[string][]model.ConfigParam{}
)

// Panics on duplicate registration, empty name, or nil factory.
// This matches database/sql behavior to catch configuration errors early.
func Register(name string, factory FactoryFunc) {
	if name == "" || factory == nil {
		panic("processor: Register called with empty name or nil factory")
	}

	factoriesMu.Lock()
	defer factoriesMu.Unlock()
	if _, exists := factoryFunctions[name]; exists {
		panic("processor: Register called twice for " + name)
	}
	factoryFunctions[name] = factory
}

// RegisterConfigParams registers the configuration parameters for a processor.
// Called by each processor package in its init() function to declare its config.
func RegisterConfigParams(processorName string, params []model.ConfigParam) {
	configParamsMu.Lock()
	defer configParamsMu.Unlock()
	configParamsMap[processorName] = params
}

// GetConfigParams retrieves the configuration parameters for a specific processor by name.
// Returns nil if processor not found.
func GetConfigParams(processorName string) []model.ConfigParam {
	configParamsMu.RLock()
	defer configParamsMu.RUnlock()
	return configParamsMap[processorName]
}

// NewProcessor creates a processor based on the configured type.
// by looking up the registered factory function and invoking it.
func NewProcessor(ctx context.Context, logger *slog.Logger, settings model.Settings) (Processor, error) {
	logger.Info("creating processor", "type", settings.ProcessorType)

	factoriesMu.RLock()
	factory := factoryFunctions[settings.ProcessorType]
	factoriesMu.RUnlock()

	if factory != nil {
		return factory(ctx, logger, settings)
	}

	return nil, fmt.Errorf("processor not registered: %s", settings.ProcessorType)
}

// GetAllRegisteredProcessors returns a map
// The key is the processorType (kafka, pulsar, noop)
// and the value is an array of ConfigParam
// that are valid for the given Processor
func GetAllRegisteredProcessors() map[string][]model.ConfigParam {
	configParamsMu.RLock()
	defer configParamsMu.RUnlock()

	result := make(map[string][]model.ConfigParam)
	for name, params := range configParamsMap {
		result[name] = params
	}
	return result
}
