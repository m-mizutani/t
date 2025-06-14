package command

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/t/pkg/action"
	"github.com/m-mizutani/t/pkg/config"
)

// RunAction implements command.run action
type RunAction struct{}

// Name returns the action name
func (a *RunAction) Name() string {
	return "command.run"
}

// Description returns a human-readable description
func (a *RunAction) Description() string {
	return "Execute a shell command"
}

// Execute runs the command.run action
func (a *RunAction) Execute(ctx context.Context, actx *action.Context, step config.StepConfig) (*action.Result, error) {
	logger := actx.Logger(ctx)
	logger.Debug("Executing command.run action")

	// Get command from step args
	cmdArg, exists := step.Args["cmd"]
	if !exists {
		return nil, goerr.New("cmd argument is required for command.run action")
	}

	cmdStr, ok := cmdArg.(string)
	if !ok {
		return nil, goerr.New("cmd argument must be string")
	}

	// Process template
	command, err := action.ProcessTemplate(cmdStr, actx, "command.run.cmd")
	if err != nil {
		return nil, goerr.Wrap(err, "failed to process command template")
	}

	// Get working directory (optional)
	var workDir string
	if dirArg, exists := step.Args["dir"]; exists {
		if dirStr, ok := dirArg.(string); ok {
			workDir, err = action.ProcessTemplate(dirStr, actx, "command.run.dir")
			if err != nil {
				return nil, goerr.Wrap(err, "failed to process dir template")
			}
		}
	}

	logger.Debug("Running command",
		slog.String("command", command),
		slog.String("workdir", workDir),
	)

	// Create command
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	if workDir != "" {
		cmd.Dir = workDir
	}

	// Set environment variables
	cmd.Env = os.Environ()
	for key, value := range actx.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run command
	err = cmd.Run()

	logger.Debug("Command execution completed",
		slog.String("command", command),
		slog.Bool("success", err == nil),
		slog.Int("stdout_length", stdout.Len()),
		slog.Int("stderr_length", stderr.Len()),
	)

	// Prepare result
	result := &action.Result{
		Output: stdout.String(),
		Metadata: map[string]interface{}{
			"command":       command,
			"workdir":       workDir,
			"stdout":        stdout.String(),
			"stderr":        stderr.String(),
			"success":       err == nil,
			"stdout_length": stdout.Len(),
			"stderr_length": stderr.Len(),
		},
	}

	if err != nil {
		// Include error information but don't fail the action
		if exitError, ok := err.(*exec.ExitError); ok {
			result.Metadata["exit_code"] = exitError.ExitCode()
		}
		result.Metadata["error"] = err.Error()

		logger.Debug("Command failed",
			slog.String("command", command),
			slog.String("error", err.Error()),
			slog.String("stderr", stderr.String()),
		)
	}

	return result, nil
}

// getArgValue extracts and processes template for argument value
func (a *RunAction) getArgValue(args map[string]interface{}, key string, actx *action.Context) (string, error) {
	value, exists := args[key]
	if !exists {
		return "", nil
	}

	valueStr, ok := value.(string)
	if !ok {
		return "", goerr.New("argument must be string", goerr.Value("key", key), goerr.Value("value", value))
	}

	return action.ProcessTemplate(valueStr, actx, "command.run."+key)
}
