# options-gen

[![Go Reference](https://pkg.go.dev/badge/github.com/kazhuravlev/options-gen.svg)](https://pkg.go.dev/github.com/kazhuravlev/options-gen)
[![License](https://img.shields.io/github/license/kazhuravlev/options-gen?color=blue)](https://github.com/kazhuravlev/options-gen/blob/master/LICENSE)
[![Build Status](https://github.com/kazhuravlev/options-gen/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/kazhuravlev/options-gen/actions/workflows/go.yml?query=branch%3Amaster)
[![Go Report Card](https://goreportcard.com/badge/github.com/kazhuravlev/options-gen)](https://goreportcard.com/report/github.com/kazhuravlev/options-gen)
[![CodeCov](https://codecov.io/gh/kazhuravlev/options-gen/branch/master/graph/badge.svg?token=tNKcOjlxLo)](https://codecov.io/gh/kazhuravlev/options-gen)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge.svg)](https://github.com/avelino/awesome-go#go-generate-tools)

Code-generator that allows you to create a functional options like
[Dave Cheney's post](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis).

Generate the options for your service/client/etc. All that you need is to define a
struct with fields, that can be applied as Option then embed this struct into
yours.

## Installation

```bash
go install github.com/kazhuravlev/options-gen/cmd/options-gen@latest
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
	if err := opts.Validate(); err != nil {    // always add only these lines for all your constructors
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

## More Examples

### Real-world HTTP Client

```go
//go:generate options-gen -from-struct=Options
type Options struct {
	httpClient  *http.Client  `option:"mandatory"`
	baseURL     string        `option:"mandatory" validate:"required,url"`
	token       string        `validate:"required"`
	timeout     time.Duration `default:"30s" validate:"min=1s,max=5m"`
	maxRetries  int           `default:"3" validate:"min=0,max=10"`
	userAgent   string        `default:"options-gen/1.0"`
}

type Client struct {
	opts Options
}

func NewClient(opts Options) (*Client, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validate options: %w", err)
	}
	
	return &Client{opts: opts}, nil
}

// Usage:
client, err := NewClient(NewOptions(
	httpClient,
	"https://api.example.com",
	WithToken("secret-token"),
	WithTimeout(10*time.Second),
))
```

### Field Exclusion

```go
//go:generate options-gen -from-struct=Options -exclude="internal*;debug*"
type Options struct {
	addr         string `option:"mandatory" validate:"required,hostname_port"`
	internalConn net.Conn // excluded: matches "internal*"
	debugMode    bool     // excluded: matches "debug*"
	logLevel     string   // included
}
```

### Custom Option Type Name

```go
//go:generate options-gen -from-struct=Options -out-setter-name=HTTPOption
type Options struct {
	timeout time.Duration
	headers map[string]string
}

// Generated code will use HTTPOption instead of OptOptionsSetter:
// type HTTPOption func(*Options)
```

See more [examples](./examples) for additional use cases.

## Configuration

To configure this tool you should know two things: how to work with cli tool
and how to define options in your `Options` struct.

### CLI tool

All the tool needs is the information about source and target files and packages.
Tool can be invoked by `options-gen` (after [Installation](#Installation)) and
it will have the following arguments:

- `filename` - is a source filename that contains `Options` struct relative to the current dir. For example
  `./pkg/github-client/options.go`.

  Default: `$GOFILE` (file where you placed `//go:generate`).
- `pkg` - name of output filename package. In most cases we can just use the same package as the `filename` file. For
  example `githubclient`.

  Default: `$GOPACKAGE`. Package name same as file where you placed `//go:generate`.
- `from-struct` - name of structure that contains our options. For example `Options`.
- `out-filename` - specifies an output filename. This filename will be rewritten with options-gen specific content. For
  example `./pkg/github-client/options_generated.go`.

  Default: `$GOFILE_generated.go`
- `all-variadic` - generate variadic functions for all fields with slice type.

  Default: `false` - functions that accept a slice are generated.
- `with-isset` - generate additional method `IsSet(field optField) bool` that allows checking whether a field was
  explicitly set.

  Default: `false`. Will not produce the `IsSet` functions.
- `constructor` - specifies the type and whether to generate a function to build your structure with parameters.
  Possible values:
    - `public` - generate public constructor
    - `private` - generate private constructor
    - `no` - not generate any constructor

  Default: `public`
- `defaults-from` - specifies how default values are determined for option fields. Possible values:
    - `tag[=TagName]` - use tag values (default TagName is `default`)
    - `var[=VariableName]` - use variable of Options type (default VariableName is `default<StructName>`)
    - `func[=FunctionName]` - use function that returns Options (default FunctionName is `getDefault<StructName>`)
    - `none` - disable defaults

  Default: `tag=default`
- `mute-warnings` - suppress warning messages during code generation.

  Default: `false` - warnings are displayed
- `out-prefix` - add prefix to the generated file. Useful when you have multiple Options structs in the same package.

  Default: empty string
- `out-setter-name` - name for the option setter type (function alias).
  If not specified, the `Opt[StructName]Setter` template is used.

  Default: `Opt[StructName]Setter`

- `exclude` - list of masks for field names excluded from generation, semicolon-separated

  Default: ''

### Using out-prefix for multiple Options structs

When you have multiple option structs in the same package, the `out-prefix` flag helps avoid naming conflicts:

```go
// client_options.go
//go:generate options-gen -from-struct=ClientOptions -out-prefix=Client -out-filename=client_options_generated.go
type ClientOptions struct {
	timeout time.Duration `option:"mandatory"`
	retries int          `default:"3"`
}

// server_options.go  
//go:generate options-gen -from-struct=ServerOptions -out-prefix=Server -out-filename=server_options_generated.go
type ServerOptions struct {
	listenAddr string `option:"mandatory" validate:"required,hostname_port"`
	maxConns   int    `default:"100"`
}

// Generated functions will be:
// - NewClientOptions() and ClientOption
// - NewServerOptions() and ServerOption
```

You can find a complete example in [this directory](./examples/go-generate-2options-1pkg/).

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
	field1 string `option:"mandatory"`
}

// options-gen will generate constructor like this
func NewOptions(field1 string, otherOptions ...option)...
```

But, if we do not want to force the user to pass each argument - we can remove
the `option:"mandatory"` feature for this field and get something like this:

```go
// Do not mark Field1 as mandatory
type Options struct {
	field1 string
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

#### Default values

`options-gen` provide several ways to define defaults for options. You can
choose which mechanism you need by providing a flag `-defaults-from`. By
default, this flag is set to `tag=default`.

- `tag[=TagName]`. This mechanism will try to find a tag `TagName` in field
  tags. By default `TagName` is equal to `default`
- `var[=VariableName]`. This mechanism will copy variable `VariableName` fields
  to your `Options` instance. By default `VariableName` is equal
  to `default<StructName>`. This variable should contain `Options` struct.
- `func[=FunctionName]`. The same as `var`, but for the function name.
  Function `FunctionName` will be called once per `NewOptions` constructor. This
  function should return an `Options` struct.
- `none` to disable defaults.

##### Using tag

For numbers, strings, and `time.Duration` you can set the default value:

```go
// simple example
//go:generate options-gen -from-struct=Options
type Options struct {
	pingPeriod  time.Duration `default:"3s" validate:"min=100ms,max=30s"`
	name        string        `default:"unknown" validate:"required"`
	maxAttempts int           `default:"10" validate:"min=1,max=10"`
	eps         float32       `default:"0.0001" validate:"gt=0"`
}
```

```go
// custom default tag
//go:generate options-gen -from-struct=Options --default-from=tag=mydefaulttag
type Options struct {
	pingPeriod  time.Duration `mydefaulttag:"3s" validate:"min=100ms,max=30s"`
	name        string        `mydefaulttag:"unknown" validate:"required"`
	maxAttempts int           `mydefaulttag:"10" validate:"min=1,max=10"`
	eps         float32       `mydefaulttag:"0.0001" validate:"gt=0"`
}
```

It would be relevant if the field were not filled either explicitly or through
functional option.

The default value must be valid for the field type and must satisfy validation
rules.

##### Using variable

Tags allow you to define defaults for simple types like `string`, `number`
, `time.Duration`. When you want to define a variable with prefilled values -
you can do this like that:

```go
// simple example
//go:generate options-gen -from-struct=Options -defaults-from=var
type Options struct {
	httpClient *http.Client
}

var defaultOptions = Options{
	httpClient: &http.Client{},
}
```

```go
// custom variable name
//go:generate options-gen -from-struct=Options -defaults-from=var=myDefaults
type Options struct {
	httpClient *http.Client
}

var myDefaults = Options{
	httpClient: &http.Client{},
}
```

##### Using function

The same as variable. See an examples:

```go
// simple example
//go:generate options-gen -from-struct=Options -defaults-from=func
type Options struct {
	httpClient *http.Client
}

func getDefaultOptions() Options {
	return Options{
		httpClient: &http.Client{},
	}
}
```

```go
// custom function name
//go:generate options-gen -from-struct=Options -defaults-from=func=myDefaults
type Options struct {
	httpClient *http.Client
}

func myDefaults() Options {
	return Options{
		httpClient: &http.Client{},
	}
}
```

##### Disable defaults

If you want to be sure that defaults will not be parsed - you can specify
the `none` for `-defaults-from` flag.

```go
// defaults will now be parsed at all
//go:generate options-gen -from-struct=Options -defaults-from=none
type Options struct {
	name string `default:"joe"`
}
```

#### Which fields are set?

`options-gen` can produce additional code that allows you to check which fields were set. To do this, simply add
the `-with-isset` flag to `options-gen`.

For example, this code with the specified option...

```go
package app

//go:generate options-gen -from-struct=Options -with-isset
type Options struct {
	name string
}
```

...will produce function `func (o *Options) IsSet(field optField) bool{...}`.

You can then use it to check if a field was explicitly set:

```go
opts := NewOptions(WithName("alice"))
if opts.IsSet(optFieldName) {
	// name was explicitly set to "alice"
}
``` 

### Custom validator

You can override `options-gen` validator for specific struct by implementing
the `Validator()` method:

```go
import (
	"github.com/go-playground/validator/v10"
)

type Options struct {
	age      int    `validate:"adult"`
	username string `validate:"required,alphanum"`
}

func (Options) Validator() *validator.Validate {
	v := validator.New()
	v.RegisterValidation("adult", func(fl validator.FieldLevel) bool {
		return fl.Field().Int() >= 18
	})
	return v
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
	must(Validator.RegisterValidation( /* ... */))
	must(Validator.RegisterAlias( /* ... */))

	optsValidator.Set(Validator)
}
```

### Variadic setters

You can generate variadic functions for slice type variables. By default, functions that accept a slice are generated.

To generate variadic functions, you can use the `-all-variadic=true` option or specify tag `option:"variadic=true"` for
specific fields. You can also generate a fallback variadic function when `-all-variadic=true` is included using the
`option:"variadic=false"` tag.

```go
//go:generate options-gen -from-struct=Options
type Options struct {
	// Default: accepts []string
	tags []string
	// Variadic: accepts multiple string arguments
	labels []string `option:"variadic=true"`
}

// Usage:
opts := NewOptions(
	WithTags([]string{"api", "v2"}),
	WithLabels("prod", "us-east-1", "critical"),
)
```

### Skip fields

If you don't need to generate a setter for a specific field, you can specify this using the tag `option:"-"`.

```go
type Options struct {
	name string `option:"-"`
}
```

### Generating Options for External Package Structs

While you cannot directly use `--from-struct=github.com/user/pkg.StructName`, you can generate options for structs from external packages using type aliases:

```go
package myapp

import (
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

// Create a type alias for the external struct
type WebsocketOptions = websocket.Dialer

//go:generate options-gen -from-struct=WebsocketOptions
// This will generate setters for all exported fields from websocket.Dialer

// You can also alias and extend external structs
type RedisOptions redis.Options

//go:generate options-gen -from-struct=RedisOptions
// This generates setters for all fields from redis.Options

// Example with generic external types
type CacheOptions[T any] = genericpkg.Cache[T]

//go:generate options-gen -from-struct=CacheOptions
```

**Important notes about external structs:**
- Only exported (public) fields from the external struct will have setters generated
- The external package must be available during `go generate` 
- Generated imports are automatically managed
- Validation tags from the original struct are not preserved (you'd need to wrap the struct instead)

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
