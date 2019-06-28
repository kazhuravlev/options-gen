# options

## Usage

```go
//go:generate options-gen -filename=$GOFILE -out-filename=options_generated.go -pkg=mypkg
type Options struct {
	logFactory log.Factory `option:"required"`
	listenAddr string      `option:"required,not-empty"`
	redis      IRedis      `option:"not-empty"`
}
```
