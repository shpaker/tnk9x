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
	CheckLeftBoundaryCollision(entity types.IEntityCollider) bool
	CheckRightBoundaryCollision(entity types.IEntityCollider) bool
	CheckTopBoundaryCollision(entity types.IEntityCollider) bool
	CheckBottomBoundaryCollision(entity types.IEntityCollider) bool
}

// IWallCollisionService интерфейс для сервиса коллизий со стенами
type IWallCollisionService interface {
	CheckTankWallCollision(
		tank *types.TankEntity,
		block *types.BlockEntity,
	) bool
}

// IBulletCollisionService интерфейс для сервиса коллизий пуль
type IBulletCollisionService interface {
	CheckBulletBlockCollision(
		bullet *types.BulletEntity,
		block *types.BlockEntity,
	) bool
	CheckBulletTankCollision(
		bullet *types.BulletEntity,
		tank *types.TankEntity,
	) bool
	CheckBulletHQCollision(
		bullet *types.BulletEntity,
		hq *types.HQEntity,
	) bool
}

// IEntitiesCollisionService интерфейс для сервиса проверки коллизий между сущностями
type IEntitiesCollisionService interface {
	CheckColliders(obj1 types.IEntityCollider, obj2 types.IEntityCollider) bool
	ResolveCollisionPosition(
		entity types.IEntityCollider,
		obstacle types.IEntityCollider,
		direction types.Direction,
	) (types.Position, error)
}

// IImageService интерфейс для сервиса работы с изображениями
// Использует interface{} для абстракции, конкретные реализации могут работать с конкретными типами
type IImageService interface {
	RotateImage(image interface{}, direction types.Direction) interface{}
	RotateImageByAngle(image interface{}, angle float64) (interface{}, error)
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
	CreateAnimationTileFromTileset(
		tilesetType string,
		id string,
	) (*image_providers.AnimationProvider, error)
}

// IAITypeConverter интерфейс для конвертации типов между Go и Lua (Application Layer)
type IAITypeConverter interface {
	// TankToLua конвертирует TankEntity в Lua таблицу
	TankToLua(tank *types.TankEntity) (*lua.LTable, error)

	// LuaToDecision конвертирует результаты Lua функции в EnemyAIDecision
	LuaToDecision(results []lua.LValue) (types.EnemyAIDecision, error)
}
