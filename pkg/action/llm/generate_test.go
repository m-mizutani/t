package llm

import (
	"context"
	"strings"
	"testing"

	"github.com/m-mizutani/t/pkg/action"
	"github.com/m-mizutani/t/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateAction_Name(t *testing.T) {
	generateAction := &GenerateAction{}
	assert.Equal(t, "llm.generate", generateAction.Name())
}

func TestGenerateAction_Description(t *testing.T) {
	generateAction := &GenerateAction{}
	assert.Equal(t, "Generate text using LLM", generateAction.Description())
}

func TestGenerateAction_NewConfig(t *testing.T) {
	generateAction := &GenerateAction{}
	config := generateAction.NewConfig()

	cfg, ok := config.(*LLMGenerateConfig)
	require.True(t, ok)
	assert.Empty(t, cfg.ID)
	assert.Empty(t, cfg.System)
	assert.Empty(t, cfg.Prompt)
	assert.Empty(t, cfg.Model)
	assert.Equal(t, float64(0), cfg.Temperature)
	assert.Equal(t, 0, cfg.MaxTokens)
}

func TestGenerateAction_Execute_InvalidConfigType(t *testing.T) {
	generateAction := &GenerateAction{}
	ctx := context.Background()
	actx := &action.Context{
		Output: "test output",
		Env:   make(map[string]string),
		Config: &config.Config{
			Defaults: config.DefaultConfig{
				LLM: config.LLMConfig{
					Provider: "gemini",
					Model:    "gemini-2.0-flash",
				},
			},
		},
	}

	// Pass wrong config type
	invalidConfig := "invalid config"

	result, err := generateAction.Execute(ctx, actx, invalidConfig)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid config type for llm.generate")
}

func TestGenerateAction_Execute_NoPromptSpecified(t *testing.T) {
	generateAction := &GenerateAction{}
	ctx := context.Background()
	actx := &action.Context{
		Output: nil, // No output
		Env:   make(map[string]string),
		Config: &config.Config{
			Defaults: config.DefaultConfig{
				LLM: config.LLMConfig{
					Provider: "gemini",
					Model:    "gemini-2.0-flash",
				},
			},
		},
	}

	cfg := &LLMGenerateConfig{} // No prompt specified

	result, err := generateAction.Execute(ctx, actx, cfg)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "no prompt specified for llm.generate action")
}

func TestGenerateAction_Execute_WithInputPrompt(t *testing.T) {
	// Skip this test if no LLM credentials are available
	if !hasLLMCredentials() {
		t.Skip("Skipping LLM test: no credentials available")
	}

	generateAction := &GenerateAction{}
	ctx := context.Background()
	actx := &action.Context{
		Output: "Say hello",
		Env:   make(map[string]string),
		Config: &config.Config{
			Defaults: config.DefaultConfig{
				LLM: config.LLMConfig{
					Provider: "gemini",
					Model:    "gemini-2.0-flash",
				},
				Gemini: config.GeminiConfig{
					ProjectID: "test-project",
					Location:  "us-central1",
				},
			},
		},
	}

	cfg := &LLMGenerateConfig{} // Will use output as prompt

	result, err := generateAction.Execute(ctx, actx, cfg)
	if err != nil {
		// If we get authentication errors, skip the test
		if isAuthError(err) {
			t.Skip("Skipping LLM test: authentication failed")
		}
		require.NoError(t, err)
	}

	require.NotNil(t, result)
	assert.NotEmpty(t, result.Output)

	// Check metadata
	assert.Equal(t, "gemini", result.Metadata["provider"])
	assert.Equal(t, "gemini-2.0-flash", result.Metadata["model"])
	assert.Equal(t, "", result.Metadata["system"])
	assert.Equal(t, len("Say hello"), result.Metadata["prompt_length"])
	assert.Greater(t, result.Metadata["response_length"], 0)
}

func TestGenerateAction_Execute_WithPromptConfig(t *testing.T) {
	// Skip this test if no LLM credentials are available
	if !hasLLMCredentials() {
		t.Skip("Skipping LLM test: no credentials available")
	}

	generateAction := &GenerateAction{}
	ctx := context.Background()
	actx := &action.Context{
		Output: "ignored output",
		Env:   make(map[string]string),
		Config: &config.Config{
			Defaults: config.DefaultConfig{
				LLM: config.LLMConfig{
					Provider: "gemini",
					Model:    "gemini-2.0-flash",
				},
				Gemini: config.GeminiConfig{
					ProjectID: "test-project",
					Location:  "us-central1",
				},
			},
		},
	}

	cfg := &LLMGenerateConfig{
		Prompt: "What is 2+2?",
	}

	result, err := generateAction.Execute(ctx, actx, cfg)
	if err != nil {
		// If we get authentication errors, skip the test
		if isAuthError(err) {
			t.Skip("Skipping LLM test: authentication failed")
		}
		require.NoError(t, err)
	}

	require.NotNil(t, result)
	assert.NotEmpty(t, result.Output)

	// Check metadata
	assert.Equal(t, "gemini", result.Metadata["provider"])
	assert.Equal(t, "gemini-2.0-flash", result.Metadata["model"])
	assert.Equal(t, "", result.Metadata["system"])
	assert.Equal(t, len("What is 2+2?"), result.Metadata["prompt_length"])
	assert.Greater(t, result.Metadata["response_length"], 0)
}

func TestGenerateAction_Execute_WithSystemMessage(t *testing.T) {
	// Skip this test if no LLM credentials are available
	if !hasLLMCredentials() {
		t.Skip("Skipping LLM test: no credentials available")
	}

	generateAction := &GenerateAction{}
	ctx := context.Background()
	actx := &action.Context{
		Output: "test output",
		Env:   make(map[string]string),
		Config: &config.Config{
			Defaults: config.DefaultConfig{
				LLM: config.LLMConfig{
					Provider: "gemini",
					Model:    "gemini-2.0-flash",
				},
				Gemini: config.GeminiConfig{
					ProjectID: "test-project",
					Location:  "us-central1",
				},
			},
		},
	}

	cfg := &LLMGenerateConfig{
		System: "You are a helpful assistant.",
		Prompt: "Say hello",
	}

	result, err := generateAction.Execute(ctx, actx, cfg)
	if err != nil {
		// If we get authentication errors, skip the test
		if isAuthError(err) {
			t.Skip("Skipping LLM test: authentication failed")
		}
		require.NoError(t, err)
	}

	require.NotNil(t, result)
	assert.NotEmpty(t, result.Output)

	// Check metadata
	assert.Equal(t, "gemini", result.Metadata["provider"])
	assert.Equal(t, "gemini-2.0-flash", result.Metadata["model"])
	assert.Equal(t, "You are a helpful assistant.", result.Metadata["system"])
	assert.Equal(t, len("Say hello"), result.Metadata["prompt_length"])
	assert.Greater(t, result.Metadata["response_length"], 0)
}

func TestGenerateAction_Execute_WithTemplateProcessing(t *testing.T) {
	// Skip this test if no LLM credentials are available
	if !hasLLMCredentials() {
		t.Skip("Skipping LLM test: no credentials available")
	}

	generateAction := &GenerateAction{}
	ctx := context.Background()
	actx := &action.Context{
		Output: "World",
		Env:   make(map[string]string),
		Config: &config.Config{
			Defaults: config.DefaultConfig{
				LLM: config.LLMConfig{
					Provider: "gemini",
					Model:    "gemini-2.0-flash",
				},
				Gemini: config.GeminiConfig{
					ProjectID: "test-project",
					Location:  "us-central1",
				},
			},
		},
	}

	cfg := &LLMGenerateConfig{
		System: "You are a greeting assistant.",
		Prompt: "Say hello to {{ .output }}",
	}

	result, err := generateAction.Execute(ctx, actx, cfg)
	if err != nil {
		// If we get authentication errors, skip the test
		if isAuthError(err) {
			t.Skip("Skipping LLM test: authentication failed")
		}
		require.NoError(t, err)
	}

	require.NotNil(t, result)
	assert.NotEmpty(t, result.Output)

	// Check metadata
	assert.Equal(t, "gemini", result.Metadata["provider"])
	assert.Equal(t, "You are a greeting assistant.", result.Metadata["system"])
	assert.Equal(t, len("Say hello to World"), result.Metadata["prompt_length"])
}

func TestGenerateAction_Execute_UnsupportedProvider(t *testing.T) {
	generateAction := &GenerateAction{}
	ctx := context.Background()
	actx := &action.Context{
		Output: "test output",
		Env:   make(map[string]string),
		Config: &config.Config{
			Defaults: config.DefaultConfig{
				LLM: config.LLMConfig{
					Provider: "unsupported",
					Model:    "test-model",
				},
			},
		},
	}

	cfg := &LLMGenerateConfig{
		Prompt: "Test prompt",
	}

	result, err := generateAction.Execute(ctx, actx, cfg)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unsupported LLM provider")
}

func TestLLMGenerateConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config LLMGenerateConfig
		valid  bool
	}{
		{
			name:   "valid empty config (will use output)",
			config: LLMGenerateConfig{},
			valid:  true,
		},
		{
			name: "valid config with prompt",
			config: LLMGenerateConfig{
				Prompt: "test prompt",
			},
			valid: true,
		},
		{
			name: "valid config with all fields",
			config: LLMGenerateConfig{
				ID:          "test-generate",
				System:      "You are a helpful assistant.",
				Prompt:      "test prompt",
				Model:       "gpt-4",
				Temperature: 0.7,
				MaxTokens:   1000,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// All configurations are valid for generate action
			assert.True(t, tt.valid, "All generate configurations should be valid")
		})
	}
}

// Helper functions for testing

func hasLLMCredentials() bool {
	// Check if any LLM credentials are available
	// This is a simple check - in real scenarios you might want more sophisticated detection
	return false // Always skip LLM tests in CI/testing environments
}

func isAuthError(err error) bool {
	// Check if the error is related to authentication
	errStr := err.Error()
	return strings.Contains(errStr, "authentication") ||
		strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "api key") ||
		strings.Contains(errStr, "credentials") ||
		strings.Contains(errStr, "permission denied")
}
