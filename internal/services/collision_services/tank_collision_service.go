package collision_services

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.ITankCollisionService = (*TankCollisionService)(nil)

type TankCollisionService struct {
	// Whether enemies pick up bonuses (game.enemy_bonus_pickup).
	enemyBonusPickup bool
}

func NewTankCollisionService(enemyBonusPickup bool) *TankCollisionService {
	return &TankCollisionService{enemyBonusPickup: enemyBonusPickup}
}

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
		if other.Data.Name == types.Water && tank.CanSail() {
			return types.CollisionIgnore
		}
		return types.CollisionSlide
	case *types.HQEntity:
		// An intact eagle blocks; a destroyed one is passable.
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
