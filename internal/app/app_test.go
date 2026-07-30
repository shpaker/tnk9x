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
) (*types.MapEntity, error) {
	return nil, errLevelUnavailable
}

func (r *failingMapsRepo) GetLevelsCount() (int, error) { return 0, nil }

// newAppTestEnv собирает минимальный App без тяжёлой инфраструктуры:
// приватные поля доступны, так как тест живёт в пакете app
func newAppTestEnv() (*App, *stubGameState) {
	state := &stubGameState{}
	app := &App{
		config:         &Config{TileBaseSize: 8, BaseSizePx: 16},
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

// Параметры уровня записываются в сессию забега до сборки состояния:
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

	if got := app.session.RunSession().GetStage(); got != 7 {
		t.Errorf("этап забега %d, ожидался 7", got)
	}
	// Лимит врагов берётся из конфига; нулевое значение конфига
	// нормализуется сессией к канону NES (4)
	stageSession := app.session.StageSession()
	if got := stageSession.GetMaxActiveEnemies(); got != 4 {
		t.Errorf("максимум врагов %d, ожидалось 4", got)
	}

	// Состояние при ошибке сборки не меняется
	if app.state != state || app.stageState != nil {
		t.Error("состояние заменено при ошибке сборки уровня")
	}
}

// Полный цикл на настоящем графе зависимостей:
// титул -> шторка -> уровень -> итоги -> шторка -> game over -> титул
func TestApp_ApplyTransition_FullApp(t *testing.T) {
	app := newFullApp(t)

	if _, ok := app.state.(*states.TitleState); !ok {
		t.Fatalf(
			"начальное состояние %T, ожидалось TitleState",
			app.state,
		)
	}

	err := app.applyTransition(types.StateTransition{
		Target:      types.TransitionToCurtain,
		Level:       2,
		PlayerCount: 1,
		NewRun:      true,
	})
	if err != nil {
		t.Fatalf("переход на шторку: %v", err)
	}
	if _, ok := app.state.(*states.StageCurtainState); !ok {
		t.Fatalf("состояние %T, ожидалось StageCurtainState", app.state)
	}
	if got := app.session.RunSession().GetStage(); got != 2 {
		t.Errorf("этап забега %d, ожидался 2", got)
	}

	err = app.applyTransition(types.StateTransition{
		Target: types.TransitionToStage,
		Level:  2,
	})
	if err != nil {
		t.Fatalf("переход на уровень: %v", err)
	}
	stageState, ok := app.state.(*states.StageState)
	if !ok {
		t.Fatalf("состояние %T, ожидалось StageState", app.state)
	}
	if app.stageState != stageState {
		t.Error("ссылка на StageState не сохранена")
	}

	err = app.applyTransition(types.StateTransition{
		Target:   types.TransitionToScore,
		StageWon: true,
	})
	if err != nil {
		t.Fatalf("переход на итоги: %v", err)
	}
	if _, ok := app.state.(*states.ScoreState); !ok {
		t.Fatalf("состояние %T, ожидалось ScoreState", app.state)
	}
	if app.stageState != nil {
		t.Error("ссылка на StageState не очищена")
	}

	err = app.applyTransition(types.StateTransition{
		Target: types.TransitionToCurtain,
		Level:  3,
	})
	if err != nil {
		t.Fatalf("переход на шторку следующего этапа: %v", err)
	}
	if _, ok := app.state.(*states.StageCurtainState); !ok {
		t.Fatalf("состояние %T, ожидалось StageCurtainState", app.state)
	}
	if got := app.session.RunSession().GetStage(); got != 3 {
		t.Errorf("этап забега %d, ожидался 3", got)
	}

	if err := app.applyTransition(types.StateTransition{
		Target: types.TransitionToGameOver,
	}); err != nil {
		t.Fatalf("переход на game over: %v", err)
	}
	if _, ok := app.state.(*states.GameOverState); !ok {
		t.Fatalf("состояние %T, ожидалось GameOverState", app.state)
	}

	if err := app.applyTransition(types.StateTransition{
		Target: types.TransitionToTitle,
	}); err != nil {
		t.Fatalf("возврат на титул: %v", err)
	}
	if _, ok := app.state.(*states.TitleState); !ok {
		t.Fatalf("состояние %T, ожидалось TitleState", app.state)
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
