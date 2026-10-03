package states_test

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/tnk9x/internal/states"
	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
)

// testShopProducts: без рекламы, две пачки и два пакета жетонов —
// строки NO ADS, PACK, TOKENS и BACK
var testShopProducts = []types.ProductSpec{
	{ID: "no_ads", Kind: types.ProductKindNoAds},
	{ID: "pack_2", Kind: types.ProductKindPack, Pack: 2},
	{ID: "pack_3", Kind: types.ProductKindPack, Pack: 3},
	{ID: "tokens_5", Kind: types.ProductKindTokens, Amount: 5},
	{ID: "tokens_15", Kind: types.ProductKindTokens, Amount: 15},
}

func testShopCampaign() *types.CampaignEntity {
	return types.NewCampaignEntity("TEST", []types.PackSpec{
		{Name: "ONE", Levels: []int{1}},
		{Name: "TWO", Levels: []int{2, 3}},
		{Name: "THREE", Levels: []int{4, 5}},
	})
}

// nopShopRenderer — рендер магазина без отрисовки; строка под
// точкой — Y
type nopShopRenderer struct{}

func (nopShopRenderer) Draw(*ebiten.Image, types.ShopViewData) {}

func (nopShopRenderer) HitRow(position types.Position) (int, bool) {
	return hitRowByY(position)
}

// newShopEnv — магазин, где площадка продаёт все тестовые товары
func newShopEnv() *menusEnv {
	env := newMenusEnv()
	env.purchase.Available = true
	for _, product := range testShopProducts {
		env.purchase.Catalog = append(env.purchase.Catalog, types.ProductOffer{
			ID: product.ID, Price: "10 YAN",
		})
	}
	env.shopOverlay.Open()
	return env
}

// Влево-вправо выбирает пачку, выбор открывает окно оплаты
// выбранной; пока оно открыто, ввод не принимается
func TestShopOverlay_BuyVariant(t *testing.T) {
	env := newShopEnv()

	env.frame(env.shopOverlay.Update, testutil.FakeMenuInput{Down: true})
	env.frame(env.shopOverlay.Update, testutil.FakeMenuInput{Side: 1})
	env.frame(env.shopOverlay.Update, testutil.FakeMenuInput{Confirm: true})
	if len(env.purchase.Requests) != 1 || env.purchase.Requests[0] != "pack_3" {
		t.Fatalf("requests %v, want [pack_3]", env.purchase.Requests)
	}

	env.purchase.Status = types.PurchaseStatusPending
	env.frame(env.shopOverlay.Update, testutil.FakeMenuInput{BackPressed: true})
	if !env.shopOverlay.IsOpen() {
		t.Fatal("shop must wait for the payment")
	}

	// Оплачено: пачка зачислена сразу
	env.purchase.Status = types.PurchaseStatusGranted
	env.purchase.Purchases = []types.Purchase{
		{ProductID: "pack_3", Token: "t1"},
	}
	env.frame(env.shopOverlay.Update, testutil.FakeMenuInput{})
	if !env.inventory.IsPackOwned(3) {
		t.Error("paid pack must be owned")
	}

	// Купленное повторно не продаётся
	env.frame(env.shopOverlay.Update, testutil.FakeMenuInput{Confirm: true})
	if len(env.purchase.Requests) != 1 {
		t.Errorf("owned pack requested again: %v", env.purchase.Requests)
	}
}

// Жетоны зачисляются при оплате, покупка списывается у площадки
func TestShopOverlay_BuyTokens(t *testing.T) {
	env := newShopEnv()

	env.frame(env.shopOverlay.Update, testutil.FakeMenuInput{Down: true})
	env.frame(env.shopOverlay.Update, testutil.FakeMenuInput{Down: true})
	env.frame(env.shopOverlay.Update, testutil.FakeMenuInput{Confirm: true})
	env.purchase.Status = types.PurchaseStatusGranted
	env.purchase.Purchases = []types.Purchase{
		{ProductID: "tokens_5", Token: "t1"},
	}
	env.frame(env.shopOverlay.Update, testutil.FakeMenuInput{})

	if env.inventory.GetTokens() != 5 {
		t.Errorf("tokens %d, want 5", env.inventory.GetTokens())
	}
	if len(env.purchase.Consumed) != 1 {
		t.Errorf("consumed %v, want [t1]", env.purchase.Consumed)
	}
}

// Отмена оплаты: ничего не зачислено, магазин снова принимает ввод;
// BACK закрывает магазин
func TestShopOverlay_DeniedAndBack(t *testing.T) {
	env := newShopEnv()

	env.frame(env.shopOverlay.Update, testutil.FakeMenuInput{Confirm: true})
	env.purchase.Status = types.PurchaseStatusDenied
	env.frame(env.shopOverlay.Update, testutil.FakeMenuInput{})
	if env.inventory.HasNoAds() {
		t.Error("denied purchase must not be granted")
	}

	env.frame(env.shopOverlay.Update, clickAt(3, 0))
	if env.shopOverlay.IsOpen() {
		t.Error("BACK must close the shop")
	}
}

// SHOP в главном меню — между режимами и настройками; пока магазин
// открыт, ввод уходит в него
func TestMainMenuState_Shop(t *testing.T) {
	env := newShopEnv()
	env.shopOverlay.Update() // закрывается ниже через BACK
	menu := states.NewMainMenuState(states.MainMenuStateDependencies{
		SettingsUseCases: newSettingsUseCases(),
		Renderer:         nopMainMenuRenderer{},
		MenuInput:        env.input,
		Scene:            &countingScene{},
		SettingsOverlay:  env.settingsOverlay,
		ShopOverlay:      env.shopOverlay,
		Settings:         env.settings,
		ShopAvailable:    true,
	})
	env.frame(
		func() { menu.Update() },
		testutil.FakeMenuInput{BackPressed: true},
	)
	env.frame(
		func() { menu.Update() },
		testutil.FakeMenuInput{BackPressed: true},
	)
	if env.shopOverlay.IsOpen() {
		t.Fatal("shop must be closed")
	}

	env.frame(func() { menu.Update() }, clickAt(2, 0))
	if !env.shopOverlay.IsOpen() {
		t.Fatal("SHOP must open the shop")
	}
	env.frame(func() { menu.Update() }, testutil.FakeMenuInput{Confirm: true})
	if len(env.purchase.Requests) != 1 {
		t.Errorf("input must go to the shop: %v", env.purchase.Requests)
	}
}
