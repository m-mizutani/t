package llm

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/gollem"
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
	var system string
	var err error
	if systemStr := getStringArg(step.Args, "system", ""); systemStr != "" {
		system, err = action.ProcessTemplate(systemStr, actx, "llm.generate.system")
		if err != nil {
			return nil, goerr.Wrap(err, "failed to process system template from args")
		}
	} else if step.System != "" {
		system, err = action.ProcessTemplate(step.System, actx, "llm.generate.system")
		if err != nil {
			return nil, goerr.Wrap(err, "failed to process system template")
		}
	}

	// Get prompt - check args.prompt first, then step.Prompt
	var prompt string
	if promptStr := getStringArg(step.Args, "prompt", ""); promptStr != "" {
		prompt, err = action.ProcessTemplate(promptStr, actx, "llm.generate.prompt")
		if err != nil {
			return nil, goerr.Wrap(err, "failed to process prompt template from args")
		}
	} else if step.Prompt != "" {
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

	// Create LLM client based on provider
	clientInfo, err := createLLMClient(ctx, actx, step)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to create LLM client")
	}

	// Create session
	session, err := clientInfo.Client.NewSession(ctx)
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
			"provider":        clientInfo.Provider,
			"model":           clientInfo.Model,
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
