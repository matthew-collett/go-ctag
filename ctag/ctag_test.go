package ctag

import (
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

	// Find the first tag with a non-nil field
	foundTag := tags.Find(func(tag CTag) bool {
		return tag.Field != nil
	})

	assert.Nil(t, foundTag)
}

// TestToSlice tests the ToSlice method of CTags.
func TestToSlice(t *testing.T) {
	tags := CTags{
		{Key: "body", Name: "xml", Field: 10},
		{Key: "path", Name: "param", Field: 42},
	}

	// Convert CTags to []CTag
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
