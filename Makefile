.PHONY: build
build:
	go build -o options-gen cmd/options-gen/main.go

.PHONY: gen
gen:
	go generate generator/generator.go

.PHONY: all
all: gen build

.PHONY: test
test:
	go test -v ./...
