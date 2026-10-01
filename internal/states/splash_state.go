package states

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// Тайминги сплеша в тиках (60 в секунду)
const (
	splashFadeInTicks  = 18
	splashMinTicks     = 120
	splashFadeOutTicks = 30
)

// Loader — загрузка ресурсов игры по шагу за вызов: доля готового
// и завершена ли загрузка; контракт определён у потребителя
type Loader interface {
	Step() (progress float64, done bool)
}

// SplashRenderer — контракт рендера сплеша, определён у потребителя
type SplashRenderer interface {
	Draw(screen *ebiten.Image, view types.SplashViewData)
}

// SplashState — сплеш-экран: выходит из чёрного, грузит ресурсы
// по шагу за кадр с полоской прогресса и, когда загрузка готова
// и прошло минимальное время, гаснет в чёрный и уступает главному
// меню. После загрузки ожидание можно пропустить
type SplashState struct {
	// Adapters
	renderer  SplashRenderer
	menuInput interfaces.IMenuInputAdapter
	loader    Loader

	ticks    uint
	progress float64
	loaded   bool
	// fadeOutTicks — тиков затухания; 0 — ещё не уходим
	fadeOutTicks uint
}

func NewSplashState(
	loader Loader,
	renderer SplashRenderer,
	menuInput interfaces.IMenuInputAdapter,
) *SplashState {
	return &SplashState{
		loader:    loader,
		renderer:  renderer,
		menuInput: menuInput,
	}
}

func (s *SplashState) Update() types.StateTransition {
	s.ticks++
	if !s.loaded {
		s.progress, s.loaded = s.loader.Step()
	}

	if s.fadeOutTicks > 0 {
		s.fadeOutTicks++
		if s.fadeOutTicks > splashFadeOutTicks {
			return types.StateTransition{Target: types.TransitionToMainMenu}
		}
		return types.StateTransition{}
	}

	skip := s.menuInput.Confirmed() || s.menuInput.Back()
	if s.loaded && (s.ticks >= splashMinTicks || skip) {
		s.fadeOutTicks = 1
	}
	return types.StateTransition{}
}

func (s *SplashState) Draw(screen *ebiten.Image) {
	fade := 0.0
	if s.ticks < splashFadeInTicks {
		fade = 1 - float64(s.ticks)/splashFadeInTicks
	}
	if s.fadeOutTicks > 0 {
		fade = float64(s.fadeOutTicks) / splashFadeOutTicks
	}
	s.renderer.Draw(screen, types.SplashViewData{
		Progress: s.progress,
		Fade:     fade,
	})
}
