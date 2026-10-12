package collision_services

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

func newTestTank(role types.TankRole, state types.TankState) *types.TankEntity {
	tank := types.NewDefaultTankEntity(role, types.DirectionUp)
	tank.State = state
	tank.Altitude = types.SURFACE
	return &tank
}

func newTestBullet(owner *types.TankEntity) *types.BulletEntity {
	return types.NewBulletEntity(
		types.Position{},
		types.Size{Width: 4, Height: 4},
		types.SURFACE,
		nil,
		types.DirectionUp,
		nil,
		owner,
	)
}

func TestBulletCollisionService_Response(t *testing.T) {
	player := newTestTank(types.TankRolePlayer1, types.TankStateActive)
	player2 := newTestTank(types.TankRolePlayer2, types.TankStateActive)
	enemy := newTestTank(types.TankRoleEnemy, types.TankStateActive)
	enemy2 := newTestTank(types.TankRoleEnemy, types.TankStateActive)
	spawning := newTestTank(types.TankRoleEnemy, types.TankStateSpawning)
	playerBullet := newTestBullet(player)
	enemyBullet := newTestBullet(enemy)

	tests := []struct {
		name   string
		bullet *types.BulletEntity
		other  types.IEntityCollider
		want   types.CollisionResponse
	}{
		{
			"кирпич",
			playerBullet,
			types.NewBlockEntity("brick", 0, 0, 8, nil),
			types.CollisionCross,
		},
		{
			"сталь",
			playerBullet,
			types.NewBlockEntity("steel", 0, 0, 8, nil),
			types.CollisionCross,
		},
		{
			"вода",
			playerBullet,
			types.NewBlockEntity("water", 0, 0, 8, nil),
			types.CollisionIgnore,
		},
		{
			"лес",
			playerBullet,
			types.NewBlockEntity("forest", 0, 0, 8, nil),
			types.CollisionIgnore,
		},
		{
			"лёд",
			playerBullet,
			types.NewBlockEntity("ice", 0, 0, 8, nil),
			types.CollisionIgnore,
		},
		{
			"край карты",
			playerBullet,
			&types.BorderEntity{},
			types.CollisionCross,
		},
		{"штаб", playerBullet, &types.HQEntity{}, types.CollisionCross},
		{
			"штаб, пуля без владельца",
			newTestBullet(nil),
			&types.HQEntity{},
			types.CollisionIgnore,
		},
		{"враг", playerBullet, enemy, types.CollisionCross},
		{"второй игрок", playerBullet, player2, types.CollisionCross},
		{"свой танк", playerBullet, player, types.CollisionIgnore},
		{"враг по врагу", enemyBullet, enemy2, types.CollisionIgnore},
		{"враг по игроку", enemyBullet, player, types.CollisionCross},
		{"танк на спавне", playerBullet, spawning, types.CollisionIgnore},
		{"пуля", playerBullet, enemyBullet, types.CollisionCross},
	}

	service := NewBulletCollisionService()
	for _, tt := range tests {
		if got := service.Response(tt.bullet, tt.other); got != tt.want {
			t.Errorf("%s: %v, ожидалось %v", tt.name, got, tt.want)
		}
	}
}
