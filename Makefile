SHELL := /bin/bash
SERVICE_NAME:=moneypenny
VERSION:=$(shell git describe --exact-match 2> /dev/null || git log -1 --pretty=format:commit_%h)

vet:
	go vet ./cmd/... ./internal/...

lint:
	revive -config lintconfig.toml -formatter friendly cmd/... internal/... tests/...

staticcheck:
	staticcheck ./cmd/... ./internal/... .tests/...

tidy:
	go mod tidy -go=1.25

checks: tidy lint vet staticcheck

run:
	go run cmd/cli/main.go

build: checks
	env GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -v -ldflags "-X main.sha1ver=${VERSION} -X main.buildTime=${NOW} -s -w" -o ./bin/moneypenny ./cmd/cli/.

test: checks
	gotestsum ./internal/... -race -covermode=atomic -test.short -v

git-hooks-install:
	pre-commit install

fmt-imports:
	find -name '*.go' -exec goimports -w {} \;

# Install all dependencies needed to execute make targets
install-deps:
	@if ! command -v gotestsum > /dev/null; then go install gotest.tools/gotestsum@latest; else echo "skipping gotestsum"; fi;
	@if ! command -v revive > /dev/null; then go install github.com/mgechev/revive@latest; else echo "skipping revive"; fi;
	@if ! command -v staticcheck > /dev/null; then go install honnef.co/go/tools/cmd/staticcheck@latest; else echo "skipping staticcheck"; fi;