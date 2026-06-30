BINARY := dev
PREFIX ?= $(HOME)/.local
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo 0.1.0)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build install uninstall test vet fmt clean run

build: ## Build the dev binary into ./bin
	@mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/dev
	@echo "built bin/$(BINARY) ($(VERSION))"

install: ## Build and install to $(PREFIX)/bin
	@mkdir -p $(PREFIX)/bin
	go build -ldflags "$(LDFLAGS)" -o $(PREFIX)/bin/$(BINARY) ./cmd/dev
	@echo "installed $(PREFIX)/bin/$(BINARY)"
	@echo "make sure $(PREFIX)/bin is on your PATH"

uninstall: ## Remove the installed binary
	rm -f $(PREFIX)/bin/$(BINARY)

test: ## Run tests
	go test ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Format the code
	gofmt -w .

release: ## Cross-compile binaries for common OS/arch into ./dist
	@mkdir -p dist
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-darwin-arm64  ./cmd/dev
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-darwin-amd64  ./cmd/dev
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-linux-arm64   ./cmd/dev
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-linux-amd64   ./cmd/dev
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-windows-amd64.exe ./cmd/dev
	@echo "built:" && ls -1 dist

clean: ## Remove build artifacts
	rm -rf bin dist

run: build ## Build then run (use ARGS=...)
	./bin/$(BINARY) $(ARGS)
