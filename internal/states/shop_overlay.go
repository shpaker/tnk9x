package states

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// ShopRenderer — контракт рендера магазина, определён у потребителя
type ShopRenderer interface {
	Draw(screen *ebiten.Image, view types.ShopViewData)
	// HitRow — строка последней отрисовки под точкой
	HitRow(position types.Position) (int, bool)
}

// ShopOverlay — магазин поверх главного меню; один на приложение.
// Строка — товары одного вида: вариант (пачка, пакет жетонов)
// выбирается влево-вправо, под списком — что именно получает игрок.
// Выбор некупленного товара открывает окно оплаты площадки; пока
// оно открыто, ввод не принимается
type ShopOverlay struct {
	// Use Cases
	shopUseCases      interfaces.IShopUseCases
	inventoryUseCases interfaces.IInventoryUseCases
	// Adapters
	renderer    ShopRenderer
	menuInput   interfaces.IMenuInputAdapter
	soundPlayer interfaces.ISoundPlayerAdapter

	open        bool
	pending     bool
	activeIndex int
	// variants — выбранный вариант строки по виду товара
	variants map[types.ProductKind]int
}

func NewShopOverlay(
	shopUseCases interfaces.IShopUseCases,
	inventoryUseCases interfaces.IInventoryUseCases,
	renderer ShopRenderer,
	menuInput interfaces.IMenuInputAdapter,
	soundPlayer interfaces.ISoundPlayerAdapter,
) *ShopOverlay {
	return &ShopOverlay{
		shopUseCases:      shopUseCases,
		inventoryUseCases: inventoryUseCases,
		renderer:          renderer,
		menuInput:         menuInput,
		soundPlayer:       soundPlayer,
		variants:          make(map[types.ProductKind]int),
	}
}

// Open показывает магазин с курсором на первой строке
func (o *ShopOverlay) Open() {
	o.open = true
	o.activeIndex = 0
}

func (o *ShopOverlay) IsOpen() bool {
	return o.open
}

// Update: вверх-вниз или наведение — выбор строки, влево-вправо
// или колесо — вариант товара, выбор или клик — покупка; Esc, B,
// Start, правая кнопка мыши или BACK закрывают магазин
func (o *ShopOverlay) Update() {
	if o.pending {
		o.pollPurchase()
		return
	}
	if o.menuInput.Back() {
		o.open = false
		return
	}

	rows := o.shopUseCases.GetRows()
	// Последняя строка — BACK
	count := len(rows) + 1
	o.activeIndex = min(o.activeIndex, count-1)
	moveUp, moveDown := o.menuInput.Steps()
	o.activeIndex = stepIndex(o.activeIndex, count, moveUp, moveDown)
	activeIndex, clicked := pointerIndex(
		o.menuInput, o.renderer.HitRow, o.activeIndex,
	)
	// Строки могли пропасть с прошлой отрисовки (куплены все уровни)
	o.activeIndex = min(activeIndex, count-1)
	confirmed := o.menuInput.Confirmed() || clicked

	if o.activeIndex == len(rows) {
		if confirmed {
			o.open = false
		}
		return
	}

	row := rows[o.activeIndex]
	if step := o.menuInput.SideStep(); step != 0 {
		o.variants[row.Kind] = wrapIndex(
			o.variant(row)+step, len(row.Offers),
		)
	}
	offer := row.Offers[o.variant(row)]
	if confirmed && !offer.Owned {
		o.pending = true
		o.shopUseCases.RequestPurchase(offer.Product.ID)
	}
}

func (o *ShopOverlay) Draw(screen *ebiten.Image) {
	rows := o.shopUseCases.GetRows()
	view := types.ShopViewData{
		Rows:        make([]types.ShopRowView, 0, len(rows)+1),
		ActiveIndex: min(o.activeIndex, len(rows)),
		Pending:     o.pending,
		Tokens:      o.inventoryUseCases.GetTokens(),
	}
	for _, row := range rows {
		view.Rows = append(view.Rows, types.ShopRowView{
			Offer:    row.Offers[o.variant(row)],
			Variants: len(row.Offers),
		})
	}
	view.Rows = append(view.Rows, types.ShopRowView{Back: true})
	o.renderer.Draw(screen, view)
}

// pollPurchase ждёт исхода оплаты: оплаченное уже зачислено,
// сигнал подтверждает покупку
func (o *ShopOverlay) pollPurchase() {
	status, err := o.shopUseCases.PollPurchase()
	if err != nil {
		log.Printf("save inventory: %v", err)
	}
	switch status {
	case types.PurchaseStatusPending:
		return
	case types.PurchaseStatusGranted:
		if err := o.soundPlayer.Play(types.SoundIDBonus); err != nil {
			log.Printf("sound %q: %v", types.SoundIDBonus, err)
		}
	}
	o.pending = false
}

// variant — выбранный вариант строки в пределах её товаров: после
// покупки всех уровней пачки пропадают, вариантов может стать меньше
func (o *ShopOverlay) variant(row types.ShopRow) int {
	return min(o.variants[row.Kind], len(row.Offers)-1)
}

// wrapIndex — индекс по кругу в пределах count
func wrapIndex(index, count int) int {
	return ((index % count) + count) % count
}
