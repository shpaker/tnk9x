package game

import (
	"github.com/shpaker/tnk25/internal/types"
)

type TanksRepository struct {
	players []*types.TankEntity
	enemies []*types.TankEntity
}

func NewTanksRepository() *TanksRepository {
	return &TanksRepository{
		players: make([]*types.TankEntity, 2),
		enemies: make([]*types.TankEntity, 0),
	}
}

func (tr *TanksRepository) SetPlayer(
	num types.PlayerTankNum,
	player *types.TankEntity,
) {
	if int(num) >= 0 && int(num) < len(tr.players) {
		tr.players[num] = player
	}
}

func (tr *TanksRepository) GetPlayer(
	num types.PlayerTankNum,
) *types.TankEntity {
	if int(num) >= 0 && int(num) < len(tr.players) {
		return tr.players[num]
	}
	return nil
}

func (tr *TanksRepository) HasPlayer(num types.PlayerTankNum) bool {
	return tr.GetPlayer(num) != nil
}

func (tr *TanksRepository) GetAllPlayers() []*types.TankEntity {
	return tr.players
}

func (tr *TanksRepository) AddEnemy(enemy *types.TankEntity) {
	tr.enemies = append(tr.enemies, enemy)
}

func (tr *TanksRepository) GetAllEnemies() []*types.TankEntity {
	return tr.enemies
}

func (tr *TanksRepository) GetAllTanks() []*types.TankEntity {
	var allTanks []*types.TankEntity

	for _, player := range tr.players {
		if player != nil {
			allTanks = append(allTanks, player)
		}
	}

	allTanks = append(allTanks, tr.enemies...)

	return allTanks
}

func (tr *TanksRepository) AddTank(tank *types.TankEntity) {
	tr.enemies = append(tr.enemies, tank)
}
