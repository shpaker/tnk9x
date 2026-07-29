package interfaces

import (
	"github.com/shpaker/tnk9x/internal/types"
)

type IConfigProvider interface {
	GetEnemySpawners() []types.Position
	GetPlayer1Spawn() types.Position
	GetPlayer2Spawn() types.Position
	GetHQPosition() [2]int
	GetAIUpdateIntervalTicks() int
	GetEnemyRespawnDelayTicks() uint
	GetBaseSizePx() uint
	GetMapBlocksCount() types.Size
	GetTileBaseSize() uint
	GetTitleFontSize() uint
	GetSubtitleFontSize() uint
	GetRegularFontSize() uint
	GetGameTitle() string
	GetVolume() float64
}

type IScreenConfig interface {
	ScreenWidth() int
	ScreenHeight() int
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

// IAIScriptEngine — контракт движка AI-скриптов; реализация инкапсулирует
// скриптовый рантайм, наружу выходят только доменные типы.
type IAIScriptEngine interface {
	LoadScript(source string) error
	SetGlobalNumber(name string, value float64)
	UpdateEnemyAI(
		x, y float64,
		direction, state int,
	) (types.EnemyAIDecision, error)
	Close()
}

type ISoundPlayerAdapter interface {
	Play(soundID types.SoundID) error
	PlayLoop(soundID types.SoundID) error
	Stop(soundID types.SoundID) error
	StopAll()
	Update() error
}
