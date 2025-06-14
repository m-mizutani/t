package clipboard

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"golang.design/x/clipboard"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/t/pkg/action"
	"github.com/m-mizutani/t/pkg/config"
)

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

// Execute runs the clipboard.write action
func (a *WriteAction) Execute(ctx context.Context, actx *action.Context, step config.StepConfig) (*action.Result, error) {
	logger := actx.Logger(ctx)
	logger.Debug("Executing clipboard.write action")

	// Get content to write
	var content string
	if contentArg, exists := step.Args["content"]; exists {
		content = fmt.Sprintf("%v", contentArg)
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
