# options

## Install cli

```go
go install github.com/kazhuravlev/options-gen/cmd/options-gen
```

## Usage

```go
package mypkg

import (
    "io"
    "log"
)

//go:generate options-gen -filename=$GOFILE -out-filename=options_generated.go -pkg=mypkg -from-struct=Options
type Options struct {
	logFactory log.Logger `option:"required"`
	listenAddr string     `option:"required,not-empty"`
	closer     io.Closer  `option:"not-empty"`
}
```
