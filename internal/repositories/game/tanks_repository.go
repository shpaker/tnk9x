package game

import (
	"fmt"

	"github.com/shpaker/gonflict/internal/types"
)

type TanksRepository struct {
	tanks []*types.TankEntity
}

func NewTanksRepository() *TanksRepository {
	return &TanksRepository{
		tanks: make([]*types.TankEntity, 0),
	}
}

// AddTank добавляет танк в репозиторий
func (tr *TanksRepository) AddTank(tank *types.TankEntity) {
	tr.tanks = append(tr.tanks, tank)
}

// GetAllTanks возвращает все танки
func (tr *TanksRepository) GetAllTanks() []*types.TankEntity {
	return tr.tanks
}

// RemoveTank удаляет танк по индексу
func (tr *TanksRepository) RemoveTank(index int) error {
	if index < 0 || index >= len(tr.tanks) {
		return fmt.Errorf("индекс танка %d вне диапазона [0, %d)", index, len(tr.tanks))
	}
	tr.tanks = append(tr.tanks[:index], tr.tanks[index+1:]...)
	return nil
}

// GetTank возвращает танк по индексу
func (tr *TanksRepository) GetTank(index int) (*types.TankEntity, error) {
	if index < 0 || index >= len(tr.tanks) {
		return nil, fmt.Errorf("индекс танка %d вне диапазона [0, %d)", index, len(tr.tanks))
	}
	return tr.tanks[index], nil
}
