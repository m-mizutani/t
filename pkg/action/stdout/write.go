package stdout

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/t/pkg/action"
	"github.com/m-mizutani/t/pkg/config"
)

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

// Execute runs the stdout.write action
func (a *WriteAction) Execute(ctx context.Context, actx *action.Context, step config.StepConfig) (*action.Result, error) {
	logger := actx.Logger(ctx)
	logger.Debug("Executing stdout.write action")

	var content string

	// Get content from args, input, or step configuration
	if contentArg, exists := step.Args["content"]; exists {
		contentStr := fmt.Sprintf("%v", contentArg)
		// Process template
		var err error
		content, err = action.ProcessTemplate(contentStr, actx, "stdout")
		if err != nil {
			return nil, goerr.Wrap(err, "failed to process content template")
		}
	} else if actx.Input != nil {
		content = fmt.Sprintf("%v", actx.Input)
	} else {
		return nil, goerr.New("no input provided for stdout.write action")
	}

	// Write to stdout
	if _, err := fmt.Fprint(os.Stdout, content); err != nil {
		return nil, goerr.Wrap(err, "failed to write to stdout")
	}

	// Add newline if content doesn't end with one
	if len(content) > 0 && content[len(content)-1] != '\n' {
		fmt.Fprintln(os.Stdout)
	}

	logger.Debug("Content written to stdout",
		slog.Int("length", len(content)),
	)

	return &action.Result{
		Output: content,
		Metadata: map[string]interface{}{
			"length": len(content),
		},
	}, nil
}
