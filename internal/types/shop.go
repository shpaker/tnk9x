package types

// ProductKind — вид товара магазина: что получает игрок
type ProductKind int

const (
	// ProductKindNoAds — без межуровневой рекламы; разовая покупка
	ProductKindNoAds ProductKind = iota
	// ProductKindAllLevels — все уровни кампании; разовая покупка
	ProductKindAllLevels
	// ProductKindPack — все уровни одной пачки; разовая покупка
	ProductKindPack
	// ProductKindTokens — жетоны вместо рекламы за вознаграждение;
	// расходуемая покупка
	ProductKindTokens
)

// IsConsumable — покупка расходуется: зачисляется и списывается
// у площадки, повторно её можно купить
func (k ProductKind) IsConsumable() bool {
	return k == ProductKindTokens
}

// ProductSpec — товар из конфигурации игры. ID совпадает с ID товара
// в каталоге площадки; Pack — порядковый номер пачки кампании с 1
// (только у пачки), Amount — число жетонов (только у жетонов)
type ProductSpec struct {
	ID     string
	Kind   ProductKind
	Pack   int
	Amount uint
}

// ProductOffer — товар каталога площадки: цена уже отформатирована
// площадкой вместе с валютой
type ProductOffer struct {
	ID    string
	Price string
}

// Purchase — покупка игрока у площадки; расходуемая остаётся в списке,
// пока её не спишут по Token
type Purchase struct {
	ProductID string
	Token     string
}

// PurchaseStatus — исход запроса покупки
type PurchaseStatus int

const (
	// PurchaseStatusNone — запроса нет или его исход уже получен
	PurchaseStatusNone PurchaseStatus = iota
	// PurchaseStatusPending — окно оплаты площадки ещё открыто
	PurchaseStatusPending
	// PurchaseStatusGranted — оплачено, покупка у площадки
	PurchaseStatusGranted
	// PurchaseStatusDenied — отмена, ошибка или покупка недоступна
	PurchaseStatusDenied
)

// ShopOffer — товар в магазине: спецификация, цена площадки
// и уровни, которые он открывает (пачка — её уровни, все уровни —
// уровни кампании); Owned — разовая покупка уже сделана
type ShopOffer struct {
	Product ProductSpec
	Price   string
	Owned   bool
	// PackName — название пачки из файла кампании (только у пачки)
	PackName string
	Levels   []int
}

// ShopRow — строка магазина: товары одного вида, вариант
// (пачка, размер пакета жетонов) выбирается влево-вправо
type ShopRow struct {
	Kind   ProductKind
	Offers []ShopOffer
}

// ShopRowView — строка магазина для отрисовки: выбранный вариант
// и число вариантов; Back — строка BACK
type ShopRowView struct {
	Back     bool
	Offer    ShopOffer
	Variants int
}

// ShopViewData — магазин для отрисовки: строки, курсор, идёт ли
// оплата и сколько жетонов у игрока
type ShopViewData struct {
	Rows        []ShopRowView
	ActiveIndex int
	Pending     bool
	Tokens      uint
}
