package config

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/m-mizutani/goerr/v2"
	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Defaults DefaultConfig `yaml:"defaults"`
	MCP      []MCPConfig   `yaml:"mcp"`
	Tasks    TasksConfig   `yaml:"tasks"`
}

// DefaultConfig represents default settings
type DefaultConfig struct {
	LLM LLMConfig `yaml:"llm"`
}

// LLMConfig represents LLM configuration
type LLMConfig struct {
	Provider         string  `yaml:"provider"`
	Model            string  `yaml:"model"`
	Temperature      float64 `yaml:"temperature"`
	MaxTokens        int     `yaml:"max_tokens,omitempty"`
	TopP             float64 `yaml:"top_p,omitempty"`
	TopK             int     `yaml:"top_k,omitempty"`
	FrequencyPenalty float64 `yaml:"frequency_penalty,omitempty"`
	PresencePenalty  float64 `yaml:"presence_penalty,omitempty"`
	Timeout          int     `yaml:"timeout,omitempty"`
	APIKey           string  `yaml:"api_key,omitempty"`
}

// MCPConfig represents MCP server configuration
type MCPConfig struct {
	Name      string            `yaml:"name"`
	Transport string            `yaml:"transport"`
	Path      string            `yaml:"path"`
	Args      []string          `yaml:"args"`
	Envs      map[string]string `yaml:"envs"`
}

// TasksConfig represents all tasks configuration
type TasksConfig map[string]TaskConfig

// TaskConfig represents a single task configuration
type TaskConfig struct {
	Aliases []string     `yaml:"aliases"`
	Steps   []StepConfig `yaml:"steps"`
}

// StepConfig represents a single step configuration
type StepConfig struct {
	ID     string                 `yaml:"id,omitempty"`
	Action string                 `yaml:"action"`
	Input  interface{}            `yaml:"input,omitempty"`
	Path   string                 `yaml:"path,omitempty"`
	System string                 `yaml:"system,omitempty"`
	Prompt string                 `yaml:"prompt,omitempty"`
	Args   map[string]interface{} `yaml:"args,omitempty"`
}

// Load loads configuration from file
func Load(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = GetDefaultConfigPath()
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to read config file", goerr.Value("path", configPath))
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, goerr.Wrap(err, "failed to parse YAML config", goerr.Value("path", configPath))
	}

	return &config, nil
}

// LoadWithFallback loads configuration with fallback to default
func LoadWithFallback(configPath string) (*Config, error) {
	config, err := Load(configPath)
	if err != nil {
		// If default config file doesn't exist, return minimal config
		if os.IsNotExist(err) && configPath == GetDefaultConfigPath() {
			return GetDefaultConfig(), nil
		}
		return nil, err
	}
	return config, nil
}

// GetDefaultConfigPath returns the default configuration file path
func GetDefaultConfigPath() string {
	var configDir string

	if runtime.GOOS == "windows" {
		// Windows: use %APPDATA%
		appData := os.Getenv("APPDATA")
		if appData != "" {
			configDir = filepath.Join(appData, "t")
		} else {
			// Fallback to user home directory
			home, err := os.UserHomeDir()
			if err != nil {
				return "task.yaml"
			}
			configDir = filepath.Join(home, "AppData", "Roaming", "t")
		}
	} else {
		// Unix-like systems: use ~/.config/t
		home, err := os.UserHomeDir()
		if err != nil {
			return "task.yaml"
		}
		configDir = filepath.Join(home, ".config", "t")
	}

	return filepath.Join(configDir, "task.yaml")
}

// GetDefaultConfig returns a minimal default configuration
func GetDefaultConfig() *Config {
	return &Config{
		Defaults: DefaultConfig{
			LLM: LLMConfig{
				Provider:         "openai",
				Model:            "gpt-4",
				Temperature:      0.3,
				MaxTokens:        1000,
				TopP:             1.0,
				TopK:             40,
				FrequencyPenalty: 0.0,
				PresencePenalty:  0.0,
				Timeout:          60,
			},
		},
		MCP:   []MCPConfig{},
		Tasks: TasksConfig{},
	}
}

// GetTask returns a task by name or alias
func (c *Config) GetTask(name string) (*TaskConfig, error) {
	// Direct name match
	if task, exists := c.Tasks[name]; exists {
		return &task, nil
	}

	// Search by alias
	for _, task := range c.Tasks {
		for _, alias := range task.Aliases {
			if alias == name {
				return &task, nil
			}
		}
	}

	return nil, goerr.New("task not found", goerr.Value("name", name))
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate LLM config
	if c.Defaults.LLM.Provider == "" {
		return goerr.New("LLM provider is required")
	}

	// Validate MCP configs
	for i, mcp := range c.MCP {
		if mcp.Name == "" {
			return goerr.New("MCP name is required", goerr.Value("index", i))
		}
		if mcp.Transport == "" {
			return goerr.New("MCP transport is required", goerr.Value("name", mcp.Name))
		}
		if mcp.Transport == "stdio" && mcp.Path == "" {
			return goerr.New("MCP path is required for stdio transport", goerr.Value("name", mcp.Name))
		}
	}

	// Validate tasks
	for taskName, task := range c.Tasks {
		if len(task.Steps) == 0 {
			return goerr.New("task must have at least one step", goerr.Value("task", taskName))
		}

		for i, step := range task.Steps {
			if step.Action == "" {
				return goerr.New("step action is required",
					goerr.Value("task", taskName),
					goerr.Value("step", i),
				)
			}
		}
	}

	return nil
}
