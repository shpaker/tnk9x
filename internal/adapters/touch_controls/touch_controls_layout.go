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
	// controlsGapDp — минимальный зазор между крестовиной и огнём
	// одной колонки: палец с крестовины не задевает огонь
	controlsGapDp = 24
	maxDPadDp     = 220
	maxFireDp     = 120

	// Желаемые размеры — доли меньшей стороны экрана
	dpadFraction = 0.42
	fireFraction = 0.24
)

// playersCount — наборов контролов в режиме на двоих
const playersCount = 2

// PlayerControls — крестовина и огонь одного игрока; пустые
// прямоугольники — у игрока нет контролов (одиночная игра)
type PlayerControls struct {
	DPad image.Rectangle
	Fire image.Rectangle
}

// ControlsLayout — прямоугольники контролов в пикселях финального
// экрана: свой набор у каждого игрока, пауза общая; Fits=false —
// контроллы не помещаются в полях целиком (сигнал к уменьшению
// игрового экрана на шаг масштаба)
type ControlsLayout struct {
	Players [playersCount]PlayerControls
	Pause   image.Rectangle
	Fits    bool
}

// controlsSizes — желаемые размеры контролов и отступ в пикселях
type controlsSizes struct {
	dpad, fire, pad, gap, minTarget float64
}

// computeControlsLayout размещает контроллы в свободных полях вокруг
// игрового экрана: в портрете — в полосах под и над игрой, в
// ландшафте — в боковых; на двоих у каждого игрока своя сторона
// устройства. dsf — device scale factor (физические пиксели на dp)
func computeControlsLayout(
	screenW, screenH, gameX, gameY, gameW, gameH int,
	dsf float64,
	players int,
) ControlsLayout {
	if dsf <= 0 {
		dsf = 1
	}
	minTarget := minTargetDp * dsf
	minDim := math.Min(float64(screenW), float64(screenH))
	sizes := controlsSizes{
		dpad:      clampf(dpadFraction*minDim, 3*minTarget, maxDPadDp*dsf),
		fire:      clampf(fireFraction*minDim, 1.5*minTarget, maxFireDp*dsf),
		pad:       paddingDp * dsf,
		gap:       controlsGapDp * dsf,
		minTarget: minTarget,
	}

	portrait := screenH > screenW
	var layout ControlsLayout
	switch {
	case players >= playersCount && portrait:
		layout = portraitTwoPlayersLayout(screenW, screenH, gameY, gameH, sizes)
	case players >= playersCount:
		layout = landscapeTwoPlayersLayout(
			screenW,
			screenH,
			gameX,
			gameW,
			sizes,
		)
	case portrait:
		layout.Players[0] = portraitPlayerControls(
			screenW, gameY+gameH, screenH, sizes,
		)
		layout.Pause = cornerPauseRect(screenW, sizes)
	default:
		layout = landscapeLayout(screenW, screenH, gameX, gameW, sizes)
		layout.Pause = cornerPauseRect(screenW, sizes)
	}

	layout.Fits = true
	for player := range min(max(players, 1), playersCount) {
		controls := layout.Players[player]
		if float64(controls.DPad.Dx()) < 3*minTarget ||
			float64(controls.Fire.Dx()) < 1.5*minTarget ||
			layout.Pause.Overlaps(controls.DPad) ||
			layout.Pause.Overlaps(controls.Fire) {
			layout.Fits = false
		}
	}
	return layout
}

// portraitPlayerControls — крестовина слева и огонь справа
// в горизонтальной полосе между top и bottom
func portraitPlayerControls(
	screenW, top, bottom int,
	sizes controlsSizes,
) PlayerControls {
	band := float64(bottom - top)
	dpad := math.Min(sizes.dpad, band-2*sizes.pad)
	fire := math.Min(sizes.fire, band-2*sizes.pad)
	centerY := float64(top) + band/2

	return PlayerControls{
		DPad: rectAround(sizes.pad+dpad/2, centerY, dpad),
		Fire: rectAround(float64(screenW)-sizes.pad-fire/2, centerY, fire),
	}
}

// portraitTwoPlayersLayout — P1 в полосе под игрой, P2 в полосе над
// ней; пауза в верхней полосе между крестовиной и огнём P2
func portraitTwoPlayersLayout(
	screenW, screenH, gameY, gameH int,
	sizes controlsSizes,
) ControlsLayout {
	var layout ControlsLayout
	layout.Players[0] = portraitPlayerControls(
		screenW, gameY+gameH, screenH, sizes,
	)
	second := portraitPlayerControls(screenW, 0, gameY, sizes)
	layout.Players[1] = second
	layout.Pause = rectAround(
		float64(second.DPad.Max.X+second.Fire.Min.X)/2,
		float64(gameY)/2,
		sizes.minTarget,
	)
	return layout
}

// landscapeLayout — крестовина в левой полосе, огонь в правой;
// вертикальный центр смещён вниз, под большие пальцы
func landscapeLayout(
	screenW, screenH, gameX, gameW int,
	sizes controlsSizes,
) ControlsLayout {
	bandW := float64(gameX)
	dpad := math.Min(sizes.dpad, bandW-2*sizes.pad)
	fire := math.Min(sizes.fire, bandW-2*sizes.pad)
	centerY := clampf(
		0.62*float64(screenH),
		sizes.pad+dpad/2,
		float64(screenH)-sizes.pad-dpad/2,
	)
	gameRight := float64(gameX + gameW)
	fireCenterX := gameRight + (float64(screenW)-gameRight)/2

	var layout ControlsLayout
	layout.Players[0] = PlayerControls{
		DPad: rectAround(float64(gameX)/2, centerY, dpad),
		Fire: rectAround(fireCenterX, centerY, fire),
	}
	return layout
}

// landscapeTwoPlayersLayout — P1 в левой полосе, P2 в правой:
// крестовина внизу, огонь над ней; сверху место под паузу
func landscapeTwoPlayersLayout(
	screenW, screenH, gameX, gameW int,
	sizes controlsSizes,
) ControlsLayout {
	gameRight := float64(gameX + gameW)
	var layout ControlsLayout
	layout.Players[0] = landscapeColumn(
		float64(gameX)/2, float64(gameX), screenH, sizes,
	)
	layout.Players[1] = landscapeColumn(
		gameRight+(float64(screenW)-gameRight)/2,
		float64(screenW)-gameRight,
		screenH,
		sizes,
	)
	layout.Pause = cornerPauseRect(screenW, sizes)
	return layout
}

// landscapeColumn — контроллы одного игрока в боковой полосе:
// крестовина у нижнего края, сверху ряд паузы, огонь по центру
// между ними не ближе зазора к крестовине; не помещаясь по высоте,
// оба уменьшаются пропорционально
func landscapeColumn(
	centerX, bandW float64,
	screenH int,
	sizes controlsSizes,
) PlayerControls {
	dpad := math.Min(sizes.dpad, bandW-2*sizes.pad)
	fire := math.Min(sizes.fire, bandW-2*sizes.pad)
	pauseRow := 2*sizes.pad + sizes.minTarget
	available := float64(screenH) - pauseRow - 2*sizes.pad - sizes.gap
	if dpad+fire > available && dpad+fire > 0 {
		shrink := available / (dpad + fire)
		dpad *= shrink
		fire *= shrink
	}
	dpadCenterY := float64(screenH) - sizes.pad - dpad/2
	dpadTop := dpadCenterY - dpad/2
	fireCenterY := math.Min(
		(pauseRow+dpadTop)/2,
		dpadTop-sizes.gap-fire/2,
	)

	return PlayerControls{
		DPad: rectAround(centerX, dpadCenterY, dpad),
		Fire: rectAround(centerX, fireCenterY, fire),
	}
}

// cornerPauseRect — кнопка паузы в правом верхнем углу экрана
func cornerPauseRect(screenW int, sizes controlsSizes) image.Rectangle {
	return image.Rect(
		int(float64(screenW)-sizes.pad-sizes.minTarget),
		int(sizes.pad),
		int(float64(screenW)-sizes.pad),
		int(sizes.pad+sizes.minTarget),
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
