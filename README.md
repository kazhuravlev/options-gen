# options-gen

Generate the options for your service/client/etc. All that you need is to define a
struct with fields, that can be applied as Option then embed this struct into
yours.

## Installation

```bash
go install github.com/kazhuravlev/options-gen/cmd/options-gen
```

## Usage

```go
package mypkg

import (
	"errors"
	"io"
	"log"
)

var ErrInvalidOption = errors.New("invalid option")

//go:generate options-gen -filename=$GOFILE -out-filename=options_generated.go -pkg=mypkg -from-struct=Options
type Options struct {
	logger     log.Logger `option:"required"`
	listenAddr string     `option:"required,not-empty"`
	closer     io.Closer  `option:"not-empty"`
}
```

```bash
go generate ./...
```

This will generate `out-filename` file with options constructor. Like this:

```go
// options_generated.go
package mypkg

import (
	"log"
)

func NewOptions(
// required options. you cannot ignore or forget them because they are 
//  arguments.
	logger log.Logger, listenAddr string,

// optional: you can leave them empty or not.
	other ...Option,
) {
	// ...
}

// Validate will check that all options are in desired state
func (o *Options) Validate() error {
	// ...
}
```

And you can use generated options as follows:

```go
package mypkg

import "fmt"

type Component struct {
	opts Options // struct that you define as struct with options 
}

func New(opts Options) (*Component, error) { // constructor of your service/client/component
	if err := opts.Validate(); err != nil {  // always add only these lines for all your constructors
		return nil, fmt.Errorf("cannot validate options: %w", err)
	}

	return &Component{opts: opts}, nil // embed options into your component
}
```

And after that you can use new constructor in (for ex.) `main.go`:

```go
package main

func main() {
	c, err := mypkg.New(mypkg.NewOptions( /* ... */))
	if err != nil {
		panic(err)
	}
}
```

## Examples

See an [examples](./examples) to get real-world examples.

## Configuration

To configure this tool you should know two things: how to work with cli tool
and how to define options in your `Options` struct.

### CLI tool

All the tool needs is the information about source and target files and packages.
Tool can be invoked by `options-gen` (after [Installation](#Installation)) and 
it will have the following arguments:

- `filename` - is a source filename that contains `Options` struct relative
  to the current dir. For example `./pkg/github-client/options.go`.
- `from-struct` - name of structure that contains our options. For
  example `Options`.
- `out-filename` - specifies an output filename. This filename will be rewritten
  with options-gen specific content. For
  example `./pkg/github-client/options_generated.go`.
- `pkg` - name of output filename package. In most cases we can just use
  the same package as the `filename` file. For example `githubclient`.

See an [Examples](#Examples).

### Option tag

To define which options should be detected by options-gen and which of them
should be `required` you can use special field tag, named `option`.

```go
type Options struct {
    // this option should be present and should not be empty. 
    MyOption string `option:"required,not-empty"`
    // this option should be present but can be empty.
    MyOption string `option:"required"`
}
```
