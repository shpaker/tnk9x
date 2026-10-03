package use_cases

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.IShopUseCases = (*ShopUseCases)(nil)

// ShopUseCases — магазин: товары конфигурации, которые продаёт
// площадка, покупка и зачисление купленного в инвентарь. Купленное
// общее для всех режимов, поэтому пачка снимается с продажи, только
// когда открыта в каждом из них
type ShopUseCases struct {
	// Configuration
	products []types.ProductSpec
	// Entities
	campaign  *types.CampaignEntity
	inventory *types.InventoryEntity
	// Use Cases
	progressions []interfaces.IProgressionUseCases
	// Adapters
	purchaseAdapter interfaces.IPurchaseAdapter
	// Repositories
	inventoryRepository interfaces.IInventoryRepository
}

func NewShopUseCases(
	products []types.ProductSpec,
	campaign *types.CampaignEntity,
	inventory *types.InventoryEntity,
	progressions []interfaces.IProgressionUseCases,
	purchaseAdapter interfaces.IPurchaseAdapter,
	inventoryRepository interfaces.IInventoryRepository,
) *ShopUseCases {
	return &ShopUseCases{
		products:            products,
		campaign:            campaign,
		inventory:           inventory,
		progressions:        progressions,
		purchaseAdapter:     purchaseAdapter,
		inventoryRepository: inventoryRepository,
	}
}

// IsAvailable реализует IShopUseCases
func (uc *ShopUseCases) IsAvailable() bool {
	return uc.purchaseAdapter.IsPurchaseAvailable() && len(uc.GetRows()) > 0
}

// GetRows реализует IShopUseCases: товары одного вида — одна строка
// на месте первого из них в конфигурации
func (uc *ShopUseCases) GetRows() []types.ShopRow {
	prices := make(map[string]string)
	for _, offer := range uc.purchaseAdapter.GetCatalog() {
		prices[offer.ID] = offer.Price
	}

	var rows []types.ShopRow
	rowIndex := make(map[types.ProductKind]int)
	for _, product := range uc.products {
		price, ok := prices[product.ID]
		if !ok || !uc.isForSale(product) {
			continue
		}
		offer := uc.newOffer(product, price)
		index, ok := rowIndex[product.Kind]
		if !ok {
			rowIndex[product.Kind] = len(rows)
			rows = append(rows, types.ShopRow{Kind: product.Kind})
			index = len(rows) - 1
		}
		rows[index].Offers = append(rows[index].Offers, offer)
	}
	return rows
}

// IsOnSale реализует IShopUseCases
func (uc *ShopUseCases) IsOnSale(kind types.ProductKind) bool {
	for _, row := range uc.GetRows() {
		if row.Kind == kind {
			return true
		}
	}
	return false
}

// RequestPurchase реализует IShopUseCases
func (uc *ShopUseCases) RequestPurchase(productID string) {
	uc.purchaseAdapter.RequestPurchase(productID)
}

// PollPurchase реализует IShopUseCases
func (uc *ShopUseCases) PollPurchase() (types.PurchaseStatus, error) {
	status := uc.purchaseAdapter.PollPurchase()
	if status != types.PurchaseStatusGranted {
		return status, nil
	}
	return status, uc.Sync()
}

// Sync реализует IShopUseCases. Расходуемая покупка сначала
// зачисляется и сохраняется, и только потом списывается у площадки:
// сбой между шагами не теряет покупку, а отметка о зачислении
// не даёт зачислить её второй раз
func (uc *ShopUseCases) Sync() error {
	changed := false
	var consume []string
	for _, purchase := range uc.purchaseAdapter.GetPurchases() {
		product, ok := uc.findProduct(purchase.ProductID)
		if !ok {
			continue
		}
		if product.Kind.IsConsumable() {
			if !uc.inventory.IsCredited(purchase.Token) {
				uc.inventory.AddTokens(product.Amount)
				uc.inventory.MarkCredited(purchase.Token)
				changed = true
			}
			consume = append(consume, purchase.Token)
			continue
		}
		if !uc.isOwned(product) {
			uc.grant(product)
			changed = true
		}
	}

	if changed {
		if err := uc.inventoryRepository.SaveInventory(uc.inventory); err != nil {
			return err
		}
	}
	for _, token := range consume {
		uc.purchaseAdapter.ConsumePurchase(token)
	}
	return nil
}

// newOffer — товар магазина: уровни, которые он открывает
func (uc *ShopUseCases) newOffer(
	product types.ProductSpec,
	price string,
) types.ShopOffer {
	offer := types.ShopOffer{
		Product: product,
		Price:   price,
		Owned:   uc.isOwned(product),
	}
	switch product.Kind {
	case types.ProductKindPack:
		pack := uc.campaign.GetPacks()[product.Pack-1]
		offer.PackName = pack.Name
		offer.Levels = pack.Levels
	case types.ProductKindAllLevels:
		for _, pack := range uc.campaign.GetPacks() {
			offer.Levels = append(offer.Levels, pack.Levels...)
		}
	}
	return offer
}

// isForSale: пачка не продаётся после покупки всех уровней и когда
// все её уровни уже открыты в каждом режиме
func (uc *ShopUseCases) isForSale(product types.ProductSpec) bool {
	if product.Kind != types.ProductKindPack {
		return true
	}
	packs := uc.campaign.GetPacks()
	if product.Pack < 1 || product.Pack > len(packs) ||
		uc.inventory.HasAllLevels() {
		return false
	}
	// Купленная пачка остаётся в магазине с пометкой «куплено»
	if uc.inventory.IsPackOwned(product.Pack) {
		return true
	}
	for _, progression := range uc.progressions {
		for _, level := range packs[product.Pack-1].Levels {
			if !progression.IsLevelUnlocked(level) {
				return true
			}
		}
	}
	return false
}

// isOwned — разовая покупка уже есть в инвентаре
func (uc *ShopUseCases) isOwned(product types.ProductSpec) bool {
	switch product.Kind {
	case types.ProductKindNoAds:
		return uc.inventory.HasNoAds()
	case types.ProductKindAllLevels:
		return uc.inventory.HasAllLevels()
	case types.ProductKindPack:
		return uc.inventory.IsPackOwned(product.Pack)
	default:
		return false
	}
}

// grant зачисляет разовую покупку
func (uc *ShopUseCases) grant(product types.ProductSpec) {
	switch product.Kind {
	case types.ProductKindNoAds:
		uc.inventory.SetNoAds(true)
	case types.ProductKindAllLevels:
		uc.inventory.SetAllLevels(true)
	case types.ProductKindPack:
		uc.inventory.SetPackOwned(product.Pack)
	}
}

func (uc *ShopUseCases) findProduct(id string) (types.ProductSpec, bool) {
	for _, product := range uc.products {
		if product.ID == id {
			return product, true
		}
	}
	return types.ProductSpec{}, false
}
