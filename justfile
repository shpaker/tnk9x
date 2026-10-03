# Go параметры
gocmd := "go"
binary_name := "tnk9x"
binary_unix := binary_name + "_unix"
max_line_length := "80"

# Версия сборки: APP_VERSION из CI или описание коммита от последнего тега
version := env("APP_VERSION", `git describe --tags --always --dirty --exclude '*a*' 2>/dev/null || echo dev`)

# Релизные теги — только MAJOR.MINOR; альфы вида 0.1a1 не учитываются
release_tag := '^[0-9]+\.[0-9]+$'
# Типы коммитов, поднимающие минорную версию, в порядке changelog
release_types := "feat fix perf refactor"
# Признаки ломающего изменения: `type!:` в теме или футер BREAKING CHANGE
breaking_subject := '^[a-z]+(\([^)]*\))?!:'
breaking_body := '^BREAKING[ -]CHANGE:'

# Основные команды
default:
    @just --list

build:
    #!/bin/bash
    set -euo pipefail
    VERSION="{{version}}"
    echo "Building application (version $VERSION)..."
    {{gocmd}} build -ldflags "-X github.com/shpaker/tnk9x/internal/app.Version=${VERSION}" -o {{binary_name}} -v ./cmd
    echo "Build completed: {{binary_name}} (version $VERSION)"

build-macos:
    #!/bin/bash
    set -euo pipefail
    out_dir="_build/macos"
    rm -rf "$out_dir"
    mkdir -p "$out_dir"
    VERSION="{{version}}"
    release_output="{{binary_name}}_darwin_arm64"
    echo "Building macOS (Apple Silicon) release $VERSION -> $out_dir/$release_output"
    GOOS="darwin" GOARCH="arm64" CGO_ENABLED=1 {{gocmd}} build -trimpath -ldflags "-s -w -X github.com/shpaker/tnk9x/internal/app.Version=${VERSION}" -o "$out_dir/$release_output" ./cmd
    echo "macOS builds stored in $out_dir"

build-windows:
    #!/bin/bash
    set -euo pipefail
    out_dir="_build/windows"
    rm -rf "$out_dir"
    mkdir -p "$out_dir"
    VERSION="{{version}}"
    release_output="{{binary_name}}_windows_amd64.exe"
    echo "Building Windows (x64) release $VERSION -> $out_dir/$release_output"
    GOOS="windows" GOARCH="amd64" CGO_ENABLED=0 {{gocmd}} build -trimpath -ldflags "-s -w -X github.com/shpaker/tnk9x/internal/app.Version=${VERSION}" -o "$out_dir/$release_output" ./cmd
    echo "Windows builds stored in $out_dir"

build-all: build-macos build-windows
    #!/bin/bash
    echo "macOS и Windows сборки готовы"

# Файлы web/<target>/ кладутся поверх общих web/common/
# Веб-сборка под площадку: pages (GitHub Pages) или yandex (Яндекс Игры)
build-web target="pages":
    #!/bin/bash
    set -euo pipefail
    if [ ! -d "web/{{target}}" ]; then
        echo "Unknown web target '{{target}}': web/{{target}}/ not found"
        exit 1
    fi
    out_dir="dist/{{target}}"
    rm -rf "$out_dir"
    mkdir -p "$out_dir"
    VERSION="{{version}}"
    echo "Building web/{{target}} (version $VERSION) -> $out_dir"
    GOOS="js" GOARCH="wasm" {{gocmd}} build -trimpath -ldflags "-s -w -X github.com/shpaker/tnk9x/internal/app.Version=${VERSION}" -o "$out_dir/{{binary_name}}.wasm" ./cmd
    cp "$({{gocmd}} env GOROOT)/lib/wasm/wasm_exec.js" "$out_dir/"
    cp web/common/* "$out_dir/"
    cp web/{{target}}/* "$out_dir/"
    echo "Web build stored in $out_dir"

# Локальный сервер веб-сборки площадки
serve-web target="pages": (build-web target)
    #!/bin/bash
    set -euo pipefail
    py=python3
    "$py" -c "" 2>/dev/null || py=python
    echo "Serving dist/{{target}} at http://localhost:8000 ..."
    "$py" -m http.server 8000 -d "dist/{{target}}"

# Ключей и ID игры сборке не нужно: архив загружается в консоль вручную
# Архив для консоли Яндекс Игр с index.html в корне
package-yandex:
    #!/bin/bash
    set -euo pipefail
    export APP_VERSION="{{version}}"
    "{{just_executable()}}" build-web yandex
    mkdir -p _build
    archive="_build/{{binary_name}}_yandex_${APP_VERSION}.zip"
    rm -f "$archive"
    py=python3
    "$py" -c "" 2>/dev/null || py=python
    (cd dist/yandex && "$py" -m zipfile -c "../../$archive" *)
    echo "Archive created: $archive"

# ID игры не хранится в репозитории: YANDEX_APP_ID из окружения
# или локального .env; без него — dev-режим с моками SDK
# Запуск сборки Яндекса через sdk-dev-proxy (нужен Node.js)
serve-yandex: (build-web "yandex")
    #!/bin/bash
    set -euo pipefail
    if [ -f .env ]; then
        set -a
        . ./.env
        set +a
    fi
    if [ -n "${YANDEX_APP_ID:-}" ]; then
        npx @yandex-games/sdk-dev-proxy -p dist/yandex --app-id="$YANDEX_APP_ID"
    else
        npx @yandex-games/sdk-dev-proxy -p dist/yandex --dev-mode=true
    fi

package-macos:
    #!/bin/bash
    set -euo pipefail
    out_dir="_build/macos"
    if [ ! -f "$out_dir/{{binary_name}}_darwin_arm64" ]; then
        echo "Error: macOS binary not found. Run 'just build-macos' first."
        exit 1
    fi
    VERSION="{{version}}"
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
    VERSION="{{version}}"
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

# Порог покрытия ядра игровой логики (internal/use_cases)
test-coverage-check:
    #!/bin/bash
    set -euo pipefail
    threshold=70
    {{gocmd}} test -coverprofile=coverage-use-cases.out ./internal/use_cases/...
    total=$({{gocmd}} tool cover -func=coverage-use-cases.out | awk '/^total:/ {gsub("%", "", $3); print $3}')
    rm -f coverage-use-cases.out
    echo "internal/use_cases coverage: ${total}% (threshold ${threshold}%)"
    awk -v total="$total" -v threshold="$threshold" 'BEGIN { exit (total < threshold) ? 1 : 0 }'

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
    VERSION="{{version}}"
    echo "Running in development mode (version $VERSION)..."
    {{gocmd}} run -ldflags "-X github.com/shpaker/tnk9x/internal/app.Version=${VERSION}" ./cmd

# Версионирование по Conventional Commits
# Пустой вывод — с последнего релиза нет релизных коммитов
# Следующая версия MAJOR.MINOR по коммитам с последнего релизного тега
next-version:
    #!/bin/bash
    set -euo pipefail
    last=$(git tag --merged HEAD | grep -E '{{release_tag}}' | sort -V | tail -n1 || true)
    range="${last:+$last..}HEAD"
    subjects=$(git log --format='%s' "$range")
    bodies=$(git log --format='%b' "$range")
    types='{{release_types}}'
    major="${last%%.*}"
    minor="${last#*.}"
    if [ -z "$last" ]; then
        major=0
        minor=0
    fi
    if grep -qE '{{breaking_subject}}' <<< "$subjects" \
        || grep -qE '{{breaking_body}}' <<< "$bodies"; then
        echo "$((major + 1)).0"
    elif grep -qE "^(${types// /|})(\([^)]*\))?:" <<< "$subjects"; then
        echo "${major}.$((minor + 1))"
    fi

# Только релизные коммиты: ломающие первыми, затем по типам
# Changelog релиза ref относительно предыдущего релизного тега
changelog ref="HEAD":
    #!/bin/bash
    set -euo pipefail
    ref='{{ref}}'
    last=$(git tag --merged "$ref" --no-contains "$ref" | grep -E '{{release_tag}}' | sort -V | tail -n1 || true)
    subjects=$(git log --reverse --format='%s' "${last:+$last..}$ref")
    {
        grep -E '{{breaking_subject}}' <<< "$subjects" || true
        for type in {{release_types}}; do
            grep -E "^${type}(\([^)]*\))?:" <<< "$subjects" || true
        done
    } | sed 's/^/- /'

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

# Файлы с тегом js видны линтеру только при GOOS=js
# Линтинг js/wasm-сборки
lint-wasm:
    #!/bin/bash
    GOBIN_PATH="$({{gocmd}} env GOPATH)/bin"
    GOOS=js GOARCH=wasm "$GOBIN_PATH/golangci-lint" run

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
