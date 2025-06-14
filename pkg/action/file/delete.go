package file

import (
	"context"
	"log/slog"
	"os"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/t/pkg/action"
	"github.com/m-mizutani/t/pkg/config"
)

// DeleteAction implements file.delete action
type DeleteAction struct{}

// Name returns the action name
func (a *DeleteAction) Name() string {
	return "file.delete"
}

// Description returns a human-readable description
func (a *DeleteAction) Description() string {
	return "Delete a file or directory"
}

// Execute runs the file.delete action
func (a *DeleteAction) Execute(ctx context.Context, actx *action.Context, step config.StepConfig) (*action.Result, error) {
	logger := actx.Logger(ctx)
	logger.Debug("Executing file.delete action")

	// Get path from Args
	pathArg, exists := step.Args["path"]
	if !exists {
		return nil, goerr.New("path argument is required for file.delete action")
	}

	pathStr, ok := pathArg.(string)
	if !ok {
		return nil, goerr.New("path argument must be string")
	}

	// Process template
	path, err := action.ProcessTemplate(pathStr, actx, "file.delete")
	if err != nil {
		return nil, goerr.Wrap(err, "failed to process path template")
	}

	if path == "" {
		return nil, goerr.New("path cannot be empty")
	}

	logger.Debug("Deleting file/directory",
		slog.String("path", path),
	)

	// Check if path exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		// File doesn't exist, consider it successful
		logger.Debug("File/directory already does not exist",
			slog.String("path", path),
		)
		return &action.Result{
			Output: path,
			Metadata: map[string]interface{}{
				"path":    path,
				"existed": false,
			},
		}, nil
	}
	if err != nil {
		return nil, goerr.Wrap(err, "failed to stat path", goerr.Value("path", path))
	}

	// Delete the file or directory
	var deleteErr error
	if info.IsDir() {
		deleteErr = os.RemoveAll(path)
	} else {
		deleteErr = os.Remove(path)
	}

	if deleteErr != nil {
		return nil, goerr.Wrap(deleteErr, "failed to delete", goerr.Value("path", path))
	}

	logger.Debug("File/directory deleted successfully",
		slog.String("path", path),
		slog.Bool("was_directory", info.IsDir()),
	)

	return &action.Result{
		Output: path,
		Metadata: map[string]interface{}{
			"path":          path,
			"existed":       true,
			"was_directory": info.IsDir(),
		},
	}, nil
}
