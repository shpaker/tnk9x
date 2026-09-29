package touch_controls

import (
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
	adapter := NewTouchControlsAdapter(256, 224)
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
	adapter := NewTouchControlsAdapter(256, 224)
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
	if adapter.FireJustPressed() || adapter.PauseJustPressed() {
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
	dpad := adapter.layout.DPad
	fire := adapter.layout.Fire

	// Палец 1 держит правое плечо крестовины
	dpadCenterX := float64(dpad.Min.X+dpad.Max.X) / 2
	dpadCenterY := float64(dpad.Min.Y+dpad.Max.Y) / 2
	touches.positions[1] = logicalAt(
		float64(dpad.Max.X)-1, dpadCenterY,
	)
	touches.active = []ebiten.TouchID{1}
	touches.justPressed = []ebiten.TouchID{1}
	adapter.Update()

	direction, ok := adapter.DPadDirection()
	if !ok || direction != types.DirectionRight {
		t.Fatalf(
			"ожидалось направление вправо, получено (%v,%v)",
			direction, ok,
		)
	}
	if adapter.FireJustPressed() {
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

	if !adapter.FireJustPressed() {
		t.Error("тап по кнопке огня должен дать выстрел")
	}
	direction, ok = adapter.DPadDirection()
	if !ok || direction != types.DirectionRight {
		t.Error("крестовина должна продолжать держать направление")
	}

	// FireJustPressed — событие одного кадра
	touches.justPressed = nil
	adapter.Update()
	if adapter.FireJustPressed() {
		t.Error("выстрел не должен повторяться при удержании")
	}

	// Отпускание пальца 1 останавливает крестовину
	touches.active = []ebiten.TouchID{2}
	adapter.Update()
	if _, ok := adapter.DPadDirection(); ok {
		t.Error("после отпускания направление должно сброситься")
	}

	// Палец в мёртвой зоне у центра не задаёт направление
	touches.positions[3] = logicalAt(dpadCenterX+1, dpadCenterY)
	touches.active = []ebiten.TouchID{2, 3}
	touches.justPressed = []ebiten.TouchID{3}
	adapter.Update()
	if _, ok := adapter.DPadDirection(); ok {
		t.Error("в мёртвой зоне направления быть не должно")
	}
}

// Нажатие крестовины — шаг меню одного кадра; смена направления
// без отпускания — новый шаг
func TestTouchControlsAdapter_DPadJustPressed(t *testing.T) {
	touches := &fakeTouches{positions: map[ebiten.TouchID][2]int{}}
	adapter := newTestAdapter(touches)
	dpad := adapter.layout.DPad
	centerX := float64(dpad.Min.X+dpad.Max.X) / 2
	centerY := float64(dpad.Min.Y+dpad.Max.Y) / 2

	touches.positions[1] = logicalAt(centerX, centerY-float64(dpad.Dy())/3)
	touches.active = []ebiten.TouchID{1}
	touches.justPressed = []ebiten.TouchID{1}
	adapter.Update()
	if direction, ok := adapter.DPadJustPressed(); !ok ||
		direction != types.DirectionUp {
		t.Fatalf("ожидался шаг вверх, получено %v %v", direction, ok)
	}

	touches.justPressed = nil
	adapter.Update()
	if _, ok := adapter.DPadJustPressed(); ok {
		t.Error("удержание не должно давать новых шагов")
	}

	touches.positions[1] = logicalAt(centerX, centerY+float64(dpad.Dy())/3)
	adapter.Update()
	if direction, ok := adapter.DPadJustPressed(); !ok ||
		direction != types.DirectionDown {
		t.Errorf("смена направления — новый шаг, получено %v %v", direction, ok)
	}

	// Касание мимо контролов ничего не делает
	touches.positions[2] = logicalAt(centerX+float64(dpad.Dx())*3, centerY)
	touches.active = []ebiten.TouchID{2}
	touches.justPressed = []ebiten.TouchID{2}
	adapter.Update()
	if _, ok := adapter.DPadJustPressed(); ok || adapter.FireJustPressed() {
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
