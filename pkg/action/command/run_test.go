package command

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/m-mizutani/t/pkg/action"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunAction_Name(t *testing.T) {
	action := &RunAction{}
	assert.Equal(t, "command.run", action.Name())
}

func TestRunAction_Description(t *testing.T) {
	action := &RunAction{}
	assert.Equal(t, "Execute a shell command", action.Description())
}

func TestRunAction_NewConfig(t *testing.T) {
	action := &RunAction{}
	config := action.NewConfig()

	cfg, ok := config.(*CommandRunConfig)
	require.True(t, ok)
	assert.Empty(t, cfg.ID)
	assert.Empty(t, cfg.Cmd)
	assert.Empty(t, cfg.Dir)
}

func TestRunAction_Execute_SimpleCommand(t *testing.T) {
	runAction := &RunAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	cfg := &CommandRunConfig{
		Cmd: "echo 'hello world'",
	}

	result, err := runAction.Execute(ctx, actx, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "hello world\n", result.Output)
	assert.Equal(t, true, result.Metadata["success"])
	assert.Equal(t, "echo 'hello world'", result.Metadata["command"])
	assert.Equal(t, "hello world\n", result.Metadata["stdout"])
	assert.Equal(t, "", result.Metadata["stderr"])
}

func TestRunAction_Execute_WithWorkingDirectory(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	runAction := &RunAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	cfg := &CommandRunConfig{
		Cmd: "cat test.txt",
		Dir: tempDir,
	}

	result, err := runAction.Execute(ctx, actx, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "test content", result.Output)
	assert.Equal(t, true, result.Metadata["success"])
	assert.Equal(t, tempDir, result.Metadata["workdir"])
}

func TestRunAction_Execute_WithTemplateProcessing(t *testing.T) {
	runAction := &RunAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "world",
		Env:   make(map[string]string),
	}

	cfg := &CommandRunConfig{
		Cmd: "echo 'hello {{ .input }}'",
	}

	result, err := runAction.Execute(ctx, actx, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, "hello world\n", result.Output)
	assert.Equal(t, true, result.Metadata["success"])
}

func TestRunAction_Execute_WithEnvironmentVariables(t *testing.T) {
	runAction := &RunAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env: map[string]string{
			"TEST_VAR": "test_value",
		},
	}

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "echo %TEST_VAR%"
	} else {
		cmd = "echo $TEST_VAR"
	}

	cfg := &CommandRunConfig{
		Cmd: cmd,
	}

	result, err := runAction.Execute(ctx, actx, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Contains(t, result.Output, "test_value")
	assert.Equal(t, true, result.Metadata["success"])
}

func TestRunAction_Execute_CommandFailure(t *testing.T) {
	runAction := &RunAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	cfg := &CommandRunConfig{
		Cmd: "exit 1",
	}

	result, err := runAction.Execute(ctx, actx, cfg)
	require.NoError(t, err) // Action should not fail even if command fails
	require.NotNil(t, result)

	assert.Equal(t, false, result.Metadata["success"])
	assert.Equal(t, 1, result.Metadata["exit_code"])
	assert.NotEmpty(t, result.Metadata["error"])
}

func TestRunAction_Execute_CommandWithStderr(t *testing.T) {
	runAction := &RunAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "echo error message 1>&2"
	} else {
		cmd = "echo 'error message' >&2"
	}

	cfg := &CommandRunConfig{
		Cmd: cmd,
	}

	result, err := runAction.Execute(ctx, actx, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Contains(t, result.Metadata["stderr"], "error message")
}

func TestRunAction_Execute_InvalidConfigType(t *testing.T) {
	runAction := &RunAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	// Pass wrong config type
	invalidConfig := "invalid config"

	result, err := runAction.Execute(ctx, actx, invalidConfig)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid config type for command.run")
}

func TestRunAction_Execute_TemplateWithMissingVariable(t *testing.T) {
	runAction := &RunAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	cfg := &CommandRunConfig{
		Cmd: "echo '{{ .nonexistent }}'", // This will output <no value>
	}

	result, err := runAction.Execute(ctx, actx, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Template processing should succeed but output <no value>
	assert.Contains(t, result.Output, "<no value>")
	assert.Equal(t, true, result.Metadata["success"])
}

func TestRunAction_Execute_NonexistentWorkingDirectory(t *testing.T) {
	runAction := &RunAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	cfg := &CommandRunConfig{
		Cmd: "echo 'hello'",
		Dir: "/nonexistent/directory",
	}

	result, err := runAction.Execute(ctx, actx, cfg)
	require.NoError(t, err) // Action should not fail
	require.NotNil(t, result)

	// Command should fail due to nonexistent directory
	assert.Equal(t, false, result.Metadata["success"])
	assert.NotEmpty(t, result.Metadata["error"])
}

func TestRunAction_Execute_LongRunningCommand(t *testing.T) {
	runAction := &RunAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "timeout /t 1 /nobreak >nul && echo done"
	} else {
		cmd = "sleep 0.1 && echo done"
	}

	cfg := &CommandRunConfig{
		Cmd: cmd,
	}

	result, err := runAction.Execute(ctx, actx, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Contains(t, result.Output, "done")
	assert.Equal(t, true, result.Metadata["success"])
}

func TestRunAction_Execute_CancelledContext(t *testing.T) {
	runAction := &RunAction{}
	ctx, cancel := context.WithCancel(context.Background())
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	// Cancel context immediately
	cancel()

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "timeout /t 10"
	} else {
		cmd = "sleep 10"
	}

	cfg := &CommandRunConfig{
		Cmd: cmd,
	}

	result, err := runAction.Execute(ctx, actx, cfg)
	require.NoError(t, err) // Action should not fail
	require.NotNil(t, result)

	// Command should fail due to cancelled context
	assert.Equal(t, false, result.Metadata["success"])
	assert.NotEmpty(t, result.Metadata["error"])
}

func TestRunAction_Execute_MultilineOutput(t *testing.T) {
	runAction := &RunAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	var cmd string
	if runtime.GOOS == "windows" {
		cmd = "echo line1 && echo line2 && echo line3"
	} else {
		cmd = "echo 'line1'; echo 'line2'; echo 'line3'"
	}

	cfg := &CommandRunConfig{
		Cmd: cmd,
	}

	result, err := runAction.Execute(ctx, actx, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	lines := strings.Split(strings.TrimSpace(result.Output.(string)), "\n")
	assert.Len(t, lines, 3)
	assert.Contains(t, lines[0], "line1")
	assert.Contains(t, lines[1], "line2")
	assert.Contains(t, lines[2], "line3")
	assert.Equal(t, true, result.Metadata["success"])
}

func TestCommandRunConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config CommandRunConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: CommandRunConfig{
				Cmd: "echo hello",
			},
			valid: true,
		},
		{
			name: "valid config with dir",
			config: CommandRunConfig{
				Cmd: "echo hello",
				Dir: "/tmp",
			},
			valid: true,
		},
		{
			name: "valid config with id",
			config: CommandRunConfig{
				ID:  "test-command",
				Cmd: "echo hello",
			},
			valid: true,
		},
		{
			name: "empty command should be invalid",
			config: CommandRunConfig{
				Cmd: "",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				assert.NotEmpty(t, tt.config.Cmd, "Valid config should have non-empty Cmd")
			} else {
				assert.Empty(t, tt.config.Cmd, "Invalid config should have empty Cmd")
			}
		})
	}
}
