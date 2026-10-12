package interfaces

import (
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
)

// ITankCollisionService решает, как едущий танк встречает другую
// сущность
type ITankCollisionService interface {
	Response(
		tank *types.TankEntity,
		other types.IEntityCollider,
	) types.CollisionResponse
}

// IBulletCollisionService решает, что пуля встречает на пути
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

// INavigationService отвечает на навигационные запросы AI по сетке карты:
// поиск пути и первое препятствие на линии огня
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

// IRandomService — источник случайных чисел; подменяется в тестах
type IRandomService interface {
	// Float64 возвращает число из [0, 1)
	Float64() float64
}
