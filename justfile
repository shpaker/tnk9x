# Go параметры
gocmd := "go"
binary_name := "gonflict"
binary_unix := binary_name + "_unix"
max_line_length := "80"

# Основные команды
default:
    @just --list

build:
    #!/bin/bash
    echo "Building application..."
    {{gocmd}} build -o {{binary_name}} -v ./cmd
    echo "Build completed: {{binary_name}}"

build-linux:
    #!/bin/bash
    echo "Building for Linux..."
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 {{gocmd}} build -o {{binary_unix}} -v ./cmd
    echo "Linux build completed: {{binary_unix}}"

clean:
    #!/bin/bash
    echo "Cleaning build artifacts..."
    {{gocmd}} clean
    rm -f {{binary_name}}
    rm -f {{binary_unix}}
    echo "Clean completed"

test:
    #!/bin/bash
    echo "Running tests..."
    {{gocmd}} test -v ./...

test-coverage:
    #!/bin/bash
    echo "Running tests with coverage..."
    {{gocmd}} test -v -coverprofile=coverage.out ./...
    {{gocmd}} tool cover -html=coverage.out -o coverage.html
    echo "Coverage report generated: coverage.html"

deps:
    #!/bin/bash
    echo "Downloading dependencies..."
    {{gocmd}} mod download
    {{gocmd}} mod tidy
    echo "Dependencies updated"

run: build
    #!/bin/bash
    echo "Running application..."
    ./{{binary_name}}

dev:
    #!/bin/bash
    echo "Running in development mode..."
    {{gocmd}} run ./cmd

# Форматирование
fmt:
    #!/bin/bash
    GOBIN_PATH="$({{gocmd}} env GOPATH)/bin"
    "$GOBIN_PATH/gofumpt" -l -w .
    "$GOBIN_PATH/golines" -w --max-len={{max_line_length}} .

# Линтинг
lint:
    #!/bin/bash
    GOBIN_PATH="$({{gocmd}} env GOPATH)/bin"
    "$GOBIN_PATH/golangci-lint" run

lint-fix:
    #!/bin/bash
    GOBIN_PATH="$({{gocmd}} env GOPATH)/bin"
    "$GOBIN_PATH/golangci-lint" run --fix

# Установка инструментов
install-tools:
    #!/bin/bash
    echo "Installing development tools..."
    {{gocmd}} install mvdan.cc/gofumpt@latest
    {{gocmd}} install github.com/segmentio/golines@latest
    {{gocmd}} install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    echo "Tools installed. Make sure \$GOBIN or \$(go env GOPATH)/bin is in your PATH"

# Проверки качества кода
check: fmt lint test
    #!/bin/bash
    echo "All checks completed successfully"

# Полная сборка с проверками
ci: deps check build
    #!/bin/bash
    echo "CI pipeline completed successfully"

# Отладка
debug: build
    #!/bin/bash
    echo "Running in debug mode..."
    GORACE="history_size=7" ./{{binary_name}}
