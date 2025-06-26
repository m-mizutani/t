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
)

// CommandRunConfig represents configuration for command.run action
type CommandRunConfig struct {
	ID  string `yaml:"id,omitempty"`
	Cmd string `yaml:"cmd" validate:"required"`
	Dir string `yaml:"dir,omitempty"`
}

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

// NewConfig returns a new instance of CommandRunConfig
func (a *RunAction) NewConfig() interface{} {
	return &CommandRunConfig{}
}

// Execute runs the command.run action with typed configuration
func (a *RunAction) Execute(ctx context.Context, actx *action.Context, config interface{}) (*action.Result, error) {
	logger := actx.Logger(ctx)
	logger.Debug("Executing command.run action")

	// Type assertion to get our config
	cfg, ok := config.(*CommandRunConfig)
	if !ok {
		return nil, goerr.New("invalid config type for command.run")
	}

	// Process template for command
	command, err := action.ProcessTemplate(cfg.Cmd, actx, "command.run.cmd")
	if err != nil {
		return nil, goerr.Wrap(err, "failed to process command template")
	}

	// Process template for working directory if provided
	var workDir string
	if cfg.Dir != "" {
		workDir, err = action.ProcessTemplate(cfg.Dir, actx, "command.run.dir")
		if err != nil {
			return nil, goerr.Wrap(err, "failed to process dir template")
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
