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

// playerTouches — касания и состояние кадра контролов одного игрока
type playerTouches struct {
	// Владение тачами: касание, начавшееся в зоне контрола, следует
	// за ним до отпускания (в порядке нажатия)
	dpadTouchIDs []ebiten.TouchID
	fireTouchIDs []ebiten.TouchID

	direction    types.Direction
	hasDirection bool
	// directionJust — крестовину нажали в этом кадре или сменили
	// направление: шаг по пунктам меню
	directionJust bool
	fireJust      bool
}

// TouchControlsAdapter — общий источник сенсорного ввода: опрашивает
// тачи, ведёт геометрию игрового экрана и контролов (единый источник
// правды для отрисовки и хит-тестов) и превращает касания в события
// управления; в режиме на двоих у каждого игрока свой набор
type TouchControlsAdapter struct {
	logicalW, logicalH int
	screenW, screenH   int

	// Entities
	settings *types.SettingsEntity

	// Защёлка «тач замечен»: включает экранные контроллы навсегда
	touchSeen bool

	gameX, gameY, gameScale int
	layout                  ControlsLayout

	// Состояние кадра
	players   [playersCount]playerTouches
	pauseJust bool
	// Тап по игровому экрану в логических координатах
	tapJust     bool
	tapPosition types.Position

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

// NewTouchControlsAdapter — контроллы для логического экрана
// logicalW x logicalH; число наборов берётся из настроек
func NewTouchControlsAdapter(
	logicalW, logicalH int,
	settings *types.SettingsEntity,
) *TouchControlsAdapter {
	return &TouchControlsAdapter{
		logicalW:                  logicalW,
		logicalH:                  logicalH,
		settings:                  settings,
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
	a.pauseJust = false
	a.tapJust = false
	for i := range a.players {
		a.players[i].fireJust = false
		a.players[i].directionJust = false
	}

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
	for i := range a.players {
		a.updateDirection(i)
	}
}

func (a *TouchControlsAdapter) IsTouchActive() bool {
	return a.touchSeen
}

func (a *TouchControlsAdapter) DPadDirection(
	player types.PlayerTankNum,
) (types.Direction, bool) {
	touches, ok := a.player(player)
	if !ok {
		return types.DirectionUp, false
	}
	return touches.direction, touches.hasDirection
}

func (a *TouchControlsAdapter) DPadJustPressed(
	player types.PlayerTankNum,
) (types.Direction, bool) {
	touches, ok := a.player(player)
	if !ok {
		return types.DirectionUp, false
	}
	return touches.direction, touches.directionJust
}

func (a *TouchControlsAdapter) FireJustPressed(
	player types.PlayerTankNum,
) bool {
	touches, ok := a.player(player)
	return ok && touches.fireJust
}

func (a *TouchControlsAdapter) PauseJustPressed() bool {
	return a.pauseJust
}

func (a *TouchControlsAdapter) TapJustPressed() (types.Position, bool) {
	return a.tapPosition, a.tapJust
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

// player — состояние контролов игрока
func (a *TouchControlsAdapter) player(
	player types.PlayerTankNum,
) (*playerTouches, bool) {
	if int(player) < 0 || int(player) >= len(a.players) {
		return nil, false
	}
	return &a.players[player], true
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
		int(a.settings.GetPlayers()),
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
		int(a.settings.GetPlayers()),
	)
}

// handleJustPressed раздаёт новые касания по зонам: контроллы
// забирают тач во владение, касание игрового экрана вне контролов —
// тап по экранным кнопкам
func (a *TouchControlsAdapter) handleJustPressed() {
	for _, id := range a.justPressedIDs {
		sx, sy := a.screenTouchPosition(id)
		point := image.Pt(int(sx), int(sy))
		if point.In(a.layout.Pause) {
			a.pauseJust = true
			continue
		}
		claimed := false
		for i, controls := range a.layout.Players {
			touches := &a.players[i]
			switch {
			case point.In(controls.DPad):
				touches.dpadTouchIDs = append(touches.dpadTouchIDs, id)
				claimed = true
			case point.In(controls.Fire):
				touches.fireTouchIDs = append(touches.fireTouchIDs, id)
				touches.fireJust = true
				claimed = true
			}
		}
		if !claimed {
			a.handleTap(sx, sy)
		}
	}
}

// handleTap — касание игрового экрана вне контролов; касания полей
// вне игры не действуют
func (a *TouchControlsAdapter) handleTap(sx, sy float64) {
	position, ok := a.screenToGame(sx, sy)
	if !ok {
		return
	}
	a.tapJust = true
	a.tapPosition = position
}

// GamePosition переводит координаты ebiten (тач, курсор мыши)
// в логические координаты нарисованного игрового экрана
func (a *TouchControlsAdapter) GamePosition(x, y int) (types.Position, bool) {
	if a.screenW == 0 || a.screenH == 0 {
		return types.Position{}, false
	}
	sx, sy := logicalToScreen(
		x, y, a.logicalW, a.logicalH, a.screenW, a.screenH,
	)
	return a.screenToGame(sx, sy)
}

// screenToGame переводит пиксели финального экрана в логические
// координаты игрового экрана; точки на полях — false
func (a *TouchControlsAdapter) screenToGame(
	sx, sy float64,
) (types.Position, bool) {
	if a.gameScale <= 0 {
		return types.Position{}, false
	}
	x := (sx - float64(a.gameX)) / float64(a.gameScale)
	y := (sy - float64(a.gameY)) / float64(a.gameScale)
	if x < 0 || y < 0 || x >= float64(a.logicalW) || y >= float64(a.logicalH) {
		return types.Position{}, false
	}
	return types.Position{X: x, Y: y}, true
}

func (a *TouchControlsAdapter) screenTouchPosition(
	id ebiten.TouchID,
) (float64, float64) {
	lx, ly := a.touchPosition(id)

	return logicalToScreen(
		lx, ly, a.logicalW, a.logicalH, a.screenW, a.screenH,
	)
}

// pruneReleased забывает отпущенные касания; контроллы, которых нет
// в текущей раскладке (режим сменился на одиночный), отпускаются
func (a *TouchControlsAdapter) pruneReleased() {
	for i := range a.players {
		touches := &a.players[i]
		if a.layout.Players[i].DPad.Empty() {
			touches.dpadTouchIDs = touches.dpadTouchIDs[:0]
			touches.fireTouchIDs = touches.fireTouchIDs[:0]
			continue
		}
		touches.dpadTouchIDs = filterActive(touches.dpadTouchIDs, a.touchIDs)
		touches.fireTouchIDs = filterActive(touches.fireTouchIDs, a.touchIDs)
	}
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

// updateDirection вычисляет направление крестовины игрока по
// последнему из удерживаемых на ней тачей: вне мёртвой зоны
// выбирается доминирующая ось, внутри — сохраняется прежнее
// направление; палец может уехать за пределы крестовины и продолжать
// рулить. Нажатие или смена направления отмечаются как шаг для меню
func (a *TouchControlsAdapter) updateDirection(player int) {
	touches := &a.players[player]
	dpadRect := a.layout.Players[player].DPad
	wasDirection, previous := touches.hasDirection, touches.direction
	defer func() {
		touches.directionJust = touches.hasDirection &&
			(!wasDirection || touches.direction != previous)
	}()

	if len(touches.dpadTouchIDs) == 0 {
		touches.hasDirection = false
		return
	}

	id := touches.dpadTouchIDs[len(touches.dpadTouchIDs)-1]
	sx, sy := a.screenTouchPosition(id)
	center := dpadRect.Min.Add(dpadRect.Max).Div(2)
	dx := sx - float64(center.X)
	dy := sy - float64(center.Y)
	deadZone := deadZoneFraction * float64(dpadRect.Dx())
	if math.Hypot(dx, dy) < deadZone {
		return
	}

	touches.hasDirection = true
	if math.Abs(dx) >= math.Abs(dy) {
		touches.direction = types.DirectionLeft
		if dx > 0 {
			touches.direction = types.DirectionRight
		}

		return
	}
	touches.direction = types.DirectionUp
	if dy > 0 {
		touches.direction = types.DirectionDown
	}
}
