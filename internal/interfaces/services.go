package interfaces

import (
	lua "github.com/yuin/gopher-lua"

	"github.com/shpaker/gonflict/internal/types"
	image_providers "github.com/shpaker/gonflict/internal/types/image_providers"
)

// ============================================================================
// Services Interfaces
// ============================================================================

// ITankBrakingService интерфейс для сервиса торможения танков
type ITankBrakingService interface {
	HandleBrakingState(tank *types.TankEntity, dt float64) error
	HandleRotateWhileBraking(tank *types.TankEntity, direction types.Direction)
}

// ICoordinateService интерфейс для сервиса работы с координатами
type ICoordinateService interface {
	RoundToNearestMultipleOf4(value float64) float64
}

// IBoundaryCollisionService интерфейс для сервиса коллизий с границами
type IBoundaryCollisionService interface {
	CheckTankBoundaryCollisions(tank *types.TankEntity, stopAndRound bool) bool
	CheckEnemyBoundaryCollisions(enemy *types.TankEntity) bool
	CheckBulletBoundaryCollisions(bullets []types.BulletEntity) []int
}

// IWallCollisionService интерфейс для сервиса коллизий со стенами
type IWallCollisionService interface {
	CheckEnemyWallCollision(
		enemy *types.TankEntity,
		level []types.BlockEntity,
	) *types.BlockEntity
	HandleEnemyWallCollision(enemy *types.TankEntity, block *types.BlockEntity)
	CheckTankWallCollision(
		tank *types.TankEntity,
		level []types.BlockEntity,
	) *types.BlockEntity
	HandleTankWallCollision(tank *types.TankEntity, block *types.BlockEntity)
}

// IBulletCollisionService интерфейс для сервиса коллизий пуль
type IBulletCollisionService interface {
	CheckBulletWallCollisions(
		bullets []types.BulletEntity,
		level []types.BlockEntity,
	) (bulletIndicesToRemove []int, blockIndicesToRemove []int)
	CheckBulletTankCollision(
		bullets []types.BulletEntity,
		tank *types.TankEntity,
	) []int
	CheckBulletEnemyCollisions(
		bullets []types.BulletEntity,
		enemies []*types.TankEntity,
	) (bulletIndicesToRemove []int, enemyIndicesToExplode map[int]int)
	CheckBulletHQCollision(
		bullets []types.BulletEntity,
		hq *types.HQEntity,
	) (bulletIndicesToRemove []int, hqDestroyed bool)
}

// IImageService интерфейс для сервиса работы с изображениями
// Использует interface{} для абстракции, конкретные реализации могут работать с конкретными типами
type IImageService interface {
	RotateImage(image interface{}, direction types.Direction) interface{}
	RotateImageByAngle(image interface{}, angle float64) (interface{}, error)
}

// ILogger интерфейс для логирования
type ILogger interface {
	Printf(format string, v ...interface{})
	Debugf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

// IAnimationService интерфейс для сервиса обновления анимаций
type IAnimationService interface {
	UpdateAnimation(animation *image_providers.AnimationProvider)
}

// ITileService интерфейс для сервиса работы с тайлами
type ITileService interface {
	GetTileAnimationFrames(id string) (types.AnimationData, error)
	GetAnimationConfig(id string) (types.AnimationConfig, error)
	CreateAnimationFromConfig(
		animationFrames types.AnimationData,
		config types.AnimationConfig,
	) *image_providers.AnimationProvider
	CreateAnimationTileFromRepo(
		repo ITilesetRepository,
		id string,
	) (*image_providers.AnimationProvider, error)
}

// IAITypeConverter интерфейс для конвертации типов между Go и Lua (Application Layer)
type IAITypeConverter interface {
	// TankToLua конвертирует TankEntity в Lua таблицу
	TankToLua(tank *types.TankEntity) (*lua.LTable, error)

	// ContextToLua конвертирует GameAiContext в Lua таблицу
	ContextToLua(context *types.GameAiContext) (*lua.LTable, error)

	// LuaToDecision конвертирует результаты Lua функции в EnemyAIDecision
	LuaToDecision(results []lua.LValue) (types.EnemyAIDecision, error)
}
