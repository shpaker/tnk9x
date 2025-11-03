package interfaces

import (
	"image"

	"github.com/shpaker/gonflict/internal/types"
	image_providers "github.com/shpaker/gonflict/internal/types/image_providers"
)

// ============================================================================
// ДРУГИЕ USE CASES
// ============================================================================

// IBulletUseCases интерфейс для операций с пулями
type IBulletUseCases interface {
	ShootBullet(tank *types.TankEntity) error
	UpdateBullets(dt float64) error
	GetBullets() []types.BulletEntity
	RemoveBullet(index int) error
}

// IMapUseCases интерфейс для операций с картой
type IMapUseCases interface {
	GetBlocks() []types.BlockEntity
	RemoveBlock(block *types.BlockEntity) error
}

// ICollisionUseCases интерфейс для операций с коллизиями
type ICollisionUseCases interface {
	UpdateCollisions() error
	CheckColliders(obj1 IMapObject, obj2 IMapObject) bool
	CheckCollidersWithArray(
		obj IMapObject,
		objects []IMapObject,
	) []IMapObject
	CheckCollidersWithArrayFirst(
		obj IMapObject,
		objects []IMapObject,
	) IMapObject
}

// ITilesUseCases интерфейс для работы с тайлами и анимациями
type ITilesUseCases interface {
	CreateStaticTile(id string) (IImageProvider, error)
	CreateAnimationTile(id string) (*image_providers.AnimationProvider, error)
	CreateSpawnAnimation() (*image_providers.AnimationProvider, error)
	CreateExplosionAnimation() (*image_providers.AnimationProvider, error)
	GetImage(id string) (image.Image, error)
	AddAnimation(animation *image_providers.AnimationProvider)
	UpdateAnimations()
	StartAnimation(animation *image_providers.AnimationProvider)
}

// ============================================================================
// TANK USE CASES INTERFACES
// ============================================================================

// ITankCommonUseCases интерфейс для общих операций с танком (движение)
type ITankCommonUseCases interface {
	Update(tank *types.TankEntity, dt float64) error
}

// ITankRenderUseCases интерфейс для графики и рендеринга танка
type ITankRenderUseCases interface {
	IsSpawnAnimationFinished(tank *types.TankEntity) bool
	IsExplosionAnimationFinished(tank *types.TankEntity) bool
}

// ITankLifecycleUseCases интерфейс для жизненного цикла танка
type ITankLifecycleUseCases interface {
	Spawn(tank *types.TankEntity) error
	Explode(tank *types.TankEntity) error
	IsSpawnFinished(tank *types.TankEntity, currentTime float64)
	IsExplosionFinished(tank *types.TankEntity)
}

// ITankActionsUseCases интерфейс для действий танка
type ITankActionsUseCases interface {
	Update(tank *types.TankEntity, dt float64) error
	Rotate(tank *types.TankEntity, direction types.Direction) error
	Move(tank *types.TankEntity) error
	Stop(tank *types.TankEntity, byCollision bool)
	IsStopped(tank *types.TankEntity) bool
	Shoot(tank *types.TankEntity) error
	ApplyDecision(tank *types.TankEntity, decision types.EnemyAIDecision)
}

// IAIUseCases интерфейс для операций с AI логикой
type IAIUseCases interface {
	ExecuteAI(
		tank *types.TankEntity,
		context *types.GameAiContext,
	) (types.EnemyAIDecision, error)
	Close()
}

// IHQUseCases интерфейс для операций с базой
type IHQUseCases interface {
	HandleBulletHit() (bulletIndicesToRemove []int, destroyed bool)
	Explode() error
	IsExplosionFinished()
}
