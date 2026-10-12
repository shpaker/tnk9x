package interfaces

import (
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
)

type ITankCollisionService interface {
	Response(
		tank *types.TankEntity,
		other types.IEntityCollider,
	) types.CollisionResponse
}

type IBulletCollisionService interface {
	Response(
		bullet *types.BulletEntity,
		other types.IEntityCollider,
	) types.CollisionResponse
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

type INavigationService interface {
	FindPath(
		grid *types.NavGridEntity,
		from, to types.Position,
		options types.NavOptions,
	) (types.PathStep, bool)
	CastRay(
		grid *types.NavGridEntity,
		origin types.Position,
		direction types.Direction,
		targets []types.RayTarget,
	) types.RayHit
}

type IRandomService interface {
	Float64() float64
}
