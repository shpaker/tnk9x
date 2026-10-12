package collision_services

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

func TestTankCollisionService_Response(t *testing.T) {
	player := newTestTank(types.TankRolePlayer1, types.TankStateActive)
	sailor := newTestTank(types.TankRolePlayer2, types.TankStateActive)
	sailor.SetBoat()
	enemy := newTestTank(types.TankRoleEnemy, types.TankStateActive)
	exploding := newTestTank(types.TankRoleEnemy, types.TankStateExploding)
	spawning := newTestTank(types.TankRoleEnemy, types.TankStateSpawning)
	bonus := types.NewBonusEntity(
		types.BonusTypeStar,
		types.Position{},
		types.Size{},
		nil,
	)

	tests := []struct {
		name   string
		tank   *types.TankEntity
		other  types.IEntityCollider
		pickup bool
		want   types.CollisionResponse
	}{
		{
			"brick",
			player,
			types.NewBlockEntity("brick", 0, 0, 8, nil),
			false,
			types.CollisionSlide,
		},
		{
			"water",
			player,
			types.NewBlockEntity("water", 0, 0, 8, nil),
			false,
			types.CollisionSlide,
		},
		{
			"water with boat",
			sailor,
			types.NewBlockEntity("water", 0, 0, 8, nil),
			false,
			types.CollisionIgnore,
		},
		{
			"forest",
			player,
			types.NewBlockEntity("forest", 0, 0, 8, nil),
			false,
			types.CollisionIgnore,
		},
		{
			"ice",
			player,
			types.NewBlockEntity("ice", 0, 0, 8, nil),
			false,
			types.CollisionIgnore,
		},
		{
			"map edge",
			player,
			&types.BorderEntity{},
			false,
			types.CollisionSlide,
		},
		{"intact HQ", player, &types.HQEntity{}, false, types.CollisionSlide},
		{
			"destroyed HQ",
			player,
			&types.HQEntity{State: types.HQStateDestroyed},
			false,
			types.CollisionIgnore,
		},
		{"tank", player, enemy, false, types.CollisionSlide},
		{"spawning tank", player, spawning, false, types.CollisionSlide},
		{"exploding tank", player, exploding, false, types.CollisionIgnore},
		{"bonus for player", player, bonus, false, types.CollisionCross},
		{"bonus for enemy", enemy, bonus, false, types.CollisionIgnore},
		{
			"bonus for enemy with pickup enabled",
			enemy,
			bonus,
			true,
			types.CollisionCross,
		},
		{"bullet", player, newTestBullet(enemy), false, types.CollisionCross},
	}

	for _, tt := range tests {
		service := NewTankCollisionService(tt.pickup)
		if got := service.Response(tt.tank, tt.other); got != tt.want {
			t.Errorf("%s: %v, want %v", tt.name, got, tt.want)
		}
	}
}
