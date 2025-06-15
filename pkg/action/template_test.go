package action

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessTemplate_EmptyTemplate(t *testing.T) {
	actx := &Context{
		Args: []string{"arg1", "arg2"},
		Env:  map[string]string{"TEST_VAR": "test_value"},
		Data: map[string]interface{}{"key": "value"},
	}

	result, err := ProcessTemplate("", actx, "test")
	require.NoError(t, err)
	assert.Equal(t, "", result)
}

func TestProcessTemplate_SimpleOutput(t *testing.T) {
	actx := &Context{
		Input: "hello world",
		Args:  []string{},
		Env:   map[string]string{},
		Data:  map[string]interface{}{},
	}

	result, err := ProcessTemplate("{{ .output }}", actx, "test")
	require.NoError(t, err)
	assert.Equal(t, "hello world", result)
}

func TestProcessTemplate_OutputWithSpaces(t *testing.T) {
	actx := &Context{
		Input: "  hello world  \n",
		Args:  []string{},
		Env:   map[string]string{},
		Data:  map[string]interface{}{},
	}

	result, err := ProcessTemplate("{{ .output }}", actx, "test")
	require.NoError(t, err)
	assert.Equal(t, "hello world", result) // Should be trimmed
}

func TestProcessTemplate_Input(t *testing.T) {
	actx := &Context{
		Input: "input_value",
		Args:  []string{},
		Env:   map[string]string{},
		Data:  map[string]interface{}{},
	}

	result, err := ProcessTemplate("{{ .input }}", actx, "test")
	require.NoError(t, err)
	assert.Equal(t, "input_value", result)
}

func TestProcessTemplate_Args(t *testing.T) {
	actx := &Context{
		Args: []string{"first", "second", "third"},
		Env:  map[string]string{},
		Data: map[string]interface{}{},
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "arg0",
			template: "{{ .arg0 }}",
			expected: "first",
		},
		{
			name:     "arg1",
			template: "{{ .arg1 }}",
			expected: "second",
		},
		{
			name:     "arg2",
			template: "{{ .arg2 }}",
			expected: "third",
		},
		{
			name:     "multiple args",
			template: "{{ .arg0 }} {{ .arg1 }} {{ .arg2 }}",
			expected: "first second third",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ProcessTemplate(tt.template, actx, "test")
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProcessTemplate_Environment(t *testing.T) {
	actx := &Context{
		Args: []string{},
		Env: map[string]string{
			"TEST_VAR":    "test_value",
			"ANOTHER_VAR": "another_value",
			"EMPTY_VAR":   "",
		},
		Data: map[string]interface{}{},
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "existing env var",
			template: "{{ .env.TEST_VAR }}",
			expected: "test_value",
		},
		{
			name:     "another env var",
			template: "{{ .env.ANOTHER_VAR }}",
			expected: "another_value",
		},
		{
			name:     "empty env var",
			template: "{{ .env.EMPTY_VAR }}",
			expected: "",
		},
		{
			name:     "multiple env vars",
			template: "{{ .env.TEST_VAR }} {{ .env.ANOTHER_VAR }}",
			expected: "test_value another_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ProcessTemplate(tt.template, actx, "test")
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProcessTemplate_Data(t *testing.T) {
	actx := &Context{
		Args: []string{},
		Env:  map[string]string{},
		Data: map[string]interface{}{
			"string_key": "string_value",
			"int_key":    42,
			"bool_key":   true,
			"float_key":  3.14,
		},
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "string data",
			template: "{{ .data.string_key }}",
			expected: "string_value",
		},
		{
			name:     "int data",
			template: "{{ .data.int_key }}",
			expected: "42",
		},
		{
			name:     "bool data",
			template: "{{ .data.bool_key }}",
			expected: "true",
		},
		{
			name:     "float data",
			template: "{{ .data.float_key }}",
			expected: "3.14",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ProcessTemplate(tt.template, actx, "test")
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProcessTemplate_ActionOutputs(t *testing.T) {
	actx := &Context{
		Input: "current_output",
		Args:  []string{},
		Env:   map[string]string{},
		Data:  map[string]interface{}{},
		ActionOutputs: map[string]interface{}{
			"action1": "output1",
			"action2": "output2",
			"action3": 123,
		},
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "current output",
			template: "{{ .output }}",
			expected: "current_output",
		},
		{
			name:     "specific action output",
			template: "{{ .output.action1 }}",
			expected: "output1",
		},
		{
			name:     "another action output",
			template: "{{ .output.action2 }}",
			expected: "output2",
		},
		{
			name:     "numeric action output",
			template: "{{ .output.action3 }}",
			expected: "123",
		},
		{
			name:     "multiple outputs",
			template: "{{ .output.action1 }} {{ .output.action2 }}",
			expected: "output1 output2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ProcessTemplate(tt.template, actx, "test")
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProcessTemplate_ComplexTemplate(t *testing.T) {
	actx := &Context{
		Input: "input_data",
		Args:  []string{"arg1", "arg2"},
		Env: map[string]string{
			"ENV_VAR": "env_value",
		},
		Data: map[string]interface{}{
			"meta_key": "meta_value",
		},
		ActionOutputs: map[string]interface{}{
			"previous_action": "previous_output",
		},
	}

	template := `Input: {{ .input }}
Args: {{ .arg0 }}, {{ .arg1 }}
Env: {{ .env.ENV_VAR }}
Data: {{ .data.meta_key }}
Previous: {{ .output.previous_action }}
Current: {{ .output }}`

	expected := `Input: input_data
Args: arg1, arg2
Env: env_value
Data: meta_value
Previous: previous_output
Current: input_data`

	result, err := ProcessTemplate(template, actx, "test")
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestProcessTemplate_MissingValues(t *testing.T) {
	actx := &Context{
		Args: []string{"arg1"},
		Env:  map[string]string{},
		Data: map[string]interface{}{},
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "missing arg",
			template: "{{ .arg5 }}",
			expected: "<no value>",
		},
		{
			name:     "missing env var",
			template: "{{ .env.NONEXISTENT }}",
			expected: "<no value>",
		},
		{
			name:     "missing data key",
			template: "{{ .data.nonexistent }}",
			expected: "<no value>",
		},
		{
			name:     "missing action output",
			template: "{{ .output.nonexistent }}",
			expected: "<no value>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ProcessTemplate(tt.template, actx, "test")
			require.NoError(t, err)
			assert.Contains(t, result, "<no value>")
		})
	}
}

func TestProcessTemplate_SpecialCharacters(t *testing.T) {
	actx := &Context{
		Input: "hello\nworld\ttab",
		Args:  []string{"arg with spaces", "arg\"with\"quotes", "arg'with'single"},
		Env: map[string]string{
			"SPECIAL_CHARS": "value\nwith\nnewlines",
		},
		Data: map[string]interface{}{
			"special": "data\twith\ttabs",
		},
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "input with newlines",
			template: "{{ .input }}",
			expected: "hello\nworld\ttab",
		},
		{
			name:     "arg with spaces",
			template: "{{ .arg0 }}",
			expected: "arg with spaces",
		},
		{
			name:     "arg with quotes",
			template: "{{ .arg1 }}",
			expected: "arg\"with\"quotes",
		},
		{
			name:     "env with newlines",
			template: "{{ .env.SPECIAL_CHARS }}",
			expected: "value\nwith\nnewlines",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ProcessTemplate(tt.template, actx, "test")
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProcessTemplate_InvalidTemplate(t *testing.T) {
	actx := &Context{
		Args: []string{},
		Env:  map[string]string{},
		Data: map[string]interface{}{},
	}

	tests := []struct {
		name     string
		template string
	}{
		{
			name:     "unclosed template",
			template: "{{ .arg0",
		},
		{
			name:     "invalid syntax",
			template: "{{ .arg0 ] }}",
		},
		{
			name:     "invalid function call",
			template: "{{ nonexistent_func }}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ProcessTemplate(tt.template, actx, "test")
			assert.Error(t, err)
		})
	}
}

func TestProcessTemplate_OutputMap(t *testing.T) {
	// Test OutputMap String method
	outputMap := OutputMap{
		"":        "current output",
		"action1": "action1 output",
	}

	assert.Equal(t, "current output", outputMap.String())

	// Test with trimming
	outputMap[""] = "  output with spaces  \n"
	assert.Equal(t, "output with spaces", outputMap.String())

	// Test with non-string value
	outputMap[""] = 123
	assert.Equal(t, "", outputMap.String())

	// Test with nil value
	outputMap[""] = nil
	assert.Equal(t, "", outputMap.String())

	// Test with no current output
	delete(outputMap, "")
	assert.Equal(t, "", outputMap.String())
}

func TestProcessTemplate_DifferentTypes(t *testing.T) {
	actx := &Context{
		Input: 42,
		Args:  []string{},
		Env:   map[string]string{},
		Data: map[string]interface{}{
			"bool_val":  true,
			"float_val": 3.14159,
			"slice_val": []string{"a", "b", "c"},
			"map_val":   map[string]interface{}{"nested": "value"},
		},
		ActionOutputs: map[string]interface{}{
			"int_output":    123,
			"bool_output":   false,
			"string_output": "test",
		},
	}

	tests := []struct {
		name     string
		template string
		expected string
	}{
		{
			name:     "int input",
			template: "{{ .input }}",
			expected: "42",
		},
		{
			name:     "bool data",
			template: "{{ .data.bool_val }}",
			expected: "true",
		},
		{
			name:     "float data",
			template: "{{ .data.float_val }}",
			expected: "3.14159",
		},
		{
			name:     "int action output",
			template: "{{ .output.int_output }}",
			expected: "123",
		},
		{
			name:     "bool action output",
			template: "{{ .output.bool_output }}",
			expected: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ProcessTemplate(tt.template, actx, "test")
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestProcessTemplate_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		actx     *Context
		template string
		expected string
	}{
		{
			name: "nil input",
			actx: &Context{
				Input: nil,
				Args:  []string{},
				Env:   map[string]string{},
				Data:  map[string]interface{}{},
			},
			template: "{{ .input }}",
			expected: "<no value>",
		},
		{
			name: "empty args",
			actx: &Context{
				Args: []string{},
				Env:  map[string]string{},
				Data: map[string]interface{}{},
			},
			template: "{{ .arg0 }}",
			expected: "<no value>",
		},
		{
			name: "nil env map",
			actx: &Context{
				Args: []string{},
				Env:  nil,
				Data: map[string]interface{}{},
			},
			template: "{{ .env.TEST }}",
			expected: "<no value>",
		},
		{
			name: "nil data map",
			actx: &Context{
				Args: []string{},
				Env:  map[string]string{},
				Data: nil,
			},
			template: "{{ .data.test }}",
			expected: "<no value>",
		},
		{
			name: "nil action outputs",
			actx: &Context{
				Args:          []string{},
				Env:           map[string]string{},
				Data:          map[string]interface{}{},
				ActionOutputs: nil,
			},
			template: "{{ .output.test }}",
			expected: "<no value>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ProcessTemplate(tt.template, tt.actx, "test")
			require.NoError(t, err)
			assert.Contains(t, result, "<no value>")
		})
	}
}

func TestProcessTemplate_Performance(t *testing.T) {
	// Test with large data sets
	actx := &Context{
		Input:         "performance test input",
		Args:          make([]string, 100),
		Env:           make(map[string]string),
		Data:          make(map[string]interface{}),
		ActionOutputs: make(map[string]interface{}),
	}

	// Fill with test data
	for i := 0; i < 100; i++ {
		actx.Args[i] = "arg" + string(rune(i))
		actx.Env["ENV_VAR_"+string(rune(i))] = "env_value_" + string(rune(i))
		actx.Data["data_key_"+string(rune(i))] = "data_value_" + string(rune(i))
		actx.ActionOutputs["action_"+string(rune(i))] = "output_" + string(rune(i))
	}

	template := "{{ .input }} {{ .arg0 }} {{ .env.ENV_VAR_0 }} {{ .data.data_key_0 }} {{ .output.action_0 }}"

	// Run multiple times to test performance
	for i := 0; i < 100; i++ {
		result, err := ProcessTemplate(template, actx, "performance_test")
		require.NoError(t, err)
		assert.Contains(t, result, "performance test input")
	}
}
