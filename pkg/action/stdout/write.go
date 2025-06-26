package stdout

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/t/pkg/action"
)

// StdoutWriteConfig represents configuration for stdout.write action
type StdoutWriteConfig struct {
	ID        string `yaml:"id,omitempty"`
	Content   string `yaml:"content,omitempty"`
	NoNewline bool   `yaml:"no_newline,omitempty"` // Don't add automatic newline
}

// WriteAction implements stdout.write action
type WriteAction struct{}

// Name returns the action name
func (a *WriteAction) Name() string {
	return "stdout.write"
}

// Description returns a human-readable description
func (a *WriteAction) Description() string {
	return "Write content to stdout"
}

// NewConfig returns a new instance of StdoutWriteConfig
func (a *WriteAction) NewConfig() interface{} {
	return &StdoutWriteConfig{}
}

// Execute runs the stdout.write action with typed configuration
func (a *WriteAction) Execute(ctx context.Context, actx *action.Context, config interface{}) (*action.Result, error) {
	logger := actx.Logger(ctx)
	logger.Debug("Executing stdout.write action")

	// Type assertion to get our config
	cfg, ok := config.(*StdoutWriteConfig)
	if !ok {
		return nil, goerr.New("invalid config type for stdout.write")
	}

	var content string

	// Get content from config or input
	if cfg.Content != "" {
		// Process template
		var err error
		content, err = action.ProcessTemplate(cfg.Content, actx, "stdout.write.content")
		if err != nil {
			return nil, goerr.Wrap(err, "failed to process content template")
		}
	} else if actx.Output != nil {
		content = fmt.Sprintf("%v", actx.Output)
	} else {
		return nil, goerr.New("no output provided for stdout.write action")
	}

	// Write to stdout
	if _, err := fmt.Fprint(os.Stdout, content); err != nil {
		return nil, goerr.Wrap(err, "failed to write to stdout")
	}

	// Add newline if content doesn't end with one and no_newline is false
	if !cfg.NoNewline && len(content) > 0 && content[len(content)-1] != '\n' {
		fmt.Fprintln(os.Stdout)
	}

	logger.Debug("Content written to stdout",
		slog.Int("length", len(content)),
		slog.Bool("no_newline", cfg.NoNewline),
	)

	return &action.Result{
		Output: content,
		Metadata: map[string]interface{}{
			"length":     len(content),
			"no_newline": cfg.NoNewline,
		},
	}, nil
}
