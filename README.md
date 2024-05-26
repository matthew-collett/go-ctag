<p align="center">
  <h1 align="center">Go CTag</h1>
  <p align="center">Custom struct tags in Go</p>
  <p align="center"> 
    <a href="https://pkg.go.dev/github.com/matthew-collett/go-ctag/ctag"><img alt="Go Reference" src="https://pkg.go.dev/badge/github.com/matthew-collett/go-ctag.svg"></a>
    <a href="https://github.com/matthew-collett/go-ctag/releases/latest"><img alt="GitHub release" src="https://img.shields.io/github/release/matthew-collett/go-ctag.svg?logo=github&color=red"></a>
    <a href="https://github.com/matthew-collett/go-ctag/actions?workflow=ci"><img alt="Test workflow" src="https://img.shields.io/github/actions/workflow/status/matthew-collett/go-ctag/.github%2Fworkflows%2Fci.yml?label=tests&logo=github"></a>
    <a href="https://github.com/matthew-collett/go-ctag/blob/main/LICENSE"><img alt="License" src="https://img.shields.io/github/license/matthew-collett/go-ctag?label=license&color=yellow"></a>
  </p>
</p>

The `ctag` package provides utilities for extracting and processing custom struct tags in Go. It supports fetching tags based on specific criteria and applying custom processing through user-defined functions.

## Features

- Extract custom tags from struct fields.
- Apply custom processing on fields based on their tags.
- Assert types to field values.
- Filter and find tags based on custom conditions.

## Installation

Install `ctag` using `go get`:

```bash
go get -u github.com/matthew-collett/ctag
```

## Usage

<details>
<summary>Extracting Tags</summary>

 You can extract tags from a struct with or without additional processing:
```go
data := ExampleStruct{
    Field1: "value1",
    Field2: 0,
    Field3: true,
}

tags, err := ctag.GetTags("ctag", data)
if err != nil {
    fmt.Printf("Error: %v\n", err)
} else {
    fmt.Printf("Tags: %+v\n", tags)
}
```
</details>

<details>
<summary>Custom Tag Processing</summary>

### Implement the `TagProcessor` interface to apply custom logic:
```go
type MyProcessor struct{}

func (p *MyProcessor) Process(field any, tag *ctag.CTag) error {
    // Custom processing logic here
    return nil
}

processor := &MyProcessor{}
processedTags, err := ctag.GetTagsAndProcess("ctag", data, processor)
if err != nil {
    fmt.Printf("Error: %v\n", err)
} else {
    fmt.Printf("Processed Tags: %+v\n", processedTags)
}
```
</details>

Take a look at the [GoDoc](https://pkg.go.dev/github.com/matthew-collett/go-ctag) for more details.

## Overview of CTag and CTags

### CTag
The `CTag` struct represents a custom tag associated with a field in a Go struct. It stores information extracted from struct tags which are defined in your Go code. Here's what each field in `CTag` represents:

- **Key**: The primary identifier used to retrieve the tag. It corresponds to the key part of the struct tag.
- **Name**: The first value associated with the key in the tag, typically used to indicate the primary purpose or content.
- **Options**: Additional comma-separated values associated with the key, providing further instructions or modifiers.
- **Field**: The actual data value of the struct field, allowing direct manipulation or examination of the field's content.

Example definition of a struct with tags:

```go
type Request struct {
    IDs []string `body:"text,comma,omitempty"`
}
var request = Request{
    IDs: []string{1, 2, 3}
}
// In this example:
// Key = "body"
// Name = "text"
// Options = []string{"comma","omitempty"}
// Field = []string{1, 2, 3}
```

### CTags
`CTags` is a type that represents a slice of `CTag` structs. It acts as a wrapper around `[]CTag`, providing additional methods to manipulate and process collections of tags efficiently. This type simplifies operations like filtering and finding tags based on specific criteria.

## Contributing
Contributions are welcome! Please feel free to submit a pull request.

## License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
