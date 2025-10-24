package adapters

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// RendererAdapter адаптер для рендеринга игры
type RendererAdapter struct {
	mapUseCases        use_cases.IMapUseCases
	playerUseCases     use_cases.IPlayerUseCases
	bulletUseCases     use_cases.IBulletUseCases
	tilesAdapter       *TilesAdapter
	playerTilesAdapter *TilesAdapter
	bulletTilesAdapter *TilesAdapter
}

// NewRendererAdapter создает новый экземпляр RendererAdapter
func NewRendererAdapter(
	mapUseCases use_cases.IMapUseCases,
	playerUseCases use_cases.IPlayerUseCases,
	bulletUseCases use_cases.IBulletUseCases,
	tilesAdapter *TilesAdapter,
	playerTilesAdapter *TilesAdapter,
	bulletTilesAdapter *TilesAdapter,
) *RendererAdapter {
	return &RendererAdapter{
		mapUseCases:        mapUseCases,
		playerUseCases:     playerUseCases,
		bulletUseCases:     bulletUseCases,
		tilesAdapter:       tilesAdapter,
		playerTilesAdapter: playerTilesAdapter,
		bulletTilesAdapter: bulletTilesAdapter,
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

		// Получаем изображение блока через TilesAdapter
		image, err := r.tilesAdapter.GetImage(&imageId, types.DirectionUp)
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

	// Получаем ID изображения танка
	imageId, err := tank.ImageGetter.GetImageId()
	if err != nil {
		return
	}

	// Получаем изображение танка через PlayerTilesAdapter
	image, err := r.playerTilesAdapter.GetImage(&imageId, tank.Direction)
	if err != nil {
		return
	}

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

			// Получаем изображение пули через BulletTilesAdapter
			image, err := r.bulletTilesAdapter.GetImage(&imageId, bullet.Direction)
			if err != nil {
				log.Printf("ERROR: Failed to get bullet image for bullet %d: %v", i, err)
				continue // Пропускаем пули с ошибками загрузки изображения
			}

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
