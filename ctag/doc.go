// Package ctag provides utilities for extracting and processing custom struct tags in Go.
//
// The ctag package offers functionality to retrieve tags from struct fields,
// apply custom processing rules, and filter or find tags based on specific conditions.
//
// Example usage:
//
// import "github.com/matthew-collett/go-ctag/ctag"
//
//	type Request struct {
//	    IDs []string `body:"text,comma,omitempty"`
//	}
//
//	request := Request{
//	    IDs: []string{"1", "2", "3"}
//	}
//
// tag, _ := ctag.GetTags("body", request)
//
// Custom processors can implement the TagProcessor interface:
//
// type Processor struct{}
//
//	func (p *Processor) Process(field any, tag *ctag.CTag) error {
//	    // Custom processing logic here
//	    return nil
//	}
//
//	request := Request{
//	    IDs: []string{"1", "2", "3"}
//	}
//
// tags, _ := ctag.GetTagsAndProcess("body", request, &Processor{})
package ctag
