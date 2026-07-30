package interfaces

import (
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
)

type ITankBrakingService interface {
	HandleBrakingState(tank *types.TankEntity, dt float64, onIce bool) error
	HandleRotateWhileBraking(tank *types.TankEntity, direction types.Direction)
}

type IBoundaryCollisionService interface {
	CheckLeftBoundaryCollision(entity types.IEntityCollider) bool
	CheckRightBoundaryCollision(entity types.IEntityCollider) bool
	CheckTopBoundaryCollision(entity types.IEntityCollider) bool
	CheckBottomBoundaryCollision(entity types.IEntityCollider) bool
}

type IWallCollisionService interface {
	CheckTankWallCollision(
		tank *types.TankEntity,
		block *types.BlockEntity,
	) bool
}

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

type IEntitiesCollisionService interface {
	CheckColliders(obj1 types.IEntityCollider, obj2 types.IEntityCollider) bool
	ResolveCollisionPosition(
		entity types.IEntityCollider,
		obstacle types.IEntityCollider,
		direction types.Direction,
	) (types.Position, error)
}

type ISpawnCollisionService interface {
	IsSpawnerBlocked(
		position types.Position,
		size types.Size,
		tanks []*types.TankEntity,
	) bool
}

type IAnimationService interface {
	UpdateAnimation(animation *image_providers.AnimationProvider)
}

type ITileService interface {
	GetTileAnimationFrames(id string) (types.AnimationData, error)
	GetAnimationConfig(id string) (types.AnimationConfig, error)
	CreateAnimationFromConfig(
		animationFrames types.AnimationData,
		config types.AnimationConfig,
	) *image_providers.AnimationProvider
	CreateAnimationTileFromTileset(
		tilesetType types.TilesetType,
		id string,
	) (*image_providers.AnimationProvider, error)
}
