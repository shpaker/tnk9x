package adapters

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
	"github.com/shpaker/gonflict/internal/utils"
)

// RendererAdapter адаптер для рендеринга игры
type RendererAdapter struct {
	mapUseCases    use_cases.IMapUseCases
	playerUseCases use_cases.IPlayerUseCases
	bulletUseCases use_cases.IBulletUseCases
}

// NewRendererAdapter создает новый экземпляр RendererAdapter
func NewRendererAdapter(
	mapUseCases use_cases.IMapUseCases,
	playerUseCases use_cases.IPlayerUseCases,
	bulletUseCases use_cases.IBulletUseCases,
) *RendererAdapter {
	return &RendererAdapter{
		mapUseCases:    mapUseCases,
		playerUseCases: playerUseCases,
		bulletUseCases: bulletUseCases,
	}
}

// drawMap отрисовывает игровую карту
func (r *RendererAdapter) drawMap(screen *ebiten.Image) {
	vector.FillRect(
		screen,
		float32(MapOffset),
		float32(MapOffset),
		float32(MapWidthHeight),
		float32(MapWidthHeight),
		color.Black,
		false,
	)

	// Draw blocks on the map
	blocks := r.mapUseCases.GetBlocks()
	for _, block := range blocks {
		if block.Image == nil {
			continue
		}
		op := &ebiten.DrawImageOptions{}
		// Предполагаем, что блоки имеют координаты X, Y в WorldPosition
		op.GeoM.Translate(
			MapOffset+block.WorldPosition.X*TileMinSize,
			MapOffset+block.WorldPosition.Y*TileMinSize,
		)
		screen.DrawImage(block.Image, op)
	}
}

// drawPlayerOne отрисовывает игрока
func (r *RendererAdapter) drawPlayerOne(screen *ebiten.Image) {
	tank, err := r.playerUseCases.GetPlayer()
	if err != nil || tank.Image == nil {
		return
	}

	// Определяем угол поворота в зависимости от направления
	var rotationAngle float64
	switch tank.Direction {
	case types.DirectionUp:
		rotationAngle = 0
	case types.DirectionRight:
		rotationAngle = math.Pi / 2
	case types.DirectionDown:
		rotationAngle = math.Pi
	case types.DirectionLeft:
		rotationAngle = 3 * math.Pi / 2
	}

	// Вычисляем позицию на экране
	screenX := MapOffset + tank.WorldPosition.X
	screenY := MapOffset + tank.WorldPosition.Y

	// Используем функцию для поворота изображения вокруг центра
	op := utils.RotateImage(tank.Image, rotationAngle, screenX, screenY)

	screen.DrawImage(tank.Image, op)
}

// drawBullets отрисовывает пули
func (r *RendererAdapter) drawBullets(screen *ebiten.Image) {
	bullets := r.bulletUseCases.GetBullets()
	for _, bullet := range bullets {
		if bullet.Image != nil {
			// Вычисляем позицию на экране
			screenX := MapOffset + bullet.WorldPosition.X
			screenY := MapOffset + bullet.WorldPosition.Y

			// Создаем опции для отрисовки
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(screenX, screenY)

			screen.DrawImage(bullet.Image, op)
		}
	}
}

// DrawAll отрисовывает все элементы игры
func (r *RendererAdapter) DrawAll(screen *ebiten.Image) {
	r.drawMap(screen)
	r.drawPlayerOne(screen)
	r.drawBullets(screen)
}
