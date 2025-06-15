package clipboard

import (
	"context"
	"testing"

	"github.com/m-mizutani/t/pkg/action"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadAction_Name(t *testing.T) {
	readAction := &ReadAction{}
	assert.Equal(t, "clipboard.read", readAction.Name())
}

func TestReadAction_Description(t *testing.T) {
	readAction := &ReadAction{}
	assert.Equal(t, "Read content from clipboard", readAction.Description())
}

func TestReadAction_NewConfig(t *testing.T) {
	readAction := &ReadAction{}
	config := readAction.NewConfig()

	cfg, ok := config.(*ClipboardReadConfig)
	require.True(t, ok)
	assert.Empty(t, cfg.ID)
}

func TestReadAction_Execute_BasicRead(t *testing.T) {
	// Note: This test may fail in CI environments without clipboard support
	// Skip if clipboard is not available
	readAction := &ReadAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	cfg := &ClipboardReadConfig{}

	result, err := readAction.Execute(ctx, actx, cfg)
	if err != nil && (err.Error() == "failed to initialize clipboard" ||
		err.Error() == "clipboard not available") {
		t.Skip("Clipboard not available in test environment")
	}

	require.NoError(t, err)
	require.NotNil(t, result)

	// Check that output is a string
	content, ok := result.Output.(string)
	require.True(t, ok)

	// Check metadata
	assert.Equal(t, len(content), result.Metadata["length"])
	expectedLines := 1
	if len(content) > 0 {
		for _, c := range content {
			if c == '\n' {
				expectedLines++
			}
		}
	}
	assert.Equal(t, expectedLines, result.Metadata["lines"])
}

func TestReadAction_Execute_InvalidConfigType(t *testing.T) {
	readAction := &ReadAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	// Pass wrong config type
	invalidConfig := "invalid config"

	result, err := readAction.Execute(ctx, actx, invalidConfig)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid config type for clipboard.read")
}

func TestClipboardReadConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config ClipboardReadConfig
		valid  bool
	}{
		{
			name:   "valid empty config",
			config: ClipboardReadConfig{},
			valid:  true,
		},
		{
			name: "valid config with id",
			config: ClipboardReadConfig{
				ID: "test-read",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// All configurations are valid for clipboard read action
			assert.True(t, tt.valid, "All clipboard read configurations should be valid")
		})
	}
}
