BIN_NAME=azctl

.PHONY: build test lint release clean

build:
	GO111MODULE=on go build -o bin/$(BIN_NAME) ./cmd/azctl

test:
	go test ./...

lint:
	go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping linting"; \
	fi

release:
	GOOS=linux GOARCH=amd64 go build -o dist/$(BIN_NAME)_linux_amd64 ./cmd/azctl
	GOOS=darwin GOARCH=amd64 go build -o dist/$(BIN_NAME)_darwin_amd64 ./cmd/azctl
	GOOS=darwin GOARCH=arm64 go build -o dist/$(BIN_NAME)_darwin_arm64 ./cmd/azctl
	GOOS=windows GOARCH=amd64 go build -o dist/$(BIN_NAME)_windows_amd64.exe ./cmd/azctl

clean:
	rm -rf bin dist


