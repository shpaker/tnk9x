package use_cases

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.IInventoryUseCases = (*InventoryUseCases)(nil)

// InventoryUseCases — купленное игроком: отключённая реклама
// и жетоны; инвентарь единственный на приложение
type InventoryUseCases struct {
	// Entities
	inventory *types.InventoryEntity
	// Repositories
	inventoryRepository interfaces.IInventoryRepository
}

func NewInventoryUseCases(
	inventory *types.InventoryEntity,
	inventoryRepository interfaces.IInventoryRepository,
) *InventoryUseCases {
	return &InventoryUseCases{
		inventory:           inventory,
		inventoryRepository: inventoryRepository,
	}
}

// HasNoAds реализует IInventoryUseCases
func (uc *InventoryUseCases) HasNoAds() bool {
	return uc.inventory.HasNoAds()
}

// GetTokens реализует IInventoryUseCases
func (uc *InventoryUseCases) GetTokens() uint {
	return uc.inventory.GetTokens()
}

// SpendToken реализует IInventoryUseCases
func (uc *InventoryUseCases) SpendToken() (bool, error) {
	if !uc.inventory.SpendToken() {
		return false, nil
	}
	return true, uc.inventoryRepository.SaveInventory(uc.inventory)
}
