package app

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/tnk9x/internal/states"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
)

// Тесты запускаются из каталога пакета, а config.yml и assets
// лежат в корне репозитория — переходим туда до запуска
func TestMain(m *testing.M) {
	if err := os.Chdir("../.."); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

type stubGameState struct{}

func (s *stubGameState) Update() types.StateTransition {
	return types.StateTransition{}
}

func (s *stubGameState) Draw(screen *ebiten.Image) {}

var errLevelUnavailable = errors.New("level unavailable")

// failingMapsRepo валит сборку уровня на первом же шаге newStageState
type failingMapsRepo struct{}

func (r *failingMapsRepo) GetLevel(
	num int,
	tileBaseSize int,
) (*types.LevelEntity, error) {
	return nil, errLevelUnavailable
}

func (r *failingMapsRepo) HasLevel(num int) bool { return false }

func (r *failingMapsRepo) GetLevelsCount() (int, error) { return 0, nil }

// newAppTestEnv собирает минимальный App без тяжёлой инфраструктуры:
// приватные поля доступны, так как тест живёт в пакете app
func newAppTestEnv() (*App, *stubGameState) {
	state := &stubGameState{}
	settings := types.NewSettingsEntity()
	settings.SetPlayers(2)
	app := &App{
		config: &Config{
			TileBaseSize: 8,
			BaseSizePx:   16,
		},
		settings:       settings,
		state:          state,
		session:        session_entities.NewGameSessionEntity(),
		mapsRepository: &failingMapsRepo{},
	}
	return app, state
}

// Полный App через New() собирается один раз на процесс:
// audio.NewContext допускает только один вызов
var (
	fullAppOnce sync.Once
	fullApp     *App
	fullAppErr  error
)

func newFullApp(t *testing.T) *App {
	t.Helper()
	fullAppOnce.Do(func() {
		cfg, err := LoadConfig()
		if err != nil {
			fullAppErr = err
			return
		}
		fullApp = New(cfg)
	})
	if fullAppErr != nil {
		t.Fatalf("конфигурация не загружена: %v", fullAppErr)
	}
	return fullApp
}

func TestApp_ApplyTransition_None(t *testing.T) {
	app, state := newAppTestEnv()

	if err := app.applyTransition(types.StateTransition{}); err != nil {
		t.Fatalf("пустой переход: %v", err)
	}
	if app.state != state {
		t.Error("состояние заменено при пустом переходе")
	}
}

// Параметры уровня записываются в сессию до сборки состояния:
// даже при ошибке сборки сессия уже обновлена
func TestApp_ApplyTransition_ToStage_SessionWrittenBeforeBuild(
	t *testing.T,
) {
	app, state := newAppTestEnv()

	err := app.applyTransition(types.StateTransition{
		Target: types.TransitionToStage,
		Level:  7,
	})
	if !errors.Is(err, errLevelUnavailable) {
		t.Fatalf("ошибка %v, ожидалась errLevelUnavailable", err)
	}

	if app.session.Level != 7 {
		t.Errorf("уровень сессии %d, ожидался 7", app.session.Level)
	}
	stageSession := app.session.StageSession()
	if got := stageSession.GetPlayerCount(); got != 2 {
		t.Errorf("players %d, want 2 from settings", got)
	}
	if got := stageSession.GetStageNumber(); got != 7 {
		t.Errorf("stage number %d, want 7", got)
	}

	// Состояние при ошибке сборки не меняется
	if app.state != state {
		t.Error("состояние заменено при ошибке сборки уровня")
	}
}

// Переход Quit завершает игровой цикл штатно, не трогая состояние
func TestApp_ApplyTransition_Quit(t *testing.T) {
	app, state := newAppTestEnv()

	err := app.applyTransition(types.StateTransition{
		Target: types.TransitionToQuit,
	})
	if !errors.Is(err, ebiten.Termination) {
		t.Fatalf("ошибка %v, ожидалась ebiten.Termination", err)
	}
	if app.state != state {
		t.Error("состояние заменено при переходе Quit")
	}
}

// Полный цикл на настоящем графе зависимостей:
// выбор уровня -> уровень -> выбор уровня
func TestApp_ApplyTransition_FullApp(t *testing.T) {
	app := newFullApp(t)

	if _, ok := app.state.(*states.LevelSelectState); !ok {
		t.Fatalf(
			"initial state %T, want LevelSelectState",
			app.state,
		)
	}

	err := app.applyTransition(types.StateTransition{
		Target: types.TransitionToStage,
		Level:  2,
	})
	if err != nil {
		t.Fatalf("переход на уровень: %v", err)
	}

	if _, ok := app.state.(*states.StageState); !ok {
		t.Fatalf("состояние %T, ожидалось StageState", app.state)
	}
	if app.session.Level != 2 {
		t.Errorf("уровень сессии %d, ожидался 2", app.session.Level)
	}

	if got := app.session.StageSession().GetTotalEnemies(); got == 0 {
		t.Error("the level waves did not reach the stage session")
	}

	err = app.applyTransition(types.StateTransition{
		Target: types.TransitionToLevelSelect,
		Level:  2,
	})
	if err != nil {
		t.Fatalf("back to stage select: %v", err)
	}
	if _, ok := app.state.(*states.LevelSelectState); !ok {
		t.Fatalf("state %T, want LevelSelectState", app.state)
	}
}

// Отмена контекста завершает игровой цикл штатно (ebiten.Termination)
func TestApp_Update_ContextCancelled(t *testing.T) {
	app, _ := newAppTestEnv()

	ctx, cancel := context.WithCancel(context.Background())
	app.ctx = ctx

	cancel()
	if err := app.Update(); !errors.Is(err, ebiten.Termination) {
		t.Fatalf("ожидался ebiten.Termination, получено %v", err)
	}
}
