package action

import (
	"strconv"
	"strings"
	"text/template"

	"github.com/m-mizutani/goerr/v2"
)

// OutputMap provides structured access to outputs with map-like behavior
type OutputMap map[string]interface{}

// String returns the current output as string (for {{ .output }})
func (o OutputMap) String() string {
	if current, exists := o[""]; exists && current != nil {
		if str, ok := current.(string); ok {
			return strings.TrimSpace(str)
		}
	}
	return ""
}

// ArgsMap provides structured access to arguments
type ArgsMap map[string]string

// ProcessTemplate processes a template string with structured dot notation
func ProcessTemplate(templateStr string, actx *Context, templateName string) (string, error) {
	if templateStr == "" {
		return "", nil
	}

	// Create output map for structured access
	outputMap := make(OutputMap)

	// Add current output with empty key (for {{ .output }})
	outputMap[""] = actx.Input

	// Add all action outputs (for {{ .output.action_id }})
	for actionID, output := range actx.ActionOutputs {
		outputMap[actionID] = output
	}

	// Create args map for structured access (for {{ index .args "0" }})
	argsMap := make(ArgsMap)
	for i, arg := range actx.Args {
		argsMap[strconv.Itoa(i)] = arg
	}

	// Create data map for template execution
	data := map[string]interface{}{
		// Structured access
		"output": outputMap,
		"args":   argsMap,

		// Input support
		"input": actx.Input,
	}

	// Add simplified args access: .arg0, .arg1, .arg2, etc.
	for i, arg := range actx.Args {
		data["arg"+strconv.Itoa(i)] = arg
	}

	// Add environment variables with dot notation support
	envData := make(map[string]string)
	for key, value := range actx.Env {
		envData[key] = value
	}
	data["env"] = envData

	// Add metadata with dot notation support
	metaData := make(map[string]interface{})
	for key, value := range actx.Data {
		metaData[key] = value
	}
	data["data"] = metaData

	// Create template with minimal custom functions (keeping only index for backward compatibility)
	tmpl := template.New(templateName).Funcs(template.FuncMap{
		"index": func(m interface{}, key interface{}) interface{} {
			switch mapVal := m.(type) {
			case map[string]string:
				if keyStr, ok := key.(string); ok {
					return mapVal[keyStr]
				}
				if keyInt, ok := key.(int); ok {
					return mapVal[strconv.Itoa(keyInt)]
				}
			case []string:
				if keyInt, ok := key.(int); ok && keyInt >= 0 && keyInt < len(mapVal) {
					return mapVal[keyInt]
				}
			case ArgsMap:
				if keyStr, ok := key.(string); ok {
					return mapVal[keyStr]
				}
				if keyInt, ok := key.(int); ok {
					return mapVal[strconv.Itoa(keyInt)]
				}
			}
			return ""
		},
	})

	// Parse and execute template
	tmpl, err := tmpl.Parse(templateStr)
	if err != nil {
		return "", goerr.Wrap(err, "failed to parse template")
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, data); err != nil {
		return "", goerr.Wrap(err, "failed to execute template")
	}

	return result.String(), nil
}
