package touch_controls

import (
	"image"
	"math"
)

// Габариты контролов в dp (умножаются на device scale factor):
// minTargetDp — минимальная зона нажатия по гайдлайнам мобильных
// платформ; крестовине нужна сетка 3x3 таких зон
const (
	minTargetDp = 48
	paddingDp   = 8
	maxDPadDp   = 220
	maxFireDp   = 120

	// Желаемые размеры — доли меньшей стороны экрана
	dpadFraction = 0.42
	fireFraction = 0.24
)

// ControlsLayout — прямоугольники контролов в пикселях финального
// экрана; Fits=false — контроллы не помещаются в полях целиком
// (сигнал к уменьшению игрового экрана на шаг масштаба)
type ControlsLayout struct {
	DPad  image.Rectangle
	Fire  image.Rectangle
	Pause image.Rectangle
	Fits  bool
}

// computeControlsLayout размещает контроллы в свободных полях вокруг
// игрового экрана: в портрете — в нижней полосе, в ландшафте — в
// боковых; dsf — device scale factor (физические пиксели на dp)
func computeControlsLayout(
	screenW, screenH, gameX, gameY, gameW, gameH int,
	dsf float64,
) ControlsLayout {
	if dsf <= 0 {
		dsf = 1
	}
	minTarget := minTargetDp * dsf
	pad := paddingDp * dsf
	minDim := math.Min(float64(screenW), float64(screenH))

	dpadSize := clampf(dpadFraction*minDim, 3*minTarget, maxDPadDp*dsf)
	fireSize := clampf(fireFraction*minDim, 1.5*minTarget, maxFireDp*dsf)

	var layout ControlsLayout
	if screenH > screenW {
		layout = portraitLayout(
			screenW, screenH, gameY+gameH, dpadSize, fireSize, pad,
		)
	} else {
		layout = landscapeLayout(
			screenW, screenH, gameX, gameW, dpadSize, fireSize, pad,
		)
	}
	layout.Fits = float64(layout.DPad.Dx()) >= 3*minTarget &&
		float64(layout.Fire.Dx()) >= 1.5*minTarget
	layout.Pause = pauseRect(screenW, minTarget, pad)

	return layout
}

// portraitLayout — крестовина слева и огонь справа в нижней полосе
// под игровым экраном
func portraitLayout(
	screenW, screenH, gameBottom int,
	dpadSize, fireSize, pad float64,
) ControlsLayout {
	band := float64(screenH - gameBottom)
	dpad := math.Min(dpadSize, band-2*pad)
	fire := math.Min(fireSize, band-2*pad)
	centerY := float64(gameBottom) + band/2

	return ControlsLayout{
		DPad: rectAround(pad+dpad/2, centerY, dpad),
		Fire: rectAround(float64(screenW)-pad-fire/2, centerY, fire),
	}
}

// landscapeLayout — крестовина в левой полосе, огонь в правой;
// вертикальный центр смещён вниз, под большие пальцы
func landscapeLayout(
	screenW, screenH, gameX, gameW int,
	dpadSize, fireSize, pad float64,
) ControlsLayout {
	bandW := float64(gameX)
	dpad := math.Min(dpadSize, bandW-2*pad)
	fire := math.Min(fireSize, bandW-2*pad)
	centerY := clampf(
		0.62*float64(screenH),
		pad+dpad/2,
		float64(screenH)-pad-dpad/2,
	)
	gameRight := float64(gameX + gameW)
	fireCenterX := gameRight + (float64(screenW)-gameRight)/2

	return ControlsLayout{
		DPad: rectAround(float64(gameX)/2, centerY, dpad),
		Fire: rectAround(fireCenterX, centerY, fire),
	}
}

// pauseRect — кнопка паузы в правом верхнем углу экрана
func pauseRect(screenW int, minTarget, pad float64) image.Rectangle {
	return image.Rect(
		int(float64(screenW)-pad-minTarget),
		int(pad),
		int(float64(screenW)-pad),
		int(pad+minTarget),
	)
}

// rectAround — квадрат заданного размера вокруг центра
func rectAround(cx, cy, size float64) image.Rectangle {
	if size <= 0 {
		return image.Rectangle{}
	}
	half := size / 2

	return image.Rect(
		int(cx-half), int(cy-half), int(cx+half), int(cy+half),
	)
}

func clampf(v, lo, hi float64) float64 {
	if hi < lo {
		hi = lo
	}

	return math.Min(math.Max(v, lo), hi)
}
