package types_test

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

func TestEffectsSettingsEntity_Toggle(t *testing.T) {
	settings := types.NewEffectsSettingsEntity(true)

	settings.Toggle()
	if settings.IsEnabled() {
		t.Fatal("после переключения эффекты должны быть выключены")
	}

	settings.Toggle()
	if !settings.IsEnabled() {
		t.Fatal("повторное переключение должно включить эффекты")
	}
}

func TestGraphicsLabel(t *testing.T) {
	if types.GraphicsLabel(true) == types.GraphicsLabel(false) {
		t.Fatal("подписи режимов графики должны различаться")
	}
}
