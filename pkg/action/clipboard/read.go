package clipboard

import (
	"context"
	"log/slog"
	"strings"

	"golang.design/x/clipboard"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/t/pkg/action"
)

// ClipboardReadConfig represents configuration for clipboard.read action
type ClipboardReadConfig struct {
	ID string `yaml:"id,omitempty"`
}

// ReadAction implements clipboard.read action
type ReadAction struct{}

// Name returns the action name
func (a *ReadAction) Name() string {
	return "clipboard.read"
}

// Description returns a human-readable description
func (a *ReadAction) Description() string {
	return "Read content from clipboard"
}

// NewConfig returns a new instance of ClipboardReadConfig
func (a *ReadAction) NewConfig() interface{} {
	return &ClipboardReadConfig{}
}

// Execute runs the clipboard.read action with typed configuration
func (a *ReadAction) Execute(ctx context.Context, actx *action.Context, config interface{}) (*action.Result, error) {
	logger := actx.Logger(ctx)
	logger.Debug("Executing clipboard.read action")

	// Type assertion to get our config
	_, ok := config.(*ClipboardReadConfig)
	if !ok {
		return nil, goerr.New("invalid config type for clipboard.read")
	}

	// Initialize clipboard
	if err := clipboard.Init(); err != nil {
		return nil, goerr.Wrap(err, "failed to initialize clipboard")
	}

	// Read clipboard content
	data := clipboard.Read(clipboard.FmtText)
	content := string(data)

	logger.Debug("Clipboard content read",
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
