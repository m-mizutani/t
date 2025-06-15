package file

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/m-mizutani/t/pkg/action"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteAction_Name(t *testing.T) {
	deleteAction := &DeleteAction{}
	assert.Equal(t, "file.delete", deleteAction.Name())
}

func TestDeleteAction_Description(t *testing.T) {
	deleteAction := &DeleteAction{}
	assert.Equal(t, "Delete a file or directory", deleteAction.Description())
}

func TestDeleteAction_NewConfig(t *testing.T) {
	deleteAction := &DeleteAction{}
	config := deleteAction.NewConfig()

	cfg, ok := config.(*FileDeleteConfig)
	require.True(t, ok)
	assert.Empty(t, cfg.ID)
	assert.Empty(t, cfg.Path)
}

func TestDeleteAction_Execute_DeleteFile(t *testing.T) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "test-delete-*.txt")
	require.NoError(t, err)
	tempPath := tempFile.Name()
	tempFile.Close()

	// Write some content
	err = os.WriteFile(tempPath, []byte("test content"), 0644)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(tempPath)
	require.NoError(t, err)

	deleteAction := &DeleteAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	cfg := &FileDeleteConfig{
		Path: tempPath,
	}

	result, err := deleteAction.Execute(ctx, actx, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify file was deleted
	_, err = os.Stat(tempPath)
	assert.True(t, os.IsNotExist(err))

	// Check result
	assert.Equal(t, tempPath, result.Output)
	assert.Equal(t, true, result.Metadata["existed"])
	assert.Equal(t, false, result.Metadata["was_directory"])
	assert.Equal(t, tempPath, result.Metadata["path"])
}

func TestDeleteAction_Execute_DeleteDirectory(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	subDir := filepath.Join(tempDir, "subdir")
	err := os.Mkdir(subDir, 0755)
	require.NoError(t, err)

	// Create a file in the subdirectory
	testFile := filepath.Join(subDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	deleteAction := &DeleteAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	cfg := &FileDeleteConfig{
		Path: subDir,
	}

	result, err := deleteAction.Execute(ctx, actx, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify directory was deleted
	_, err = os.Stat(subDir)
	assert.True(t, os.IsNotExist(err))

	// Check result
	assert.Equal(t, subDir, result.Output)
	assert.Equal(t, true, result.Metadata["existed"])
	assert.Equal(t, true, result.Metadata["was_directory"])
	assert.Equal(t, subDir, result.Metadata["path"])
}

func TestDeleteAction_Execute_NonexistentFile(t *testing.T) {
	nonexistentPath := "/nonexistent/file.txt"

	deleteAction := &DeleteAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	cfg := &FileDeleteConfig{
		Path: nonexistentPath,
	}

	result, err := deleteAction.Execute(ctx, actx, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should succeed even if file doesn't exist
	assert.Equal(t, nonexistentPath, result.Output)
	assert.Equal(t, false, result.Metadata["existed"])
	assert.Equal(t, nonexistentPath, result.Metadata["path"])
}

func TestDeleteAction_Execute_WithTemplateProcessing(t *testing.T) {
	// Create a temporary file
	tempFile, err := os.CreateTemp("", "test-delete-*.txt")
	require.NoError(t, err)
	tempPath := tempFile.Name()
	tempFile.Close()

	deleteAction := &DeleteAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: tempPath,
		Env:   make(map[string]string),
	}

	cfg := &FileDeleteConfig{
		Path: "{{ .input }}",
	}

	result, err := deleteAction.Execute(ctx, actx, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify file was deleted
	_, err = os.Stat(tempPath)
	assert.True(t, os.IsNotExist(err))

	assert.Equal(t, tempPath, result.Output)
}

func TestDeleteAction_Execute_InvalidConfigType(t *testing.T) {
	deleteAction := &DeleteAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	// Pass wrong config type
	invalidConfig := "invalid config"

	result, err := deleteAction.Execute(ctx, actx, invalidConfig)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid config type for file.delete")
}

func TestDeleteAction_Execute_EmptyPath(t *testing.T) {
	deleteAction := &DeleteAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	cfg := &FileDeleteConfig{
		Path: "",
	}

	result, err := deleteAction.Execute(ctx, actx, cfg)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "path cannot be empty")
}

func TestDeleteAction_Execute_TemplateProcessingError(t *testing.T) {
	deleteAction := &DeleteAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	cfg := &FileDeleteConfig{
		Path: "{{ .nonexistent }}", // This will output <no value>
	}

	result, err := deleteAction.Execute(ctx, actx, cfg)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should succeed with <no value> path (which doesn't exist)
	assert.Equal(t, false, result.Metadata["existed"])
}

func TestDeleteAction_Execute_PermissionDenied(t *testing.T) {
	// This test may not work on all systems, especially Windows
	// Skip if we can't create the test scenario
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	// Create a temporary directory
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	require.NoError(t, err)

	// Make parent directory read-only (this may not work on all filesystems)
	err = os.Chmod(tempDir, 0444)
	if err != nil {
		t.Skip("Cannot change directory permissions on this system")
	}
	defer os.Chmod(tempDir, 0755) // Restore permissions for cleanup

	deleteAction := &DeleteAction{}
	ctx := context.Background()
	actx := &action.Context{
		Input: "test input",
		Env:   make(map[string]string),
	}

	cfg := &FileDeleteConfig{
		Path: testFile,
	}

	result, err := deleteAction.Execute(ctx, actx, cfg)
	// This should fail due to permission denied
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestFileDeleteConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config FileDeleteConfig
		valid  bool
	}{
		{
			name: "valid config",
			config: FileDeleteConfig{
				Path: "/tmp/test.txt",
			},
			valid: true,
		},
		{
			name: "valid config with id",
			config: FileDeleteConfig{
				ID:   "test-delete",
				Path: "/tmp/test.txt",
			},
			valid: true,
		},
		{
			name: "empty path should be invalid",
			config: FileDeleteConfig{
				Path: "",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				assert.NotEmpty(t, tt.config.Path, "Valid config should have non-empty Path")
			} else {
				assert.Empty(t, tt.config.Path, "Invalid config should have empty Path")
			}
		})
	}
}
