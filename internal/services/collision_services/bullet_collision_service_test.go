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
			"brick",
			playerBullet,
			types.NewBlockEntity("brick", 0, 0, 8, nil),
			types.CollisionCross,
		},
		{
			"steel",
			playerBullet,
			types.NewBlockEntity("steel", 0, 0, 8, nil),
			types.CollisionCross,
		},
		{
			"water",
			playerBullet,
			types.NewBlockEntity("water", 0, 0, 8, nil),
			types.CollisionIgnore,
		},
		{
			"forest",
			playerBullet,
			types.NewBlockEntity("forest", 0, 0, 8, nil),
			types.CollisionIgnore,
		},
		{
			"ice",
			playerBullet,
			types.NewBlockEntity("ice", 0, 0, 8, nil),
			types.CollisionIgnore,
		},
		{
			"map edge",
			playerBullet,
			&types.BorderEntity{},
			types.CollisionCross,
		},
		{"HQ", playerBullet, &types.HQEntity{}, types.CollisionCross},
		{
			"HQ, bullet without owner",
			newTestBullet(nil),
			&types.HQEntity{},
			types.CollisionIgnore,
		},
		{"enemy", playerBullet, enemy, types.CollisionCross},
		{"second player", playerBullet, player2, types.CollisionCross},
		{"own tank", playerBullet, player, types.CollisionIgnore},
		{"enemy hits enemy", enemyBullet, enemy2, types.CollisionIgnore},
		{"enemy hits player", enemyBullet, player, types.CollisionCross},
		{"spawning tank", playerBullet, spawning, types.CollisionIgnore},
		{"bullet", playerBullet, enemyBullet, types.CollisionCross},
	}

	service := NewBulletCollisionService()
	for _, tt := range tests {
		if got := service.Response(tt.bullet, tt.other); got != tt.want {
			t.Errorf("%s: %v, want %v", tt.name, got, tt.want)
		}
	}
}
