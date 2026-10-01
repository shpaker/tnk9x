package interfaces

import (
	"github.com/shpaker/tnk9x/internal/types"
)

type IConfigProvider interface {
	GetEnemySpawners() []types.Position
	GetPlayer1Spawn() types.Position
	GetPlayer2Spawn() types.Position
	GetHQPosition() [2]int
	GetEnemyRespawnDelayTicks() uint
	GetBaseSizePx() uint
	GetMapBlocksCount() types.Size
	GetTileBaseSize() uint
	GetTitleFontSize() uint
	GetSubtitleFontSize() uint
	GetRegularFontSize() uint
	GetGameTitle() string
}

type IInputAdapter interface {
	Update(dt float64)
}

type IInputAdapterWithTank interface {
	IInputAdapter
	SetPlayerTank(tank *types.TankEntity)
}

type IAiInputAdapter interface {
	IInputAdapter
	AddTank(tank *types.TankEntity)
	RemoveTank(tank *types.TankEntity)
}

// ITouchControlsAdapter — источник событий сенсорного управления;
// Update опрашивается раз в кадр из game loop до обновления состояния.
// В режиме на двоих у каждого игрока своя крестовина и огонь,
// пауза общая
type ITouchControlsAdapter interface {
	Update()
	// IsTouchActive — «тач замечен хотя бы раз»: защёлка активации
	// экранных контролов
	IsTouchActive() bool
	// DPadDirection — текущее направление крестовины игрока
	DPadDirection(player types.PlayerTankNum) (types.Direction, bool)
	// DPadJustPressed — крестовину игрока нажали или сменили
	// направление в этом кадре: шаг по пунктам меню
	DPadJustPressed(player types.PlayerTankNum) (types.Direction, bool)
	FireJustPressed(player types.PlayerTankNum) bool
	PauseJustPressed() bool
}

// IMenuInputAdapter — ввод меню и паузы, не зависящий от раскладки
// игроков: стрелки и WASD, Enter и Space, Esc; крестовина, стик,
// A, B и Start любого геймпада; тач-контролы любого игрока.
// Update опрашивается раз в кадр из game loop до обновления состояния
type IMenuInputAdapter interface {
	Update()
	// IsTouchActive — управление с экрана: подсказки на тач-вариантах
	IsTouchActive() bool
	// Steps — шаг вверх или вниз по пунктам меню
	Steps() (up bool, down bool)
	// SideStep — сдвиг влево (-1) или вправо (+1), 0 — без нажатия
	SideStep() int
	// Confirmed — выбор пункта: Enter, Space, A или огонь
	Confirmed() bool
	// Back — выход из меню: Esc, B, Start или тач-пауза
	Back() bool
	// PauseJustPressed — пауза: Esc, Start или тач-пауза
	PauseJustPressed() bool
}

// IHotkeysAdapter — нажатые в этом кадре глобальные хоткеи
// по раскладке игрока
type IHotkeysAdapter interface {
	JustPressed() []types.HotkeyAction
}

// IBindingCaptureAdapter — ввод для назначения раскладки: нажатые
// в этом кадре клавиша или кнопка любого геймпада в текстовых именах
type IBindingCaptureAdapter interface {
	JustPressedKey() (string, bool)
	JustPressedPadButton() (string, bool)
}

// IAIScriptEngine — контракт движка AI-скриптов; реализация инкапсулирует
// скриптовый рантайм, наружу выходят только доменные типы.
type IAIScriptEngine interface {
	LoadScript(source string) error
	SetGlobalNumber(name string, value float64)
	UpdateEnemyAI(context types.EnemyAIContext) (types.EnemyAIDecision, error)
	Close()
}

type ISoundPlayerAdapter interface {
	Play(soundID types.SoundID) error
	PlayLoop(soundID types.SoundID) error
	Stop(soundID types.SoundID)
	StopAll()
	Update()
	// SetVolume меняет громкость (0..1) и уже играющих звуков
	SetVolume(volume float64)
}

// IWindowAdapter — окно приложения: полноэкранный режим
type IWindowAdapter interface {
	IsFullscreen() bool
	SetFullscreen(fullscreen bool)
}
