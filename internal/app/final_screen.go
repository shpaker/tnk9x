package app

import "github.com/hajimehoshi/ebiten/v2"

var _ ebiten.FinalScreenDrawer = (*App)(nil)

// DrawFinalScreen масштабирует логический экран 256x224 целым
// множителем (чёткие пиксели NES), центрирует и оставляет чёрные
// поля; геометрию считает адаптер тач-контролов — единый источник
// правды для отрисовки и хит-тестов касаний. Экранные контроллы
// рисуются в полях только во время уровня
func (app *App) DrawFinalScreen(
	screen ebiten.FinalScreen,
	offscreen *ebiten.Image,
	_ ebiten.GeoM,
) {
	screen.Clear()
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	app.touchControls.SetScreenSize(sw, sh)
	x, y, scale := app.touchControls.GameRect()

	op := &ebiten.DrawImageOptions{} // Filter по умолчанию — Nearest
	op.GeoM.Scale(float64(scale), float64(scale))
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(offscreen, op)

	if app.stageState != nil {
		app.touchControls.DrawControls(screen)
	}
}
