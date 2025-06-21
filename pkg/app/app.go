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
	"github.com/m-mizutani/t/pkg/action/llm"
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

	// Register typed actions (new approach)
	registry.RegisterTyped(&command.RunAction{})
	registry.RegisterTyped(&file.WriteAction{})
	registry.RegisterTyped(&file.ReadAction{})
	registry.RegisterTyped(&file.TempAction{})
	registry.RegisterTyped(&file.DeleteAction{})
	registry.RegisterTyped(&stdout.WriteAction{})
	registry.RegisterTyped(&llm.GenerateAction{})
	registry.RegisterTyped(&clipboard.ReadAction{})
	registry.RegisterTyped(&clipboard.WriteAction{})

	// Process RawSteps to typed Steps
	if err := processSteps(cfg, registry); err != nil {
		return nil, goerr.Wrap(err, "failed to process steps")
	}

	return &App{
		cfg:      cfg,
		registry: registry,
	}, nil
}

// processSteps converts RawSteps to typed Steps using the registry
func processSteps(cfg *config.Config, registry *action.Registry) error {
	// Process each task
	for taskName, task := range cfg.Tasks {
		steps := make([]config.StepConfig, 0, len(task.RawSteps))

		for i, rawStep := range task.RawSteps {
			// Parse as typed action
			typedStep, err := registry.ParseStepConfig(rawStep)
			if err != nil {
				return goerr.Wrap(err, "failed to parse step config",
					goerr.Value("task", taskName),
					goerr.Value("step", i),
					goerr.Value("action", rawStep.Action))
			}
			steps = append(steps, *typedStep)
		}

		// Update the task with processed steps
		updatedTask := task
		updatedTask.Steps = steps
		cfg.Tasks[taskName] = updatedTask
	}

	return nil
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

	// Try typed action first
	if typedAction, exists := app.registry.GetTyped(step.Action); exists {
		// Execute typed action
		result, err := typedAction.Execute(ctx, actionCtx, step.Config)
		if err != nil {
			return goerr.Wrap(err, "typed action execution failed")
		}

		return app.handleStepResult(ctx, step, result, actionCtx, logger)
	}

	return goerr.New("unknown action", goerr.Value("action", step.Action))
}

// handleStepResult handles the result of a step execution
func (app *App) handleStepResult(ctx context.Context, step config.StepConfig, result *action.Result, actionCtx *action.Context, logger *slog.Logger) error {
	// Store result in context for next step
	if result != nil && result.Output != nil {
		actionCtx.Output = result.Output

		// Store output by action ID if specified (extract from config)
		var stepID string
		if configMap, ok := step.Config.(map[string]interface{}); ok {
			if id, exists := configMap["id"]; exists {
				if idStr, ok := id.(string); ok {
					stepID = idStr
				}
			}
		}

		if stepID != "" {
			actionCtx.ActionOutputs[stepID] = result.Output
		}

		// Store metadata if available
		if result.Metadata != nil {
			for key, value := range result.Metadata {
				actionCtx.Data[key] = value
			}
		}

		logger.Debug("Step completed",
			slog.String("action", step.Action),
			slog.String("id", stepID),
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
