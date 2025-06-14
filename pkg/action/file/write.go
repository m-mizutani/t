package file

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/t/pkg/action"
	"github.com/m-mizutani/t/pkg/config"
)

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

// Execute runs the file.write action
func (a *WriteAction) Execute(ctx context.Context, actx *action.Context, step config.StepConfig) (*action.Result, error) {
	logger := actx.Logger(ctx)
	logger.Debug("Executing file.write action")

	// Get file path
	pathArg, exists := step.Args["path"]
	if !exists {
		return nil, goerr.New("path argument is required for file.write action")
	}

	pathStr, ok := pathArg.(string)
	if !ok {
		return nil, goerr.New("path argument must be string")
	}

	// Process path template
	filePath, err := action.ProcessTemplate(pathStr, actx, "file.write.path")
	if err != nil {
		return nil, goerr.Wrap(err, "failed to process path template")
	}

	// Get content to write
	var content string
	if contentArg, exists := step.Args["content"]; exists {
		if contentStr, ok := contentArg.(string); ok {
			content, err = action.ProcessTemplate(contentStr, actx, "file.write.content")
			if err != nil {
				return nil, goerr.Wrap(err, "failed to process content template")
			}
		} else {
			content = fmt.Sprintf("%v", contentArg)
		}
	} else if actx.Input != nil {
		content = fmt.Sprintf("%v", actx.Input)
	} else {
		return nil, goerr.New("no content specified for file.write action")
	}

	logger.Debug("Writing file",
		slog.String("path", filePath),
		slog.Int("content_length", len(content)),
	)

	// Create directory if it doesn't exist
	if dir := filepath.Dir(filePath); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, goerr.Wrap(err, "failed to create directory", goerr.Value("dir", dir))
		}
	}

	// Write content to file
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
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
		},
	}, nil
}
