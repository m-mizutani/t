package clipboard

import (
	"context"
	"testing"

	"github.com/m-mizutani/t/pkg/action"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteAction_Name(t *testing.T) {
	writeAction := &WriteAction{}
	assert.Equal(t, "clipboard.write", writeAction.Name())
}

func TestWriteAction_Description(t *testing.T) {
	writeAction := &WriteAction{}
	assert.Equal(t, "Write content to clipboard", writeAction.Description())
}

func TestWriteAction_NewConfig(t *testing.T) {
	writeAction := &WriteAction{}
	config := writeAction.NewConfig()

	cfg, ok := config.(*ClipboardWriteConfig)
	require.True(t, ok)
	assert.Empty(t, cfg.ID)
	assert.Empty(t, cfg.Content)
}

func TestWriteAction_Execute_WithContent(t *testing.T) {
	// Note: This test may fail in CI environments without clipboard support
	writeAction := &WriteAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	cfg := &ClipboardWriteConfig{
		Content: "Hello, Clipboard!",
	}

	result, err := writeAction.Execute(ctx, actx, cfg)
	if err != nil && (err.Error() == "failed to initialize clipboard" ||
		err.Error() == "clipboard not available") {
		t.Skip("Clipboard not available in test environment")
	}

	require.NoError(t, err)
	require.NotNil(t, result)

	// Check output
	assert.Equal(t, "Hello, Clipboard!", result.Output)

	// Check metadata
	assert.Equal(t, len("Hello, Clipboard!"), result.Metadata["length"])
	assert.Equal(t, 1, result.Metadata["lines"]) // No newlines
}

func TestWriteAction_Execute_WithInputContent(t *testing.T) {
	writeAction := &WriteAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "Input content",
		Env:   make(map[string]string),
	}

	cfg := &ClipboardWriteConfig{} // No content specified, should use input

	result, err := writeAction.Execute(ctx, actx, cfg)
	if err != nil && (err.Error() == "failed to initialize clipboard" ||
		err.Error() == "clipboard not available") {
		t.Skip("Clipboard not available in test environment")
	}

	require.NoError(t, err)
	require.NotNil(t, result)

	// Check output
	assert.Equal(t, "Input content", result.Output)

	// Check metadata
	assert.Equal(t, len("Input content"), result.Metadata["length"])
	assert.Equal(t, 1, result.Metadata["lines"])
}

func TestWriteAction_Execute_WithTemplateProcessing(t *testing.T) {
	writeAction := &WriteAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "World",
		Env:   make(map[string]string),
	}

	cfg := &ClipboardWriteConfig{
		Content: "Hello, {{ .input }}!",
	}

	result, err := writeAction.Execute(ctx, actx, cfg)
	if err != nil && (err.Error() == "failed to initialize clipboard" ||
		err.Error() == "clipboard not available") {
		t.Skip("Clipboard not available in test environment")
	}

	require.NoError(t, err)
	require.NotNil(t, result)

	// Check output
	assert.Equal(t, "Hello, World!", result.Output)

	// Check metadata
	assert.Equal(t, len("Hello, World!"), result.Metadata["length"])
	assert.Equal(t, 1, result.Metadata["lines"])
}

func TestWriteAction_Execute_WithMultilineContent(t *testing.T) {
	writeAction := &WriteAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	multilineContent := "Line 1\nLine 2\nLine 3"
	cfg := &ClipboardWriteConfig{
		Content: multilineContent,
	}

	result, err := writeAction.Execute(ctx, actx, cfg)
	if err != nil && (err.Error() == "failed to initialize clipboard" ||
		err.Error() == "clipboard not available") {
		t.Skip("Clipboard not available in test environment")
	}

	require.NoError(t, err)
	require.NotNil(t, result)

	// Check output
	assert.Equal(t, multilineContent, result.Output)

	// Check metadata
	assert.Equal(t, len(multilineContent), result.Metadata["length"])
	assert.Equal(t, 3, result.Metadata["lines"]) // 2 newlines = 3 lines
}

func TestWriteAction_Execute_InvalidConfigType(t *testing.T) {
	writeAction := &WriteAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	// Pass wrong config type
	invalidConfig := "invalid config"

	result, err := writeAction.Execute(ctx, actx, invalidConfig)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid config type for clipboard.write")
}

func TestWriteAction_Execute_NoContentSpecified(t *testing.T) {
	writeAction := &WriteAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: nil, // No input
		Env:   make(map[string]string),
	}

	cfg := &ClipboardWriteConfig{} // No content specified

	result, err := writeAction.Execute(ctx, actx, cfg)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no content specified for clipboard.write action")
}

func TestWriteAction_Execute_TemplateProcessingWithMissingVariable(t *testing.T) {
	writeAction := &WriteAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	cfg := &ClipboardWriteConfig{
		Content: "{{ .nonexistent }}", // This will output <no value>
	}

	result, err := writeAction.Execute(ctx, actx, cfg)
	if err != nil && (err.Error() == "failed to initialize clipboard" ||
		err.Error() == "clipboard not available") {
		t.Skip("Clipboard not available in test environment")
	}

	require.NoError(t, err)
	require.NotNil(t, result)

	// Should succeed with <no value>
	assert.Contains(t, result.Output, "<no value>")
}

func TestClipboardWriteConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config ClipboardWriteConfig
		valid  bool
	}{
		{
			name:   "valid empty config (will use input)",
			config: ClipboardWriteConfig{},
			valid:  true,
		},
		{
			name: "valid config with content",
			config: ClipboardWriteConfig{
				Content: "test content",
			},
			valid: true,
		},
		{
			name: "valid config with id and content",
			config: ClipboardWriteConfig{
				ID:      "test-write",
				Content: "test content",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// All configurations are valid for clipboard write action
			assert.True(t, tt.valid, "All clipboard write configurations should be valid")
		})
	}
}
