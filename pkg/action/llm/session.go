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
		clientInfo, err := createLLMClient(ctx, actx, step)
		if err != nil {
			return nil, goerr.Wrap(err, "failed to create LLM client")
		}

		// Create agent for this session
		agent = gollem.New(clientInfo.Client)
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
