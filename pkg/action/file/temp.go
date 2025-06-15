package file

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/t/pkg/action"
)

// FileTempConfig represents configuration for file.temp action
type FileTempConfig struct {
	ID      string `yaml:"id,omitempty"`
	Content string `yaml:"content,omitempty"`
	Prefix  string `yaml:"prefix,omitempty"`
	Suffix  string `yaml:"suffix,omitempty"`
	Dir     string `yaml:"dir,omitempty"`
}

// TempAction implements file.temp action
type TempAction struct{}

// Name returns the action name
func (a *TempAction) Name() string {
	return "file.temp"
}

// Description returns a human-readable description
func (a *TempAction) Description() string {
	return "Create a temporary file"
}

// NewConfig returns a new instance of FileTempConfig
func (a *TempAction) NewConfig() interface{} {
	return &FileTempConfig{}
}

// Execute runs the file.temp action with typed configuration
func (a *TempAction) Execute(ctx context.Context, actx *action.Context, config interface{}) (*action.Result, error) {
	logger := actx.Logger(ctx)
	logger.Debug("Executing file.temp action")

	// Type assertion to get our config
	cfg, ok := config.(*FileTempConfig)
	if !ok {
		return nil, goerr.New("invalid config type for file.temp")
	}

	// Set default prefix if not specified
	prefix := cfg.Prefix
	if prefix == "" {
		prefix = "t-"
	}

	// Process template for prefix if needed
	if prefix != "" {
		processedPrefix, err := action.ProcessTemplate(prefix, actx, "file.temp.prefix")
		if err != nil {
			return nil, goerr.Wrap(err, "failed to process prefix template")
		}
		prefix = processedPrefix
	}

	// Process template for suffix if needed
	suffix := cfg.Suffix
	if suffix != "" {
		processedSuffix, err := action.ProcessTemplate(suffix, actx, "file.temp.suffix")
		if err != nil {
			return nil, goerr.Wrap(err, "failed to process suffix template")
		}
		suffix = processedSuffix
	}

	// Process template for directory if needed
	dir := cfg.Dir
	if dir != "" {
		processedDir, err := action.ProcessTemplate(dir, actx, "file.temp.dir")
		if err != nil {
			return nil, goerr.Wrap(err, "failed to process dir template")
		}
		dir = processedDir
	}

	// Create temporary file
	tempFile, err := os.CreateTemp(dir, prefix+"*"+suffix)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to create temporary file")
	}
	defer tempFile.Close()

	tempPath := tempFile.Name()

	logger.Debug("Temporary file created",
		slog.String("path", tempPath),
		slog.String("prefix", prefix),
		slog.String("suffix", suffix),
		slog.String("dir", dir),
	)

	// Write content if provided
	var bytesWritten int
	if cfg.Content != "" {
		// Process template for content
		content, err := action.ProcessTemplate(cfg.Content, actx, "file.temp.content")
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
			"dir":           dir,
			"bytes_written": bytesWritten,
		},
	}, nil
}
