package use_cases

import (
	"github.com/shpaker/gonflict/internal/types"
)

// IMapObject определяет интерфейс для объектов карты
type IMapObject interface {
	GetSize() types.Size
	GetWorldPosition() types.Position
	GetScreenPosition() types.Position
}

// IPlayerUseCases интерфейс для операций с игроком
type IPlayerUseCases interface {
	MovePlayer(direction types.Direction, dt float64) error
	RotatePlayer(direction types.Direction) error
	StopPlayer(byCollision bool) error
	GetPlayer() (*types.TankEntity, error)
	GetDirection() types.Direction
	UpdateTankAnimation()
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
