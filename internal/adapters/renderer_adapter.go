package adapters

import (
	"fmt"
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
	mapUseCases             use_cases.IMapUseCases
	playerUseCases          use_cases.IPlayerUseCases
	bulletUseCases          use_cases.IBulletUseCases
	tilesetRepository       types.ITilesetRepository
	playerTilesetRepository types.ITilesetRepository
}

// NewRendererAdapter создает новый экземпляр RendererAdapter
func NewRendererAdapter(
	mapUseCases use_cases.IMapUseCases,
	playerUseCases use_cases.IPlayerUseCases,
	bulletUseCases use_cases.IBulletUseCases,
	tilesetRepository types.ITilesetRepository,
	playerTilesetRepository types.ITilesetRepository,
) *RendererAdapter {
	return &RendererAdapter{
		mapUseCases:             mapUseCases,
		playerUseCases:          playerUseCases,
		bulletUseCases:          bulletUseCases,
		tilesetRepository:       tilesetRepository,
		playerTilesetRepository: playerTilesetRepository,
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
	fmt.Printf("DEBUG: Found %d blocks to render\n", len(blocks))
	for i, block := range blocks {
		// Получаем изображение блока
		image, err := block.GetImage(r.tilesetRepository)
		if err != nil {
			// Логируем ошибку, но продолжаем рендеринг других блоков
			fmt.Printf("DEBUG: Block %d error: %v\n", i, err)
			continue
		}

		fmt.Printf("DEBUG: Rendering block %d at position (%.2f, %.2f)\n", i, block.WorldPosition.X, block.WorldPosition.Y)

		op := &ebiten.DrawImageOptions{}
		// Предполагаем, что блоки имеют координаты X, Y в WorldPosition
		op.GeoM.Translate(
			MapOffset+block.WorldPosition.X*TileMinSize,
			MapOffset+block.WorldPosition.Y*TileMinSize,
		)
		screen.DrawImage(image, op)
	}
}

// drawPlayerOne отрисовывает игрока
func (r *RendererAdapter) drawPlayerOne(screen *ebiten.Image) {
	tank, err := r.playerUseCases.GetPlayer()
	if err != nil || tank.ImageGetter == nil {
		return
	}

	// Получаем изображение танка из репозитория
	imageId := tank.ImageGetter.GetImageId()
	if imageId == "" {
		return
	}
	image, err := r.playerTilesetRepository.GetImage(imageId)
	if err != nil {
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
	op := utils.RotateImage(image, rotationAngle, screenX, screenY)

	screen.DrawImage(image, op)
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
