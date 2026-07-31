package touch_controls

import (
	"image"
	"testing"
)

// layoutFor — раскладка контролов для экрана с игрой, размещённой
// по gameRect без shrink
func layoutFor(
	t *testing.T,
	screenW, screenH int,
	dsf float64,
) (ControlsLayout, image.Rectangle) {
	t.Helper()
	x, y, scale := gameRect(256, 224, screenW, screenH, false)
	game := image.Rect(x, y, x+256*scale, y+224*scale)
	layout := computeControlsLayout(
		screenW, screenH, x, y, 256*scale, 224*scale, dsf,
	)

	return layout, game
}

func TestComputeControlsLayout_PortraitPhone(t *testing.T) {
	// iPhone 14 портрет: 1170x2532 физических пикселей, dsf=3
	layout, game := layoutFor(t, 1170, 2532, 3)

	if !layout.Fits {
		t.Fatal("контроллы должны помещаться в нижней полосе")
	}
	if layout.DPad.Overlaps(game) || layout.Fire.Overlaps(game) {
		t.Error("контроллы не должны перекрывать игровой экран")
	}
	if layout.DPad.Min.Y <= game.Max.Y {
		t.Error("в портрете крестовина должна быть под игрой")
	}
	if layout.Fire.Min.X <= layout.DPad.Max.X {
		t.Error("огонь должен быть правее крестовины")
	}
	if layout.DPad.Dx() < 3*minTargetDp*3 {
		t.Errorf(
			"крестовина %dpx меньше минимума %dpx",
			layout.DPad.Dx(), 3*minTargetDp*3,
		)
	}
}

func TestComputeControlsLayout_LandscapePhone(t *testing.T) {
	// iPhone 14 ландшафт: боковые полосы (2532-256*5)/2 = 626px
	layout, game := layoutFor(t, 2532, 1170, 3)

	if !layout.Fits {
		t.Fatal("контроллы должны помещаться в боковых полосах")
	}
	if layout.DPad.Overlaps(game) || layout.Fire.Overlaps(game) {
		t.Error("контроллы не должны перекрывать игровой экран")
	}
	if layout.DPad.Max.X > game.Min.X {
		t.Error("в ландшафте крестовина должна быть слева от игры")
	}
	if layout.Fire.Min.X < game.Max.X {
		t.Error("в ландшафте огонь должен быть справа от игры")
	}
}

func TestComputeControlsLayout_TightLandscapeNeedsShrink(t *testing.T) {
	// iPhone SE ландшафт: 1334x750, dsf=2, масштаб 2 -> полосы
	// (1334-512)/2 = 411px < 3*48*2 + отступы -> не помещается
	layout, _ := layoutFor(t, 1334, 750, 2)
	if layout.Fits {
		t.Fatal("ожидался сигнал о нехватке места")
	}

	// После shrink (масштаб 2 -> 1) полосы расширяются и места
	// хватает
	x, y, scale := gameRect(256, 224, 1334, 750, true)
	shrunk := computeControlsLayout(
		1334, 750, x, y, 256*scale, 224*scale, 2,
	)
	if !shrunk.Fits {
		t.Error("после shrink контроллы должны помещаться")
	}
}

func TestComputeControlsLayout_PauseInTopRightCorner(t *testing.T) {
	layout, _ := layoutFor(t, 1170, 2532, 3)
	if layout.Pause.Empty() {
		t.Fatal("зона паузы не должна быть пустой")
	}
	if layout.Pause.Max.X > 1170 || layout.Pause.Min.Y < 0 {
		t.Error("пауза должна быть внутри экрана")
	}
	if layout.Pause.Min.X < 1170/2 {
		t.Error("пауза должна быть в правой части экрана")
	}
}

func TestComputeControlsLayout_ZeroDSFFallsBackToOne(t *testing.T) {
	layout, _ := layoutFor(t, 1170, 2532, 0)
	if layout.DPad.Empty() {
		t.Error("нулевой dsf должен трактоваться как 1")
	}
}
