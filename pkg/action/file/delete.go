package file

import (
	"context"
	"log/slog"
	"os"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/t/pkg/action"
)

// FileDeleteConfig represents configuration for file.delete action
type FileDeleteConfig struct {
	ID   string `yaml:"id,omitempty"`
	Path string `yaml:"path" validate:"required"`
}

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

// NewConfig returns a new instance of FileDeleteConfig
func (a *DeleteAction) NewConfig() interface{} {
	return &FileDeleteConfig{}
}

// Execute runs the file.delete action with typed configuration
func (a *DeleteAction) Execute(ctx context.Context, actx *action.Context, config interface{}) (*action.Result, error) {
	logger := actx.Logger(ctx)
	logger.Debug("Executing file.delete action")

	// Type assertion to get our config
	cfg, ok := config.(*FileDeleteConfig)
	if !ok {
		return nil, goerr.New("invalid config type for file.delete")
	}

	// Process template for path
	path, err := action.ProcessTemplate(cfg.Path, actx, "file.delete.path")
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
