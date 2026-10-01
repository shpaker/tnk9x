package states

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// SettingsRenderer — контракт рендера экрана настроек,
// определён у потребителя
type SettingsRenderer interface {
	Draw(screen *ebiten.Image, view types.SettingsViewData)
}

// SettingsOverlay — экран настроек поверх выбора уровня или меню
// паузы; один на приложение. Изменения сразу применяются к звуку
// и окну и сохраняются; CONTROLS открывает экран раскладки
type SettingsOverlay struct {
	// Use Cases
	settingsUseCases interfaces.ISettingsUseCases
	// Adapters
	renderer    SettingsRenderer
	menuInput   interfaces.IMenuInputAdapter
	soundPlayer interfaces.ISoundPlayerAdapter
	window      interfaces.IWindowAdapter
	// Presentation
	controlsOverlay *ControlsOverlay
	// Entities
	settings *types.SettingsEntity

	open        bool
	activeIndex int
}

func NewSettingsOverlay(
	settingsUseCases interfaces.ISettingsUseCases,
	renderer SettingsRenderer,
	menuInput interfaces.IMenuInputAdapter,
	soundPlayer interfaces.ISoundPlayerAdapter,
	window interfaces.IWindowAdapter,
	controlsOverlay *ControlsOverlay,
	settings *types.SettingsEntity,
) *SettingsOverlay {
	return &SettingsOverlay{
		settingsUseCases: settingsUseCases,
		renderer:         renderer,
		menuInput:        menuInput,
		soundPlayer:      soundPlayer,
		window:           window,
		controlsOverlay:  controlsOverlay,
		settings:         settings,
	}
}

// Open показывает экран с курсором на первой строке; полный экран
// сверяется с окном — из него могли выйти средствами ОС или браузера
func (o *SettingsOverlay) Open() {
	o.open = true
	o.activeIndex = 0
	o.settings.SetFullscreen(o.window.IsFullscreen())
}

func (o *SettingsOverlay) IsOpen() bool {
	return o.open
}

// IsCapturing — на экране раскладки ждём нажатия для назначения
func (o *SettingsOverlay) IsCapturing() bool {
	return o.open && o.controlsOverlay.IsCapturing()
}

// Update: вверх-вниз — выбор строки, влево-вправо или выбор —
// смена значения; Esc, B, Start, пауза или BACK закрывают экран.
// Пока открыт экран раскладки, ввод уходит в него
func (o *SettingsOverlay) Update() {
	if o.controlsOverlay.IsOpen() {
		o.controlsOverlay.Update()
		return
	}
	if o.menuInput.Back() {
		o.open = false
		return
	}

	items := o.settingsUseCases.Items()
	moveUp, moveDown := o.menuInput.Steps()
	o.activeIndex = stepIndex(o.activeIndex, len(items), moveUp, moveDown)
	item := items[o.activeIndex]

	step := o.menuInput.SideStep()
	if o.menuInput.Confirmed() {
		switch item {
		case types.SettingsItemBack:
			o.open = false
			return
		case types.SettingsItemControls:
			o.controlsOverlay.Open()
			return
		}
		step = 1
	}
	if step != 0 {
		o.change(item, step)
	}
}

func (o *SettingsOverlay) Draw(screen *ebiten.Image) {
	if o.controlsOverlay.IsOpen() {
		o.controlsOverlay.Draw(screen)
		return
	}
	o.renderer.Draw(
		screen,
		o.settingsUseCases.BuildView(o.settings, o.activeIndex),
	)
}

// change меняет значение пункта и применяет его к звуку и окну;
// ошибка сохранения не мешает игре
func (o *SettingsOverlay) change(item types.SettingsItem, step int) {
	if err := o.settingsUseCases.Change(o.settings, item, step); err != nil {
		log.Printf("save settings: %v", err)
	}

	switch item {
	case types.SettingsItemVolume:
		o.soundPlayer.SetVolume(o.settings.GetVolume())
		// Короткий звук — новая громкость слышна сразу
		if err := o.soundPlayer.Play(types.SoundIDScore); err != nil {
			log.Printf("sound %q: %v", types.SoundIDScore, err)
		}
	case types.SettingsItemFullscreen:
		o.window.SetFullscreen(o.settings.IsFullscreen())
	}
}
