package touch_controls

import (
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/tnk9x/internal/types"
)

// fakeTouches — управляемый источник тачей вместо ebiten-рантайма
type fakeTouches struct {
	active      []ebiten.TouchID
	justPressed []ebiten.TouchID
	positions   map[ebiten.TouchID][2]int
}

// newTestAdapter — адаптер на экране 1170x2532 (iPhone портрет,
// dsf=1 для простых координат) с подменёнными ebiten-функциями
func newTestAdapter(touches *fakeTouches) *TouchControlsAdapter {
	adapter := NewTouchControlsAdapter(256, 224, types.NewSettingsEntity())
	adapter.appendTouchIDs = func(
		ids []ebiten.TouchID,
	) []ebiten.TouchID {
		return append(ids, touches.active...)
	}
	adapter.appendJustPressedTouchIDs = func(
		ids []ebiten.TouchID,
	) []ebiten.TouchID {
		return append(ids, touches.justPressed...)
	}
	adapter.touchPosition = func(id ebiten.TouchID) (int, int) {
		pos := touches.positions[id]

		return pos[0], pos[1]
	}
	adapter.deviceScaleFactor = func() float64 { return 1 }
	adapter.SetScreenSize(1170, 2532)

	return adapter
}

// logicalAt переводит экранные пиксели в логические координаты
// ebiten — фейковые тачи задаются позициями на экране
func logicalAt(sx, sy float64) [2]int {
	lx, ly := ebitenScreenToLogical(sx, sy, 256, 224, 1170, 2532)

	return [2]int{int(lx), int(ly)}
}

func TestTouchControlsAdapter_LatchAndInertBeforeScreen(t *testing.T) {
	touches := &fakeTouches{}
	adapter := NewTouchControlsAdapter(256, 224, types.NewSettingsEntity())
	adapter.appendTouchIDs = func(
		ids []ebiten.TouchID,
	) []ebiten.TouchID {
		return append(ids, touches.active...)
	}
	adapter.appendJustPressedTouchIDs = func(
		ids []ebiten.TouchID,
	) []ebiten.TouchID {
		return append(ids, touches.justPressed...)
	}
	adapter.touchPosition = func(ebiten.TouchID) (int, int) {
		return 0, 0
	}
	adapter.deviceScaleFactor = func() float64 { return 1 }

	if adapter.IsTouchActive() {
		t.Fatal("до первого касания защёлка должна быть выключена")
	}

	// До SetScreenSize тач взводит защёлку, но события не создаются
	touches.active = []ebiten.TouchID{1}
	touches.justPressed = []ebiten.TouchID{1}
	adapter.Update()
	if !adapter.IsTouchActive() {
		t.Error("касание должно взводить защёлку")
	}
	if adapter.FireJustPressed(types.PlayerTankNumPlayer1) ||
		adapter.PauseJustPressed() {
		t.Error("без геометрии экрана касания должны игнорироваться")
	}

	// Защёлка не сбрасывается после отпускания
	touches.active = nil
	touches.justPressed = nil
	adapter.Update()
	if !adapter.IsTouchActive() {
		t.Error("защёлка не должна сбрасываться")
	}
}

func TestTouchControlsAdapter_MultiTouchIndependence(t *testing.T) {
	touches := &fakeTouches{positions: map[ebiten.TouchID][2]int{}}
	adapter := newTestAdapter(touches)
	dpad := adapter.layout.Players[0].DPad
	fire := adapter.layout.Players[0].Fire

	// Палец 1 держит правое плечо крестовины
	dpadCenterX := float64(dpad.Min.X+dpad.Max.X) / 2
	dpadCenterY := float64(dpad.Min.Y+dpad.Max.Y) / 2
	touches.positions[1] = logicalAt(
		float64(dpad.Max.X)-1, dpadCenterY,
	)
	touches.active = []ebiten.TouchID{1}
	touches.justPressed = []ebiten.TouchID{1}
	adapter.Update()

	direction, ok := adapter.DPadDirection(types.PlayerTankNumPlayer1)
	if !ok || direction != types.DirectionRight {
		t.Fatalf(
			"ожидалось направление вправо, получено (%v,%v)",
			direction, ok,
		)
	}
	if adapter.FireJustPressed(types.PlayerTankNumPlayer1) {
		t.Error("огонь не нажимался")
	}

	// Палец 2 тапает огонь, палец 1 продолжает держать крестовину
	touches.positions[2] = logicalAt(
		float64(fire.Min.X+fire.Max.X)/2,
		float64(fire.Min.Y+fire.Max.Y)/2,
	)
	touches.active = []ebiten.TouchID{1, 2}
	touches.justPressed = []ebiten.TouchID{2}
	adapter.Update()

	if !adapter.FireJustPressed(types.PlayerTankNumPlayer1) {
		t.Error("тап по кнопке огня должен дать выстрел")
	}
	direction, ok = adapter.DPadDirection(types.PlayerTankNumPlayer1)
	if !ok || direction != types.DirectionRight {
		t.Error("крестовина должна продолжать держать направление")
	}

	// FireJustPressed — событие одного кадра
	touches.justPressed = nil
	adapter.Update()
	if adapter.FireJustPressed(types.PlayerTankNumPlayer1) {
		t.Error("выстрел не должен повторяться при удержании")
	}

	// Отпускание пальца 1 останавливает крестовину
	touches.active = []ebiten.TouchID{2}
	adapter.Update()
	if _, ok := adapter.DPadDirection(types.PlayerTankNumPlayer1); ok {
		t.Error("после отпускания направление должно сброситься")
	}

	// Палец в мёртвой зоне у центра не задаёт направление
	touches.positions[3] = logicalAt(dpadCenterX+1, dpadCenterY)
	touches.active = []ebiten.TouchID{2, 3}
	touches.justPressed = []ebiten.TouchID{3}
	adapter.Update()
	if _, ok := adapter.DPadDirection(types.PlayerTankNumPlayer1); ok {
		t.Error("в мёртвой зоне направления быть не должно")
	}
}

// Нажатие крестовины — шаг меню одного кадра; смена направления
// без отпускания — новый шаг
func TestTouchControlsAdapter_DPadJustPressed(t *testing.T) {
	touches := &fakeTouches{positions: map[ebiten.TouchID][2]int{}}
	adapter := newTestAdapter(touches)
	dpad := adapter.layout.Players[0].DPad
	centerX := float64(dpad.Min.X+dpad.Max.X) / 2
	centerY := float64(dpad.Min.Y+dpad.Max.Y) / 2

	touches.positions[1] = logicalAt(centerX, centerY-float64(dpad.Dy())/3)
	touches.active = []ebiten.TouchID{1}
	touches.justPressed = []ebiten.TouchID{1}
	adapter.Update()
	if direction, ok := adapter.DPadJustPressed(types.PlayerTankNumPlayer1); !ok ||
		direction != types.DirectionUp {
		t.Fatalf("ожидался шаг вверх, получено %v %v", direction, ok)
	}

	touches.justPressed = nil
	adapter.Update()
	if _, ok := adapter.DPadJustPressed(types.PlayerTankNumPlayer1); ok {
		t.Error("удержание не должно давать новых шагов")
	}

	touches.positions[1] = logicalAt(centerX, centerY+float64(dpad.Dy())/3)
	adapter.Update()
	if direction, ok := adapter.DPadJustPressed(types.PlayerTankNumPlayer1); !ok ||
		direction != types.DirectionDown {
		t.Errorf("смена направления — новый шаг, получено %v %v", direction, ok)
	}

	// Касание мимо контролов ничего не делает
	touches.positions[2] = logicalAt(centerX+float64(dpad.Dx())*3, centerY)
	touches.active = []ebiten.TouchID{2}
	touches.justPressed = []ebiten.TouchID{2}
	adapter.Update()
	if _, ok := adapter.DPadJustPressed(types.PlayerTankNumPlayer1); ok ||
		adapter.FireJustPressed(types.PlayerTankNumPlayer1) {
		t.Error("касание вне контролов не должно давать событий")
	}
}

func TestTouchControlsAdapter_PauseZone(t *testing.T) {
	touches := &fakeTouches{positions: map[ebiten.TouchID][2]int{}}
	adapter := newTestAdapter(touches)

	pause := adapter.layout.Pause
	touches.positions[1] = logicalAt(
		float64(pause.Min.X+pause.Max.X)/2,
		float64(pause.Min.Y+pause.Max.Y)/2,
	)
	touches.active = []ebiten.TouchID{1}
	touches.justPressed = []ebiten.TouchID{1}
	adapter.Update()

	if !adapter.PauseJustPressed() {
		t.Error("тап по зоне паузы должен дать событие паузы")
	}
}

// Касание игрового экрана — тап в логических координатах; касания
// контролов и полей тапом не считаются
func TestTouchControlsAdapter_TapOnGameScreen(t *testing.T) {
	touches := &fakeTouches{positions: map[ebiten.TouchID][2]int{}}
	adapter := newTestAdapter(touches)
	x, y, scale := adapter.GameRect()

	touches.positions[1] = logicalAt(
		float64(x+100*scale+scale/2), float64(y+50*scale+scale/2),
	)
	touches.active = []ebiten.TouchID{1}
	touches.justPressed = []ebiten.TouchID{1}
	adapter.Update()
	position, ok := adapter.TapJustPressed()
	// Тачи ebiten целочисленные в логике экрана: допуск в пиксель
	if !ok || math.Abs(position.X-100) > 1 || math.Abs(position.Y-50) > 1 {
		t.Errorf("тап (100, 50), получено %v %v", position, ok)
	}

	fire := adapter.layout.Players[0].Fire
	touches.positions[2] = logicalAt(
		float64(fire.Min.X+fire.Max.X)/2,
		float64(fire.Min.Y+fire.Max.Y)/2,
	)
	touches.active = []ebiten.TouchID{2}
	touches.justPressed = []ebiten.TouchID{2}
	adapter.Update()
	if _, ok := adapter.TapJustPressed(); ok {
		t.Error("касание огня не тап")
	}
}

// В режиме на двоих касания правой полосы управляют P2, левой — P1
func TestTouchControlsAdapter_TwoPlayers(t *testing.T) {
	touches := &fakeTouches{positions: map[ebiten.TouchID][2]int{}}
	settings := types.NewSettingsEntity()
	settings.SetPlayers(2)
	adapter := NewTouchControlsAdapter(256, 224, settings)
	adapter.appendTouchIDs = func(ids []ebiten.TouchID) []ebiten.TouchID {
		return append(ids, touches.active...)
	}
	adapter.appendJustPressedTouchIDs = func(
		ids []ebiten.TouchID,
	) []ebiten.TouchID {
		return append(ids, touches.justPressed...)
	}
	adapter.touchPosition = func(id ebiten.TouchID) (int, int) {
		pos := touches.positions[id]
		return pos[0], pos[1]
	}
	adapter.deviceScaleFactor = func() float64 { return 1 }
	adapter.SetScreenSize(2532, 1170)

	at := func(x, y float64) [2]int {
		lx, ly := ebitenScreenToLogical(x, y, 256, 224, 2532, 1170)
		return [2]int{int(lx), int(ly)}
	}
	second := adapter.layout.Players[1]
	touches.positions[1] = at(
		float64(second.DPad.Min.X)+float64(second.DPad.Dx())/6,
		float64(second.DPad.Min.Y+second.DPad.Max.Y)/2,
	)
	touches.positions[2] = at(
		float64(second.Fire.Min.X+second.Fire.Max.X)/2,
		float64(second.Fire.Min.Y+second.Fire.Max.Y)/2,
	)
	touches.active = []ebiten.TouchID{1, 2}
	touches.justPressed = []ebiten.TouchID{1, 2}
	adapter.Update()

	direction, ok := adapter.DPadDirection(types.PlayerTankNumPlayer2)
	if !ok || direction != types.DirectionLeft {
		t.Errorf("P2 должен ехать влево, получено %v %v", direction, ok)
	}
	if !adapter.FireJustPressed(types.PlayerTankNumPlayer2) {
		t.Error("огонь P2 должен сработать")
	}
	if _, ok := adapter.DPadDirection(types.PlayerTankNumPlayer1); ok ||
		adapter.FireJustPressed(types.PlayerTankNumPlayer1) {
		t.Error("касания P2 не должны управлять P1")
	}

	// Переход в одиночный режим отпускает контроллы P2
	settings.SetPlayers(1)
	adapter.SetScreenSize(2532, 1170)
	touches.justPressed = nil
	adapter.Update()
	if _, ok := adapter.DPadDirection(types.PlayerTankNumPlayer2); ok {
		t.Error("без второго игрока направления P2 нет")
	}
}

func TestTouchControlsAdapter_GamePosition(t *testing.T) {
	// До SetScreenSize геометрия неизвестна
	inert := NewTouchControlsAdapter(256, 224, types.NewSettingsEntity())
	if _, ok := inert.GamePosition(128, 112); ok {
		t.Error("до первой отрисовки позиций нет")
	}

	adapter := newTestAdapter(&fakeTouches{})
	position, ok := adapter.GamePosition(128, 112)
	if !ok || math.Abs(position.X-128) > 1 || math.Abs(position.Y-112) > 1 {
		t.Errorf("центр экрана — центр игры, получено %v, %v", position, ok)
	}

	// Угол логического экрана ebiten лежит на чёрном поле
	if _, ok := adapter.GamePosition(0, 0); ok {
		t.Error("точка на поле вне игры не переводится")
	}
}
