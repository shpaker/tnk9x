// Package tnk9x содержит встроенные ресурсы игры (assets и config.yml)
// для самодостаточных сборок: wasm и запуск бинарника без файлов рядом.
package tnk9x

import "embed"

// FS — встроенная копия каталога assets и config.yml из корня модуля;
// файлы с префиксами "." и "_" (например .DS_Store) не включаются
//
//go:embed assets config.yml
var FS embed.FS
