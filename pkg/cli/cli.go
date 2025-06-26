package cli

import (
	"bufio"
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/t/pkg/app"
	"github.com/m-mizutani/t/pkg/config"
	"github.com/m-mizutani/t/pkg/logger"
	"github.com/urfave/cli/v3"
)

//go:embed template.yaml
var configTemplate string

// Run is the main entry point for the CLI application
func Run(ctx context.Context, args []string) error {
	cmd := &cli.Command{
		Name:  "t",
		Usage: "Task runner",
		Description: `t is a task runner that integrates LLM capabilities for various automation tasks.
It supports file operations, LLM interactions, and MCP (Model Context Protocol) integration.

Usage:
  t [options] <task-name> [arguments...]

Examples:
  t hello                        # Run hello task
  t echo "Hello World"           # Run echo task with argument
  t --config custom.yaml my-task # Run with custom config file
  t --init                       # Initialize task.yaml with template
  t --init --config custom.yaml  # Initialize custom.yaml with template`,
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
			&cli.BoolFlag{
				Name:    "init",
				Aliases: []string{"i"},
				Usage:   "Initialize configuration file with template",
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
	// Handle init command
	if c.Bool("init") {
		return initConfig(ctx, c)
	}

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

// initConfig initializes configuration file with template
func initConfig(ctx context.Context, c *cli.Command) error {
	configPath := c.String("config")

	// Check if file already exists
	if _, err := os.Stat(configPath); err == nil {
		// File exists, ask for confirmation
		fmt.Printf("Configuration file already exists: %s\n", configPath)
		fmt.Print("Do you want to overwrite it? (y/N): ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return goerr.Wrap(err, "failed to read user input")
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Initialization cancelled.")
			return nil
		}
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return goerr.Wrap(err, "failed to create directory", goerr.Value("dir", dir))
	}

	// Create template content
	template := getConfigTemplate()

	// Write template to file
	if err := os.WriteFile(configPath, []byte(template), 0644); err != nil {
		return goerr.Wrap(err, "failed to write config file", goerr.Value("path", configPath))
	}

	fmt.Printf("Configuration file initialized: %s\n", configPath)
	fmt.Println("\nYou can now:")
	fmt.Println("1. Set your API keys (OpenAI, Anthropic, etc.) in environment variables")
	fmt.Println("2. Customize the tasks in the configuration file")
	fmt.Println("3. Run tasks with: t <task-name>")

	return nil
}

// getConfigTemplate returns the template configuration content
func getConfigTemplate() string {
	return configTemplate
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

		// Show aliases if available
		if len(task.Aliases) > 0 {
			fmt.Printf(" ( %s )", strings.Join(task.Aliases, ", "))
		}

		// Show description if available
		if task.Description != "" {
			fmt.Printf(" - %s", task.Description)
		}

		fmt.Printf("\n")
	}

	fmt.Printf(`
Usage:
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
