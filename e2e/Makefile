BINARY_NAME = nr-kubernetes-e2e
GOTOOLS     = github.com/kardianos/govendor

all: compile

compile: deps
	@echo "[ compile ]: Building $(BINARY_NAME)..."
	@go build -o bin/$(BINARY_NAME) ./cmd/e2e.go

deps: tools
	@echo "[ deps ]: Sync dependencies using govendor"
	@govendor sync

tools:
	@echo "[ tools ]: Installing tools required by the project..."
	@go get $(GOTOOLS)

.PHONY: all compile deps tools