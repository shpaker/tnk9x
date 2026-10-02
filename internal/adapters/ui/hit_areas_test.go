package ui

import (
	"image"
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

func TestHitAreas_Hit(t *testing.T) {
	var areas HitAreas
	areas.Add(image.Rect(0, 0, 10, 10))
	areas.Add(image.Rect(0, 10, 10, 20))

	if index, ok := areas.Hit(types.Position{X: 5, Y: 15}); !ok || index != 1 {
		t.Errorf("Hit = %d, %v; ожидался второй пункт", index, ok)
	}
	if _, ok := areas.Hit(types.Position{X: 15, Y: 5}); ok {
		t.Error("точка вне пунктов не должна попадать")
	}

	areas.Reset()
	if _, ok := areas.Hit(types.Position{X: 5, Y: 5}); ok {
		t.Error("после Reset пунктов нет")
	}
}

func TestRowRect_CoversHalfGap(t *testing.T) {
	rect := RowRect(10, 50, 20, 8, 8)
	if rect != image.Rect(10, 16, 50, 32) {
		t.Errorf("RowRect = %v", rect)
	}
}
