package adapters

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// RendererAdapter адаптер для рендеринга игры
type RendererAdapter struct {
	mapUseCases         use_cases.IMapUseCases
	playerUseCases      use_cases.IPlayerUseCases
	bulletUseCases      use_cases.IBulletUseCases
	mapTilesUseCases    *use_cases.TilesUseCases
	playerTilesUseCases *use_cases.TilesUseCases
	bulletTilesUseCases *use_cases.TilesUseCases
}

// NewRendererAdapter создает новый экземпляр RendererAdapter
func NewRendererAdapter(
	mapUseCases use_cases.IMapUseCases,
	playerUseCases use_cases.IPlayerUseCases,
	bulletUseCases use_cases.IBulletUseCases,
	mapTilesUseCases *use_cases.TilesUseCases,
	playerTilesUseCases *use_cases.TilesUseCases,
	bulletTilesUseCases *use_cases.TilesUseCases,
) *RendererAdapter {
	return &RendererAdapter{
		mapUseCases:         mapUseCases,
		playerUseCases:      playerUseCases,
		bulletUseCases:      bulletUseCases,
		mapTilesUseCases:    mapTilesUseCases,
		playerTilesUseCases: playerTilesUseCases,
		bulletTilesUseCases: bulletTilesUseCases,
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
		// Получаем ID изображения блока
		imageId, err := block.ImageGetter.GetImageId()
		if err != nil {
			fmt.Printf("DEBUG: Block %d error getting image ID: %v\n", i, err)
			continue
		}

		// Получаем изображение блока через TilesUseCases
		imageData, err := r.mapTilesUseCases.GetImage(imageId)
		if err != nil {
			// Логируем ошибку, но продолжаем рендеринг других блоков
			fmt.Printf("DEBUG: Block %d error: %v\n", i, err)
			continue
		}

		// Конвертируем image.Image в ebiten.Image
		image := ebiten.NewImageFromImage(imageData)

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

	// Получаем ID изображения танка
	imageId, err := tank.ImageGetter.GetImageId()
	if err != nil {
		return
	}

	// Получаем изображение танка через PlayerTilesUseCases
	imageData, err := r.playerTilesUseCases.GetImage(imageId)
	if err != nil {
		return
	}

	// Конвертируем image.Image в ebiten.Image
	image := ebiten.NewImageFromImage(imageData)

	// Вычисляем позицию на экране
	screenX := MapOffset + tank.WorldPosition.X
	screenY := MapOffset + tank.WorldPosition.Y

	// Создаем опции для отрисовки
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(screenX, screenY)

	screen.DrawImage(image, op)
}

// drawBullets отрисовывает пули
func (r *RendererAdapter) drawBullets(screen *ebiten.Image) {
	bullets := r.bulletUseCases.GetBullets()
	log.Printf("DEBUG: Rendering %d bullets", len(bullets))

	for i, bullet := range bullets {
		if bullet.ImageGetter != nil {
			// Получаем ID изображения пули
			imageId, err := bullet.ImageGetter.GetImageId()
			if err != nil {
				log.Printf("ERROR: Failed to get bullet image ID for bullet %d: %v", i, err)
				continue
			}

			// Получаем изображение пули через BulletTilesUseCases
			imageData, err := r.bulletTilesUseCases.GetImage(imageId)
			if err != nil {
				log.Printf("ERROR: Failed to get bullet image for bullet %d: %v", i, err)
				continue // Пропускаем пули с ошибками загрузки изображения
			}

			// Конвертируем image.Image в ebiten.Image
			image := ebiten.NewImageFromImage(imageData)

			// Вычисляем позицию на экране
			screenX := MapOffset + bullet.WorldPosition.X
			screenY := MapOffset + bullet.WorldPosition.Y

			log.Printf("DEBUG: Rendering bullet %d at position (%.2f, %.2f) direction %s", i, screenX, screenY, bullet.Direction)

			// Создаем опции для отрисовки
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(screenX, screenY)

			screen.DrawImage(image, op)
		} else {
			log.Printf("WARNING: Bullet %d has nil ImageGetter", i)
		}
	}
}

// DrawAll отрисовывает все элементы игры
func (r *RendererAdapter) DrawAll(screen *ebiten.Image) {
	r.drawMap(screen)
	r.drawPlayerOne(screen)
	r.drawBullets(screen)
}
