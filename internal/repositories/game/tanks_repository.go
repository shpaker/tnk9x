package game

import (
	"fmt"

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

// ClearPlayer удаляет танк игрока
func (tr *TanksRepository) ClearPlayer() {
	tr.player = nil
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

// RemoveEnemy удаляет врага по индексу
func (tr *TanksRepository) RemoveEnemy(index int) error {
	if index < 0 || index >= len(tr.enemies) {
		return fmt.Errorf(
			"enemy index %d out of range [0, %d)",
			index,
			len(tr.enemies),
		)
	}
	tr.enemies = append(tr.enemies[:index], tr.enemies[index+1:]...)
	return nil
}

// === Методы для обратной совместимости ===

// AddTank добавляет танк (по умолчанию как врага)
func (tr *TanksRepository) AddTank(tank *types.TankEntity) {
	tr.enemies = append(tr.enemies, tank)
}

// GetAllTanks возвращает все танки (игрок + враги)
func (tr *TanksRepository) GetAllTanks() []*types.TankEntity {
	allTanks := make([]*types.TankEntity, 0)

	// Добавляем игрока, если он есть
	if tr.player != nil {
		allTanks = append(allTanks, tr.player)
	}

	// Добавляем всех врагов
	allTanks = append(allTanks, tr.enemies...)

	return allTanks
}

// RemoveTank удаляет танк по указателю
func (tr *TanksRepository) RemoveTank(tank *types.TankEntity) error {
	if tank == nil {
		return fmt.Errorf("tank pointer cannot be nil")
	}

	// Проверяем, это игрок?
	if tr.player == tank {
		tr.player = nil
		return nil
	}

	// Ищем врага в массиве
	for i, enemy := range tr.enemies {
		if enemy == tank {
			tr.enemies = append(tr.enemies[:i], tr.enemies[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("tank not found in repository")
}
