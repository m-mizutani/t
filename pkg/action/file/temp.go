package file

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/t/pkg/action"
	"github.com/m-mizutani/t/pkg/config"
)

// TempAction implements file.temp action
type TempAction struct{}

// Name returns the action name
func (a *TempAction) Name() string {
	return "file.temp"
}

// Description returns a human-readable description
func (a *TempAction) Description() string {
	return "Create a temporary file with optional content"
}

// Execute runs the file.temp action
func (a *TempAction) Execute(ctx context.Context, actx *action.Context, step config.StepConfig) (*action.Result, error) {
	logger := actx.Logger(ctx)
	logger.Debug("Executing file.temp action")

	// Get optional prefix and suffix
	prefix := "t-"
	if prefixArg, exists := step.Args["prefix"]; exists {
		if prefixStr, ok := prefixArg.(string); ok {
			prefix = prefixStr
		}
	}

	suffix := ""
	if suffixArg, exists := step.Args["suffix"]; exists {
		if suffixStr, ok := suffixArg.(string); ok {
			suffix = suffixStr
		}
	}

	// Create temporary file
	tempFile, err := os.CreateTemp("", prefix+"*"+suffix)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to create temporary file")
	}
	defer tempFile.Close()

	tempPath := tempFile.Name()

	logger.Debug("Temporary file created",
		slog.String("path", tempPath),
		slog.String("prefix", prefix),
		slog.String("suffix", suffix),
	)

	// Write content if provided
	var bytesWritten int
	if contentArg, exists := step.Args["content"]; exists {
		contentStr := fmt.Sprintf("%v", contentArg)
		// Process template
		content, err := action.ProcessTemplate(contentStr, actx, "file.temp")
		if err != nil {
			os.Remove(tempPath) // Clean up on error
			return nil, goerr.Wrap(err, "failed to process content template")
		}
		if _, err := tempFile.WriteString(content); err != nil {
			os.Remove(tempPath) // Clean up on error
			return nil, goerr.Wrap(err, "failed to write content to temp file")
		}
		bytesWritten = len(content)

		logger.Debug("Content written to temp file",
			slog.String("path", tempPath),
			slog.Int("bytes", bytesWritten),
		)
	} else if actx.Input != nil {
		// Write input to temp file
		content := fmt.Sprintf("%v", actx.Input)
		if _, err := tempFile.WriteString(content); err != nil {
			os.Remove(tempPath) // Clean up on error
			return nil, goerr.Wrap(err, "failed to write input to temp file")
		}
		bytesWritten = len(content)

		logger.Debug("Input written to temp file",
			slog.String("path", tempPath),
			slog.Int("bytes", bytesWritten),
		)
	}

	return &action.Result{
		Output: tempPath,
		Metadata: map[string]interface{}{
			"path":          tempPath,
			"prefix":        prefix,
			"suffix":        suffix,
			"bytes_written": bytesWritten,
		},
	}, nil
}
