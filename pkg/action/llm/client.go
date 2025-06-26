package llm

import (
	"context"
	"os"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/gollem"
	"github.com/m-mizutani/gollem/llm/claude"
	"github.com/m-mizutani/gollem/llm/gemini"
	"github.com/m-mizutani/gollem/llm/openai"
	"github.com/m-mizutani/t/pkg/action"
	"github.com/m-mizutani/t/pkg/config"
)

// ClientInfo holds information about the created LLM client
type ClientInfo struct {
	Client   gollem.LLMClient
	Provider string
	Model    string
}

// createOpenAIClient creates an OpenAI client for typed configuration
func createOpenAIClient(ctx context.Context, actx *action.Context, cfg *LLMGenerateConfig, llmConfig config.LLMConfig, model string) (gollem.LLMClient, error) {
	// Prioritize environment variable over config file
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = actx.Config.Defaults.OpenAI.APIKey
	}
	if apiKey == "" {
		return nil, goerr.New("OpenAI API key not found in OPENAI_API_KEY environment variable or config")
	}

	opts := []openai.Option{}
	if model != "" {
		opts = append(opts, openai.WithModel(model))
	}
	if cfg.Temperature > 0 {
		opts = append(opts, openai.WithTemperature(float32(cfg.Temperature)))
	} else if llmConfig.Temperature > 0 {
		opts = append(opts, openai.WithTemperature(float32(llmConfig.Temperature)))
	}
	if cfg.MaxTokens > 0 {
		opts = append(opts, openai.WithMaxTokens(cfg.MaxTokens))
	} else if llmConfig.MaxTokens > 0 {
		opts = append(opts, openai.WithMaxTokens(llmConfig.MaxTokens))
	}

	client, err := openai.New(ctx, apiKey, opts...)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to create OpenAI client")
	}

	return client, nil
}

// createClaudeClient creates a Claude client for typed configuration
func createClaudeClient(ctx context.Context, actx *action.Context, cfg *LLMGenerateConfig, llmConfig config.LLMConfig, model string) (gollem.LLMClient, error) {
	// Prioritize environment variable over config file
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		apiKey = actx.Config.Defaults.Claude.APIKey
	}
	if apiKey == "" {
		return nil, goerr.New("Anthropic API key not found in ANTHROPIC_API_KEY environment variable or config")
	}

	opts := []claude.Option{}
	if model != "" {
		opts = append(opts, claude.WithModel(model))
	}
	if cfg.Temperature > 0 {
		opts = append(opts, claude.WithTemperature(cfg.Temperature))
	} else if llmConfig.Temperature > 0 {
		opts = append(opts, claude.WithTemperature(llmConfig.Temperature))
	}
	if cfg.MaxTokens > 0 {
		opts = append(opts, claude.WithMaxTokens(int64(cfg.MaxTokens)))
	} else if llmConfig.MaxTokens > 0 {
		opts = append(opts, claude.WithMaxTokens(int64(llmConfig.MaxTokens)))
	}

	client, err := claude.New(ctx, apiKey, opts...)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to create Claude client")
	}

	return client, nil
}

// createGeminiClient creates a Gemini client for typed configuration
func createGeminiClient(ctx context.Context, actx *action.Context, cfg *LLMGenerateConfig, llmConfig config.LLMConfig, model string) (gollem.LLMClient, error) {
	// Prioritize environment variable over config file
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID = actx.Config.Defaults.Gemini.ProjectID
	}
	if projectID == "" {
		return nil, goerr.New("Google Cloud Project ID not found in GOOGLE_CLOUD_PROJECT environment variable or config")
	}

	// Prioritize environment variable over config file
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")
	if location == "" {
		location = actx.Config.Defaults.Gemini.Location
	}
	if location == "" {
		location = "us-central1" // default location
	}

	opts := []gemini.Option{}
	if model != "" {
		opts = append(opts, gemini.WithModel(model))
	}
	if cfg.Temperature > 0 {
		opts = append(opts, gemini.WithTemperature(float32(cfg.Temperature)))
	} else if llmConfig.Temperature > 0 {
		opts = append(opts, gemini.WithTemperature(float32(llmConfig.Temperature)))
	}
	if cfg.MaxTokens > 0 {
		opts = append(opts, gemini.WithMaxTokens(int32(cfg.MaxTokens)))
	} else if llmConfig.MaxTokens > 0 {
		opts = append(opts, gemini.WithMaxTokens(int32(llmConfig.MaxTokens)))
	}

	client, err := gemini.New(ctx, projectID, location, opts...)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to create Gemini client")
	}

	return client, nil
}
