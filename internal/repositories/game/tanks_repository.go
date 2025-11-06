package game

import (
	"github.com/shpaker/gonflict/internal/types"
)

type TanksRepository struct {
	player  *types.TankEntity   // Игрок
	enemies []*types.TankEntity // Враги
}

func NewTanksRepository() *TanksRepository {
	return &TanksRepository{
		player:  nil,
		enemies: make([]*types.TankEntity, 0),
	}
}

// === Методы для работы с игроком ===

// SetPlayer устанавливает танк игрока
func (tr *TanksRepository) SetPlayer(player *types.TankEntity) {
	tr.player = player
}

// GetPlayer возвращает танк игрока
func (tr *TanksRepository) GetPlayer() *types.TankEntity {
	return tr.player
}

// HasPlayer возвращает true, если есть игрок
func (tr *TanksRepository) HasPlayer() bool {
	return tr.player != nil
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

// GetAllTanks возвращает все танки (игрок + враги)
func (tr *TanksRepository) GetAllTanks() []*types.TankEntity {
	var allTanks []*types.TankEntity

	// Добавляем игрока, если он есть
	if tr.player != nil {
		allTanks = append(allTanks, tr.player)
	}

	// Добавляем всех врагов
	allTanks = append(allTanks, tr.enemies...)

	return allTanks
}

// === Методы для обратной совместимости ===

// AddTank добавляет танк (по умолчанию как врага)
func (tr *TanksRepository) AddTank(tank *types.TankEntity) {
	tr.enemies = append(tr.enemies, tank)
}
