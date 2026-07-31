package touch_controls

import (
	"image"
	"math"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.ITouchControlsAdapter = (*TouchControlsAdapter)(nil)

// deadZoneFraction — доля размера крестовины вокруг её центра, внутри
// которой направление не меняется (гистерезис против дребезга)
const deadZoneFraction = 0.12

// TouchControlsAdapter — общий источник сенсорного ввода: опрашивает
// тачи, ведёт геометрию игрового экрана и контролов (единый источник
// правды для отрисовки и хит-тестов) и превращает касания в события
// управления
type TouchControlsAdapter struct {
	logicalW, logicalH int
	screenW, screenH   int

	// Защёлка «тач замечен»: включает экранные контроллы навсегда
	touchSeen bool

	gameX, gameY, gameScale int
	layout                  ControlsLayout

	// Владение тачами: касание, начавшееся в зоне контрола, следует
	// за ним до отпускания (в порядке нажатия)
	dpadTouchIDs []ebiten.TouchID
	fireTouchIDs []ebiten.TouchID

	// Состояние кадра
	direction    types.Direction
	hasDirection bool
	fireJust     bool
	pauseJust    bool
	tap          types.Position
	hasTap       bool

	// Ebiten-функции инжектируются для headless-тестов
	appendTouchIDs            func([]ebiten.TouchID) []ebiten.TouchID
	appendJustPressedTouchIDs func([]ebiten.TouchID) []ebiten.TouchID
	touchPosition             func(ebiten.TouchID) (int, int)
	deviceScaleFactor         func() float64

	// Переиспользуемые буферы опроса
	touchIDs       []ebiten.TouchID
	justPressedIDs []ebiten.TouchID

	// Кеш спрайтов контролов (см. touch_controls_renderer.go)
	whiteSprite    *ebiten.Image
	fireSprite     *ebiten.Image
	fireSpriteSize int
}

func NewTouchControlsAdapter(
	logicalW, logicalH int,
) *TouchControlsAdapter {
	return &TouchControlsAdapter{
		logicalW:                  logicalW,
		logicalH:                  logicalH,
		appendTouchIDs:            ebiten.AppendTouchIDs,
		appendJustPressedTouchIDs: inpututil.AppendJustPressedTouchIDs,
		touchPosition:             ebiten.TouchPosition,
		deviceScaleFactor: func() float64 {
			return ebiten.Monitor().DeviceScaleFactor()
		},
	}
}

// Update опрашивает тачи и вычисляет состояние кадра; вызывается раз
// в кадр из game loop до обновления игрового состояния
func (a *TouchControlsAdapter) Update() {
	a.fireJust = false
	a.pauseJust = false
	a.hasTap = false

	a.touchIDs = a.appendTouchIDs(a.touchIDs[:0])
	a.justPressedIDs = a.appendJustPressedTouchIDs(a.justPressedIDs[:0])
	if len(a.touchIDs) > 0 {
		a.touchSeen = true
	}
	// До первого DrawFinalScreen геометрия неизвестна — тачи инертны
	if a.screenW == 0 || a.screenH == 0 {
		return
	}

	a.handleJustPressed()
	a.pruneReleased()
	a.updateDirection()
}

func (a *TouchControlsAdapter) IsTouchActive() bool {
	return a.touchSeen
}

func (a *TouchControlsAdapter) DPadDirection() (types.Direction, bool) {
	return a.direction, a.hasDirection
}

func (a *TouchControlsAdapter) FireJustPressed() bool {
	return a.fireJust
}

func (a *TouchControlsAdapter) PauseJustPressed() bool {
	return a.pauseJust
}

func (a *TouchControlsAdapter) TapJustPressed() (types.Position, bool) {
	return a.tap, a.hasTap
}

// SetScreenSize фиксирует размер финального экрана (вызывается из
// DrawFinalScreen) и пересчитывает геометрию игрового экрана
// и контролов
func (a *TouchControlsAdapter) SetScreenSize(width, height int) {
	a.screenW, a.screenH = width, height
	a.refreshGeometry()
}

// GameRect — смещение и целый масштаб нарисованного игрового экрана
// для DrawFinalScreen
func (a *TouchControlsAdapter) GameRect() (x, y, scale int) {
	return a.gameX, a.gameY, a.gameScale
}

func (a *TouchControlsAdapter) refreshGeometry() {
	a.gameX, a.gameY, a.gameScale = gameRect(
		a.logicalW, a.logicalH, a.screenW, a.screenH, false,
	)
	a.layout = a.computeLayout()
	if !a.touchSeen || a.layout.Fits {
		return
	}
	// Контроллы не помещаются в полях: пробуем уменьшить игровой
	// экран на шаг масштаба; оставляем, только если места хватило
	x, y, scale := gameRect(
		a.logicalW, a.logicalH, a.screenW, a.screenH, true,
	)
	if scale == a.gameScale {
		return
	}
	shrunk := computeControlsLayout(
		a.screenW, a.screenH,
		x, y, a.logicalW*scale, a.logicalH*scale,
		a.deviceScaleFactor(),
	)
	if !shrunk.Fits {
		return
	}
	a.gameX, a.gameY, a.gameScale = x, y, scale
	a.layout = shrunk
}

func (a *TouchControlsAdapter) computeLayout() ControlsLayout {
	return computeControlsLayout(
		a.screenW, a.screenH,
		a.gameX, a.gameY,
		a.logicalW*a.gameScale, a.logicalH*a.gameScale,
		a.deviceScaleFactor(),
	)
}

// handleJustPressed раздаёт новые касания по зонам: контроллы
// забирают тач во владение, остальное — кандидат в тап по экрану
func (a *TouchControlsAdapter) handleJustPressed() {
	for _, id := range a.justPressedIDs {
		sx, sy := a.screenTouchPosition(id)
		point := image.Pt(int(sx), int(sy))
		switch {
		case point.In(a.layout.DPad):
			a.dpadTouchIDs = append(a.dpadTouchIDs, id)
		case point.In(a.layout.Fire):
			a.fireTouchIDs = append(a.fireTouchIDs, id)
			a.fireJust = true
		case point.In(a.layout.Pause):
			a.pauseJust = true
		default:
			a.registerTap(sx, sy)
		}
	}
}

func (a *TouchControlsAdapter) screenTouchPosition(
	id ebiten.TouchID,
) (float64, float64) {
	lx, ly := a.touchPosition(id)

	return logicalToScreen(
		lx, ly, a.logicalW, a.logicalH, a.screenW, a.screenH,
	)
}

// registerTap запоминает тап по нарисованному игровому экрану в его
// логических координатах (для меню и оверлеев)
func (a *TouchControlsAdapter) registerTap(sx, sy float64) {
	drawn := image.Rect(
		a.gameX,
		a.gameY,
		a.gameX+a.logicalW*a.gameScale,
		a.gameY+a.logicalH*a.gameScale,
	)
	if !image.Pt(int(sx), int(sy)).In(drawn) {
		return
	}
	lx, ly := screenToDrawnLogical(sx, sy, a.gameX, a.gameY, a.gameScale)
	a.tap = types.Position{X: lx, Y: ly}
	a.hasTap = true
}

func (a *TouchControlsAdapter) pruneReleased() {
	a.dpadTouchIDs = filterActive(a.dpadTouchIDs, a.touchIDs)
	a.fireTouchIDs = filterActive(a.fireTouchIDs, a.touchIDs)
}

// filterActive оставляет только ещё активные тачи, сохраняя порядок
// нажатия
func filterActive(owned, active []ebiten.TouchID) []ebiten.TouchID {
	result := owned[:0]
	for _, id := range owned {
		if slices.Contains(active, id) {
			result = append(result, id)
		}
	}

	return result
}

// updateDirection вычисляет направление крестовины по последнему из
// удерживаемых на ней тачей: вне мёртвой зоны выбирается
// доминирующая ось, внутри — сохраняется прежнее направление; палец
// может уехать за пределы крестовины и продолжать рулить
func (a *TouchControlsAdapter) updateDirection() {
	if len(a.dpadTouchIDs) == 0 {
		a.hasDirection = false
		return
	}
	id := a.dpadTouchIDs[len(a.dpadTouchIDs)-1]
	sx, sy := a.screenTouchPosition(id)
	center := a.layout.DPad.Min.Add(a.layout.DPad.Max).Div(2)
	dx := sx - float64(center.X)
	dy := sy - float64(center.Y)
	deadZone := deadZoneFraction * float64(a.layout.DPad.Dx())
	if math.Hypot(dx, dy) < deadZone {
		return
	}
	a.hasDirection = true
	if math.Abs(dx) >= math.Abs(dy) {
		a.direction = types.DirectionLeft
		if dx > 0 {
			a.direction = types.DirectionRight
		}

		return
	}
	a.direction = types.DirectionUp
	if dy > 0 {
		a.direction = types.DirectionDown
	}
}
