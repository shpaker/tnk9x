# Go параметры
gocmd := "go"
binary_name := "tnk9x"
binary_unix := binary_name + "_unix"
max_line_length := "80"

# Основные команды
default:
    @just --list

build:
    #!/bin/bash
    set -euo pipefail
    VERSION="dev-$(date -u +%Y-%m-%dT%H:%M)"
    echo "Building application (version $VERSION)..."
    {{gocmd}} build -ldflags "-X github.com/shpaker/tnk9x/internal.Version=${VERSION}" -o {{binary_name}} -v ./cmd
    echo "Build completed: {{binary_name}} (version $VERSION)"

build-macos:
    #!/bin/bash
    set -euo pipefail
    out_dir="_build/macos"
    rm -rf "$out_dir"
    mkdir -p "$out_dir"
    VERSION="dev-$(date -u +%Y-%m-%dT%H:%M)"
    release_output="{{binary_name}}_darwin_arm64"
    debug_output="{{binary_name}}_darwin_arm64_debug"
    echo "Building macOS (Apple Silicon) release $VERSION -> $out_dir/$release_output"
    GOOS="darwin" GOARCH="arm64" CGO_ENABLED=1 {{gocmd}} build -trimpath -ldflags "-s -w -X github.com/shpaker/tnk9x/internal.Version=${VERSION}" -o "$out_dir/$release_output" ./cmd
    echo "Building macOS (Apple Silicon) debug $VERSION -> $out_dir/$debug_output"
    GOOS="darwin" GOARCH="arm64" CGO_ENABLED=1 {{gocmd}} build -ldflags "-X github.com/shpaker/tnk9x/internal.Version=${VERSION} -X github.com/shpaker/tnk9x/internal.DebugFlag=true" -o "$out_dir/$debug_output" ./cmd
    echo "macOS builds stored in $out_dir"

build-windows:
    #!/bin/bash
    set -euo pipefail
    out_dir="_build/windows"
    rm -rf "$out_dir"
    mkdir -p "$out_dir"
    VERSION="dev-$(date -u +%Y-%m-%dT%H:%M)"
    release_output="{{binary_name}}_windows_amd64.exe"
    debug_output="{{binary_name}}_windows_amd64_debug.exe"
    echo "Building Windows (x64) release $VERSION -> $out_dir/$release_output"
    GOOS="windows" GOARCH="amd64" CGO_ENABLED=0 {{gocmd}} build -trimpath -ldflags "-s -w -X github.com/shpaker/tnk9x/internal.Version=${VERSION}" -o "$out_dir/$release_output" ./cmd
    echo "Building Windows (x64) debug $VERSION -> $out_dir/$debug_output"
    GOOS="windows" GOARCH="amd64" CGO_ENABLED=0 {{gocmd}} build -ldflags "-X github.com/shpaker/tnk9x/internal.Version=${VERSION} -X github.com/shpaker/tnk9x/internal.DebugFlag=true" -o "$out_dir/$debug_output" ./cmd
    echo "Windows builds stored in $out_dir"

build-all: build-macos build-windows
    #!/bin/bash
    echo "macOS и Windows сборки готовы"

package-macos:
    #!/bin/bash
    set -euo pipefail
    out_dir="_build/macos"
    if [ ! -f "$out_dir/{{binary_name}}_darwin_arm64" ]; then
        echo "Error: macOS binary not found. Run 'just build-macos' first."
        exit 1
    fi
    VERSION="dev-$(date -u +%Y-%m-%dT%H:%M)"
    archive_name="{{binary_name}}_darwin_arm64_${VERSION}.tar.gz"
    echo "Creating macOS archive: $archive_name"
    cd "$out_dir"
    cp ../README.md .
    cp ../LICENSE .
    tar -czf "$archive_name" {{binary_name}}_darwin_arm64 README.md LICENSE
    rm README.md LICENSE
    mv "$archive_name" ..
    echo "Archive created: _build/$archive_name"

package-windows:
    #!/bin/bash
    set -euo pipefail
    out_dir="_build/windows"
    if [ ! -f "$out_dir/{{binary_name}}_windows_amd64.exe" ]; then
        echo "Error: Windows binary not found. Run 'just build-windows' first."
        exit 1
    fi
    VERSION="dev-$(date -u +%Y-%m-%dT%H:%M)"
    archive_name="{{binary_name}}_windows_amd64_${VERSION}.zip"
    echo "Creating Windows archive: $archive_name"
    cd "$out_dir"
    cp ../README.md .
    cp ../LICENSE .
    zip -q "$archive_name" {{binary_name}}_windows_amd64.exe README.md LICENSE
    rm README.md LICENSE
    mv "$archive_name" ..
    echo "Archive created: _build/$archive_name"

package-all: package-macos package-windows
    #!/bin/bash
    echo "All packages created in _build/"

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
    set -euo pipefail
    VERSION="dev-$(date -u +%Y-%m-%dT%H:%M)"
    echo "Running in development mode (version $VERSION)..."
    {{gocmd}} run -ldflags "-X github.com/shpaker/tnk9x/internal.Version=${VERSION}" ./cmd

# Форматирование
fmt:
    #!/bin/bash
    GOBIN_PATH="$({{gocmd}} env GOPATH)/bin"
    "$GOBIN_PATH/gofumpt" -l -w .
    "$GOBIN_PATH/golines" -w --max-len={{max_line_length}} .

fmt-check:
    #!/bin/bash
    echo "Checking code formatting..."
    GOBIN_PATH="$({{gocmd}} env GOPATH)/bin"
    "$GOBIN_PATH/gofumpt" -l .
    "$GOBIN_PATH/golines" -l --max-len={{max_line_length}} .
    echo "Formatting check passed"

# Линтинг
lint:
    #!/bin/bash
    GOBIN_PATH="$({{gocmd}} env GOPATH)/bin"
    "$GOBIN_PATH/golangci-lint" run

lint-notests:
    #!/bin/bash
    GOBIN_PATH="$({{gocmd}} env GOPATH)/bin"
    "$GOBIN_PATH/golangci-lint" run --tests=false

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
    {{gocmd}} install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
    echo "Tools installed. Make sure \$GOBIN or \$(go env GOPATH)/bin is in your PATH"

# Проверки качества кода
check:
    #!/bin/bash
    echo "Running code quality checks..."
    @just fmt-check
    @just lint
    @just test
    echo "All checks completed successfully"

# Отладка
debug: build
    #!/bin/bash
    echo "Running in debug mode..."
    GORACE="history_size=7" ./{{binary_name}}
