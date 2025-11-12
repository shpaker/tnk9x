package game

import (
	"github.com/shpaker/gonflict/internal/types"
)

type TanksRepository struct {
	players []*types.TankEntity // Игроки (массив из двух элементов)
	enemies []*types.TankEntity // Враги
}

func NewTanksRepository() *TanksRepository {
	return &TanksRepository{
		players: make([]*types.TankEntity, 2),
		enemies: make([]*types.TankEntity, 0),
	}
}

// === Методы для работы с игроками по номеру ===

// SetPlayer устанавливает танк игрока по номеру
func (tr *TanksRepository) SetPlayer(
	num types.PlayerTankNum,
	player *types.TankEntity,
) {
	if int(num) >= 0 && int(num) < len(tr.players) {
		tr.players[num] = player
	}
}

// GetPlayer возвращает танк игрока по номеру
func (tr *TanksRepository) GetPlayer(
	num types.PlayerTankNum,
) *types.TankEntity {
	if int(num) >= 0 && int(num) < len(tr.players) {
		return tr.players[num]
	}
	return nil
}

// HasPlayer возвращает true, если есть игрок с указанным номером
func (tr *TanksRepository) HasPlayer(num types.PlayerTankNum) bool {
	return tr.GetPlayer(num) != nil
}

// GetAllPlayers возвращает всех игроков
func (tr *TanksRepository) GetAllPlayers() []*types.TankEntity {
	return tr.players
}

// === Методы для работы с врагами ===

// AddEnemy добавляет врага в репозиторий
func (tr *TanksRepository) AddEnemy(enemy *types.TankEntity) {
	tr.enemies = append(tr.enemies, enemy)
}

// GetAllEnemies возвращает всех врагов
func (tr *TanksRepository) GetAllEnemies() []*types.TankEntity {
	return tr.enemies
}

// === Методы для работы со всеми танками ===

// GetAllTanks возвращает все танки (игроки + враги)
func (tr *TanksRepository) GetAllTanks() []*types.TankEntity {
	var allTanks []*types.TankEntity

	// Добавляем всех игроков через итерацию
	for _, player := range tr.players {
		if player != nil {
			allTanks = append(allTanks, player)
		}
	}

	// Добавляем всех врагов
	allTanks = append(allTanks, tr.enemies...)

	return allTanks
}

// AddTank добавляет танк (по умолчанию как врага)
func (tr *TanksRepository) AddTank(tank *types.TankEntity) {
	tr.enemies = append(tr.enemies, tank)
}
