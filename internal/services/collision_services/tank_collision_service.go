package collision_services

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.ITankCollisionService = (*TankCollisionService)(nil)

// TankCollisionService решает, как едущий танк встречает сущности:
// стены, край карты, целый штаб и другие танки его останавливают,
// бонусы и пули встречаются на ходу
type TankCollisionService struct {
	// enemyBonusPickup — враги подбирают бонусы (game.enemy_bonus_pickup)
	enemyBonusPickup bool
}

func NewTankCollisionService(enemyBonusPickup bool) *TankCollisionService {
	return &TankCollisionService{enemyBonusPickup: enemyBonusPickup}
}

// Response реализует ITankCollisionService
func (s *TankCollisionService) Response(
	tank *types.TankEntity,
	other types.IEntityCollider,
) types.CollisionResponse {
	if other.GetAltitude() != tank.GetAltitude() {
		return types.CollisionIgnore
	}
	switch other := other.(type) {
	case *types.BorderEntity:
		return types.CollisionSlide
	case *types.BlockEntity:
		if other.Data == nil {
			return types.CollisionIgnore
		}
		// Танк с лодкой проходит по воде
		if other.Data.Name == types.Water && tank.CanSail() {
			return types.CollisionIgnore
		}
		return types.CollisionSlide
	case *types.HQEntity:
		// Целый орёл — препятствие, разрушенный проходим
		if other.IsIntact() {
			return types.CollisionSlide
		}
	case *types.TankEntity:
		if other != tank && !other.IsDestroyed() {
			return types.CollisionSlide
		}
	case *types.BonusEntity:
		if !tank.IsEnemy() || s.enemyBonusPickup {
			return types.CollisionCross
		}
	case *types.BulletEntity:
		return types.CollisionCross
	}
	return types.CollisionIgnore
}
