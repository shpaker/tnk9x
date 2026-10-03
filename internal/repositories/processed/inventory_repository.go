package processed

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// InventoryKey — ключ сохранения купленного: один на все режимы
const InventoryKey = "inventory"

// inventoryFormatVersion — версия формата сохранения
const inventoryFormatVersion = 1

var _ interfaces.IInventoryRepository = (*InventoryRepository)(nil)

// InventoryRepository сериализует купленное игроком в JSON поверх
// пользовательского хранилища
type InventoryRepository struct {
	storage interfaces.IStorageRepository
}

func NewInventoryRepository(
	storage interfaces.IStorageRepository,
) *InventoryRepository {
	return &InventoryRepository{storage: storage}
}

type inventorySchema struct {
	Version   int      `json:"version"`
	NoAds     bool     `json:"no_ads"`
	AllLevels bool     `json:"all_levels"`
	Packs     []int    `json:"packs"`
	Tokens    uint     `json:"tokens"`
	Credited  []string `json:"credited"`
}

// GetInventory читает купленное; отсутствие сохранения — пусто
func (r *InventoryRepository) GetInventory() (*types.InventoryEntity, error) {
	data, err := r.storage.Load(InventoryKey)
	if err != nil {
		return types.NewInventoryEntity(), err
	}
	if len(data) == 0 {
		return types.NewInventoryEntity(), nil
	}

	var schema inventorySchema
	if err := json.Unmarshal(data, &schema); err != nil {
		return types.NewInventoryEntity(),
			fmt.Errorf("malformed inventory: %w", err)
	}
	if schema.Version != inventoryFormatVersion {
		return types.NewInventoryEntity(), fmt.Errorf(
			"unsupported inventory version %d", schema.Version,
		)
	}

	inventory := types.NewInventoryEntity()
	inventory.SetNoAds(schema.NoAds)
	inventory.SetAllLevels(schema.AllLevels)
	for _, pack := range schema.Packs {
		inventory.SetPackOwned(pack)
	}
	inventory.AddTokens(schema.Tokens)
	for _, token := range schema.Credited {
		inventory.MarkCredited(token)
	}
	return inventory, nil
}

func (r *InventoryRepository) SaveInventory(
	inventory *types.InventoryEntity,
) error {
	packs := inventory.GetPacks()
	sort.Ints(packs)
	credited := inventory.GetCredited()
	sort.Strings(credited)

	data, err := json.Marshal(inventorySchema{
		Version:   inventoryFormatVersion,
		NoAds:     inventory.HasNoAds(),
		AllLevels: inventory.HasAllLevels(),
		Packs:     packs,
		Tokens:    inventory.GetTokens(),
		Credited:  credited,
	})
	if err != nil {
		return err
	}
	return r.storage.Save(InventoryKey, data)
}
