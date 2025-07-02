package ctag

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func assertTags(t *testing.T, expected CTags, tags CTags, err error) {
	assert.NoError(t, err)
	assert.Len(t, tags, len(expected))

	for i, tag := range tags {
		assert.Equal(t, expected[i].Key, tag.Key)
		assert.Equal(t, expected[i].Name, tag.Name)
		assert.Equal(t, expected[i].Options, tag.Options)
		assert.Equal(t, expected[i].Field, tag.Field)
	}
}

func TestGetTags(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		tagKey   string
		expected CTags
	}{
		{
			name: "basic",
			input: struct {
				ID   int    `query:"id"`
				Name string `query:"name"`
			}{
				ID:   1,
				Name: "John",
			},
			tagKey: "query",
			expected: CTags{
				{Key: "query", Name: "id", Field: 1},
				{Key: "query", Name: "name", Field: "John"},
			},
		},
		{
			name: "omitempty",
			input: struct {
				Email string `query:"email,omitempty"`
				Age   int    `query:"age,omitempty"`
			}{
				Email: "",
				Age:   0,
			},
			tagKey:   "query",
			expected: CTags{},
		},
		{
			name: "nested",
			input: struct {
				Simple struct {
					ID   int    `query:"id"`
					Name string `query:"name"`
				} `query:"simple"`
				Extra string `query:"extra"`
			}{
				Simple: struct {
					ID   int    `query:"id"`
					Name string `query:"name"`
				}{ID: 2, Name: "Jane"},
				Extra: "Additional Info",
			},
			tagKey: "query",
			expected: CTags{
				{Key: "query", Name: "simple", Field: struct {
					ID   int    `query:"id"`
					Name string `query:"name"`
				}{ID: 2, Name: "Jane"}},
				{Key: "query", Name: "id", Field: 2},
				{Key: "query", Name: "name", Field: "Jane"},
				{Key: "query", Name: "extra", Field: "Additional Info"},
			},
		},
		{
			name: "slice",
			input: struct {
				IDs []int `query:"ids"`
			}{
				IDs: []int{1, 2, 3},
			},
			tagKey: "query",
			expected: CTags{
				{Key: "query", Name: "ids", Field: []int{1, 2, 3}},
			},
		},
		{
			name: "map",
			input: struct {
				Properties map[string]string `query:"properties"`
			}{
				Properties: map[string]string{"key1": "value1", "key2": "value2"},
			},
			tagKey: "query",
			expected: CTags{
				{Key: "query", Name: "properties", Field: map[string]string{"key1": "value1", "key2": "value2"}},
			},
		},
		{
			name: "private",
			input: struct {
				PublicField  string `query:"public"`
				privateField string `query:"private"`
			}{
				PublicField:  "public value",
				privateField: "private value",
			},
			tagKey: "query",
			expected: CTags{
				{Key: "query", Name: "public", Field: "public value"},
			},
		},
		{
			name: "pointer",
			input: struct {
				PtrInt *int  `query:"ptr"`
				PtrNil *bool `query:"ptr"`
			}{
				PtrInt: func() *int { v := 42; return &v }(),
				PtrNil: func() *bool { return nil }(),
			},
			tagKey: "query",
			expected: CTags{
				{Key: "query", Name: "ptr", Field: 42},
				{Key: "query", Name: "ptr", Field: nil},
			},
		},
		{
			name: "embedded structs",
			input: struct {
				ID       int `query:"id"`
				Embedded struct {
					Description string `query:"description"`
					Title       string `query:"title"`
					Version     int    `query:"version"`
				}
			}{
				ID: 1,
				Embedded: struct {
					Description string `query:"description"`
					Title       string `query:"title"`
					Version     int    `query:"version"`
				}{
					Description: "Test description",
					Version:     1,
				},
			},
			tagKey: "query",
			expected: CTags{
				{Key: "query", Name: "id", Field: 1},
				{Key: "query", Name: "description", Field: "Test description"},
				{Key: "query", Name: "title", Field: ""},
				{Key: "query", Name: "version", Field: 1},
			},
		},
		{
			name: "'-'",
			input: struct {
				ID   int    `query:"-"`
				Name string `query:"name"`
			}{
				ID:   1,
				Name: "John",
			},
			tagKey: "query",
			expected: CTags{
				{Key: "query", Name: "name", Field: "John"},
			},
		},
		{
			name: "options",
			input: struct {
				ID      []int  `body:"json,comma,omitempty"`
				WithOpt bool   `body:"opt,true,xml"`
				Name    string `path:"{username},encode,url"`
			}{
				ID:      []int{1, 2, 3},
				WithOpt: true,
				Name:    "John",
			},
			tagKey: "body",
			expected: CTags{
				{Key: "body", Name: "json", Options: []string{"comma", "omitempty"}, Field: []int{1, 2, 3}},
				{Key: "body", Name: "opt", Options: []string{"true", "xml"}, Field: true},
			},
		},
	}
	for _, tt := range tests {
		tags, err := GetTags(tt.tagKey, tt.input)
		assertTags(t, tt.expected, tags, err)
	}
}

func TestGetTagsAndProcess(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		tagKey   string
		expected CTags
	}{
		{
			name: "basic",
			input: struct {
				ID   int    `query:"id"`
				Name string `query:"name"`
			}{
				ID:   1,
				Name: "John",
			},
			tagKey: "query",
			expected: CTags{
				{Key: "query", Name: "processed_id", Field: 1},
				{Key: "query", Name: "processed_name", Field: "John"},
			},
		},
		{
			name: "omitempty",
			input: struct {
				Email string `query:"email,omitempty"`
				Age   int    `query:"age,omitempty"`
			}{
				Email: "",
				Age:   0,
			},
			tagKey:   "query",
			expected: CTags{},
		},
		{
			name: "nested",
			input: struct {
				Simple struct {
					ID   int    `query:"id"`
					Name string `query:"name"`
				} `query:"simple"`
				Extra string `query:"extra"`
			}{
				Simple: struct {
					ID   int    `query:"id"`
					Name string `query:"name"`
				}{ID: 2, Name: "Jane"},
				Extra: "Additional Info",
			},
			tagKey: "query",
			expected: CTags{
				{Key: "query", Name: "processed_simple", Field: struct {
					ID   int    `query:"id"`
					Name string `query:"name"`
				}{ID: 2, Name: "Jane"}},
				{Key: "query", Name: "processed_id", Field: 2},
				{Key: "query", Name: "processed_name", Field: "Jane"},
				{Key: "query", Name: "processed_extra", Field: "Additional Info"},
			},
		},
		{
			name: "slice",
			input: struct {
				IDs []int `query:"ids"`
			}{
				IDs: []int{1, 2, 3},
			},
			tagKey: "query",
			expected: CTags{
				{Key: "query", Name: "processed_ids", Field: []int{1, 2, 3}},
			},
		},
		{
			name: "map",
			input: struct {
				Properties map[string]string `query:"properties"`
			}{
				Properties: map[string]string{"key1": "value1", "key2": "value2"},
			},
			tagKey: "query",
			expected: CTags{
				{Key: "query", Name: "processed_properties", Field: map[string]string{"key1": "value1", "key2": "value2"}},
			},
		},
		{
			name: "private",
			input: struct {
				PublicField  string `query:"public"`
				privateField string `query:"private"`
			}{
				PublicField:  "public value",
				privateField: "private value",
			},
			tagKey: "query",
			expected: CTags{
				{Key: "query", Name: "processed_public", Field: "public value"},
			},
		},
		{
			name: "pointer",
			input: struct {
				PtrInt *int  `query:"ptr"`
				PtrNil *bool `query:"ptr"`
			}{
				PtrInt: func() *int { v := 42; return &v }(),
				PtrNil: func() *bool { return nil }(),
			},
			tagKey: "query",
			expected: CTags{
				{Key: "query", Name: "processed_ptr", Field: 42},
				{Key: "query", Name: "processed_ptr", Field: nil},
			},
		},
		{
			name: "embedded structs",
			input: struct {
				ID       int `query:"id"`
				Embedded struct {
					Description string `query:"description"`
					Title       string `query:"title"`
					Version     int    `query:"version"`
				}
			}{
				ID: 1,
				Embedded: struct {
					Description string `query:"description"`
					Title       string `query:"title"`
					Version     int    `query:"version"`
				}{
					Description: "Test description",
					Version:     1,
				},
			},
			tagKey: "query",
			expected: CTags{
				{Key: "query", Name: "processed_id", Field: 1},
				{Key: "query", Name: "processed_description", Field: "Test description"},
				{Key: "query", Name: "processed_title", Field: ""},
				{Key: "query", Name: "processed_version", Field: 1},
			},
		},
		{
			name: "'-'",
			input: struct {
				ID   int    `query:"-"`
				Name string `query:"name"`
			}{
				ID:   1,
				Name: "John",
			},
			tagKey: "query",
			expected: CTags{
				{Key: "query", Name: "processed_name", Field: "John"},
			},
		},
		{
			name: "options",
			input: struct {
				ID      []int  `body:"json,comma,omitempty"`
				WithOpt bool   `body:"opt,true,xml"`
				Name    string `path:"{username},encode,url"`
			}{
				ID:      []int{1, 2, 3},
				WithOpt: true,
				Name:    "John",
			},
			tagKey: "body",
			expected: CTags{
				{Key: "body", Name: "processed_json", Options: []string{"comma", "omitempty"}, Field: []int{1, 2, 3}},
				{Key: "body", Name: "processed_opt", Options: []string{"true", "xml"}, Field: true},
			},
		},
	}

	for _, tt := range tests {
		tags, err := GetTagsAndProcess(tt.tagKey, tt.input, &testProcessor{})
		assertTags(t, tt.expected, tags, err)
	}
}

func TestFilter(t *testing.T) {
	tags := CTags{
		{Key: "body", Name: "json", Field: "1,2,3,4"},
		{Key: "query", Name: "id", Field: 42},
		{Key: "url", Name: "path", Field: nil},
	}

	filteredTags := tags.Filter(func(tag CTag) bool {
		return tag.Field != nil
	})

	expectedTags := CTags{
		{Key: "body", Name: "json", Field: "1,2,3,4"},
		{Key: "query", Name: "id", Field: 42},
	}

	assert.Equal(t, expectedTags, filteredTags)
}

func TestFind(t *testing.T) {
	tags := CTags{
		{Key: "body", Name: "text,omitempty", Field: 69},
		{Key: "query", Name: "id", Field: 42},
		{Key: "url", Name: "path", Field: nil},
	}

	foundTag := tags.Find(func(tag CTag) bool {
		return tag.Key == "body"
	})

	expectedTag := &CTag{Key: "body", Name: "text,omitempty", Field: 69}

	assert.Equal(t, expectedTag, foundTag)
}

func TestFindNil(t *testing.T) {
	tags := CTags{
		{Key: "body", Name: "xml", Field: nil},
		{Key: "query", Name: "id", Field: nil},
	}

	foundTag := tags.Find(func(tag CTag) bool {
		return tag.Field != nil
	})

	assert.Nil(t, foundTag)
}

func TestToSlice(t *testing.T) {
	tags := CTags{
		{Key: "body", Name: "xml", Field: 10},
		{Key: "path", Name: "param", Field: 42},
	}

	tagSlice := tags.ToSlice()

	expectedSlice := []CTag{
		{Key: "body", Name: "xml", Field: 10},
		{Key: "path", Name: "param", Field: 42},
	}

	assert.Equal(t, expectedSlice, tagSlice)
}

func TestAssertFieldType(t *testing.T) {
	tests := []struct {
		name       string
		tag        *CTag
		targetType any
		expectErr  bool
		expected   any
	}{
		{
			name: "assert string",
			tag: &CTag{
				Key:   "query",
				Name:  "name",
				Field: "John",
			},
			targetType: new(string),
			expectErr:  false,
			expected:   "John",
		},
		{
			name: "assert int failure",
			tag: &CTag{
				Key:   "query",
				Name:  "name",
				Field: "John",
			},
			targetType: new(int),
			expectErr:  true,
		},
		{
			name: "assert slice",
			tag: &CTag{
				Key:   "query",
				Name:  "ids",
				Field: []int{1, 2, 3},
			},
			targetType: new([]int),
			expectErr:  false,
			expected:   []int{1, 2, 3},
		},
		{
			name: "assert map",
			tag: &CTag{
				Key:   "query",
				Name:  "properties",
				Field: map[string]string{"key1": "value1", "key2": "value2"},
			},
			targetType: new(map[string]string),
			expectErr:  false,
			expected:   map[string]string{"key1": "value1", "key2": "value2"},
		},
		{
			name: "assert pointer",
			tag: &CTag{
				Key:   "query",
				Name:  "ptr",
				Field: func() *int { v := 42; return &v }(),
			},
			targetType: new(*int),
			expectErr:  false,
			expected:   func() *int { v := 42; return &v }(),
		},
		{
			name: "assert struct",
			tag: &CTag{
				Key:   "query",
				Name:  "simple",
				Field: struct{ Name string }{Name: "John"},
			},
			targetType: new(struct{ Name string }),
			expectErr:  false,
			expected:   struct{ Name string }{Name: "John"},
		},
	}

	for _, tt := range tests {
		switch tt.targetType.(type) {
		case string:
			tag, err := AssertFieldType[string](tt.tag)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, tag.Field)
			}
		case int:
			_, err := AssertFieldType[int](tt.tag)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		case []int:
			tag, err := AssertFieldType[[]int](tt.tag)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, tag.Field)
			}
		case map[string]string:
			tag, err := AssertFieldType[map[string]string](tt.tag)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, tag.Field)
			}
		case *int:
			tag, err := AssertFieldType[*int](tt.tag)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, *tt.expected.(*int), tag.Field)
			}
		case struct{ Name string }:
			tag, err := AssertFieldType[struct{ Name string }](tt.tag)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, tag.Field)
			}
		}
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name     string
		tag      *CTag
		expected string
	}{
		{
			name: "assert string",
			tag: &CTag{
				Key:     "query",
				Name:    "name",
				Options: []string{"omitempty"},
				Field:   "John",
			},
			expected: "CTag(Key=query, Name=name, Options=[omitempty], Field=John)",
		},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.tag.String())
	}
}

type testProcessor struct{}

func (p *testProcessor) Process(field any, tag *CTag) error {
	tag.Name = "processed_" + tag.Name
	return nil
}

type setFieldProcessor struct{}

func (p *setFieldProcessor) Process(field any, tag *CTag) error {
	if tag.Name == "name" {
		return SetField(field, "test_value")
	}
	return nil
}

func TestSetField(t *testing.T) {
	tests := []struct {
		name        string
		field       any
		value       any
		expected    any
		expectError bool
		errorMsg    string
	}{
		// Basic string operations
		{
			name:     "string to string",
			field:    func() any { var s string; return &s }(),
			value:    "hello",
			expected: "hello",
		},
		{
			name:     "int to string",
			field:    func() any { var s string; return &s }(),
			value:    42,
			expected: "42",
		},
		{
			name:     "bool to string",
			field:    func() any { var s string; return &s }(),
			value:    true,
			expected: "true",
		},

		// Basic numeric operations
		{
			name:     "string to int",
			field:    func() any { var i int; return &i }(),
			value:    "42",
			expected: 42,
		},
		{
			name:     "int to int",
			field:    func() any { var i int; return &i }(),
			value:    42,
			expected: 42,
		},
		{
			name:     "float to int",
			field:    func() any { var i int; return &i }(),
			value:    42.7,
			expected: 42,
		},
		{
			name:     "string to float",
			field:    func() any { var f float64; return &f }(),
			value:    "3.14",
			expected: 3.14,
		},
		{
			name:     "int to float",
			field:    func() any { var f float64; return &f }(),
			value:    42,
			expected: 42.0,
		},

		// Boolean operations
		{
			name:     "string true to bool",
			field:    func() any { var b bool; return &b }(),
			value:    "true",
			expected: true,
		},
		{
			name:     "string false to bool",
			field:    func() any { var b bool; return &b }(),
			value:    "false",
			expected: false,
		},
		{
			name:     "string 1 to bool",
			field:    func() any { var b bool; return &b }(),
			value:    "1",
			expected: true,
		},

		// Pointer operations
		{
			name:     "string to string pointer",
			field:    func() any { var s *string; return &s }(),
			value:    "hello",
			expected: func() *string { s := "hello"; return &s }(),
		},
		{
			name:     "int to int pointer",
			field:    func() any { var i *int; return &i }(),
			value:    42,
			expected: func() *int { i := 42; return &i }(),
		},
		{
			name:     "string to int pointer",
			field:    func() any { var i *int; return &i }(),
			value:    "42",
			expected: func() *int { i := 42; return &i }(),
		},

		// Slice operations
		{
			name:     "string slice to string slice",
			field:    func() any { var s []string; return &s }(),
			value:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "comma-separated string to string slice",
			field:    func() any { var s []string; return &s }(),
			value:    "a,b,c",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "comma-separated string to int slice",
			field:    func() any { var s []int; return &s }(),
			value:    "1,2,3",
			expected: []int{1, 2, 3},
		},
		{
			name:     "single value to slice",
			field:    func() any { var s []string; return &s }(),
			value:    "single",
			expected: []string{"single"},
		},
		{
			name:     "empty string to slice",
			field:    func() any { var s []string; return &s }(),
			value:    "",
			expected: []string{},
		},

		// Map operations
		{
			name:     "map to map",
			field:    func() any { var m map[string]string; return &m }(),
			value:    map[string]string{"key": "value"},
			expected: map[string]string{"key": "value"},
		},

		// Interface operations
		{
			name:     "any to interface",
			field:    func() any { var i interface{}; return &i }(),
			value:    "hello",
			expected: "hello",
		},

		// Nil value operations
		{
			name:     "nil to string",
			field:    func() any { var s string; return &s }(),
			value:    nil,
			expected: "",
		},
		{
			name:     "nil to int",
			field:    func() any { var i int; return &i }(),
			value:    nil,
			expected: 0,
		},
		{
			name:     "nil to pointer",
			field:    func() any { var s *string; return &s }(),
			value:    nil,
			expected: (*string)(nil),
		},

		// Error cases
		{
			name:        "non-pointer field",
			field:       "not a pointer",
			value:       "hello",
			expectError: true,
			errorMsg:    "field must be a pointer",
		},
		{
			name:        "nil pointer field",
			field:       (*string)(nil),
			value:       "hello",
			expectError: true,
			errorMsg:    "field pointer is nil",
		},
		{
			name:        "invalid string to int",
			field:       func() any { var i int; return &i }(),
			value:       "not a number",
			expectError: true,
			errorMsg:    "cannot parse",
		},
		{
			name:        "invalid string to bool",
			field:       func() any { var b bool; return &b }(),
			value:       "not a bool",
			expectError: true,
			errorMsg:    "cannot parse",
		},
		{
			name:        "invalid string to float",
			field:       func() any { var f float64; return &f }(),
			value:       "not a float",
			expectError: true,
			errorMsg:    "cannot parse",
		},
		{
			name:        "incompatible map types",
			field:       func() any { var m map[string]string; return &m }(),
			value:       map[int]int{1: 2},
			expectError: true,
			errorMsg:    "cannot convert",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetField(tt.field, tt.value)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				return
			}

			assert.NoError(t, err)

			fieldVal := reflect.ValueOf(tt.field).Elem()
			actual := fieldVal.Interface()

			if fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil() {
				assert.Nil(t, tt.expected)
			} else if fieldVal.Kind() == reflect.Ptr && !fieldVal.IsNil() {
				expectedPtr := reflect.ValueOf(tt.expected)
				if expectedPtr.Kind() == reflect.Ptr && !expectedPtr.IsNil() {
					assert.Equal(t, expectedPtr.Elem().Interface(), fieldVal.Elem().Interface())
				} else {
					assert.Equal(t, tt.expected, fieldVal.Elem().Interface())
				}
			} else {
				assert.Equal(t, tt.expected, actual)
			}
		})
	}
}

func TestSetFieldComplexSlices(t *testing.T) {
	tests := []struct {
		name        string
		field       any
		value       any
		expected    any
		expectError bool
	}{
		{
			name:     "comma-separated with spaces",
			field:    func() any { var s []string; return &s }(),
			value:    " a , b , c ",
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "comma-separated bool slice",
			field:    func() any { var s []bool; return &s }(),
			value:    "true,false,1,0",
			expected: []bool{true, false, true, false},
		},
		{
			name:     "comma-separated float slice",
			field:    func() any { var s []float64; return &s }(),
			value:    "1.1,2.2,3.3",
			expected: []float64{1.1, 2.2, 3.3},
		},
		{
			name:        "invalid comma-separated int slice",
			field:       func() any { var s []int; return &s }(),
			value:       "1,not_a_number,3",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetField(tt.field, tt.value)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			fieldVal := reflect.ValueOf(tt.field).Elem()
			actual := fieldVal.Interface()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestSetFieldNumericConversions(t *testing.T) {
	tests := []struct {
		name     string
		field    any
		value    any
		expected any
	}{
		// Int conversions
		{
			name:     "uint to int",
			field:    func() any { var i int; return &i }(),
			value:    uint(42),
			expected: 42,
		},
		{
			name:     "int32 to int64",
			field:    func() any { var i int64; return &i }(),
			value:    int32(42),
			expected: int64(42),
		},
		{
			name:     "float64 to int",
			field:    func() any { var i int; return &i }(),
			value:    42.9,
			expected: 42,
		},

		// Uint conversions
		{
			name:     "int to uint",
			field:    func() any { var u uint; return &u }(),
			value:    42,
			expected: uint(42),
		},
		{
			name:     "float to uint",
			field:    func() any { var u uint; return &u }(),
			value:    42.9,
			expected: uint(42),
		},

		// Float conversions
		{
			name:     "int to float32",
			field:    func() any { var f float32; return &f }(),
			value:    42,
			expected: float32(42),
		},
		{
			name:     "uint to float64",
			field:    func() any { var f float64; return &f }(),
			value:    uint(42),
			expected: float64(42),
		},
		{
			name:     "float32 to float64",
			field:    func() any { var f float64; return &f }(),
			value:    float32(3.14),
			expected: float64(float32(3.14)), // Account for precision loss
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetField(tt.field, tt.value)
			assert.NoError(t, err)

			fieldVal := reflect.ValueOf(tt.field).Elem()
			actual := fieldVal.Interface()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestSetFieldWithProcessor(t *testing.T) {
	type TestStruct struct {
		Name string `test:"name"`
		Age  int    `test:"age"`
	}

	tests := []struct {
		name     string
		input    TestStruct
		expected TestStruct
	}{
		{
			name:     "string field",
			input:    TestStruct{Name: "original", Age: 25},
			expected: TestStruct{Name: "test_value", Age: 25},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := &setFieldProcessor{}

			_, err := GetTagsAndProcess("test", &tt.input, processor)
			assert.NoError(t, err)

			assert.Equal(t, "test_value", tt.input.Name)
			assert.Equal(t, 25, tt.input.Age)
		})
	}
}

func TestSetFieldMapToStruct(t *testing.T) {
	type NestedStruct struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name     string
		field    any
		value    any
		expected any
	}{
		{
			name:  "map to struct with json tags",
			field: func() any { var s NestedStruct; return &s }(),
			value: map[string]interface{}{
				"name": "John",
				"age":  float64(30),
			},
			expected: NestedStruct{Name: "John", Age: 30},
		},
		{
			name:  "map to struct with field names",
			field: func() any { var s NestedStruct; return &s }(),
			value: map[string]interface{}{
				"Name": "Jane",
				"Age":  25,
			},
			expected: NestedStruct{Name: "Jane", Age: 25},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetField(tt.field, tt.value)
			assert.NoError(t, err)

			fieldVal := reflect.ValueOf(tt.field).Elem()
			actual := fieldVal.Interface()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestSetFieldSliceConversion(t *testing.T) {
	tests := []struct {
		name     string
		field    any
		value    any
		expected any
	}{
		{
			name:     "interface slice to string slice",
			field:    func() any { var s []string; return &s }(),
			value:    []interface{}{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "interface slice to int slice",
			field:    func() any { var s []int; return &s }(),
			value:    []interface{}{float64(1), float64(2), float64(3)}, // JSON numbers
			expected: []int{1, 2, 3},
		},
		{
			name:     "interface slice to bool slice",
			field:    func() any { var s []bool; return &s }(),
			value:    []interface{}{true, false, true},
			expected: []bool{true, false, true},
		},
		{
			name:     "mixed interface slice to string slice",
			field:    func() any { var s []string; return &s }(),
			value:    []interface{}{42, true, "test"},
			expected: []string{"42", "true", "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetField(tt.field, tt.value)
			assert.NoError(t, err)

			fieldVal := reflect.ValueOf(tt.field).Elem()
			actual := fieldVal.Interface()
			assert.Equal(t, tt.expected, actual)
		})
	}
}
