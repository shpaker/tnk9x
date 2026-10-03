// Package shop рисует магазин: строки товаров с ценами и описание
// выбранного товара — что именно получает игрок.
package shop

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"

	"github.com/shpaker/tnk9x/internal/adapters/ui"
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// Раскладка магазина в логических координатах 256x224: заголовок,
// строки товаров и под ними описание выбранного товара
const (
	titleTop = 24
	rowsTop  = 64
	rowStep  = 16
	// descriptionGap — отступ описания от последней строки
	descriptionGap  = 12
	descriptionStep = 12
	// priceGap — зазор между колонками названий и цен
	priceGap = 16
)

// Цвета магазина
var (
	rowColor         = color.NRGBA{R: 150, G: 150, B: 150, A: 255}
	activeRowColor   = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	ownedColor       = color.NRGBA{R: 120, G: 200, B: 120, A: 255}
	descriptionColor = color.NRGBA{R: 230, G: 200, B: 90, A: 255}
)

// ShopRendererAdapter рисует магазин оверлеем поверх главного меню
type ShopRendererAdapter struct {
	texts interfaces.ITextsAdapter
	font  ui.MenuFont
	// hits — строки последней отрисовки для мыши и тапов
	hits ui.HitAreas
}

func NewShopRendererAdapter(
	texts interfaces.ITextsAdapter,
	fontFace text.Face,
	titleFontSize int,
	regularFontSize int,
) *ShopRendererAdapter {
	return &ShopRendererAdapter{
		texts: texts,
		font: ui.MenuFont{
			Face:            fontFace,
			TitleFontSize:   titleFontSize,
			RegularFontSize: regularFontSize,
		},
	}
}

// shopLine — строка для отрисовки: название и цена или «куплено»
type shopLine struct {
	label string
	value string
	owned bool
}

func (r *ShopRendererAdapter) Draw(
	screen *ebiten.Image,
	view types.ShopViewData,
) {
	width := float64(screen.Bounds().Dx())
	ui.DrawOverlay(screen, r.font, r.texts.Get(types.TextShopTitle), titleTop)

	lines := make([]shopLine, len(view.Rows))
	var labelsWidth, valuesWidth float64
	for i, row := range view.Rows {
		lines[i] = r.line(row, i == view.ActiveIndex)
		if row.Back {
			continue
		}
		labelsWidth = max(labelsWidth, r.font.TextWidth(lines[i].label))
		valuesWidth = max(valuesWidth, r.font.TextWidth(lines[i].value))
	}
	// Ширина названий — с запасом под стрелки выбора варианта
	labelsWidth = max(labelsWidth, r.widestArrowedLabel(view))
	left := math.Round((width - labelsWidth - priceGap - valuesWidth) / 2)
	valuesRight := left + labelsWidth + priceGap + valuesWidth

	rowHeight := float64(r.font.RegularFontSize)
	r.hits.Reset()
	for i, line := range lines {
		top := float64(rowsTop + i*rowStep)
		textColor := rowColor
		if i == view.ActiveIndex {
			textColor = activeRowColor
		}

		// BACK — по центру всей ширины
		if view.Rows[i].Back {
			labelWidth := r.font.TextWidth(line.label)
			labelLeft := math.Round((width - labelWidth) / 2)
			r.hits.Add(ui.RowRect(
				min(left, labelLeft), max(valuesRight, labelLeft+labelWidth),
				top, rowHeight, rowStep-rowHeight,
			))
			r.font.Draw(screen, line.label, labelLeft, top, textColor)
			continue
		}

		r.hits.Add(ui.RowRect(
			left, valuesRight, top, rowHeight, rowStep-rowHeight,
		))
		r.font.Draw(screen, line.label, left, top, textColor)
		valueColor := textColor
		if line.owned {
			valueColor = ownedColor
		}
		r.font.Draw(
			screen, line.value,
			math.Round(valuesRight-r.font.TextWidth(line.value)), top,
			valueColor,
		)
	}

	top := float64(rowsTop+(len(lines)-1)*rowStep) + rowHeight +
		descriptionGap
	for i, description := range r.description(view) {
		r.font.Draw(
			screen, description,
			math.Round((width-r.font.TextWidth(description))/2),
			top+float64(i*descriptionStep),
			descriptionColor,
		)
	}
}

// HitRow — строка последней отрисовки под точкой
func (r *ShopRendererAdapter) HitRow(position types.Position) (int, bool) {
	return r.hits.Hit(position)
}

// line — название и цена строки; у выбранной строки с вариантами
// название в стрелках
func (r *ShopRendererAdapter) line(
	row types.ShopRowView,
	active bool,
) shopLine {
	if row.Back {
		return shopLine{label: r.texts.Get(types.TextShopBack)}
	}
	label := r.label(row.Offer)
	if active && row.Variants > 1 {
		label = "< " + label + " >"
	}
	if row.Offer.Owned {
		return shopLine{
			label: label,
			value: r.texts.Get(types.TextShopOwned),
			owned: true,
		}
	}
	return shopLine{label: label, value: row.Offer.Price}
}

// widestArrowedLabel — ширина самого длинного названия в стрелках
// среди строк с вариантами: колонка не прыгает при выборе строки
func (r *ShopRendererAdapter) widestArrowedLabel(
	view types.ShopViewData,
) float64 {
	var widest float64
	for _, row := range view.Rows {
		if row.Back || row.Variants <= 1 {
			continue
		}
		widest = max(
			widest, r.font.TextWidth("< "+r.label(row.Offer)+" >"),
		)
	}
	return widest
}

// label — название товара: пачка — её название, жетоны — размер
// пакета
func (r *ShopRendererAdapter) label(offer types.ShopOffer) string {
	switch offer.Product.Kind {
	case types.ProductKindNoAds:
		return r.texts.Get(types.TextShopNoAds)
	case types.ProductKindAllLevels:
		return r.texts.Get(types.TextShopAllLevels)
	case types.ProductKindPack:
		return r.texts.GetOr(
			types.PackNameTextKey(offer.Product.Pack), offer.PackName,
		)
	default:
		return r.texts.Plural(
			types.TextShopTokens, int(offer.Product.Amount), nil,
		)
	}
}

// description — что получает игрок за выбранный товар; пока открыто
// окно оплаты — ожидание
func (r *ShopRendererAdapter) description(view types.ShopViewData) []string {
	if view.Pending {
		return []string{r.texts.Get(types.TextShopPending)}
	}
	if view.ActiveIndex < 0 || view.ActiveIndex >= len(view.Rows) ||
		view.Rows[view.ActiveIndex].Back {
		return nil
	}

	offer := view.Rows[view.ActiveIndex].Offer
	switch offer.Product.Kind {
	case types.ProductKindNoAds:
		return []string{
			r.texts.Get(types.TextShopDescNoAds),
			r.texts.Get(types.TextShopDescNoAdsRewarded),
		}
	case types.ProductKindAllLevels:
		return []string{
			r.texts.Format(types.TextShopDescAllLevels, types.TextArgs{
				"Count": len(offer.Levels),
			}),
			r.texts.Get(types.TextShopDescAllLevelsAtOnce),
		}
	case types.ProductKindPack:
		if len(offer.Levels) == 0 {
			return nil
		}
		return []string{
			r.texts.Format(types.TextShopDescPack, types.TextArgs{
				"From": offer.Levels[0],
				"To":   offer.Levels[len(offer.Levels)-1],
			}),
			r.texts.Get(types.TextShopDescPackAll),
		}
	default:
		return []string{
			r.texts.Get(types.TextShopDescTokens),
			r.texts.Get(types.TextShopDescTokensNoAd),
			r.texts.Format(types.TextShopDescBalance, types.TextArgs{
				"Count": view.Tokens,
			}),
		}
	}
}
