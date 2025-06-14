package file

import (
	"context"
	"log/slog"
	"os"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/t/pkg/action"
	"github.com/m-mizutani/t/pkg/config"
)

// ReadAction implements file.read action
type ReadAction struct{}

// Name returns the action name
func (a *ReadAction) Name() string {
	return "file.read"
}

// Description returns a human-readable description
func (a *ReadAction) Description() string {
	return "Read content from a file"
}

// Execute runs the file.read action
func (a *ReadAction) Execute(ctx context.Context, actx *action.Context, step config.StepConfig) (*action.Result, error) {
	logger := actx.Logger(ctx)
	logger.Debug("Executing file.read action")

	// Get file path
	pathArg, exists := step.Args["path"]
	if !exists {
		return nil, goerr.New("path argument is required for file.read action")
	}

	pathStr, ok := pathArg.(string)
	if !ok {
		return nil, goerr.New("path argument must be string")
	}

	// Process template
	filePath, err := action.ProcessTemplate(pathStr, actx, "file.read")
	if err != nil {
		return nil, goerr.Wrap(err, "failed to process path template")
	}

	logger.Debug("Reading file",
		slog.String("path", filePath),
	)

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, goerr.New("file not found", goerr.Value("path", filePath))
		}
		return nil, goerr.Wrap(err, "failed to read file", goerr.Value("path", filePath))
	}

	contentStr := string(content)

	// Get file info
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get file info", goerr.Value("path", filePath))
	}

	logger.Debug("File read successfully",
		slog.String("path", filePath),
		slog.Int("size", len(content)),
	)

	return &action.Result{
		Output: contentStr,
		Metadata: map[string]interface{}{
			"path":     filePath,
			"size":     len(content),
			"mod_time": info.ModTime(),
			"mode":     info.Mode().String(),
		},
	}, nil
}
