package processed

import (
	"testing"
)

// Купленное переживает сохранение и чтение
func TestInventoryRepository_RoundTrip(t *testing.T) {
	storage := &memoryStorage{data: map[string][]byte{}}
	repository := NewInventoryRepository(storage)

	inventory, err := repository.GetInventory()
	if err != nil || inventory.GetTokens() != 0 || inventory.HasNoAds() {
		t.Fatalf("empty storage: %+v, %v", inventory, err)
	}

	inventory.SetNoAds(true)
	inventory.SetPackOwned(3)
	inventory.AddTokens(15)
	inventory.MarkCredited("t1")
	if err := repository.SaveInventory(inventory); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := repository.GetInventory()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if !loaded.HasNoAds() || loaded.HasAllLevels() ||
		!loaded.IsPackOwned(3) || loaded.GetTokens() != 15 ||
		!loaded.IsCredited("t1") {
		t.Errorf("loaded %+v", loaded)
	}
}

// Повреждённое сохранение — пустой инвентарь и ошибка
func TestInventoryRepository_Malformed(t *testing.T) {
	for _, data := range []string{"{", `{"version": 99}`} {
		repository := NewInventoryRepository(&memoryStorage{
			data: map[string][]byte{InventoryKey: []byte(data)},
		})
		inventory, err := repository.GetInventory()
		if err == nil || inventory == nil || inventory.GetTokens() != 0 {
			t.Errorf("%q: %+v, %v", data, inventory, err)
		}
	}
}
