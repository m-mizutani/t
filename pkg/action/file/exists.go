package file

import (
	"context"
	"log/slog"
	"os"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/t/pkg/action"
	"github.com/m-mizutani/t/pkg/config"
)

// ExistsAction implements file.exists action
type ExistsAction struct{}

// Name returns the action name
func (a *ExistsAction) Name() string {
	return "file.exists"
}

// Description returns a human-readable description
func (a *ExistsAction) Description() string {
	return "Check if a file or directory exists"
}

// Execute runs the file.exists action
func (a *ExistsAction) Execute(ctx context.Context, actx *action.Context, step config.LegacyStepConfig) (*action.Result, error) {
	logger := actx.Logger(ctx)
	logger.Debug("Executing file.exists action")

	// Get path from Args
	pathArg, exists := step.Args["path"]
	if !exists {
		return nil, goerr.New("path argument is required for file.exists action")
	}

	pathStr, ok := pathArg.(string)
	if !ok {
		return nil, goerr.New("path argument must be string")
	}

	// Process template
	path, err := action.ProcessTemplate(pathStr, actx, "file.exists")
	if err != nil {
		return nil, goerr.Wrap(err, "failed to process path template")
	}

	logger.Debug("Checking file existence",
		slog.String("path", path),
	)

	// Check if file/directory exists
	info, err := os.Stat(path)
	exists = err == nil

	var metadata map[string]interface{}
	if exists {
		metadata = map[string]interface{}{
			"path":         path,
			"exists":       true,
			"is_directory": info.IsDir(),
			"size":         info.Size(),
			"mode":         info.Mode().String(),
			"mod_time":     info.ModTime(),
		}
	} else {
		metadata = map[string]interface{}{
			"path":   path,
			"exists": false,
		}
	}

	logger.Debug("File existence check completed",
		slog.String("path", path),
		slog.Bool("exists", exists),
	)

	return &action.Result{
		Output:   exists,
		Metadata: metadata,
	}, nil
}
