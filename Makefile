INTEGRATION  := $(shell basename $(shell pwd))
BINARY_NAME   = nr-$(INTEGRATION)
E2E_BINARY_NAME = $(BINARY_NAME)-e2e
GO_PKGS      := $(shell go list ./... | grep -v "/vendor/")
GO_FILES     := $(shell find src -type f -name "*.go")
GOTOOLS       = github.com/kardianos/govendor \
		gopkg.in/alecthomas/gometalinter.v2 \
		github.com/axw/gocov/gocov \
		github.com/AlekSi/gocov-xml \
		go.datanerd.us/p/ohai/papers-go/... \

all: build

build: clean validate test-nocov compile

clean:
	@echo "=== $(INTEGRATION) === [ clean ]: Removing binaries and coverage file..."
	@rm -rfv bin coverage.xml

tools:
	@echo "=== $(INTEGRATION) === [ tools ]: Installing tools required by the project..."
	@go get $(GOTOOLS)
	@gometalinter.v2 --install

tools-update:
	@echo "=== $(INTEGRATION) === [ tools-update ]: Updating tools required by the project..."
	@go get -u $(GOTOOLS)
	@gometalinter.v2 --install

deps: tools
	@echo "=== $(INTEGRATION) === [ deps ]: Installing package dependencies required by the project..."
	@govendor sync

validate: lint license-check
validate-all: lint-all license-check

lint: deps
	@echo "=== $(INTEGRATION) === [ validate ]: Validating source code running gometalinter..."
	@gometalinter.v2 --config=.gometalinter.json ./...

lint-all: deps
	@echo "=== $(INTEGRATION) === [ validate ]: Validating source code running gometalinter..."
	@gometalinter.v2 --config=.gometalinter.json --enable=interfacer --enable=gosimple ./...

license-check:
	@echo "=== $(INTEGRATION) === [ validate ]: Validating licenses of package dependencies required by the project..."
	@papers-go validate -c ../../.papers_config.yml

compile: deps
	@echo "=== $(INTEGRATION) === [ compile ]: Building $(BINARY_NAME)..."
	@go build -o bin/$(BINARY_NAME) ./src

compile-dev: deps
	@echo "=== $(INTEGRATION) === [ compile-dev ]: Building $(BINARY_NAME) for development environment..."
	@GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME) ./src

deploy-dev: compile-dev
	@echo "=== $(INTEGRATION) === [ deploy-dev ]: Deploying dev container image containing $(BINARY_NAME) in Kubernetes..."
	@skaffold run

test: deps
	@echo "=== $(INTEGRATION) === [ test ]: Running unit tests with coverage (gocov)..."
	@gocov test $(GO_PKGS) | gocov-xml > coverage.xml

test-nocov: deps
	@echo "=== $(INTEGRATION) === [ test ]: Running unit tests..."
	@go test ./...

guard-%:
	@ if [ "${${*}}" = "" ]; then \
		echo "Environment variable $* not set"; \
		exit 1; \
	fi

e2e-compile: deps
	@echo "[ compile E2E binary]: Building $(E2E_BINARY_NAME)..."
	# CGO_ENABLED=0 is needed since the binary is compiled in a non alpine linux.
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/$(E2E_BINARY_NAME) ./e2e/cmd/e2e.go

e2e-compile-only:
	@echo "[ compile E2E binary]: Building $(E2E_BINARY_NAME)..."
	# CGO_ENABLED=0 is needed since the binary is compiled in a non alpine linux.
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/$(E2E_BINARY_NAME) ./e2e/cmd/e2e.go

.PHONY: all build clean tools tools-update deps validate compile test
