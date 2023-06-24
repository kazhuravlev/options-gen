## CLI example

Code you write is - `client.go` and `options.go`. To
generate `options_generated.go` you can just `go generate` like that:

```bash
go install github.com/kazhuravlev/options-gen/cmd/options-gen@latest

git clone git@github.com:kazhuravlev/options-gen.git
cd options-gen/examples/go-generate-generics

go generate ./...
```
