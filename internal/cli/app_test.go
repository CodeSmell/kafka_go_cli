package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

// setupCmdWithOutput creates a root command and sets up output capture.
// Returns both the root command and a buffer containing any output (stdout or stderr).
func setupCmdWithOutput(t *testing.T) (*bytes.Buffer, func(args ...string) error) {
	rootCmd := NewRootCommand()
	output := &bytes.Buffer{}
	rootCmd.SetOut(output)
	rootCmd.SetErr(output)

	exec := func(args ...string) error {
		rootCmd.SetArgs(args)
		return rootCmd.Execute()
	}
	return output, exec
}

func isCommandHelpShown(output *bytes.Buffer) bool {
	return bytes.Contains(output.Bytes(), []byte("Usage:"))
}

func TestScanSilenceUsageOnRuntimeErrors(t *testing.T) {
	output, exec := setupCmdWithOutput(t)

	// Execute with invalid directory
	err := exec("--message-location", "/nonexistent/dir")

	// We expect an error (directory doesn't exist)
	assert.Error(t, err)

	assert.False(t, isCommandHelpShown(output))
}

func TestCheckOK(t *testing.T) {
	tmpDir := t.TempDir()

	output, exec := setupCmdWithOutput(t)
	err := exec("--message-location", tmpDir, "--check")

	assert.NoError(t, err)
	assert.False(t, isCommandHelpShown(output))
	assert.Contains(t, output.String(), "OK")
}

func TestCheckInvalidDirExitsNonZero(t *testing.T) {
	output, exec := setupCmdWithOutput(t)
	err := exec("--message-location", "/nonexistent/path", "--check")

	assert.Error(t, err)
	assert.False(t, isCommandHelpShown(output))
	assert.Contains(t, output.String(), "PROBLEMS:")
}

func TestRunWithValidDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	output, exec := setupCmdWithOutput(t)
	// Set max-cycles to 1 to prevent infinite polling in tests
	err := exec("--message-location", tmpDir, "--max-cycles", "1")

	assert.NoError(t, err)
	assert.False(t, isCommandHelpShown(output))
}

func TestRunMissingRequiredFlag(t *testing.T) {
	output, exec := setupCmdWithOutput(t)
	// no message-location which is required
	err := exec()

	assert.Error(t, err)
	assert.True(t, isCommandHelpShown(output))
}

func TestCheckMissingRequiredFlagExitsNonZero(t *testing.T) {
	output, exec := setupCmdWithOutput(t)
	// no message-location which is required
	err := exec("--check")

	assert.Error(t, err)
	assert.False(t, isCommandHelpShown(output))
	assert.Contains(t, output.String(), "PROBLEMS:")
}
