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
	// TapJustPressed — новое касание игрового экрана вне контролов
	// в логических координатах экрана
	TapJustPressed() (types.Position, bool)
	// GamePosition переводит координаты ebiten (тач, курсор мыши)
	// в логические координаты нарисованного игрового экрана;
	// точки на полях вне игры — false
	GamePosition(x, y int) (types.Position, bool)
}

// IMenuInputAdapter — ввод меню и паузы, не зависящий от раскладки
// игроков: стрелки и WASD, Enter и Space, Esc; крестовина, стик,
// A, B и Start любого геймпада; тач-контролы любого игрока; мышь.
// Update опрашивается раз в кадр из game loop до обновления состояния
type IMenuInputAdapter interface {
	Update()
	// IsTouchActive — управление с экрана: подсказки на тач-вариантах
	IsTouchActive() bool
	// Steps — шаг вверх или вниз по пунктам меню
	Steps() (up bool, down bool)
	// SideStep — сдвиг влево (-1) или вправо (+1), 0 — без нажатия;
	// колесо мыши тоже даёт шаг
	SideStep() int
	// Confirmed — выбор пункта: Enter, Space, A или огонь
	Confirmed() bool
	// Back — выход из меню: Esc, B, Start, тач-пауза или правая
	// кнопка мыши
	Back() bool
	// PauseJustPressed — пауза: Esc, Start или тач-пауза
	PauseJustPressed() bool
	// Tapped — тап или клик левой кнопкой мыши по игровому экрану
	// в логических координатах: нажатие пунктов и экранных кнопок
	Tapped() (types.Position, bool)
	// Pointed — курсор мыши сдвинулся в этом кадре над игровым
	// экраном: наведение на пункт меню. Неподвижный курсор не
	// перебивает навигацию с клавиатуры
	Pointed() (types.Position, bool)
	// IsPointerActive — «мышь замечена хотя бы раз»: экранные кнопки
	// меню, которые без мыши и тача не нужны
	IsPointerActive() bool
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

// IWindowAdapter — окно приложения: полноэкранный режим и курсор
type IWindowAdapter interface {
	IsFullscreen() bool
	SetFullscreen(fullscreen bool)
	// SetCursorVisible показывает или прячет системный курсор мыши
	SetCursorVisible(visible bool)
}

// IPlatformAdapter — площадка, на которой запущена игра: десктоп,
// веб-страница или игровой портал. Контракт нейтрален к площадке:
// чего площадка не умеет, то её адаптер не делает. Опрашивается
// раз в кадр
type IPlatformAdapter interface {
	// Update опрашивает площадку; вызывается первым в кадре
	Update()
	// IsSuspended — площадка приостановила игру в этом кадре
	// (реклама, скрытая вкладка, собственная пауза площадки)
	IsSuspended() bool
	// IsJustSuspended — приостановка началась с прошлого кадра,
	// в том числе пропущенная целиком, пока кадры не шли
	IsJustSuspended() bool
	// Ready сообщает, что игра загружена и готова к игроку;
	// повторные вызовы игнорируются
	Ready()
	// SetGameplayActive размечает активный геймплей; вызывается
	// каждый кадр, площадке передаются только изменения
	SetGameplayActive(active bool)
	// RequestIntermission — логическая пауза игры: площадка может
	// показать межуровневую рекламу и на это время приостановить игру
	RequestIntermission()
	// GetLanguage — язык интерфейса площадки как его отдаёт площадка
	// ("ru", "ru-RU", "ru_RU.UTF-8"); пустая строка — неизвестен
	GetLanguage() string
}

// IRewardAdapter — реклама за вознаграждение по действию игрока
type IRewardAdapter interface {
	// IsRewardAvailable — площадка умеет показывать такую рекламу
	IsRewardAvailable() bool
	// RequestReward показывает рекламу; исход — через PollReward
	RequestReward()
	// PollReward — исход последнего запроса: Pending, пока реклама
	// показывается, итоговый Granted/Denied отдаётся один раз,
	// дальше — None
	PollReward() types.RewardStatus
}

// ITextsAdapter — тексты интерфейса на активном языке из файлов
// локалей. Ключа нет в языке — текст языка по умолчанию, нет нигде —
// сам ключ: пропуск виден на экране
type ITextsAdapter interface {
	// GetLanguage — активный язык
	GetLanguage() types.Language
	// SetLanguage делает активным язык с файлом локали
	SetLanguage(language types.Language) error
	// Get — текст без параметров
	Get(key types.TextKey) string
	// Format — текст с именованными параметрами шаблона
	Format(key types.TextKey, args types.TextArgs) string
	// Plural — текст в форме множественного числа для count;
	// count доступен шаблону как {{.Count}}
	Plural(key types.TextKey, count int, args types.TextArgs) string
	// GetOr — текст ключа, а без перевода — fallback: названия
	// уровней и пачек из файлов карт и кампании
	GetOr(key types.TextKey, fallback string) string
	// GetLanguageName — название языка на нём самом
	GetLanguageName(language types.Language) string
}
