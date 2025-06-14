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

// SessionAction implements llm.session action
type SessionAction struct {
	agents map[string]*gollem.Agent // Store agents by session ID
}

// Name returns the action name
func (a *SessionAction) Name() string {
	return "llm.session"
}

// Description returns a human-readable description
func (a *SessionAction) Description() string {
	return "Send a prompt to an LLM session with conversation history"
}

// Execute runs the llm.session action
func (a *SessionAction) Execute(ctx context.Context, actx *action.Context, step config.StepConfig) (*action.Result, error) {
	logger := actx.Logger(ctx)
	logger.Debug("Executing llm.session action")

	// Initialize sessions map if not done
	if a.agents == nil {
		a.agents = make(map[string]*gollem.Agent)
	}

	// Get session ID (optional, defaults to "default")
	sessionID, err := a.getArgValue(step.Args, "session_id", actx)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get session ID")
	}
	if sessionID == "" {
		sessionID = "default"
	}

	// Get system message
	system, err := action.ProcessTemplate(step.System, actx, "llm.session.system")
	if err != nil {
		return nil, goerr.Wrap(err, "failed to process system template")
	}

	// Get prompt
	var prompt string
	if step.Prompt != "" {
		prompt, err = action.ProcessTemplate(step.Prompt, actx, "llm.session.prompt")
		if err != nil {
			return nil, goerr.Wrap(err, "failed to process prompt template")
		}
	} else if actx.Input != nil {
		prompt = fmt.Sprintf("%v", actx.Input)
	} else {
		return nil, goerr.New("no prompt specified for llm.session action")
	}

	logger.Debug("Sending prompt to LLM session",
		slog.String("session_id", sessionID),
		slog.String("system", system),
		slog.Int("prompt_length", len(prompt)),
	)

	// Get LLM configuration
	llmConfig := actx.Config.Defaults.LLM
	provider := getStringArg(step.Args, "provider", llmConfig.Provider)
	model := getStringArg(step.Args, "model", llmConfig.Model)

	// Get or create agent for this session
	agent, exists := a.agents[sessionID]
	if !exists {
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

		// Create agent for this session
		agent = gollem.New(client)
		a.agents[sessionID] = agent
	}

	// Create full prompt with system message if provided
	fullPrompt := prompt
	if system != "" {
		fullPrompt = "System: " + system + "\n\n" + prompt
	}

	// Generate response with history
	history, err := agent.Prompt(ctx, fullPrompt)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to generate LLM response with session")
	}

	// Extract response text
	response := ""
	if history != nil {
		// Try to get the last response
		response = fmt.Sprintf("%v", history)
	}

	logger.Debug("LLM session response received",
		slog.String("session_id", sessionID),
		slog.Int("response_length", len(response)),
	)

	return &action.Result{
		Output: response,
		Metadata: map[string]interface{}{
			"session_id":      sessionID,
			"provider":        provider,
			"model":           model,
			"system":          system,
			"prompt_length":   len(prompt),
			"response_length": len(response),
		},
	}, nil
}

// getArgValue extracts and processes template for argument value
func (a *SessionAction) getArgValue(args map[string]interface{}, key string, actx *action.Context) (string, error) {
	value, exists := args[key]
	if !exists {
		return "", nil
	}

	valueStr, ok := value.(string)
	if !ok {
		return "", goerr.New("argument must be string", goerr.Value("key", key), goerr.Value("value", value))
	}

	return action.ProcessTemplate(valueStr, actx, "llm.session."+key)
}
