package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/t/pkg/action"
	"github.com/m-mizutani/t/pkg/action/clipboard"
	"github.com/m-mizutani/t/pkg/action/command"
	"github.com/m-mizutani/t/pkg/action/file"
	llmAction "github.com/m-mizutani/t/pkg/action/llm"
	"github.com/m-mizutani/t/pkg/action/stdout"
	"github.com/m-mizutani/t/pkg/config"
	"github.com/m-mizutani/t/pkg/logger"
)

// App represents the main application
type App struct {
	cfg      *config.Config
	registry *action.Registry
}

// Config represents application configuration options
type Config struct {
	ConfigPath string
}

// New creates a new application instance
func New(appCfg Config) (*App, error) {
	// Load configuration
	cfg, err := config.LoadWithFallback(appCfg.ConfigPath)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to load configuration")
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, goerr.Wrap(err, "invalid configuration")
	}

	// Create action registry and register built-in actions
	registry := action.NewRegistry()
	registry.Register(&stdout.WriteAction{})
	registry.Register(&file.ReadAction{})
	registry.Register(&file.WriteAction{})
	registry.Register(&file.TempAction{})
	registry.Register(&file.ExistsAction{})
	registry.Register(&file.DeleteAction{})
	registry.Register(&file.CopyAction{})
	registry.Register(&command.RunAction{})
	registry.Register(&llmAction.GenerateAction{})
	registry.Register(&llmAction.SessionAction{})
	registry.Register(&clipboard.ReadAction{})
	registry.Register(&clipboard.WriteAction{})

	return &App{
		cfg:      cfg,
		registry: registry,
	}, nil
}

// Run executes a task with the given name and arguments
func (app *App) Run(ctx context.Context, taskName string, args []string) error {
	logger := logger.FromContext(ctx)
	logger.Debug("Running task", slog.String("task", taskName), slog.Any("args", args))

	// Get task configuration
	task, err := app.cfg.GetTask(taskName)
	if err != nil {
		return goerr.Wrap(err, "failed to get task")
	}

	// Create action context
	actionCtx := action.NewContext(app.cfg, args)

	// Execute each step
	for i, step := range task.Steps {
		logger.Debug("Executing step", slog.Int("step", i), slog.String("action", step.Action))

		if err := app.executeStep(ctx, step, actionCtx); err != nil {
			return goerr.Wrap(err, "failed to execute step", goerr.Value("step", i), goerr.Value("action", step.Action))
		}
	}

	logger.Info("Task completed successfully", slog.String("task", taskName))
	return nil
}

// executeStep executes a single step
func (app *App) executeStep(ctx context.Context, step config.StepConfig, actionCtx *action.Context) error {
	logger := logger.FromContext(ctx)

	// Get action from registry
	actionImpl, exists := app.registry.Get(step.Action)
	if !exists {
		return goerr.New("unknown action", goerr.Value("action", step.Action))
	}

	// Execute action
	result, err := actionImpl.Execute(ctx, actionCtx, step)
	if err != nil {
		return goerr.Wrap(err, "action execution failed")
	}

	// Store result in context for next step
	if result != nil && result.Output != nil {
		actionCtx.Input = result.Output

		// Store output by action ID if specified
		if step.ID != "" {
			actionCtx.ActionOutputs[step.ID] = result.Output
		}

		// Store metadata if available
		if result.Metadata != nil {
			for key, value := range result.Metadata {
				actionCtx.Data[key] = value
			}
		}

		logger.Debug("Step completed",
			slog.String("action", step.Action),
			slog.String("id", step.ID),
			slog.Any("output_type", fmt.Sprintf("%T", result.Output)),
		)
	}

	return nil
}

// GetConfig returns the loaded configuration
func (app *App) GetConfig() *config.Config {
	return app.cfg
}

// GetRegistry returns the action registry
func (app *App) GetRegistry() *action.Registry {
	return app.registry
}
