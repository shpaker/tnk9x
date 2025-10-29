package use_cases

import (
	"image"

	"github.com/shpaker/gonflict/internal/types"
)

// IMapObject определяет интерфейс для объектов карты
type IMapObject interface {
	GetSize() types.Size
	GetPosition() types.Position
	GetAltitude() types.Altitude
}

// ITankUseCasesRef интерфейс для базовых операций с танками
type ITankUseCasesRef interface {
	StartSpawn() error
	GetTank() *types.TankEntity
	Rotate(direction types.Direction) error
	Move() error
	StopTank(byCollision bool) error
	Update(dt float64) error
	StartExplosion() error
	IsActive() bool
	IsStopped() bool
	GetImageId() (string, error)
	GetAnimationGetter() types.IImageIdGetter
	Shoot() error
}

// IPlayerUseCases интерфейс для операций с игроком
type IPlayerUseCases interface {
	GetTank() *types.TankEntity
	UpdateSpawn(currentTime float64)
	GetTankImageId() (string, error)
}

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
	CheckCollidersWithArray(obj IMapObject, objects []IMapObject) []IMapObject
	CheckCollidersWithArrayFirst(obj IMapObject, objects []IMapObject) IMapObject
}

// ITilesUseCases определяет интерфейс для работы с тайлами и анимациями
type ITilesUseCases interface {
	CreateStaticTile(id string) (types.IImageIdGetter, error)
	CreateAnimationTile(id string) (*types.TileAnimationEntity, error)
	GetImage(id string) (image.Image, error)
	GetTileAnimationFrames(id string) (types.AnimationData, error)
	AddAnimation(animation *types.TileAnimationEntity)
	UpdateAnimations()
	StartAnimation(animation *types.TileAnimationEntity)
	StopAnimation(animation *types.TileAnimationEntity)
	CreateSpawnAnimation() (*types.TileAnimationEntity, error)
	CreateExplosionAnimation() (*types.TileAnimationEntity, error)
}
