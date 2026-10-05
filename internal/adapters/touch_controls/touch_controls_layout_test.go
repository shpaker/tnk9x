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
		screenW, screenH, x, y, 256*scale, 224*scale, dsf, 1,
	)

	return layout, game
}

func TestComputeControlsLayout_PortraitPhone(t *testing.T) {
	// iPhone 14 портрет: 1170x2532 физических пикселей, dsf=3
	layout, game := layoutFor(t, 1170, 2532, 3)

	if !layout.Fits {
		t.Fatal("контроллы должны помещаться в нижней полосе")
	}
	if layout.Players[0].DPad.Overlaps(game) ||
		layout.Players[0].Fire.Overlaps(game) {
		t.Error("контроллы не должны перекрывать игровой экран")
	}
	if layout.Players[0].DPad.Min.Y <= game.Max.Y {
		t.Error("в портрете крестовина должна быть под игрой")
	}
	if layout.Players[0].Fire.Min.X <= layout.Players[0].DPad.Max.X {
		t.Error("огонь должен быть правее крестовины")
	}
	if layout.Players[0].DPad.Dx() < 3*minTargetDp*3 {
		t.Errorf(
			"крестовина %dpx меньше минимума %dpx",
			layout.Players[0].DPad.Dx(), 3*minTargetDp*3,
		)
	}
}

func TestComputeControlsLayout_LandscapePhone(t *testing.T) {
	// iPhone 14 ландшафт: боковые полосы (2532-256*5)/2 = 626px
	layout, game := layoutFor(t, 2532, 1170, 3)

	if !layout.Fits {
		t.Fatal("контроллы должны помещаться в боковых полосах")
	}
	if layout.Players[0].DPad.Overlaps(game) ||
		layout.Players[0].Fire.Overlaps(game) {
		t.Error("контроллы не должны перекрывать игровой экран")
	}
	if layout.Players[0].DPad.Max.X > game.Min.X {
		t.Error("в ландшафте крестовина должна быть слева от игры")
	}
	if layout.Players[0].Fire.Min.X < game.Max.X {
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
		1334, 750, x, y, 256*scale, 224*scale, 2, 1,
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
	if layout.Players[0].DPad.Empty() {
		t.Error("нулевой dsf должен трактоваться как 1")
	}
}

// twoPlayersLayoutFor — раскладка на двоих для экрана с игрой,
// размещённой по gameRect без shrink
func twoPlayersLayoutFor(
	t *testing.T,
	screenW, screenH int,
	dsf float64,
) (ControlsLayout, image.Rectangle) {
	t.Helper()
	x, y, scale := gameRect(256, 224, screenW, screenH, false)
	game := image.Rect(x, y, x+256*scale, y+224*scale)
	layout := computeControlsLayout(
		screenW, screenH, x, y, 256*scale, 224*scale, dsf, 2,
	)

	return layout, game
}

// assertNoOverlaps — контроллы не перекрывают игру, друг друга
// и паузу
func assertNoOverlaps(
	t *testing.T,
	layout ControlsLayout,
	game image.Rectangle,
) {
	t.Helper()
	rects := []image.Rectangle{layout.Pause}
	for _, controls := range layout.Players {
		rects = append(rects, controls.DPad, controls.Fire)
	}
	for i, rect := range rects {
		if rect.Empty() {
			t.Errorf("прямоугольник %d пустой", i)
		}
		if i > 0 && rect.Overlaps(game) {
			t.Errorf("прямоугольник %d перекрывает игру", i)
		}
		for _, other := range rects[i+1:] {
			if rect.Overlaps(other) {
				t.Errorf("прямоугольник %d перекрывает соседа", i)
			}
		}
	}
}

func TestComputeControlsLayout_TwoPlayersLandscape(t *testing.T) {
	layout, game := twoPlayersLayoutFor(t, 2532, 1170, 3)

	if !layout.Fits {
		t.Fatal("контроллы двоих должны помещаться в боковых полосах")
	}
	assertNoOverlaps(t, layout, game)
	first, second := layout.Players[0], layout.Players[1]
	if first.DPad.Max.X > game.Min.X || first.Fire.Max.X > game.Min.X {
		t.Error("контроллы P1 — в левой полосе")
	}
	if second.DPad.Min.X < game.Max.X || second.Fire.Min.X < game.Max.X {
		t.Error("контроллы P2 — в правой полосе")
	}
	for i, controls := range layout.Players {
		if gap := controls.DPad.Min.Y - controls.Fire.Max.Y; gap < controlsGapDp*3 {
			t.Errorf(
				"P%d: огонь над крестовиной с зазором не меньше %d, а не %d",
				i+1,
				controlsGapDp*3,
				gap,
			)
		}
	}
}

func TestComputeControlsLayout_TwoPlayersLandscapeLowScreenKeepsGap(
	t *testing.T,
) {
	layout, game := twoPlayersLayoutFor(t, 1600, 620, 2)
	assertNoOverlaps(t, layout, game)
	for i, controls := range layout.Players {
		if gap := controls.DPad.Min.Y - controls.Fire.Max.Y; gap < controlsGapDp*2-1 {
			t.Errorf("P%d: на низком экране зазор %d меньше %d",
				i+1, gap, controlsGapDp*2)
		}
	}
}

func TestComputeControlsLayout_TwoPlayersPortrait(t *testing.T) {
	layout, game := twoPlayersLayoutFor(t, 1170, 2532, 3)

	if !layout.Fits {
		t.Fatal("контроллы двоих должны помещаться над и под игрой")
	}
	assertNoOverlaps(t, layout, game)
	first, second := layout.Players[0], layout.Players[1]
	if first.DPad.Min.Y < game.Max.Y || first.Fire.Min.Y < game.Max.Y {
		t.Error("контроллы P1 — под игрой")
	}
	if second.DPad.Max.Y > game.Min.Y || second.Fire.Max.Y > game.Min.Y {
		t.Error("контроллы P2 — над игрой")
	}
	if layout.Pause.Max.Y > game.Min.Y {
		t.Error("пауза — в верхней полосе")
	}
}

func TestComputeControlsLayout_SinglePlayerHasNoSecondSet(t *testing.T) {
	layout, _ := layoutFor(t, 2532, 1170, 3)
	if !layout.Players[1].DPad.Empty() || !layout.Players[1].Fire.Empty() {
		t.Error("в одиночной игре контролов P2 нет")
	}
}
