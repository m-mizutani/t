package action

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structs for validation
type TestRequiredString struct {
	Name string `validate:"required"`
}

type TestRequiredSlice struct {
	Items []string `validate:"required"`
}

type TestRequiredMap struct {
	Data map[string]interface{} `validate:"required"`
}

type TestRequiredPointer struct {
	Ptr *string `validate:"required"`
}

type TestRequiredInterface struct {
	Value interface{} `validate:"required"`
}

type TestMultipleFields struct {
	Name     string   `validate:"required"`
	Items    []string `validate:"required"`
	Optional string   // No validation tag
}

type TestNestedStruct struct {
	Inner TestRequiredString `validate:"required"`
}

type TestNoValidation struct {
	Name string
	Age  int
}

func TestValidator_NewValidator(t *testing.T) {
	validator := NewValidator()
	assert.NotNil(t, validator)
}

func TestValidator_Validate_RequiredString(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		input   TestRequiredString
		wantErr bool
	}{
		{
			name:    "valid string",
			input:   TestRequiredString{Name: "test"},
			wantErr: false,
		},
		{
			name:    "empty string should fail",
			input:   TestRequiredString{Name: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "field is required")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_Validate_RequiredSlice(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		input   TestRequiredSlice
		wantErr bool
	}{
		{
			name:    "valid slice",
			input:   TestRequiredSlice{Items: []string{"item1", "item2"}},
			wantErr: false,
		},
		{
			name:    "empty slice should fail",
			input:   TestRequiredSlice{Items: []string{}},
			wantErr: true,
		},
		{
			name:    "nil slice should fail",
			input:   TestRequiredSlice{Items: nil},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "field is required")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_Validate_RequiredMap(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		input   TestRequiredMap
		wantErr bool
	}{
		{
			name:    "valid map",
			input:   TestRequiredMap{Data: map[string]interface{}{"key": "value"}},
			wantErr: false,
		},
		{
			name:    "empty map should fail",
			input:   TestRequiredMap{Data: map[string]interface{}{}},
			wantErr: true,
		},
		{
			name:    "nil map should fail",
			input:   TestRequiredMap{Data: nil},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "field is required")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_Validate_RequiredPointer(t *testing.T) {
	validator := NewValidator()

	validStr := "test"
	tests := []struct {
		name    string
		input   TestRequiredPointer
		wantErr bool
	}{
		{
			name:    "valid pointer",
			input:   TestRequiredPointer{Ptr: &validStr},
			wantErr: false,
		},
		{
			name:    "nil pointer should fail",
			input:   TestRequiredPointer{Ptr: nil},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "field is required")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_Validate_RequiredInterface(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		input   TestRequiredInterface
		wantErr bool
	}{
		{
			name:    "valid interface with value",
			input:   TestRequiredInterface{Value: "test"},
			wantErr: false,
		},
		{
			name:    "valid interface with zero value",
			input:   TestRequiredInterface{Value: 0},
			wantErr: false,
		},
		{
			name:    "nil interface should fail",
			input:   TestRequiredInterface{Value: nil},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "field is required")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_Validate_MultipleFields(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		input   TestMultipleFields
		wantErr bool
	}{
		{
			name: "all valid",
			input: TestMultipleFields{
				Name:     "test",
				Items:    []string{"item1"},
				Optional: "optional",
			},
			wantErr: false,
		},
		{
			name: "optional field can be empty",
			input: TestMultipleFields{
				Name:     "test",
				Items:    []string{"item1"},
				Optional: "",
			},
			wantErr: false,
		},
		{
			name: "required name empty should fail",
			input: TestMultipleFields{
				Name:     "",
				Items:    []string{"item1"},
				Optional: "optional",
			},
			wantErr: true,
		},
		{
			name: "required items empty should fail",
			input: TestMultipleFields{
				Name:     "test",
				Items:    []string{},
				Optional: "optional",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "field is required")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_Validate_NoValidation(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name  string
		input TestNoValidation
	}{
		{
			name:  "struct with no validation tags",
			input: TestNoValidation{Name: "test", Age: 25},
		},
		{
			name:  "struct with empty fields",
			input: TestNoValidation{Name: "", Age: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			assert.NoError(t, err)
		})
	}
}

func TestValidator_Validate_NonStruct(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name  string
		input interface{}
	}{
		{
			name:  "string",
			input: "test",
		},
		{
			name:  "int",
			input: 42,
		},
		{
			name:  "slice",
			input: []string{"item1", "item2"},
		},
		{
			name:  "map",
			input: map[string]interface{}{"key": "value"},
		},
		{
			name:  "nil",
			input: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			assert.NoError(t, err)
		})
	}
}

func TestValidator_Validate_Pointer(t *testing.T) {
	validator := NewValidator()

	valid := TestRequiredString{Name: "test"}
	invalid := TestRequiredString{Name: ""}

	tests := []struct {
		name    string
		input   *TestRequiredString
		wantErr bool
	}{
		{
			name:    "valid pointer to struct",
			input:   &valid,
			wantErr: false,
		},
		{
			name:    "invalid pointer to struct",
			input:   &invalid,
			wantErr: true,
		},
		{
			name:    "nil pointer",
			input:   nil,
			wantErr: false, // nil pointer should not be validated
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "field is required")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_Validate_UnknownValidationRule(t *testing.T) {
	validator := NewValidator()

	type TestUnknownRule struct {
		Name string `validate:"unknown_rule"`
	}

	input := TestUnknownRule{Name: "test"}
	err := validator.Validate(input)
	assert.NoError(t, err) // Unknown rules should be ignored
}

func TestValidator_Validate_CommaDelimitedRules(t *testing.T) {
	validator := NewValidator()

	type TestCommaRules struct {
		Name string `validate:"required,unknown_rule"`
	}

	tests := []struct {
		name    string
		input   TestCommaRules
		wantErr bool
	}{
		{
			name:    "valid with multiple rules",
			input:   TestCommaRules{Name: "test"},
			wantErr: false,
		},
		{
			name:    "invalid with multiple rules",
			input:   TestCommaRules{Name: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "field is required")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_validateRequired_EdgeCases(t *testing.T) {
	validator := NewValidator()

	// Test with different slice/array types
	type TestDifferentTypes struct {
		IntSlice    []int               `validate:"required"`
		StringArray [3]string           `validate:"required"`
		StringMap   map[string]string   `validate:"required"`
		IntMap      map[int]interface{} `validate:"required"`
	}

	tests := []struct {
		name    string
		input   TestDifferentTypes
		wantErr bool
	}{
		{
			name: "all valid",
			input: TestDifferentTypes{
				IntSlice:    []int{1, 2, 3},
				StringArray: [3]string{"a", "b", "c"},
				StringMap:   map[string]string{"key": "value"},
				IntMap:      map[int]interface{}{1: "value"},
			},
			wantErr: false,
		},
		{
			name: "empty int slice should fail",
			input: TestDifferentTypes{
				IntSlice:    []int{},
				StringArray: [3]string{"a", "b", "c"},
				StringMap:   map[string]string{"key": "value"},
				IntMap:      map[int]interface{}{1: "value"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "field is required")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_Validate_Performance(t *testing.T) {
	validator := NewValidator()

	// Test with struct containing many fields
	type LargeStruct struct {
		Field1  string `validate:"required"`
		Field2  string `validate:"required"`
		Field3  string `validate:"required"`
		Field4  string `validate:"required"`
		Field5  string `validate:"required"`
		Field6  string
		Field7  string
		Field8  string
		Field9  string
		Field10 string
	}

	input := LargeStruct{
		Field1:  "value1",
		Field2:  "value2",
		Field3:  "value3",
		Field4:  "value4",
		Field5:  "value5",
		Field6:  "value6",
		Field7:  "value7",
		Field8:  "value8",
		Field9:  "value9",
		Field10: "value10",
	}

	// Run validation multiple times to ensure performance
	for i := 0; i < 100; i++ {
		err := validator.Validate(input)
		require.NoError(t, err)
	}
}
