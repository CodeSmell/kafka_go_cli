package processor

import (
	"context"

	"kafka_go_cli/internal/model"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// avoid cicular imports by defining a mock processor here for testing the factory
// instead of the testutil package which is used by other packages and cannot import processor
type mockProcessorForTesting struct {
}

func (m *mockProcessorForTesting) Process(ctx context.Context, message model.Message) error {
	return nil
}

func (m *mockProcessorForTesting) Close() error {
	return nil
}

func TestFactoryRegister(t *testing.T) {

	Register("mock", func(ctx context.Context, logger *slog.Logger, settings model.Settings) (Processor, error) {
		return &mockProcessorForTesting{}, nil
	})

	// Verify that the factory was registered and can be invoked
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	settings := model.Settings{ProcessorType: "mock"}

	proc, err := NewProcessor(ctx, logger, settings)
	assert.NoError(t, err)
	assert.IsType(t, &mockProcessorForTesting{}, proc)
}

func TestFactoryRegisterNoName(t *testing.T) {
	// Register with empty name should panic
	assert.Panics(t, func() {
		Register("", func(ctx context.Context, logger *slog.Logger, settings model.Settings) (Processor, error) {
			return &mockProcessorForTesting{}, nil
		})
	})
}

func TestFactoryRegisterNoFactoryFunc(t *testing.T) {
	// Register with nil factory function should panic
	assert.Panics(t, func() {
		Register("mock", nil)
	})
}

func TestFactoryNewProcessorNoneRegistered(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	settings := model.Settings{ProcessorType: "nonexistent"}

	_, err := NewProcessor(ctx, logger, settings)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "processor not registered: nonexistent")
}

func TestRegisterAndGetConfigParams(t *testing.T) {
	processorName := "mock"
	params := []model.ConfigParam{
		{Name: "param1", Flag: "mock-param1", Description: "mock param1"},
		{Name: "param2", Flag: "mock-param2", Description: "mock param2"},
		{Name: "param3", Flag: "mock-param3", Description: "mock param3"},
	}

	RegisterConfigParams(processorName, params)

	configParams := GetConfigParams(processorName)
	assert.Equal(t, params, configParams)

	registeredProcessors := GetAllRegisteredProcessors()
	assert.Contains(t, registeredProcessors, processorName)
	assert.Equal(t, params, registeredProcessors[processorName])
}
