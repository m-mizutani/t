package file

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/t/pkg/action"
)

// FileWriteConfig represents configuration for file.write action
type FileWriteConfig struct {
	ID      string `yaml:"id,omitempty"`
	Path    string `yaml:"path" validate:"required"`
	Content string `yaml:"content,omitempty"`
	Mode    string `yaml:"mode,omitempty"` // File permissions (octal string like "0644")
}

// WriteAction implements file.write action
type WriteAction struct{}

// Name returns the action name
func (a *WriteAction) Name() string {
	return "file.write"
}

// Description returns a human-readable description
func (a *WriteAction) Description() string {
	return "Write content to a file"
}

// NewConfig returns a new instance of FileWriteConfig
func (a *WriteAction) NewConfig() interface{} {
	return &FileWriteConfig{}
}

// Execute runs the file.write action with typed configuration
func (a *WriteAction) Execute(ctx context.Context, actx *action.Context, config interface{}) (*action.Result, error) {
	logger := actx.Logger(ctx)
	logger.Debug("Executing file.write action")

	// Type assertion to get our config
	cfg, ok := config.(*FileWriteConfig)
	if !ok {
		return nil, goerr.New("invalid config type for file.write")
	}

	// Process path template
	filePath, err := action.ProcessTemplate(cfg.Path, actx, "file.write.path")
	if err != nil {
		return nil, goerr.Wrap(err, "failed to process path template")
	}

	// Get content to write
	var content string
	if cfg.Content != "" {
		content, err = action.ProcessTemplate(cfg.Content, actx, "file.write.content")
		if err != nil {
			return nil, goerr.Wrap(err, "failed to process content template")
		}
	} else if actx.Input != nil {
		content = fmt.Sprintf("%v", actx.Input)
	} else {
		return nil, goerr.New("no content specified for file.write action")
	}

	// Parse file mode
	fileMode := os.FileMode(0644) // Default permission
	if cfg.Mode != "" {
		// TODO: Parse octal mode string (e.g., "0644" -> 0644)
		// For now, use default
	}

	logger.Debug("Writing file",
		slog.String("path", filePath),
		slog.Int("content_length", len(content)),
		slog.String("mode", fmt.Sprintf("%o", fileMode)),
	)

	// Create directory if it doesn't exist
	if dir := filepath.Dir(filePath); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, goerr.Wrap(err, "failed to create directory", goerr.Value("dir", dir))
		}
	}

	// Write content to file
	if err := os.WriteFile(filePath, []byte(content), fileMode); err != nil {
		return nil, goerr.Wrap(err, "failed to write file", goerr.Value("path", filePath))
	}

	logger.Debug("File written successfully",
		slog.String("path", filePath),
		slog.Int("bytes_written", len(content)),
	)

	return &action.Result{
		Output: filePath,
		Metadata: map[string]interface{}{
			"path":          filePath,
			"bytes_written": len(content),
			"mode":          fmt.Sprintf("%o", fileMode),
		},
	}, nil
}
