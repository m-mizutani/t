package action

import (
	"context"
	"log/slog"
	"os"
	"reflect"
	"strings"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/t/pkg/config"
	"github.com/m-mizutani/t/pkg/logger"
	"gopkg.in/yaml.v3"
)

// Context provides execution context for actions
type Context struct {
	Config        *config.Config
	Args          []string
	Env           map[string]string
	Output        interface{}
	Data          map[string]interface{}
	ActionOutputs map[string]interface{} // Store outputs by action ID
}

// Result represents the result of an action execution
type Result struct {
	Output   interface{}
	Metadata map[string]interface{}
}

// TypedAction defines the new interface for typed actions
type TypedAction interface {
	Name() string
	Description() string
	// NewConfig returns a new instance of the action's configuration struct
	NewConfig() interface{}
	// Execute runs the action with typed configuration
	Execute(ctx context.Context, actx *Context, config interface{}) (*Result, error)
}

// Validator provides simple validation functionality
type Validator struct{}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// Validate validates a struct using reflection and validate tags
func (v *Validator) Validate(s interface{}) error {
	rv := reflect.ValueOf(s)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return nil // Only validate structs
	}

	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rt.Field(i)

		// Check validate tag
		validateTag := fieldType.Tag.Get("validate")
		if validateTag == "" {
			continue
		}

		// Parse validation rules
		rules := strings.Split(validateTag, ",")
		for _, rule := range rules {
			rule = strings.TrimSpace(rule)

			switch rule {
			case "required":
				if err := v.validateRequired(field, fieldType.Name); err != nil {
					return err
				}
			default:
				// Ignore unknown validation rules for now
			}
		}
	}

	return nil
}

// validateRequired checks if a field is not empty
func (v *Validator) validateRequired(field reflect.Value, fieldName string) error {
	switch field.Kind() {
	case reflect.String:
		if field.String() == "" {
			return goerr.New("field is required", goerr.Value("field", fieldName))
		}
	case reflect.Slice, reflect.Array, reflect.Map, reflect.Chan:
		if field.Len() == 0 {
			return goerr.New("field is required", goerr.Value("field", fieldName))
		}
	case reflect.Ptr, reflect.Interface:
		if field.IsNil() {
			return goerr.New("field is required", goerr.Value("field", fieldName))
		}
	}
	return nil
}

// Registry manages action registration and lookup
type Registry struct {
	typedActions map[string]TypedAction
	validator    *Validator
}

// NewRegistry creates a new action registry
func NewRegistry() *Registry {
	return &Registry{
		typedActions: make(map[string]TypedAction),
		validator:    NewValidator(),
	}
}

// RegisterTyped registers a typed action
func (r *Registry) RegisterTyped(action TypedAction) {
	r.typedActions[action.Name()] = action
}

// GetTyped retrieves a typed action by name
func (r *Registry) GetTyped(name string) (TypedAction, bool) {
	action, exists := r.typedActions[name]
	return action, exists
}

// ParseStepConfig converts raw step configuration to typed configuration
func (r *Registry) ParseStepConfig(raw config.RawStepConfig) (*config.StepConfig, error) {
	// Check if this is a typed action
	typedAction, exists := r.typedActions[raw.Action]
	if !exists {
		return nil, goerr.New("unknown action", goerr.Value("action", raw.Action))
	}

	// Create action-specific config struct
	configStruct := typedAction.NewConfig()

	// Marshal raw data to YAML then unmarshal to typed struct
	data, err := yaml.Marshal(raw.Raw)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to marshal raw config")
	}

	if err := yaml.Unmarshal(data, configStruct); err != nil {
		return nil, goerr.Wrap(err, "failed to unmarshal to typed config")
	}

	// Validate the configuration
	if err := r.validator.Validate(configStruct); err != nil {
		return nil, goerr.Wrap(err, "validation failed", goerr.Value("action", raw.Action))
	}

	return &config.StepConfig{
		Action: raw.Action,
		Config: configStruct,
	}, nil
}

// List returns all registered action names
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.typedActions))
	for name := range r.typedActions {
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
