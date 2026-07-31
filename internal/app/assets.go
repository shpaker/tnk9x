package app

import (
	"io/fs"
	"os"

	"github.com/shpaker/tnk9x"
)

// assetsFS выбирает источник ресурсов: каталог assets на диске
// (dev-режим, распакованный релиз), иначе — встроенная копия
// (wasm, самодостаточный бинарник)
func assetsFS() fs.FS {
	if info, err := os.Stat("assets"); err == nil && info.IsDir() {
		return os.DirFS("assets")
	}
	sub, err := fs.Sub(tnk9x.FS, "assets")
	if err != nil {
		panic(err) // недостижимо: каталог гарантирован go:embed
	}
	return sub
}
