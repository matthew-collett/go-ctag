package ctag

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// CTag represents a parsed tag associated with a struct field.
// It holds the tag's key, name, and additional options along with the field's actual value.
//
// Fields:
//
//	Key     - The primary identifier in a struct tag, used to retrieve the tag.
//	Name    - The first value associated with the Key in the tag, typically used to describe the purpose or content.
//	Options - Additional comma-separated values associated with the Key, providing further instructions or modifiers.
//	Field   - The actual data value of the struct field.
//
// Example:
//
//	type Request struct {
//	    IDs []string `body:"text,comma,omitempty"`
//	}
//
// The tag associated with the IDs field is parsed as:
//
//	Key = "body"
//	Name = "text"
//	Options = ["comma", "omitempty"]
//	Field contains the actual data of the string field 'IDs'.
type CTag struct {
	Key     string   // Key is the primary identifier in a struct tag.
	Name    string   // Name is the first value associated with Key in the tag.
	Options []string // Options are additional values associated with Key.
	Field   any      // Field is the data value of the struct field.
}

// TagProcessor defines an interface for custom processing of fields based on their associated tags.
// It is intended to be implemented by clients who wish to apply additional processing
// to fields of a struct during tag processing.
//
// The Process method is called for each field that is successfully retrieved and parsed.
// It allows for the modification, validation, or enhancement of the field.
type TagProcessor interface {
	Process(field any, tag *CTag) error // Process applies a custom processing rule to a tagged field.
}

// CTags represents a slice of CTag structures.
//
// This type is a convenient wrapper for []CTag used to define methods
// on a slice of CTag objects. By defining methods on this type, we can perform
// common operations on the collection of tags in a more idiomatic and readable way.
type CTags []CTag

// GetTags retrieves all tags from a struct without additional processing.
// This function is a convenience wrapper around GetTagsWithProcessor, using nil as the processor
// to perform no additional processing after parsing the tag.
//
// A field is skipped if:
//   - The tag name is an empty string
//   - The tag name is "-", indicating the field should not be serialized or processed.
//   - The tag contains "omitempty" and the field's value is the zero value for that type
//
// Parameters:
//
//	key  - the tag key to search for in the struct tags
//	data - the struct from which tags should be extracted, must be a struct
//
// Returns:
//
//	A slice of CTag containing all tags, or an error if the input is not a struct.
//
// Example usage:
//
//	type ExampleStruct struct {
//	    Field1 string `json:"field1"`
//	    Field2 int    `json:"field2,omitempty"`
//	    Field3 bool   `json:"-"`
//	}
//
//	data := ExampleStruct{
//	    Field1: "value1",
//	    Field2: 0,
//	    Field3: true,
//	}
//
//	tags, err := GetTags("json", data)
//	if err != nil {
//	    fmt.Printf("Error: %v\n", err)
//	} else {
//	    fmt.Printf("Tags: %+v\n", tags)
//	}
func GetTags(key string, data any) (CTags, error) {
	return GetTagsAndProcess(key, data, nil)
}

// GetTagsAndProcess retrieves and processes all tags from a struct.
// It allows for custom processing of each tag via a provided TagProcessor.
//
// A field is skipped if:
//   - The tag name is an empty string
//   - The tag name is "-", indicating the field should not be serialized or processed.
//   - The tag contains "omitempty" and the field's value is the zero value for that type
//
// Parameters:
//
//	key       - the tag key to search for in the struct tags
//	data      - the struct from which tags should be extracted, must be a struct
//	processor - a TagProcessor to apply custom processing to each extracted tag
//
// Returns:
//
//	A slice of CTag containing all processed tags, or an error if the input is not a struct or the processing fails.
//
// Example usage:
//
//	type ExampleStruct struct {
//	    Field1 string `json:"field1"`
//	    Field2 int    `json:"field2,omitempty"`
//	    Field3 bool   `json:"-"`
//	}
//
//	type MyProcessor struct{}
//
//	func (p *MyProcessor) Process(tag CTag) error {
//	    // Custom processing logic here
//	    return nil
//	}
//
//	data := ExampleStruct{
//	    Field1: "value1",
//	    Field2: 0,
//	    Field3: true,
//	}
//
//	processor := &MyProcessor{}
//	tags, err := GetTagsAndProcess("json", data, processor)
//	if err != nil {
//	    fmt.Printf("Error: %v\n", err)
//	} else {
//	    fmt.Printf("Processed Tags: %+v\n", tags)
//	}
func GetTagsAndProcess(key string, data any, processor TagProcessor) (CTags, error) {
	v := reflect.Indirect(reflect.ValueOf(data))
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("ctag: expected input to be a struct; got: %T", data)
	}
	return getTags(key, v, processor)
}

// Filter returns a new CTags slice containing only the tags that satisfy the
// provided predicate function.
//
// Parameters:
//
//	predicate - a function that takes a CTag and returns a boolean indicating
//	            whether the tag should be included in the resulting slice.
//
// Returns:
//
//	A new CTags slice containing only the tags that satisfy the provided predicate function.
//
// Example usage:
//
//	tags := CTags{
//	    {Key: "query", Name: "ptr_int", Field: 42},
//	    {Key: "query", Name: "ptr_nil", Field: nil},
//	}
//
//	// Filter tags to include only those with non-nil field values
//	nonNilTags := tags.Filter(func(tag CTag) bool {
//	    return tag.Field != nil
//	})
func (ct CTags) Filter(predicate func(CTag) bool) CTags {
	var ftags CTags
	for _, t := range ct {
		if predicate(t) {
			ftags = append(ftags, t)
		}
	}
	return ftags
}

// Find returns the first CTag that matches the provided predicate function.
// If no tag matches, it returns nil.
//
// Parameters:
//
//	predicate - a function that takes a CTag and returns a boolean indicating
//	            whether the tag matches the condition.
//
// Returns:
//
//	A pointer to the first CTag that matches the provided predicate function,
//	or nil if no tag matches.
//
// Example usage:
//
//	tags := CTags{
//	    {Key: "query", Name: "ptr_int", Field: 42},
//	    {Key: "query", Name: "ptr_nil", Field: nil},
//	}
//
//	// Find the first tag with a non-nil field value
//	firstNonNilTag := tags.Find(func(tag CTag) bool {
//	    return tag.Field != nil
//	})
//
//	if firstNonNilTag != nil {
//	    fmt.Println("First non-nil tag:", *firstNonNilTag)
//	} else {
//	    fmt.Println("No non-nil tags found")
//	}
func (ct CTags) Find(predicate func(CTag) bool) *CTag {
	for _, t := range ct {
		if predicate(t) {
			return &t
		}
	}
	return nil
}

// ToSlice converts the CTags type back to a slice of CTag.
//
// Returns:
//
//	A slice of CTag containing all elements in the CTags.
//
// Example usage:
//
//	tags := CTags{
//	    {Key: "query", Name: "ptr_int", Field: 42},
//	    {Key: "query", Name: "ptr_nil", Field: nil},
//	}
//
//	tagSlice := tags.ToSlice()
func (ct CTags) ToSlice() []CTag {
	return []CTag(ct)
}

// AssertFieldType asserts the type of the Field in a CTag to the specified generic type T.
// If the assertion is successful, it returns the updated CTag with the Field set to the asserted type.
// Otherwise, it returns an error indicating the type assertion failure.
//
// Type Parameters:
//
//	T - the type to which the Field should be asserted
//
// Parameters:
//
//	tag - the CTag whose Field needs to be type asserted
//
// Returns:
//
//	The CTag with the Field asserted to type T or an error if the type assertion fails.
//
// Example usage:
//
//	tag := CTag{Key: "query", Name: "int", Field: 42}
//
//	// Assert that the Field is of type int
//	assertedTag, err := AssertFieldType[int](&tag)
//	if err != nil {
//	    fmt.Printf("Error: %v\n", err)
//	} else {
//	    fmt.Printf("Asserted tag: %+v\n", assertedTag)
//	}
//
//	// Attempt to assert that the Field is of type string
//	_, err = AssertFieldType[string](&tag)
//	if err != nil {
//	    fmt.Printf("Error: %v\n", err)
//	}
func AssertFieldType[T any](tag *CTag) (*CTag, error) {
	if val, ok := tag.Field.(T); ok {
		tag.Field = val
		return tag, nil
	}
	return nil, fmt.Errorf("type assertion to %T failed for field %v", (*T)(nil), tag.Field)
}

// String returns a string representation of the CTag.
//
// This method formats the CTag's key, name, options, and field into a readable string.
// It is useful for debugging and logging purposes, providing a clear
// representation of the CTag's contents.
//
// Parameters:
//
//	None
//
// Returns:
//
//	A string representation of the CTag.
//
// Example usage:
//
//	tag := CTag{
//	    Key:     "query",
//	    Name:    "ptr_int",
//	    Options: []string{"opt1", "opt2"},
//	    Field:   42,
//	}
//
//	fmt.Println(tag.String()) // Output: CTag(Key=query, Name=ptr_int, Options=[opt1, opt2], Field=42)
func (t *CTag) String() string {
	options := strings.Join(t.Options, ", ")
	return fmt.Sprintf("CTag(Key=%s, Name=%s, Options=[%s], Field=%+v)", t.Key, t.Name, options, t.Field)
}

// SetField attempts to convert a value to the appropriate type and set it on the field pointer.
// This helper function reduces boilerplate code in TagProcessor implementations by handling
// common type conversions automatically.
//
// The field parameter should be a pointer to the actual struct field (same as TagProcessor.Process receives).
// The value parameter can be any type - the function will attempt to convert it to match the field's type.
//
// Supported conversions:
//   - String to any basic type (int, float, bool, etc.)
//   - Numeric types to other numeric types
//   - Any type to string via fmt.Sprintf
//   - Slice and map types (direct assignment if types match)
//   - Pointer types (allocates new pointer if needed)
//   - Interface types (direct assignment)
//
// Parameters:
//
//	field - a pointer to the struct field to set (must be a pointer)
//	value - the value to convert and assign to the field
//
// Returns:
//
//	An error if the conversion fails or if field is not a pointer.
//
// Example usage:
//
//	type QueryProcessor struct {
//	    req *http.Request
//	}
//
//	func (p *QueryProcessor) Process(field any, tag *CTag) error {
//	    value := p.req.URL.Query().Get(tag.Name)
//	    if value == "" {
//	        return nil
//	    }
//	    return ctag.SetField(field, value)
//	}
//
//	// Instead of writing 20+ type cases manually:
//	switch f := field.(type) {
//	case *string:
//	    *f = value
//	case *int:
//	    intVal, err := strconv.Atoi(value)
//	    // ... handle error and assignment
//	// ... 18+ more cases
//	}
func SetField(field any, value any) error {
	fieldVal := reflect.ValueOf(field)
	if fieldVal.Kind() != reflect.Ptr {
		return fmt.Errorf("ctag: field must be a pointer, got %T", field)
	}

	if fieldVal.IsNil() {
		return fmt.Errorf("ctag: field pointer is nil")
	}

	fieldElem := fieldVal.Elem()
	if !fieldElem.CanSet() {
		return fmt.Errorf("ctag: field is not settable")
	}

	return setValue(fieldElem, value)
}

// setValue handles the actual type conversion and assignment
func setValue(fieldVal reflect.Value, value any) error {
	if value == nil {
		// Set to zero value for the type
		fieldVal.Set(reflect.Zero(fieldVal.Type()))
		return nil
	}

	valueVal := reflect.ValueOf(value)
	fieldType := fieldVal.Type()

	// Direct assignment if types match
	if valueVal.Type().AssignableTo(fieldType) {
		fieldVal.Set(valueVal)
		return nil
	}

	// Handle pointer types
	if fieldType.Kind() == reflect.Ptr {
		return setPointerValue(fieldVal, value)
	}

	// Handle interface types
	if fieldType.Kind() == reflect.Interface {
		fieldVal.Set(valueVal)
		return nil
	}

	// Handle slice types
	if fieldType.Kind() == reflect.Slice {
		return setSliceValue(fieldVal, value)
	}

	// Handle map types
	if fieldType.Kind() == reflect.Map {
		if valueVal.Type().AssignableTo(fieldType) {
			fieldVal.Set(valueVal)
			return nil
		}
		return fmt.Errorf("ctag: cannot convert %T to %v", value, fieldType)
	}

	// Convert from string
	if valueVal.Kind() == reflect.String {
		return setFromString(fieldVal, valueVal.String())
	}

	// Convert between numeric types
	if isNumeric(valueVal.Kind()) && isNumeric(fieldType.Kind()) {
		return setNumericValue(fieldVal, valueVal)
	}

	// Convert any type to string
	if fieldType.Kind() == reflect.String {
		fieldVal.SetString(fmt.Sprintf("%v", value))
		return nil
	}

	return fmt.Errorf("ctag: cannot convert %T to %v", value, fieldType)
}

// setPointerValue handles setting pointer field values
func setPointerValue(fieldVal reflect.Value, value any) error {
	fieldType := fieldVal.Type()
	elemType := fieldType.Elem()

	// Create new pointer if field is nil
	if fieldVal.IsNil() {
		newPtr := reflect.New(elemType)
		fieldVal.Set(newPtr)
	}

	return setValue(fieldVal.Elem(), value)
}

// setSliceValue handles setting slice field values
func setSliceValue(fieldVal reflect.Value, value any) error {
	valueVal := reflect.ValueOf(value)

	// Direct assignment if types match
	if valueVal.Type().AssignableTo(fieldVal.Type()) {
		fieldVal.Set(valueVal)
		return nil
	}

	// Convert string to slice (comma-separated)
	if valueVal.Kind() == reflect.String {
		return setSliceFromString(fieldVal, valueVal.String())
	}

	// Convert single value to slice
	elemType := fieldVal.Type().Elem()
	if valueVal.Type().AssignableTo(elemType) {
		slice := reflect.MakeSlice(fieldVal.Type(), 1, 1)
		slice.Index(0).Set(valueVal)
		fieldVal.Set(slice)
		return nil
	}

	return fmt.Errorf("ctag: cannot convert %T to %v", value, fieldVal.Type())
}

// setSliceFromString converts a comma-separated string to a slice
func setSliceFromString(fieldVal reflect.Value, str string) error {
	if str == "" {
		fieldVal.Set(reflect.MakeSlice(fieldVal.Type(), 0, 0))
		return nil
	}

	parts := strings.Split(str, ",")
	slice := reflect.MakeSlice(fieldVal.Type(), len(parts), len(parts))

	for i, part := range parts {
		part = strings.TrimSpace(part)
		elem := slice.Index(i)
		if err := setValue(elem, part); err != nil {
			return fmt.Errorf("ctag: error converting slice element %d: %w", i, err)
		}
	}

	fieldVal.Set(slice)
	return nil
}

// setFromString converts a string value to the target field type
func setFromString(fieldVal reflect.Value, str string) error {
	switch fieldVal.Kind() {
	case reflect.String:
		fieldVal.SetString(str)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return fmt.Errorf("ctag: cannot parse %q as int: %w", str, err)
		}
		fieldVal.SetInt(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return fmt.Errorf("ctag: cannot parse %q as uint: %w", str, err)
		}
		fieldVal.SetUint(val)
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return fmt.Errorf("ctag: cannot parse %q as float: %w", str, err)
		}
		fieldVal.SetFloat(val)
	case reflect.Bool:
		val, err := strconv.ParseBool(str)
		if err != nil {
			return fmt.Errorf("ctag: cannot parse %q as bool: %w", str, err)
		}
		fieldVal.SetBool(val)
	default:
		return fmt.Errorf("ctag: cannot convert string to %v", fieldVal.Type())
	}
	return nil
}

// setNumericValue converts between numeric types
func setNumericValue(fieldVal reflect.Value, valueVal reflect.Value) error {
	switch fieldVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch valueVal.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fieldVal.SetInt(valueVal.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			fieldVal.SetInt(int64(valueVal.Uint()))
		case reflect.Float32, reflect.Float64:
			fieldVal.SetInt(int64(valueVal.Float()))
		default:
			return fmt.Errorf("ctag: cannot convert %v to %v", valueVal.Type(), fieldVal.Type())
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch valueVal.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fieldVal.SetUint(uint64(valueVal.Int()))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			fieldVal.SetUint(valueVal.Uint())
		case reflect.Float32, reflect.Float64:
			fieldVal.SetUint(uint64(valueVal.Float()))
		default:
			return fmt.Errorf("ctag: cannot convert %v to %v", valueVal.Type(), fieldVal.Type())
		}
	case reflect.Float32, reflect.Float64:
		switch valueVal.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			fieldVal.SetFloat(float64(valueVal.Int()))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			fieldVal.SetFloat(float64(valueVal.Uint()))
		case reflect.Float32, reflect.Float64:
			fieldVal.SetFloat(valueVal.Float())
		default:
			return fmt.Errorf("ctag: cannot convert %v to %v", valueVal.Type(), fieldVal.Type())
		}
	default:
		return fmt.Errorf("ctag: unsupported numeric type %v", fieldVal.Type())
	}
	return nil
}

// isNumeric checks if a reflect.Kind represents a numeric type
func isNumeric(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// getTags is a helper function that recursively fetches and optionally processes tags from struct fields.
func getTags(key string, v reflect.Value, p TagProcessor) (CTags, error) {
	var embedded []reflect.Value
	var tags CTags
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		f := t.Field(i)
		fv := v.Field(i)

		// unexported field
		if f.PkgPath != "" && !f.Anonymous {
			continue
		}

		// dereference pointers
		for fv.Kind() == reflect.Ptr {
			if fv.IsNil() {
				break
			}
			fv = fv.Elem()
		}

		tagStr := f.Tag.Get(key)

		// skip "-", "omitempty" if field is zero value
		if tagStr == "-" || (strings.Contains(tagStr, "omitempty") && fv.IsZero()) {
			continue
		}

		// embedded structs
		if f.Anonymous {
			if fv.IsValid() && fv.Kind() == reflect.Struct {
				embedded = append(embedded, fv)
			}
			continue
		}

		// parse tag and apply processor
		if tagStr != "" {
			tag := parse(key, tagStr, fv)
			if p != nil {
				if err := p.Process(tag.Field, &tag); err != nil {
					return nil, fmt.Errorf("error processing field: %w", err)
				}
			}
			tags = append(tags, tag)
		}

		// nested structs
		if fv.Kind() == reflect.Struct && !f.Anonymous {
			nestedTags, err := getTags(key, fv, p)
			if err != nil {
				return nil, err
			}
			tags = append(tags, nestedTags...)
		}
	}

	// resolve embedded fields
	for _, f := range embedded {
		etags, err := getTags(key, f, p)
		if err != nil {
			return nil, err
		}
		tags = append(tags, etags...)
	}
	return tags, nil
}

// parse converts a raw struct tag string into a CTag struct.
func parse(key string, tagStr string, fv reflect.Value) CTag {
	v := reflect.Indirect(fv)
	tag := CTag{
		Key: key,
	}
	if v.IsValid() {
		tag.Field = v.Interface()
	} else {
		tag.Field = nil
	}
	parts := strings.SplitN(tagStr, ",", 2)
	tag.Name = parts[0]
	if len(parts) > 1 {
		tag.Options = strings.Split(parts[1], ",")
	}
	return tag
}
