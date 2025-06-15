package clipboard

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"golang.design/x/clipboard"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/t/pkg/action"
)

// ClipboardWriteConfig represents configuration for clipboard.write action
type ClipboardWriteConfig struct {
	ID      string `yaml:"id,omitempty"`
	Content string `yaml:"content,omitempty"`
}

// WriteAction implements clipboard.write action
type WriteAction struct{}

// Name returns the action name
func (a *WriteAction) Name() string {
	return "clipboard.write"
}

// Description returns a human-readable description
func (a *WriteAction) Description() string {
	return "Write content to clipboard"
}

// NewConfig returns a new instance of ClipboardWriteConfig
func (a *WriteAction) NewConfig() interface{} {
	return &ClipboardWriteConfig{}
}

// Execute runs the clipboard.write action with typed configuration
func (a *WriteAction) Execute(ctx context.Context, actx *action.Context, config interface{}) (*action.Result, error) {
	logger := actx.Logger(ctx)
	logger.Debug("Executing clipboard.write action")

	// Type assertion to get our config
	cfg, ok := config.(*ClipboardWriteConfig)
	if !ok {
		return nil, goerr.New("invalid config type for clipboard.write")
	}

	// Get content to write
	var content string
	if cfg.Content != "" {
		// Process template for content
		processedContent, err := action.ProcessTemplate(cfg.Content, actx, "clipboard.write.content")
		if err != nil {
			return nil, goerr.Wrap(err, "failed to process content template")
		}
		content = processedContent
	} else if actx.Input != nil {
		content = fmt.Sprintf("%v", actx.Input)
	} else {
		return nil, goerr.New("no content specified for clipboard.write action")
	}

	logger.Debug("Writing content to clipboard",
		slog.Int("length", len(content)),
	)

	// Initialize clipboard
	if err := clipboard.Init(); err != nil {
		return nil, goerr.Wrap(err, "failed to initialize clipboard")
	}

	// Write content to clipboard
	clipboard.Write(clipboard.FmtText, []byte(content))

	logger.Debug("Content written to clipboard",
		slog.Int("length", len(content)),
	)

	return &action.Result{
		Output: content,
		Metadata: map[string]interface{}{
			"length": len(content),
			"lines":  strings.Count(content, "\n") + 1,
		},
	}, nil
}
