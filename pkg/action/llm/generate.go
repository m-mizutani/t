package llm

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/gollem"
	"github.com/m-mizutani/t/pkg/action"
)

// LLMGenerateConfig represents configuration for llm.generate action
type LLMGenerateConfig struct {
	ID          string  `yaml:"id,omitempty"`
	System      string  `yaml:"system,omitempty"`
	Prompt      string  `yaml:"prompt,omitempty"`
	Model       string  `yaml:"model,omitempty"`
	Temperature float64 `yaml:"temperature,omitempty"`
	MaxTokens   int     `yaml:"max_tokens,omitempty"`
}

// GenerateAction implements llm.generate action
type GenerateAction struct{}

// Name returns the action name
func (a *GenerateAction) Name() string {
	return "llm.generate"
}

// Description returns a human-readable description
func (a *GenerateAction) Description() string {
	return "Generate text using LLM"
}

// NewConfig returns a new instance of LLMGenerateConfig
func (a *GenerateAction) NewConfig() interface{} {
	return &LLMGenerateConfig{}
}

// Execute runs the llm.generate action with typed configuration
func (a *GenerateAction) Execute(ctx context.Context, actx *action.Context, config interface{}) (*action.Result, error) {
	logger := actx.Logger(ctx)
	logger.Debug("Executing llm.generate action")

	// Type assertion to get our config
	cfg, ok := config.(*LLMGenerateConfig)
	if !ok {
		return nil, goerr.New("invalid config type for llm.generate")
	}

	// Process system message template
	var system string
	var err error
	if cfg.System != "" {
		system, err = action.ProcessTemplate(cfg.System, actx, "llm.generate.system")
		if err != nil {
			return nil, goerr.Wrap(err, "failed to process system template")
		}
	}

	// Process prompt template
	var prompt string
	if cfg.Prompt != "" {
		prompt, err = action.ProcessTemplate(cfg.Prompt, actx, "llm.generate.prompt")
		if err != nil {
			return nil, goerr.Wrap(err, "failed to process prompt template")
		}
	} else if actx.Input != nil {
		prompt = fmt.Sprintf("%v", actx.Input)
	} else {
		return nil, goerr.New("no prompt specified for llm.generate action")
	}

	logger.Debug("Sending prompt to LLM",
		slog.String("system", system),
		slog.Int("prompt_length", len(prompt)),
		slog.String("model", cfg.Model),
		slog.Float64("temperature", cfg.Temperature),
		slog.Int("max_tokens", cfg.MaxTokens),
	)

	// Get LLM configuration
	clientInfo, err := getLLMClientInfo(ctx, actx, cfg)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get LLM client info")
	}

	// Create session options
	var sessionOptions []gollem.SessionOption
	if system != "" {
		sessionOptions = append(sessionOptions, gollem.WithSessionSystemPrompt(system))
	}

	// Create session from LLM client
	session, err := clientInfo.Client.NewSession(ctx, sessionOptions...)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to create LLM session")
	}

	// Create input for the session
	input := gollem.Text(prompt)

	// Generate response using session
	response, err := session.GenerateContent(ctx, input)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to generate response")
	}

	// Check if response is nil
	if response == nil {
		return nil, goerr.New("received nil response from LLM")
	}

	// Extract text from response
	var responseText string
	if len(response.Texts) > 0 {
		responseText = strings.Join(response.Texts, "")
	}

	logger.Debug("LLM response generated",
		slog.Int("response_length", len(responseText)),
	)

	return &action.Result{
		Output: responseText,
		Metadata: map[string]interface{}{
			"provider":        clientInfo.Provider,
			"model":           clientInfo.Model,
			"system":          system,
			"prompt_length":   len(prompt),
			"response_length": len(responseText),
		},
	}, nil
}

// getLLMClientInfo creates LLM client info for typed configuration
func getLLMClientInfo(ctx context.Context, actx *action.Context, cfg *LLMGenerateConfig) (*ClientInfo, error) {
	llmConfig := actx.Config.Defaults.LLM
	provider := llmConfig.Provider

	model := cfg.Model
	if model == "" {
		model = llmConfig.Model
	}

	var client gollem.LLMClient
	var err error

	switch provider {
	case "openai":
		client, err = createOpenAIClientTyped(ctx, actx, cfg, llmConfig, model)
	case "claude":
		client, err = createClaudeClientTyped(ctx, actx, cfg, llmConfig, model)
	case "gemini":
		client, err = createGeminiClientTyped(ctx, actx, cfg, llmConfig, model)
	default:
		return nil, goerr.New("unsupported LLM provider", goerr.Value("provider", provider))
	}

	if err != nil {
		return nil, err
	}

	return &ClientInfo{
		Client:   client,
		Provider: provider,
		Model:    model,
	}, nil
}

// Helper functions to safely extract arguments
func getStringArg(args map[string]interface{}, key, defaultValue string) string {
	if val, exists := args[key]; exists {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

func getFloatArg(args map[string]interface{}, key string, defaultValue float64) float64 {
	if val, exists := args[key]; exists {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return defaultValue
}

func getIntArg(args map[string]interface{}, key string, defaultValue int) int {
	if val, exists := args[key]; exists {
		switch v := val.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		case float32:
			return int(v)
		}
	}
	return defaultValue
}
