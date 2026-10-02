package states

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// helpStage — уровень, перед первым запуском которого показывается
// страница «Как играть»
const helpStage = 1

// HelpRenderer — контракт рендера страницы «Как играть»,
// определён у потребителя
type HelpRenderer interface {
	Draw(screen *ebiten.Image, view types.HelpViewData)
}

// HelpOverlay — страница «Как играть»: цель и управление перед
// самым первым запуском уровня 1; после закрытия больше не
// показывается. Одна на приложение
type HelpOverlay struct {
	// Use Cases
	settingsUseCases interfaces.ISettingsUseCases
	controlsUseCases interfaces.IControlsUseCases
	// Adapters
	renderer  HelpRenderer
	menuInput interfaces.IMenuInputAdapter
	// Entities
	settings *types.SettingsEntity
	controls *types.ControlsEntity

	players uint
	ticks   uint
}

func NewHelpOverlay(
	settingsUseCases interfaces.ISettingsUseCases,
	controlsUseCases interfaces.IControlsUseCases,
	renderer HelpRenderer,
	menuInput interfaces.IMenuInputAdapter,
	settings *types.SettingsEntity,
	controls *types.ControlsEntity,
) *HelpOverlay {
	return &HelpOverlay{
		settingsUseCases: settingsUseCases,
		controlsUseCases: controlsUseCases,
		renderer:         renderer,
		menuInput:        menuInput,
		settings:         settings,
		controls:         controls,
	}
}

// IsDue — страницу нужно показать перед запуском уровня stage
func (o *HelpOverlay) IsDue(stage uint) bool {
	return stage == helpStage && !o.settings.IsHelpShown()
}

// Update: выбор, «назад», пауза или тап закрывают страницу
// и запоминают, что она показана
func (o *HelpOverlay) Update(players uint) {
	o.players = players
	o.ticks++
	closed := o.menuInput.Confirmed() || o.menuInput.Back() ||
		o.menuInput.PauseJustPressed() || tappedAnywhere(o.menuInput)
	if !closed {
		return
	}
	if err := o.settingsUseCases.MarkHelpShown(o.settings); err != nil {
		log.Printf("save settings: %v", err)
	}
	o.ticks = 0
}

func (o *HelpOverlay) Draw(screen *ebiten.Image) {
	o.renderer.Draw(screen, types.HelpViewData{
		Players:     o.players,
		TouchActive: o.menuInput.IsTouchActive(),
		Rows:        o.controlsUseCases.HelpRows(o.controls),
		Ticks:       o.ticks,
	})
}
