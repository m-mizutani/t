package cli

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/t/pkg/app"
	"github.com/m-mizutani/t/pkg/config"
	"github.com/m-mizutani/t/pkg/logger"
	"github.com/urfave/cli/v3"
)

// Run is the main entry point for the CLI application
func Run(ctx context.Context, args []string) error {
	cmd := &cli.Command{
		Name:  "t",
		Usage: "LLM-powered task runner",
		Description: `t is a task runner that integrates LLM capabilities for various automation tasks.
It supports file operations, LLM interactions, and MCP (Model Context Protocol) integration.

Usage:
  t [options] <task-name> [arguments...]

Examples:
  t hello                        # Run hello task
  t echo "Hello World"           # Run echo task with argument
  t --config custom.yaml my-task # Run with custom config file`,
		ArgsUsage: "<task-name> [arguments...]",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Configuration file path",
				Value:   config.GetDefaultConfigPath(),
			},
			&cli.StringFlag{
				Name:    "log-level",
				Aliases: []string{"l"},
				Usage:   "Log level (debug, info, warn, error)",
				Value:   "info",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Enable verbose output",
			},
		},
		Before: func(ctx context.Context, c *cli.Command) (context.Context, error) {
			// Initialize logger
			loggerCfg := logger.Config{
				Level:   logger.LogLevel(c.String("log-level")),
				Verbose: c.Bool("verbose"),
			}
			appLogger := logger.New(loggerCfg)

			// Embed logger into context
			return logger.WithLogger(ctx, appLogger), nil
		},
		Action: runTaskAction,
	}

	return cmd.Run(ctx, args)
}

// runTaskAction handles task execution
func runTaskAction(ctx context.Context, c *cli.Command) error {
	// Get logger from context
	appLogger := logger.FromContext(ctx)

	// Check if task name is provided
	taskName := c.Args().First()
	if taskName == "" {
		return showTasks(ctx, c)
	}

	// Create application instance
	appInstance, err := createApp(c, appLogger)
	if err != nil {
		return goerr.Wrap(err, "failed to create application")
	}

	// Get all arguments except the task name
	taskArgs := c.Args().Slice()[1:]

	// Run the task
	if err := appInstance.Run(ctx, taskName, taskArgs); err != nil {
		return goerr.Wrap(err, "task execution failed", goerr.Value("task", taskName))
	}

	return nil
}

// createApp creates an application instance from CLI flags
func createApp(c *cli.Command, appLogger *slog.Logger) (*app.App, error) {
	appCfg := app.Config{
		ConfigPath: c.String("config"),
	}

	return app.New(appCfg)
}

// showTasks displays available tasks when no task is specified
func showTasks(ctx context.Context, c *cli.Command) error {
	// Create application instance to access configuration
	appInstance, err := createApp(c, logger.FromContext(ctx))
	if err != nil {
		return goerr.Wrap(err, "failed to create application")
	}

	cfg := appInstance.GetConfig()

	if len(cfg.Tasks) == 0 {
		fmt.Printf(`%s

No tasks available.

Please define tasks in your configuration file (%s).

Usage:
  %s [options] <task-name> [arguments...]

Options:
  --config, -c    Configuration file path
  --log-level, -l Log level (debug, info, warn, error) (default: "info")
  --verbose, -v   Enable verbose output
  --help, -h      Show help

Examples:
  %s hello                        # Run hello task
  %s echo "Hello World"           # Run echo task with argument
  %s --config custom.yaml my-task # Run with custom config file
`,
			c.Usage,
			c.String("config"),
			c.Name,
			c.Name,
			c.Name,
			c.Name,
		)
		return nil
	}

	fmt.Printf(`%s

Available tasks:

`, c.Usage)

	// Display tasks
	for taskName, task := range cfg.Tasks {
		fmt.Printf("  %s", taskName)
		if len(task.Aliases) > 0 {
			fmt.Printf(" (aliases: %v)", task.Aliases)
		}
		fmt.Printf("\n")

		// Show step count
		fmt.Printf("    Steps: %d\n", len(task.Steps))

		// Show first few actions
		if len(task.Steps) > 0 {
			fmt.Printf("    Actions: ")
			maxActions := 3
			for i, step := range task.Steps {
				if i >= maxActions {
					fmt.Printf("...")
					break
				}
				if i > 0 {
					fmt.Printf(" → ")
				}
				fmt.Printf("%s", step.Action)
			}
			fmt.Printf("\n")
		}
		fmt.Printf("\n")
	}

	fmt.Printf(`Usage:
  %s [options] <task-name> [arguments...]

Examples:
  %s hello                        # Run hello task
  %s echo "Hello World"           # Run echo task with argument
  %s --config custom.yaml my-task # Run with custom config file
`,
		c.Name,
		c.Name,
		c.Name,
		c.Name,
	)

	return nil
}
