package app

import "github.com/hajimehoshi/ebiten/v2"

var _ ebiten.FinalScreenDrawer = (*App)(nil)

// DrawFinalScreen масштабирует логический экран 256x224 целым
// множителем (чёткие пиксели NES), центрирует и оставляет чёрные
// поля; на десктопе окно 768x672 даёт прежний множитель 3
func (app *App) DrawFinalScreen(
	screen ebiten.FinalScreen,
	offscreen *ebiten.Image,
	_ ebiten.GeoM,
) {
	screen.Clear()
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	ow, oh := offscreen.Bounds().Dx(), offscreen.Bounds().Dy()
	scale := min(sw/ow, sh/oh) // целочисленное деление = floor
	if scale < 1 {
		scale = 1
	}
	op := &ebiten.DrawImageOptions{} // Filter по умолчанию — Nearest
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(
		float64(sw-ow*scale)/2,
		float64(sh-oh*scale)/2,
	)
	screen.DrawImage(offscreen, op)
}
