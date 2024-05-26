// Package ctag provides utilities for extracting and processing custom struct tags in Go.
//
// The ctag package offers functionality to retrieve tags from struct fields,
// apply custom processing rules, and filter or find tags based on specific conditions.
//
// Example usage:
//
//	type Request struct {
//	    IDs []string `body:"text,comma,omitempty"`
//	}
//
//	tag, _ := ctag.GetTags("body", Request{IDs: []string{"id1", "id2"}})
//
// Custom processors can implement the TagProcessor interface:
//
//	type MyProcessor struct{}
//
//	func (p *MyProcessor) Process(field any, tag *ctag.CTag) error {
//	    // Custom processing logic here
//	    return nil
//	}
//
//	tags, _ := ctag.GetTagsAndProcess("body", Request{IDs: []string{"id1", "id2"}}, &MyProcessor{})
package ctag
