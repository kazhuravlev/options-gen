# options-gen

[![Go Reference](https://pkg.go.dev/badge/github.com/kazhuravlev/options-gen.svg)](https://pkg.go.dev/github.com/kazhuravlev/options-gen)
[![License](https://img.shields.io/github/license/kazhuravlev/options-gen?color=blue)](https://github.com/kazhuravlev/options-gen/blob/master/LICENSE)
[![Build Status](https://github.com/kazhuravlev/options-gen/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/kazhuravlev/options-gen/actions/workflows/go.yml?query=branch%3Amaster)
[![Go Report Card](https://goreportcard.com/badge/github.com/kazhuravlev/options-gen)](https://goreportcard.com/report/github.com/kazhuravlev/options-gen)
[![CodeCov](https://codecov.io/gh/kazhuravlev/options-gen/branch/master/graph/badge.svg?token=tNKcOjlxLo)](https://codecov.io/gh/kazhuravlev/options-gen)

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
  "io"
  "log"
)

//go:generate options-gen -out-filename=options_generated.go -from-struct=Options
type Options struct {
  logger     log.Logger `option:"mandatory"`
  listenAddr string     `option:"mandatory" validate:"required,hostname_port"`
  closer     io.Closer  `validate:"required"`
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
  // mandatory options. you cannot ignore or forget them because they are arguments.
  logger log.Logger, 
  listenAddr string,
  // optional: you can leave them empty or not.
  other ...Option,
) {
  ...
}

// Validate will check that all options are in desired state
func (o *Options) Validate() error {
  ...
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

## Usage with Generics

```go
package mypkg

//go:generate options-gen -from-struct=Options
type Options[T any] struct {
  addr string   `option:"mandatory" validate:"required,hostname_port"`
  ch   <-chan T `option:"mandatory"`
}
```

And just `go generate ./...`.

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
  Default: `$GOFILE` (file where you placed `//go:generate`).
- `pkg` - name of output filename package. In most cases we can just use
  the same package as the `filename` file. For example `githubclient`.
  Default: `$GOPACKAGE`. Package name same as file where you placed `//go:generate`.
- `from-struct` - name of structure that contains our options. For
  example `Options`.
- `out-filename` - specifies an output filename. This filename will be rewritten
  with options-gen specific content. For
  example `./pkg/github-client/options_generated.go`.

See an [Examples](#Examples).

### Option tag

You can control two important things. The first is about the options constructor
- how `options-gen` will generate `NewOptions` constructor. The second is about
how to validate data, that has been passed as value for this field.

#### Control the constructor

`options-gen` can generate a constructor that can receive all option fields as
separate arguments. It will force the user to pass each (or someone) option
field to the constructor. Like this:

```go
// Mark Field1 as mandatory
type Options struct {
  Field1 string `option:"mandatory"`
}

// options-gen will generate constructor like this
func NewOptions(field1 string, otherOptions ...option)...
```

But, if we do not want to force the user to pass each argument - we can remove
the `option:"mandatory"` feature for this field and get something like this:

```go
// Do not mark Field1 as mandatory
type Options struct {
  Field1 string
}

// options-gen will generate constructor like this
func NewOptions(otherOptions ...option)...
```

So, this allows setting only those options fields that user is want to set.

#### Validate field data

After we define the fields, we want to restrict the values of these fields. To
do that we can use a well-known library [validator](https://github.com/go-playground/validator)

Just read the docs for `validator` library and add tag to fields like this:

```go
type Options struct {
  maxDbConn int `validate:"required,min=1,max=16"`
}
```

### Custom validator

You can override `options-gen` validator for specific struct by implementing the `Validator()` method:
```go
import "github.com/mycoolmodule/internal/validator"

// ...

func (Options) Validator() *validator.Validate {
  return validator.Validator
}
```

Or you can override `options-gen` validator globally:
```go
package validator

import (
  goplvalidator "github.com/go-playground/validator/v10"
  optsValidator "github.com/kazhuravlev/options-gen/pkg/validator"
)

var Validator = goplvalidator.New()

func init() {
  must(Validator.RegisterValidation(/* ... */))
  must(Validator.RegisterAlias(/* ... */))

  optsValidator.Validator = Validator
}
```

## Contributing

The development process is pretty simple:

- [Fork](https://docs.github.com/en/get-started/quickstart/fork-a-repo) the repo
  on GitHub
- [Clone](https://docs.github.com/en/get-started/quickstart/fork-a-repo#cloning-your-forked-repository)
  your copy of the repo
- [Create a new branch](https://git-scm.com/book/en/v2/Git-Branching-Basic-Branching-and-Merging)
  for your goals
- Install the [Task](https://taskfile.dev/installation/). It's like `Make`, but
  simple
- Check that your working copy is ready to start development by
  running `task check` in repo workdir
- Reach your goals!
- Check that all is ok by `task check`
- [Create](https://docs.github.com/en/pull-requests/collaborating-with-pull-requests/proposing-changes-to-your-work-with-pull-requests/creating-a-pull-request)
  a Pull Request
