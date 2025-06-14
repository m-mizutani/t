package llm

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/gollem"
	"github.com/m-mizutani/gollem/llm/claude"
	"github.com/m-mizutani/gollem/llm/gemini"
	"github.com/m-mizutani/gollem/llm/openai"
	"github.com/m-mizutani/t/pkg/action"
	"github.com/m-mizutani/t/pkg/config"
)

// GenerateAction implements llm.generate action
type GenerateAction struct{}

// Name returns the action name
func (a *GenerateAction) Name() string {
	return "llm.generate"
}

// Description returns a human-readable description
func (a *GenerateAction) Description() string {
	return "Generate text using a Large Language Model"
}

// Execute runs the llm.generate action
func (a *GenerateAction) Execute(ctx context.Context, actx *action.Context, step config.StepConfig) (*action.Result, error) {
	logger := actx.Logger(ctx)
	logger.Debug("Executing llm.generate action")

	// Get system message
	system, err := action.ProcessTemplate(step.System, actx, "llm.generate.system")
	if err != nil {
		return nil, goerr.Wrap(err, "failed to process system template")
	}

	// Get prompt
	var prompt string
	if step.Prompt != "" {
		prompt, err = action.ProcessTemplate(step.Prompt, actx, "llm.generate.prompt")
		if err != nil {
			return nil, goerr.Wrap(err, "failed to process prompt template")
		}
	} else if actx.Input != nil {
		prompt = fmt.Sprintf("%v", actx.Input)
	} else {
		return nil, goerr.New("no prompt specified for llm.generate action")
	}

	logger.Debug("Generating LLM content",
		slog.String("system", system),
		slog.Int("prompt_length", len(prompt)),
	)

	// Get LLM configuration - step args can override defaults
	llmConfig := actx.Config.Defaults.LLM
	provider := getStringArg(step.Args, "provider", llmConfig.Provider)
	model := getStringArg(step.Args, "model", llmConfig.Model)

	// Create LLM client based on provider
	var client gollem.LLMClient
	switch provider {
	case "openai":
		apiKey := getStringArg(step.Args, "api_key", llmConfig.APIKey)
		if apiKey == "" {
			apiKey = os.Getenv("OPENAI_API_KEY")
		}
		if apiKey == "" {
			return nil, goerr.New("OpenAI API key not found in config or OPENAI_API_KEY environment variable")
		}

		opts := []openai.Option{}
		if model != "" {
			opts = append(opts, openai.WithModel(model))
		}
		if temp := getFloatArg(step.Args, "temperature", llmConfig.Temperature); temp > 0 {
			opts = append(opts, openai.WithTemperature(float32(temp)))
		}
		if maxTokens := getIntArg(step.Args, "max_tokens", llmConfig.MaxTokens); maxTokens > 0 {
			opts = append(opts, openai.WithMaxTokens(maxTokens))
		}

		client, err = openai.New(ctx, apiKey, opts...)
		if err != nil {
			return nil, goerr.Wrap(err, "failed to create OpenAI client")
		}

	case "claude":
		apiKey := getStringArg(step.Args, "api_key", llmConfig.APIKey)
		if apiKey == "" {
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
		}
		if apiKey == "" {
			return nil, goerr.New("Anthropic API key not found in config or ANTHROPIC_API_KEY environment variable")
		}

		opts := []claude.Option{}
		if model != "" {
			opts = append(opts, claude.WithModel(model))
		}
		if temp := getFloatArg(step.Args, "temperature", llmConfig.Temperature); temp > 0 {
			opts = append(opts, claude.WithTemperature(temp))
		}
		if maxTokens := getIntArg(step.Args, "max_tokens", llmConfig.MaxTokens); maxTokens > 0 {
			opts = append(opts, claude.WithMaxTokens(int64(maxTokens)))
		}

		client, err = claude.New(ctx, apiKey, opts...)
		if err != nil {
			return nil, goerr.Wrap(err, "failed to create Claude client")
		}

	case "gemini":
		projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
		if projectID == "" {
			return nil, goerr.New("Google Cloud Project ID not found in GOOGLE_CLOUD_PROJECT environment variable")
		}
		location := os.Getenv("GOOGLE_CLOUD_LOCATION")
		if location == "" {
			location = "us-central1" // default location
		}

		opts := []gemini.Option{}
		if model != "" {
			opts = append(opts, gemini.WithModel(model))
		}
		if temp := getFloatArg(step.Args, "temperature", llmConfig.Temperature); temp > 0 {
			opts = append(opts, gemini.WithTemperature(float32(temp)))
		}
		if maxTokens := getIntArg(step.Args, "max_tokens", llmConfig.MaxTokens); maxTokens > 0 {
			opts = append(opts, gemini.WithMaxTokens(int32(maxTokens)))
		}

		client, err = gemini.New(ctx, projectID, location, opts...)
		if err != nil {
			return nil, goerr.Wrap(err, "failed to create Gemini client")
		}

	default:
		return nil, goerr.New("unsupported LLM provider", goerr.Value("provider", provider))
	}

	// Create session
	session, err := client.NewSession(ctx)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to create LLM session")
	}

	// Create prompt content
	fullPrompt := prompt
	if system != "" {
		fullPrompt = "System: " + system + "\n\n" + prompt
	}

	// Generate response
	result, err := session.GenerateContent(ctx, gollem.Text(fullPrompt))
	if err != nil {
		return nil, goerr.Wrap(err, "failed to generate LLM response")
	}

	response := ""
	if len(result.Texts) > 0 {
		response = result.Texts[0]
	}

	logger.Debug("LLM response generated",
		slog.Int("response_length", len(response)),
	)

	return &action.Result{
		Output: response,
		Metadata: map[string]interface{}{
			"provider":        provider,
			"model":           model,
			"system":          system,
			"prompt_length":   len(prompt),
			"response_length": len(response),
		},
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
