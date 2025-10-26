package use_cases

import (
	"image"

	"github.com/shpaker/gonflict/internal/types"
)

// IMapObject определяет интерфейс для объектов карты
type IMapObject interface {
	GetSize() types.Size
	GetWorldPosition() types.Position
	GetScreenPosition() types.Position
	GetAltitude() types.Altitude
}

// ITankUseCases интерфейс для операций с танком
type ITankUseCases interface {
	MoveTank(direction types.Direction, dt float64) error
	RotateTank(direction types.Direction) error
	StopTank(byCollision bool) error
	GetTank() (*types.TankEntity, error)
	GetDirection() types.Direction
	StartSpawn(spawnStartTime float64)
	UpdateSpawn(currentTime float64)
	IsSpawning() bool
	GetSpawnAnimation() *types.TileAnimationEntity
	GetTankImageId() (string, error)
	ShouldShowTank() bool
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

// ITilesUseCases определяет интерфейс для работы с тайлами
type ITilesUseCases interface {
	CreateStaticTile(id string) (types.IImageIdGetter, error)
	CreateAnimationTile(id string) (*types.TileAnimationEntity, error)
	GetImage(id string) (image.Image, error)
	GetTileAnimationFrames(id string) (types.AnimationData, error)
}

// IAnimationUseCases интерфейс для управления анимацией
type IAnimationUseCases interface {
	AddAnimation(animation *types.TileAnimationEntity)
	UpdateAnimations()
	StartAnimation(animation *types.TileAnimationEntity)
	StopAnimation(animation *types.TileAnimationEntity)
}

// IEnemyUseCases интерфейс для работы с врагами
type IEnemyUseCases interface {
	GetEnemies() []*types.TankEntity
	RemoveEnemy(index int) error
	SpawnEnemy(position types.Position) error
	InitEnemies(enemySpawners [][]int) error
	UpdateEnemiesSpawn(currentTime float64)
	GetEnemySpawnAnimation(enemyIndex int) *types.TileAnimationEntity
}
