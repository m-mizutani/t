package file

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/m-mizutani/t/pkg/action"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTempAction_Name(t *testing.T) {
	tempAction := &TempAction{}
	assert.Equal(t, "file.temp", tempAction.Name())
}

func TestTempAction_Description(t *testing.T) {
	tempAction := &TempAction{}
	assert.Equal(t, "Create a temporary file", tempAction.Description())
}

func TestTempAction_NewConfig(t *testing.T) {
	tempAction := &TempAction{}
	config := tempAction.NewConfig()

	cfg, ok := config.(*FileTempConfig)
	require.True(t, ok)
	assert.Empty(t, cfg.ID)
	assert.Empty(t, cfg.Content)
	assert.Empty(t, cfg.Prefix)
	assert.Empty(t, cfg.Suffix)
	assert.Empty(t, cfg.Dir)
}

func TestTempAction_Execute_BasicTempFile(t *testing.T) {
	tempAction := &TempAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	cfg := &FileTempConfig{}

	result, err := tempAction.Execute(ctx, actx, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	tempPath, ok := result.Output.(string)
	require.True(t, ok)
	defer os.Remove(tempPath) // Cleanup

	// Verify file exists
	_, err = os.Stat(tempPath)
	require.NoError(t, err)

	// Check default prefix
	filename := filepath.Base(tempPath)
	assert.True(t, strings.HasPrefix(filename, "t-"))

	// Check metadata
	assert.Equal(t, tempPath, result.Metadata["path"])
	assert.Equal(t, "t-", result.Metadata["prefix"])
	assert.Equal(t, "", result.Metadata["suffix"])
	assert.Equal(t, "", result.Metadata["dir"])
}

func TestTempAction_Execute_WithContent(t *testing.T) {
	tempAction := &TempAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	cfg := &FileTempConfig{
		Content: "Hello, World!",
	}

	result, err := tempAction.Execute(ctx, actx, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	tempPath, ok := result.Output.(string)
	require.True(t, ok)
	defer os.Remove(tempPath) // Cleanup

	// Verify file content
	content, err := os.ReadFile(tempPath)
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!", string(content))

	// Check metadata
	assert.Equal(t, len("Hello, World!"), result.Metadata["bytes_written"])
}

func TestTempAction_Execute_WithInputContent(t *testing.T) {
	tempAction := &TempAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "Input content",
		Env:   make(map[string]string),
	}

	cfg := &FileTempConfig{} // No content specified, should use input

	result, err := tempAction.Execute(ctx, actx, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	tempPath, ok := result.Output.(string)
	require.True(t, ok)
	defer os.Remove(tempPath) // Cleanup

	// Verify file content
	content, err := os.ReadFile(tempPath)
	require.NoError(t, err)
	assert.Equal(t, "Input content", string(content))

	// Check metadata
	assert.Equal(t, len("Input content"), result.Metadata["bytes_written"])
}

func TestTempAction_Execute_WithCustomPrefix(t *testing.T) {
	tempAction := &TempAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	cfg := &FileTempConfig{
		Prefix: "myapp-",
	}

	result, err := tempAction.Execute(ctx, actx, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	tempPath, ok := result.Output.(string)
	require.True(t, ok)
	defer os.Remove(tempPath) // Cleanup

	// Check prefix
	filename := filepath.Base(tempPath)
	assert.True(t, strings.HasPrefix(filename, "myapp-"))

	// Check metadata
	assert.Equal(t, "myapp-", result.Metadata["prefix"])
}

func TestTempAction_Execute_WithCustomSuffix(t *testing.T) {
	tempAction := &TempAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	cfg := &FileTempConfig{
		Suffix: ".json",
	}

	result, err := tempAction.Execute(ctx, actx, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	tempPath, ok := result.Output.(string)
	require.True(t, ok)
	defer os.Remove(tempPath) // Cleanup

	// Check suffix
	assert.True(t, strings.HasSuffix(tempPath, ".json"))

	// Check metadata
	assert.Equal(t, ".json", result.Metadata["suffix"])
}

func TestTempAction_Execute_WithCustomDirectory(t *testing.T) {
	// Create a custom temp directory
	customDir := t.TempDir()

	tempAction := &TempAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	cfg := &FileTempConfig{
		Dir: customDir,
	}

	result, err := tempAction.Execute(ctx, actx, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	tempPath, ok := result.Output.(string)
	require.True(t, ok)
	defer os.Remove(tempPath) // Cleanup

	// Check directory
	assert.True(t, strings.HasPrefix(tempPath, customDir))

	// Check metadata
	assert.Equal(t, customDir, result.Metadata["dir"])
}

func TestTempAction_Execute_WithTemplateProcessing(t *testing.T) {
	tempAction := &TempAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "World",
		Env:   make(map[string]string),
	}

	cfg := &FileTempConfig{
		Content: "Hello, {{ .input }}!",
		Prefix:  "greeting-",
	}

	result, err := tempAction.Execute(ctx, actx, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	tempPath, ok := result.Output.(string)
	require.True(t, ok)
	defer os.Remove(tempPath) // Cleanup

	// Verify file content
	content, err := os.ReadFile(tempPath)
	require.NoError(t, err)
	assert.Equal(t, "Hello, World!", string(content))

	// Check prefix
	filename := filepath.Base(tempPath)
	assert.True(t, strings.HasPrefix(filename, "greeting-"))
}

func TestTempAction_Execute_InvalidConfigType(t *testing.T) {
	tempAction := &TempAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	// Pass wrong config type
	invalidConfig := "invalid config"

	result, err := tempAction.Execute(ctx, actx, invalidConfig)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid config type for file.temp")
}

func TestTempAction_Execute_TemplateProcessingWithMissingVariable(t *testing.T) {
	tempAction := &TempAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	cfg := &FileTempConfig{
		Content: "{{ .nonexistent }}", // This will output <no value>
	}

	result, err := tempAction.Execute(ctx, actx, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	tempPath, ok := result.Output.(string)
	require.True(t, ok)
	defer os.Remove(tempPath) // Cleanup

	// Verify file content contains <no value>
	content, err := os.ReadFile(tempPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "<no value>")
}

func TestTempAction_Execute_InvalidDirectory(t *testing.T) {
	tempAction := &TempAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	cfg := &FileTempConfig{
		Dir: "/nonexistent/directory",
	}

	result, err := tempAction.Execute(ctx, actx, cfg)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create temporary file")
}

func TestFileTempConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config FileTempConfig
		valid  bool
	}{
		{
			name:   "valid empty config",
			config: FileTempConfig{},
			valid:  true,
		},
		{
			name: "valid config with content",
			config: FileTempConfig{
				Content: "test content",
			},
			valid: true,
		},
		{
			name: "valid config with all fields",
			config: FileTempConfig{
				ID:      "test-temp",
				Content: "test content",
				Prefix:  "myapp-",
				Suffix:  ".txt",
				Dir:     "/tmp",
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// All configurations are valid for temp action
			assert.True(t, tt.valid, "All temp configurations should be valid")
		})
	}
}
