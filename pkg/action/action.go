package action

import (
	"context"
	"log/slog"
	"os"

	"github.com/m-mizutani/t/pkg/config"
	"github.com/m-mizutani/t/pkg/logger"
)

// Context provides execution context for actions
type Context struct {
	Config        *config.Config
	Args          []string
	Env           map[string]string
	Input         interface{}
	Data          map[string]interface{}
	ActionOutputs map[string]interface{} // Store outputs by action ID
}

// Result represents the result of an action execution
type Result struct {
	Output   interface{}
	Metadata map[string]interface{}
}

// Action defines the interface for all actions
type Action interface {
	Name() string
	Description() string
	Execute(ctx context.Context, actx *Context, step config.StepConfig) (*Result, error)
}

// Registry manages action registration and lookup
type Registry struct {
	actions map[string]Action
}

// NewRegistry creates a new action registry
func NewRegistry() *Registry {
	return &Registry{
		actions: make(map[string]Action),
	}
}

// Register registers an action
func (r *Registry) Register(action Action) {
	r.actions[action.Name()] = action
}

// Get retrieves an action by name
func (r *Registry) Get(name string) (Action, bool) {
	action, exists := r.actions[name]
	return action, exists
}

// List returns all registered action names
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.actions))
	for name := range r.actions {
		names = append(names, name)
	}
	return names
}

// NewContext creates a new action context
func NewContext(cfg *config.Config, args []string) *Context {
	// Get environment variables
	env := make(map[string]string)
	for _, e := range os.Environ() {
		if len(e) > 0 {
			for i := 0; i < len(e); i++ {
				if e[i] == '=' {
					env[e[:i]] = e[i+1:]
					break
				}
			}
		}
	}

	return &Context{
		Config:        cfg,
		Args:          args,
		Env:           env,
		Data:          make(map[string]interface{}),
		ActionOutputs: make(map[string]interface{}),
	}
}

// Logger returns the logger from context
func (c *Context) Logger(ctx context.Context) *slog.Logger {
	return logger.FromContext(ctx)
}
