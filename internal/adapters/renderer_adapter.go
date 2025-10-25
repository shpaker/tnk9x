package adapters

import (
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
	"github.com/shpaker/gonflict/internal/utils"
)

// RendererAdapter адаптер для рендеринга игры
type RendererAdapter struct {
	mapUseCases         use_cases.IMapUseCases
	tankUseCases        use_cases.ITankUseCases
	bulletUseCases      use_cases.IBulletUseCases
	mapTilesUseCases    *use_cases.TilesUseCases
	playerTilesUseCases *use_cases.TilesUseCases
	bulletTilesUseCases *use_cases.TilesUseCases
	spawnerTilesUseCases *use_cases.TilesUseCases
}

// NewRendererAdapter создает новый экземпляр RendererAdapter
func NewRendererAdapter(
	mapUseCases use_cases.IMapUseCases,
	tankUseCases use_cases.ITankUseCases,
	bulletUseCases use_cases.IBulletUseCases,
	mapTilesUseCases *use_cases.TilesUseCases,
	playerTilesUseCases *use_cases.TilesUseCases,
	bulletTilesUseCases *use_cases.TilesUseCases,
	spawnerTilesUseCases *use_cases.TilesUseCases,
) *RendererAdapter {
	return &RendererAdapter{
		mapUseCases:         mapUseCases,
		tankUseCases:        tankUseCases,
		bulletUseCases:      bulletUseCases,
		mapTilesUseCases:    mapTilesUseCases,
		playerTilesUseCases: playerTilesUseCases,
		bulletTilesUseCases: bulletTilesUseCases,
		spawnerTilesUseCases: spawnerTilesUseCases,
	}
}

// drawMap отрисовывает игровую карту
func (r *RendererAdapter) drawMap(screen *ebiten.Image) {
	vector.FillRect(
		screen,
		float32(use_cases.MapOffset),
		float32(use_cases.MapOffset),
		float32(use_cases.MapWidthHeight),
		float32(use_cases.MapWidthHeight),
		color.Black,
		false,
	)

	// Draw blocks on the map
	blocks := r.mapUseCases.GetBlocks()
	for i, block := range blocks {
		// Получаем ID изображения блока
		imageId, err := block.ImageGetter.GetImageId()
		if err != nil {
			log.Printf("ERROR: Block %d error getting image ID: %v", i, err)
			continue
		}

		// Получаем изображение блока через TilesUseCases
		imageData, err := r.mapTilesUseCases.GetImage(imageId)
		if err != nil {
			log.Printf("ERROR: Block %d error loading image '%s': %v", i, imageId, err)
			continue
		}

		// Конвертируем image.Image в ebiten.Image
		image := ebiten.NewImageFromImage(imageData)

		op := &ebiten.DrawImageOptions{}
		// Предполагаем, что блоки имеют координаты X, Y в WorldPosition
		op.GeoM.Translate(
			use_cases.MapOffset+block.WorldPosition.X*use_cases.TileMinSize,
			use_cases.MapOffset+block.WorldPosition.Y*use_cases.TileMinSize,
		)
		screen.DrawImage(image, op)
	}
}

// drawTank отрисовывает танк
func (r *RendererAdapter) drawTank(screen *ebiten.Image) {
	tank, err := r.tankUseCases.GetTank()
	if err != nil {
		return
	}

	// Если танк в процессе спавна, отображаем анимацию спавна
	if r.tankUseCases.IsSpawning() {
		r.drawSpawnAnimation(screen, tank)
		return
	}

	// Если танк не заспавнен, не отображаем его
	if !r.tankUseCases.ShouldShowTank() {
		return
	}

	// Получаем ID изображения танка
	imageId, err := tank.AnimationGetter.GetImageId()
	if err != nil {
		log.Printf("ERROR: Tank error getting image ID: %v", err)
		return
	}

	// Получаем изображение танка через TankTilesUseCases
	imageData, err := r.playerTilesUseCases.GetImage(imageId)
	if err != nil {
		log.Printf("ERROR: Tank error loading image '%s': %v", imageId, err)
		return
	}

	// Конвертируем image.Image в ebiten.Image
	image := ebiten.NewImageFromImage(imageData)

	// Поворачиваем изображение в зависимости от направления танка
	rotationAngle := getRotationAngle(tank.Direction)
	rotatedImage, err := utils.RotateImageByAngle(image, rotationAngle)
	if err != nil {
		log.Printf("ERROR: Tank error rotating image: %v", err)
		return
	}

	// Вычисляем позицию на экране
	screenX := use_cases.MapOffset + tank.WorldPosition.X
	screenY := use_cases.MapOffset + tank.WorldPosition.Y

	// Создаем опции для отрисовки
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(screenX, screenY)

	screen.DrawImage(rotatedImage, op)
}

// drawSpawnAnimation отрисовывает анимацию спавна
func (r *RendererAdapter) drawSpawnAnimation(screen *ebiten.Image, tank *types.TankEntity) {
	spawnAnimation := r.tankUseCases.GetSpawnAnimation()
	if spawnAnimation == nil {
		return
	}

	// Получаем ID изображения анимации спавна
	imageId, err := spawnAnimation.GetImageId()
	if err != nil {
		log.Printf("ERROR: Spawn animation error getting image ID: %v", err)
		return
	}

	// Получаем изображение анимации спавна через SpawnerTilesUseCases
	imageData, err := r.spawnerTilesUseCases.GetImage(imageId)
	if err != nil {
		log.Printf("ERROR: Spawn animation error loading image '%s': %v", imageId, err)
		return
	}

	// Конвертируем image.Image в ebiten.Image
	image := ebiten.NewImageFromImage(imageData)

	// Вычисляем позицию на экране (в центре позиции танка)
	screenX := use_cases.MapOffset + tank.WorldPosition.X
	screenY := use_cases.MapOffset + tank.WorldPosition.Y

	// Создаем опции для отрисовки
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(screenX, screenY)

	screen.DrawImage(image, op)
}

// getRotationAngle возвращает угол поворота в радианах для указанного направления
func getRotationAngle(direction types.Direction) float64 {
	switch direction {
	case types.DirectionUp:
		return 0
	case types.DirectionRight:
		return math.Pi / 2
	case types.DirectionDown:
		return math.Pi
	case types.DirectionLeft:
		return 3 * math.Pi / 2
	default:
		return 0
	}
}

// drawBullets отрисовывает пули
func (r *RendererAdapter) drawBullets(screen *ebiten.Image) {
	bullets := r.bulletUseCases.GetBullets()

	for i, bullet := range bullets {
		if bullet.ImageGetter != nil {
			// Получаем ID изображения пули
			imageId, err := bullet.ImageGetter.GetImageId()
			if err != nil {
				log.Printf("ERROR: Bullet %d error getting image ID: %v", i, err)
				continue
			}

			// Получаем изображение пули через BulletTilesUseCases
			imageData, err := r.bulletTilesUseCases.GetImage(imageId)
			if err != nil {
				log.Printf("ERROR: Bullet %d error loading image '%s': %v", i, imageId, err)
				continue
			}

			// Конвертируем image.Image в ebiten.Image
			image := ebiten.NewImageFromImage(imageData)

			// Поворачиваем изображение в зависимости от направления пули
			rotationAngle := getRotationAngle(bullet.Direction)
			rotatedImage, err := utils.RotateImageByAngle(image, rotationAngle)
			if err != nil {
				log.Printf("ERROR: Bullet %d error rotating image: %v", i, err)
				continue
			}

			// Вычисляем позицию на экране
			screenX := use_cases.MapOffset + bullet.WorldPosition.X
			screenY := use_cases.MapOffset + bullet.WorldPosition.Y

			// Создаем опции для отрисовки
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(screenX, screenY)

			screen.DrawImage(rotatedImage, op)
		} else {
			log.Printf("WARNING: Bullet %d has nil ImageGetter", i)
		}
	}
}

// DrawAll отрисовывает все элементы игры
func (r *RendererAdapter) DrawAll(screen *ebiten.Image) {
	r.drawMap(screen)
	r.drawTank(screen)
	r.drawBullets(screen)
}
