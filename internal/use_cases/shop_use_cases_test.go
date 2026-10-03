package use_cases_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

// shopProducts — товары для testCampaign: пачки 1 и 2
var shopProducts = []types.ProductSpec{
	{ID: "no_ads", Kind: types.ProductKindNoAds},
	{ID: "all_levels", Kind: types.ProductKindAllLevels},
	{ID: "pack_1", Kind: types.ProductKindPack, Pack: 1},
	{ID: "pack_2", Kind: types.ProductKindPack, Pack: 2},
	{ID: "tokens_5", Kind: types.ProductKindTokens, Amount: 5},
	{ID: "tokens_15", Kind: types.ProductKindTokens, Amount: 15},
}

type shopEnv struct {
	shop        *use_cases.ShopUseCases
	progression *use_cases.ProgressionUseCases
	inventory   *types.InventoryEntity
	purchase    *testutil.FakePurchase
	repository  *testutil.MemoryInventoryRepository
}

// newShopEnv — площадка продаёт все товары, прогресс пустой
func newShopEnv() *shopEnv {
	inventory := types.NewInventoryEntity()
	purchase := &testutil.FakePurchase{Available: true}
	for _, product := range shopProducts {
		purchase.Catalog = append(purchase.Catalog, types.ProductOffer{
			ID: product.ID, Price: "1 YAN",
		})
	}
	progression, _ := newProgressionWithInventory(inventory)
	repository := &testutil.MemoryInventoryRepository{}
	return &shopEnv{
		shop: use_cases.NewShopUseCases(
			shopProducts,
			testCampaign(),
			inventory,
			[]interfaces.IProgressionUseCases{progression},
			purchase,
			repository,
		),
		progression: progression,
		inventory:   inventory,
		purchase:    purchase,
		repository:  repository,
	}
}

func rowKinds(rows []types.ShopRow) []types.ProductKind {
	kinds := make([]types.ProductKind, len(rows))
	for i, row := range rows {
		kinds[i] = row.Kind
	}
	return kinds
}

// Товары одного вида — одна строка; пачка открывает свои уровни
func TestShop_GetRows(t *testing.T) {
	env := newShopEnv()

	rows := env.shop.GetRows()
	want := []types.ProductKind{
		types.ProductKindNoAds,
		types.ProductKindAllLevels,
		types.ProductKindPack,
		types.ProductKindTokens,
	}
	if !slices.Equal(rowKinds(rows), want) {
		t.Fatalf("rows %v, want %v", rowKinds(rows), want)
	}
	packs := rows[2].Offers
	if len(packs) != 2 || packs[1].Product.ID != "pack_2" ||
		!slices.Equal(packs[1].Levels, []int{4, 5}) ||
		packs[1].PackName != "TWO" {
		t.Errorf("packs %+v", packs)
	}
	if len(rows[1].Offers[0].Levels) != 5 {
		t.Errorf("all levels %v", rows[1].Offers[0].Levels)
	}
	if len(rows[3].Offers) != 2 {
		t.Errorf("token bundles %+v", rows[3].Offers)
	}
}

// Только товары каталога площадки; купленная пачка — «куплено»,
// после покупки всех уровней пачки не продаются
func TestShop_GetRows_CatalogAndOwned(t *testing.T) {
	env := newShopEnv()
	env.purchase.Catalog = env.purchase.Catalog[1:]
	env.inventory.SetPackOwned(2)

	rows := env.shop.GetRows()
	if rows[0].Kind != types.ProductKindAllLevels {
		t.Fatalf("NO ADS is not in the catalog: %v", rowKinds(rows))
	}
	if !rows[1].Offers[1].Owned {
		t.Error("purchased pack must be owned")
	}

	env.inventory.SetAllLevels(true)
	if slices.Contains(rowKinds(env.shop.GetRows()), types.ProductKindPack) {
		t.Error("packs are not sold after all levels")
	}
}

// Пачка, открытая прохождением целиком, не продаётся
func TestShop_GetRows_PackOpenedByProgress(t *testing.T) {
	env := newShopEnv()
	_ = env.progression.RecordResult(1, 1)
	_ = env.progression.RecordResult(2, 1)

	packs := env.shop.GetRows()[2].Offers
	if len(packs) != 1 || packs[0].Product.ID != "pack_2" {
		t.Errorf("packs %+v, want only pack_2", packs)
	}
}

// Без покупок у площадки магазина нет
func TestShop_IsAvailable(t *testing.T) {
	env := newShopEnv()
	if !env.shop.IsAvailable() {
		t.Error("shop must be available")
	}
	env.purchase.Available = false
	if env.shop.IsAvailable() {
		t.Error("no purchases — no shop")
	}
}

// Разовые покупки ставят флаги, жетоны зачисляются один раз
// и списываются после сохранения
func TestShop_Sync(t *testing.T) {
	env := newShopEnv()
	env.purchase.Purchases = []types.Purchase{
		{ProductID: "no_ads", Token: "a"},
		{ProductID: "pack_2", Token: "b"},
		{ProductID: "tokens_15", Token: "c"},
		{ProductID: "unknown", Token: "d"},
	}

	if err := env.shop.Sync(); err != nil {
		t.Fatalf("sync: %v", err)
	}
	if !env.inventory.HasNoAds() || !env.inventory.IsPackOwned(2) ||
		env.inventory.GetTokens() != 15 {
		t.Errorf("inventory after sync: %+v", env.inventory)
	}
	if env.repository.Saves != 1 ||
		!slices.Equal(env.purchase.Consumed, []string{"c"}) {
		t.Errorf(
			"saves %d, consumed %v",
			env.repository.Saves, env.purchase.Consumed,
		)
	}

	// Площадка снова отдала ту же покупку: повторно не зачисляется
	env.purchase.Purchases = append(
		env.purchase.Purchases,
		types.Purchase{ProductID: "tokens_15", Token: "c"},
	)
	if err := env.shop.Sync(); err != nil {
		t.Fatalf("second sync: %v", err)
	}
	if env.inventory.GetTokens() != 15 || env.repository.Saves != 1 {
		t.Errorf(
			"tokens %d, saves %d",
			env.inventory.GetTokens(), env.repository.Saves,
		)
	}
}

// Сохранение не удалось: покупка у площадки не списывается
func TestShop_Sync_SaveFailed(t *testing.T) {
	env := newShopEnv()
	env.repository.Err = errors.New("disk full")
	env.purchase.Purchases = []types.Purchase{
		{ProductID: "tokens_5", Token: "c"},
	}

	if err := env.shop.Sync(); err == nil {
		t.Fatal("save error must be returned")
	}
	if len(env.purchase.Consumed) != 0 {
		t.Errorf("consumed %v without save", env.purchase.Consumed)
	}
}

// Оплата: исход отдаётся, оплаченное зачисляется сразу
func TestShop_PurchaseFlow(t *testing.T) {
	env := newShopEnv()
	env.shop.RequestPurchase("tokens_5")
	if !slices.Equal(env.purchase.Requests, []string{"tokens_5"}) {
		t.Fatalf("requests %v", env.purchase.Requests)
	}

	env.purchase.Status = types.PurchaseStatusPending
	if status, _ := env.shop.PollPurchase(); status != types.PurchaseStatusPending {
		t.Errorf("status %v, want pending", status)
	}

	env.purchase.Status = types.PurchaseStatusGranted
	env.purchase.Purchases = []types.Purchase{
		{ProductID: "tokens_5", Token: "c"},
	}
	status, err := env.shop.PollPurchase()
	if status != types.PurchaseStatusGranted || err != nil ||
		env.inventory.GetTokens() != 5 {
		t.Errorf(
			"granted: %v, %v, tokens %d",
			status, err, env.inventory.GetTokens(),
		)
	}
}
