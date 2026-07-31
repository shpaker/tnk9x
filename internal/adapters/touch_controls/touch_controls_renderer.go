package touch_controls

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/shpaker/tnk9x/internal/types"
)

// Прозрачность контролов: едва заметные в покое, ярче при нажатии
const (
	idleAlpha    = 0.25
	pressedAlpha = 0.30
)

// DrawControls рисует экранные контроллы в полях финального экрана;
// до первого касания не рисует ничего
func (a *TouchControlsAdapter) DrawControls(screen ebiten.FinalScreen) {
	if !a.touchSeen {
		return
	}
	a.drawDPad(screen)
	a.drawFire(screen)
	a.drawPause(screen)
}

// drawDPad — крест из трёх непересекающихся прямоугольников;
// активное направление подсвечивается внешней третью плеча
func (a *TouchControlsAdapter) drawDPad(screen ebiten.FinalScreen) {
	r := a.layout.DPad
	if r.Empty() {
		return
	}
	third := r.Dx() / 3
	vertical := image.Rect(
		r.Min.X+third, r.Min.Y, r.Max.X-third, r.Max.Y,
	)
	left := image.Rect(
		r.Min.X, r.Min.Y+third, r.Min.X+third, r.Max.Y-third,
	)
	right := image.Rect(
		r.Max.X-third, r.Min.Y+third, r.Max.X, r.Max.Y-third,
	)
	a.fillRect(screen, vertical, idleAlpha)
	a.fillRect(screen, left, idleAlpha)
	a.fillRect(screen, right, idleAlpha)
	if !a.hasDirection {
		return
	}
	a.fillRect(screen, a.dpadArmRect(a.direction), pressedAlpha)
}

// dpadArmRect — внешняя треть плеча крестовины в заданном направлении
func (a *TouchControlsAdapter) dpadArmRect(
	direction types.Direction,
) image.Rectangle {
	r := a.layout.DPad
	third := r.Dx() / 3
	switch direction {
	case types.DirectionUp:
		return image.Rect(
			r.Min.X+third, r.Min.Y, r.Max.X-third, r.Min.Y+third,
		)
	case types.DirectionDown:
		return image.Rect(
			r.Min.X+third, r.Max.Y-third, r.Max.X-third, r.Max.Y,
		)
	case types.DirectionLeft:
		return image.Rect(
			r.Min.X, r.Min.Y+third, r.Min.X+third, r.Max.Y-third,
		)
	default:
		return image.Rect(
			r.Max.X-third, r.Min.Y+third, r.Max.X, r.Max.Y-third,
		)
	}
}

func (a *TouchControlsAdapter) drawFire(screen ebiten.FinalScreen) {
	r := a.layout.Fire
	if r.Empty() {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(r.Min.X), float64(r.Min.Y))
	alpha := float32(idleAlpha)
	if len(a.fireTouchIDs) > 0 {
		alpha = idleAlpha + pressedAlpha
	}
	op.ColorScale.ScaleAlpha(alpha)
	screen.DrawImage(a.fireCircle(r.Dx()), op)
}

// drawPause — две вертикальные полоски символа паузы по центру зоны
func (a *TouchControlsAdapter) drawPause(screen ebiten.FinalScreen) {
	r := a.layout.Pause
	if r.Empty() {
		return
	}
	barWidth := r.Dx() / 5
	centerX := (r.Min.X + r.Max.X) / 2
	top := r.Min.Y + r.Dy()/4
	bottom := r.Max.Y - r.Dy()/4
	left := image.Rect(
		centerX-barWidth-barWidth/2, top, centerX-barWidth/2, bottom,
	)
	right := image.Rect(
		centerX+barWidth/2, top, centerX+barWidth/2+barWidth, bottom,
	)
	a.fillRect(screen, left, idleAlpha)
	a.fillRect(screen, right, idleAlpha)
}

// fillRect — заливка прямоугольника растянутым белым пикселем:
// FinalScreen умеет только DrawImage
func (a *TouchControlsAdapter) fillRect(
	screen ebiten.FinalScreen,
	r image.Rectangle,
	alpha float32,
) {
	if r.Empty() {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(r.Dx()), float64(r.Dy()))
	op.GeoM.Translate(float64(r.Min.X), float64(r.Min.Y))
	op.ColorScale.ScaleAlpha(alpha)
	screen.DrawImage(a.whitePixel(), op)
}

// whitePixel — ленивый белый спрайт 1x1, базовый примитив заливки
func (a *TouchControlsAdapter) whitePixel() *ebiten.Image {
	if a.whiteSprite == nil {
		a.whiteSprite = ebiten.NewImage(1, 1)
		a.whiteSprite.Fill(color.White)
	}

	return a.whiteSprite
}

// fireCircle — кешированный круглый спрайт кнопки огня
func (a *TouchControlsAdapter) fireCircle(size int) *ebiten.Image {
	if a.fireSprite != nil && a.fireSpriteSize == size {
		return a.fireSprite
	}
	a.fireSprite = ebiten.NewImage(size, size)
	half := float32(size) / 2
	vector.FillCircle(
		a.fireSprite, half, half, half, color.White, true,
	)
	a.fireSpriteSize = size

	return a.fireSprite
}
