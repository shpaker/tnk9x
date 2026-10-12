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
			"кирпич",
			player,
			types.NewBlockEntity("brick", 0, 0, 8, nil),
			false,
			types.CollisionSlide,
		},
		{
			"вода",
			player,
			types.NewBlockEntity("water", 0, 0, 8, nil),
			false,
			types.CollisionSlide,
		},
		{
			"вода с лодкой",
			sailor,
			types.NewBlockEntity("water", 0, 0, 8, nil),
			false,
			types.CollisionIgnore,
		},
		{
			"лес",
			player,
			types.NewBlockEntity("forest", 0, 0, 8, nil),
			false,
			types.CollisionIgnore,
		},
		{
			"лёд",
			player,
			types.NewBlockEntity("ice", 0, 0, 8, nil),
			false,
			types.CollisionIgnore,
		},
		{
			"край карты",
			player,
			&types.BorderEntity{},
			false,
			types.CollisionSlide,
		},
		{"целый штаб", player, &types.HQEntity{}, false, types.CollisionSlide},
		{
			"разрушенный штаб",
			player,
			&types.HQEntity{State: types.HQStateDestroyed},
			false,
			types.CollisionIgnore,
		},
		{"танк", player, enemy, false, types.CollisionSlide},
		{"танк на спавне", player, spawning, false, types.CollisionSlide},
		{"взрывающийся танк", player, exploding, false, types.CollisionIgnore},
		{"бонус игроку", player, bonus, false, types.CollisionCross},
		{"бонус врагу", enemy, bonus, false, types.CollisionIgnore},
		{"бонус врагу с подбором", enemy, bonus, true, types.CollisionCross},
		{"пуля", player, newTestBullet(enemy), false, types.CollisionCross},
	}

	for _, tt := range tests {
		service := NewTankCollisionService(tt.pickup)
		if got := service.Response(tt.tank, tt.other); got != tt.want {
			t.Errorf("%s: %v, ожидалось %v", tt.name, got, tt.want)
		}
	}
}
