package collision_services

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.IBulletCollisionService = (*BulletCollisionService)(nil)

// BulletCollisionService решает, что пуля встречает на пути. Пуля
// проходит всё насквозь (CollisionCross): что из встреченного её
// гасит и в каком порядке, решают use cases
type BulletCollisionService struct{}

func NewBulletCollisionService() *BulletCollisionService {
	return &BulletCollisionService{}
}

// Response реализует IBulletCollisionService
func (s *BulletCollisionService) Response(
	bullet *types.BulletEntity,
	other types.IEntityCollider,
) types.CollisionResponse {
	switch other := other.(type) {
	case *types.BorderEntity:
		return types.CollisionCross
	case *types.BlockEntity:
		// Вода, лёд и лес пулю не останавливают
		if other.Data == nil || other.Data.Name == types.Water ||
			other.GetAltitude() != bullet.GetAltitude() {
			return types.CollisionIgnore
		}
		return types.CollisionCross
	case *types.HQEntity:
		if bullet.GetOwner() == nil ||
			other.GetAltitude() != bullet.GetAltitude() {
			return types.CollisionIgnore
		}
		return types.CollisionCross
	case *types.TankEntity:
		if s.hitsTank(bullet, other) {
			return types.CollisionCross
		}
	case *types.BulletEntity:
		if other.GetAltitude() == bullet.GetAltitude() {
			return types.CollisionCross
		}
	}
	return types.CollisionIgnore
}

// hitsTank — пуля бьёт активный танк чужой стороны: враги друг друга
// не задевают, свой танк пуля не трогает
func (s *BulletCollisionService) hitsTank(
	bullet *types.BulletEntity,
	tank *types.TankEntity,
) bool {
	owner := bullet.GetOwner()
	if owner == nil || owner == tank || !tank.IsActive() {
		return false
	}
	if tank.IsEnemy() && owner.IsEnemy() {
		return false
	}
	return tank.GetAltitude() == bullet.GetAltitude()
}
