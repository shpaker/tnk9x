package app

import "github.com/hajimehoshi/ebiten/v2"

var _ ebiten.FinalScreenDrawer = (*App)(nil)

// DrawFinalScreen масштабирует логический экран 256x224 целым
// множителем (чёткие пиксели), центрирует и оставляет чёрные
// поля; с включёнными эффектами кадр проходит bloom и CRT. Геометрию
// считает адаптер тач-контролов — единый источник правды для
// отрисовки и хит-тестов касаний. Экранные контроллы рисуются
// в полях на всех экранах: меню тоже управляются крестовиной
func (app *App) DrawFinalScreen(
	screen ebiten.FinalScreen,
	offscreen *ebiten.Image,
	_ ebiten.GeoM,
) {
	screen.Clear()
	sw, sh := screen.Bounds().Dx(), screen.Bounds().Dy()
	app.touchControls.SetScreenSize(sw, sh)
	x, y, scale := app.touchControls.GameRect()

	app.effectsRenderer.DrawFinal(
		screen,
		offscreen,
		x,
		y,
		scale,
		app.settings.IsEffectsEnabled(),
	)

	app.touchControls.DrawControls(screen)
}
