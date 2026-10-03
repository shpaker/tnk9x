//go:build js

package platform

import (
	"syscall/js"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var (
	_ interfaces.IPlatformAdapter = (*PlatformAdapter)(nil)
	_ interfaces.IRewardAdapter   = (*PlatformAdapter)(nil)
	_ interfaces.IPurchaseAdapter = (*PlatformAdapter)(nil)
)

// bridgeName — глобальный объект моста площадки; контракт описан
// в web/README.md, мост ставит web/common/bridge.js до запуска wasm
const bridgeName = "tnk9xPlatform"

// Исходы рекламы за вознаграждение в контракте моста
const (
	bridgeRewardPending = "pending"
	bridgeRewardGranted = "granted"
	bridgeRewardDenied  = "denied"
)

// Исходы покупки в контракте моста
const (
	bridgePurchasePending = "pending"
	bridgePurchaseGranted = "granted"
	bridgePurchaseDenied  = "denied"
)

// PlatformAdapter — площадка браузера за нейтральным мостом
// window.tnk9xPlatform: обычная страница или игровой портал. Что
// именно умеет площадка, решает её мост; звук при приостановке
// глушит тоже мост — на скрытой вкладке кадры не идут
type PlatformAdapter struct {
	bridge js.Value

	// Состояние кадра
	suspended     bool
	justSuspended bool
	suspensions   int

	// Уже переданное площадке
	ready          bool
	gameplayActive bool
}

// NewPlatformAdapter подключается к мосту; без моста страница
// собрана неверно — запуск останавливается сразу
func NewPlatformAdapter() *PlatformAdapter {
	bridge := js.Global().Get(bridgeName)
	if bridge.IsUndefined() || bridge.IsNull() {
		panic("platform bridge window." + bridgeName + " is missing")
	}
	return &PlatformAdapter{
		bridge:      bridge,
		suspensions: bridge.Get("suspensions").Int(),
	}
}

// Update реализует IPlatformAdapter: счётчик приостановок моста
// ловит и те, что начались и закончились, пока кадры не шли
func (a *PlatformAdapter) Update() {
	suspensions := a.bridge.Get("suspensions").Int()
	a.justSuspended = suspensions != a.suspensions
	a.suspensions = suspensions
	a.suspended = a.bridge.Get("suspended").Bool()
}

// IsSuspended реализует IPlatformAdapter
func (a *PlatformAdapter) IsSuspended() bool { return a.suspended }

// IsJustSuspended реализует IPlatformAdapter
func (a *PlatformAdapter) IsJustSuspended() bool { return a.justSuspended }

// Ready реализует IPlatformAdapter
func (a *PlatformAdapter) Ready() {
	if a.ready {
		return
	}
	a.ready = true
	a.bridge.Call("ready")
}

// SetGameplayActive реализует IPlatformAdapter
func (a *PlatformAdapter) SetGameplayActive(active bool) {
	if active == a.gameplayActive {
		return
	}
	a.gameplayActive = active
	a.bridge.Call("gameplay", active)
}

// RequestIntermission реализует IPlatformAdapter
func (a *PlatformAdapter) RequestIntermission() {
	a.bridge.Call("intermission")
}

// RequestRating реализует IPlatformAdapter
func (a *PlatformAdapter) RequestRating() {
	a.bridge.Call("rating")
}

// GetLanguage реализует IPlatformAdapter: язык, который мост
// узнал у площадки (SDK портала или браузер)
func (a *PlatformAdapter) GetLanguage() string {
	language := a.bridge.Get("language")
	if language.Type() != js.TypeString {
		return ""
	}
	return language.String()
}

// IsRewardAvailable реализует IRewardAdapter
func (a *PlatformAdapter) IsRewardAvailable() bool {
	return a.bridge.Get("rewardAvailable").Bool()
}

// RequestReward реализует IRewardAdapter
func (a *PlatformAdapter) RequestReward() {
	a.bridge.Call("requestReward")
}

// PollReward реализует IRewardAdapter
func (a *PlatformAdapter) PollReward() types.RewardStatus {
	switch a.bridge.Call("rewardStatus").String() {
	case bridgeRewardPending:
		return types.RewardStatusPending
	case bridgeRewardGranted:
		return types.RewardStatusGranted
	case bridgeRewardDenied:
		return types.RewardStatusDenied
	default:
		return types.RewardStatusNone
	}
}

// IsPurchaseAvailable реализует IPurchaseAdapter
func (a *PlatformAdapter) IsPurchaseAvailable() bool {
	return a.bridge.Get("purchasesAvailable").Bool()
}

// GetCatalog реализует IPurchaseAdapter: массив { id, price } моста
func (a *PlatformAdapter) GetCatalog() []types.ProductOffer {
	items := a.bridgeArray("catalog")
	offers := make([]types.ProductOffer, 0, len(items))
	for _, item := range items {
		offers = append(offers, types.ProductOffer{
			ID:    stringField(item, "id"),
			Price: stringField(item, "price"),
		})
	}
	return offers
}

// RequestPurchase реализует IPurchaseAdapter
func (a *PlatformAdapter) RequestPurchase(productID string) {
	a.bridge.Call("requestPurchase", productID)
}

// PollPurchase реализует IPurchaseAdapter
func (a *PlatformAdapter) PollPurchase() types.PurchaseStatus {
	switch a.bridge.Call("purchaseStatus").String() {
	case bridgePurchasePending:
		return types.PurchaseStatusPending
	case bridgePurchaseGranted:
		return types.PurchaseStatusGranted
	case bridgePurchaseDenied:
		return types.PurchaseStatusDenied
	default:
		return types.PurchaseStatusNone
	}
}

// GetPurchases реализует IPurchaseAdapter: массив { id, token } моста
func (a *PlatformAdapter) GetPurchases() []types.Purchase {
	items := a.bridgeArray("purchases")
	purchases := make([]types.Purchase, 0, len(items))
	for _, item := range items {
		purchases = append(purchases, types.Purchase{
			ProductID: stringField(item, "id"),
			Token:     stringField(item, "token"),
		})
	}
	return purchases
}

// ConsumePurchase реализует IPurchaseAdapter
func (a *PlatformAdapter) ConsumePurchase(token string) {
	a.bridge.Call("consumePurchase", token)
}

// bridgeArray — элементы массива, который возвращает метод моста;
// не массив — пусто
func (a *PlatformAdapter) bridgeArray(method string) []js.Value {
	array := a.bridge.Call(method)
	if !array.InstanceOf(js.Global().Get("Array")) {
		return nil
	}
	items := make([]js.Value, array.Length())
	for i := range items {
		items[i] = array.Index(i)
	}
	return items
}

// stringField — строковое поле объекта моста; иначе пустая строка
func stringField(object js.Value, name string) string {
	if object.Type() != js.TypeObject {
		return ""
	}
	value := object.Get(name)
	if value.Type() != js.TypeString {
		return ""
	}
	return value.String()
}
