package use_cases_test

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

// Жетон списывается и сохраняется; без жетонов — false
func TestInventory_SpendToken(t *testing.T) {
	inventory := types.NewInventoryEntity()
	inventory.AddTokens(1)
	repository := &testutil.MemoryInventoryRepository{}
	inventoryUseCases := use_cases.NewInventoryUseCases(inventory, repository)

	spent, err := inventoryUseCases.SpendToken()
	if !spent || err != nil || inventoryUseCases.GetTokens() != 0 {
		t.Fatalf("spend: %v, %v, tokens %d", spent, err, inventory.GetTokens())
	}
	if repository.Saves != 1 {
		t.Errorf("saves %d, want 1", repository.Saves)
	}

	spent, _ = inventoryUseCases.SpendToken()
	if spent || repository.Saves != 1 {
		t.Errorf("no tokens: spent %v, saves %d", spent, repository.Saves)
	}
}
