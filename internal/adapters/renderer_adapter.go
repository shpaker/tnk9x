package adapters

import (
	"image"
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
	mapUseCases            use_cases.IMapUseCases
	tankUseCases           use_cases.IPlayerUseCases
	bulletUseCases         use_cases.IBulletUseCases
	enemyUseCasesList      []*use_cases.EnemyUseCases // Массив врагов
	mapTilesUseCases       *use_cases.TilesUseCases
	playerTilesUseCases    *use_cases.TilesUseCases
	bulletTilesUseCases    *use_cases.TilesUseCases
	spawnerTilesUseCases   *use_cases.TilesUseCases
	explosionTilesUseCases *use_cases.TilesUseCases
	imageCache             map[string]*ebiten.Image // Кэш ebiten.Image
}

// NewRendererAdapter создает новый экземпляр RendererAdapter
func NewRendererAdapter(
	mapUseCases use_cases.IMapUseCases,
	tankUseCases use_cases.IPlayerUseCases,
	bulletUseCases use_cases.IBulletUseCases,
	enemyUseCasesList []*use_cases.EnemyUseCases,
	mapTilesUseCases *use_cases.TilesUseCases,
	playerTilesUseCases *use_cases.TilesUseCases,
	bulletTilesUseCases *use_cases.TilesUseCases,
	spawnerTilesUseCases *use_cases.TilesUseCases,
	explosionTilesUseCases *use_cases.TilesUseCases,
) *RendererAdapter {
	return &RendererAdapter{
		mapUseCases:            mapUseCases,
		tankUseCases:           tankUseCases,
		bulletUseCases:         bulletUseCases,
		enemyUseCasesList:      enemyUseCasesList,
		mapTilesUseCases:       mapTilesUseCases,
		playerTilesUseCases:    playerTilesUseCases,
		bulletTilesUseCases:    bulletTilesUseCases,
		spawnerTilesUseCases:   spawnerTilesUseCases,
		explosionTilesUseCases: explosionTilesUseCases,
		imageCache:             make(map[string]*ebiten.Image),
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
		// Предполагаем, что блоки имеют координаты X, Y в Position
		op.GeoM.Translate(
			use_cases.MapOffset+block.Position.X*use_cases.TileMinSize,
			use_cases.MapOffset+block.Position.Y*use_cases.TileMinSize,
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
	if tank.State == types.TankStateSpawning {
		r.drawSpawnAnimation(screen, tank)
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
	screenX := use_cases.MapOffset + tank.Position.X
	screenY := use_cases.MapOffset + tank.Position.Y

	// Создаем опции для отрисовки
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(screenX, screenY)

	screen.DrawImage(rotatedImage, op)
}

// getCachedImage возвращает закэшированное ebiten.Image или создает новое
func (r *RendererAdapter) getCachedImage(imageId string, imageData image.Image) *ebiten.Image {
	// Проверяем кэш
	if cachedImage, exists := r.imageCache[imageId]; exists {
		return cachedImage
	}

	// Проверяем размер изображения
	if imageData.Bounds().Dx() == 0 || imageData.Bounds().Dy() == 0 {
		log.Printf("ERROR: Image '%s' has zero size (bounds: %s)", imageId, imageData.Bounds())
		// Возвращаем пустое изображение 1x1 вместо nil
		ebitenImage := ebiten.NewImage(1, 1)
		return ebitenImage
	}

	// Создаем новое изображение и кэшируем его
	ebitenImage := ebiten.NewImageFromImage(imageData)
	r.imageCache[imageId] = ebitenImage
	return ebitenImage
}

// drawEnemiesWithoutExplosions отрисовывает врагов без взрывов (уровень SURFACE)
func (r *RendererAdapter) drawEnemiesWithoutExplosions(screen *ebiten.Image) {
	for i, enemyUseCases := range r.enemyUseCasesList {
		enemy := enemyUseCases.GetEnemy()

		// Пропускаем если врага нет
		if enemy == nil {
			continue
		}

		// Пропускаем взрывающихся врагов
		if enemy.State == types.TankStateExploding {
			continue
		}

		// Если враг в процессе спавна, отображаем анимацию спавна
		if enemy.State == types.TankStateSpawning {
			r.drawEnemySpawnAnimation(screen, enemy, i)
			continue
		}

		// Получаем ID изображения врага
		imageId, err := enemy.AnimationGetter.GetImageId()
		if err != nil {
			log.Printf("ERROR: Enemy error getting image ID: %v", err)
			continue
		}

		// Получаем изображение через TilesUseCases
		imageData, err := r.playerTilesUseCases.GetImage(imageId)
		if err != nil {
			log.Printf("ERROR: Enemy error loading image '%s': %v", imageId, err)
			continue
		}

		// Получаем закэшированное изображение
		img := r.getCachedImage(imageId, imageData)

		// Поворачиваем изображение в зависимости от направления
		rotatedImage := utils.RotateImage(img, enemy.Direction)

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(
			use_cases.MapOffset+enemy.Position.X,
			use_cases.MapOffset+enemy.Position.Y,
		)

		screen.DrawImage(rotatedImage, op)
	}
}

// drawEnemiesExplosions отрисовывает взрывы врагов (уровень AIR)
func (r *RendererAdapter) drawEnemiesExplosions(screen *ebiten.Image) {
	for _, enemyUseCases := range r.enemyUseCasesList {
		enemy := enemyUseCases.GetEnemy()

		// Пропускаем если врага нет или он не взрывается
		if enemy == nil || enemy.State != types.TankStateExploding {
			continue
		}

		// Получаем ID изображения взрыва
		imageId, err := enemy.AnimationGetter.GetImageId()
		if err != nil {
			log.Printf("ERROR: Enemy explosion error getting image ID: %v", err)
			continue
		}

		// Получаем изображение через explosion tileset
		imageData, err := r.explosionTilesUseCases.GetImage(imageId)
		if err != nil {
			log.Printf("ERROR: Enemy explosion error loading image '%s': %v", imageId, err)
			continue
		}

		// Получаем закэшированное изображение
		img := r.getCachedImage(imageId, imageData)

		op := &ebiten.DrawImageOptions{}

		// Применяем offset если это анимация
		var offsetX, offsetY float64 = 0, 0
		if anim, ok := enemy.AnimationGetter.(*types.TileAnimationEntity); ok {
			offsetX = anim.Offset[0]
			offsetY = anim.Offset[1]
		}

		op.GeoM.Translate(
			use_cases.MapOffset+enemy.Position.X+offsetX,
			use_cases.MapOffset+enemy.Position.Y+offsetY,
		)

		screen.DrawImage(img, op)
	}
}

// drawEnemySpawnAnimation отрисовывает анимацию спавна врага
func (r *RendererAdapter) drawEnemySpawnAnimation(screen *ebiten.Image, enemy *types.TankEntity, enemyIndex int) {
	// Получаем анимацию спавна напрямую из AnimationGetter танка
	spawnAnimation := enemy.AnimationGetter
	if spawnAnimation == nil {
		return
	}

	// Получаем ID изображения анимации спавна
	imageId, err := spawnAnimation.GetImageId()
	if err != nil {
		log.Printf("ERROR: Enemy spawn animation error getting image ID: %v", err)
		return
	}

	// Получаем изображение через TilesUseCases
	imageData, err := r.spawnerTilesUseCases.GetImage(imageId)
	if err != nil {
		log.Printf("ERROR: Enemy spawn animation error loading image '%s': %v", imageId, err)
		return
	}

	// Конвертируем image.Image в ebiten.Image
	image := ebiten.NewImageFromImage(imageData)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(
		use_cases.MapOffset+enemy.Position.X,
		use_cases.MapOffset+enemy.Position.Y,
	)

	screen.DrawImage(image, op)
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
	screenX := use_cases.MapOffset + tank.Position.X
	screenY := use_cases.MapOffset + tank.Position.Y

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
			screenX := use_cases.MapOffset + bullet.Position.X
			screenY := use_cases.MapOffset + bullet.Position.Y

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
	// Сначала отрисовываем серый фон экрана
	r.drawScreenBackground(screen)
	// Затем отрисовываем черный фон карты
	r.drawMapBackground(screen)
	// Затем отрисовываем блоки уровня GROUND
	r.drawBlocksByAltitude(screen, types.GROUND)
	// Затем отрисовываем танк игрока (если на уровне SURFACE)
	r.drawTank(screen)
	// Затем отрисовываем врагов без взрывов (если на уровне SURFACE)
	r.drawEnemiesWithoutExplosions(screen)
	// Затем отрисовываем пули (если на уровне SURFACE)
	r.drawBullets(screen)
	// Затем отрисовываем блоки уровня SURFACE (если есть)
	r.drawBlocksByAltitude(screen, types.SURFACE)
	// Затем отрисовываем взрывы врагов (на уровне AIR)
	r.drawEnemiesExplosions(screen)
	// В конце отрисовываем блоки уровня AIR (деревья)
	r.drawBlocksByAltitude(screen, types.AIR)
}

// drawScreenBackground отрисовывает серый фон экрана
func (r *RendererAdapter) drawScreenBackground(screen *ebiten.Image) {
	vector.FillRect(
		screen,
		0,
		0,
		float32(screen.Bounds().Dx()),
		float32(screen.Bounds().Dy()),
		color.Gray{Y: 128},
		false,
	)
}

// drawMapBackground отрисовывает черный фон карты
func (r *RendererAdapter) drawMapBackground(screen *ebiten.Image) {
	vector.FillRect(
		screen,
		float32(use_cases.MapOffset),
		float32(use_cases.MapOffset),
		float32(use_cases.MapWidthHeight),
		float32(use_cases.MapWidthHeight),
		color.Black,
		false,
	)
}

// drawBlocksByAltitude отрисовывает блоки на определенном уровне высоты
func (r *RendererAdapter) drawBlocksByAltitude(screen *ebiten.Image, altitude types.Altitude) {
	blocks := r.mapUseCases.GetBlocks()
	for i, block := range blocks {
		// Пропускаем блоки других уровней
		if block.Altitude != altitude {
			continue
		}

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
		// Предполагаем, что блоки имеют координаты X, Y в Position
		op.GeoM.Translate(
			use_cases.MapOffset+block.Position.X*use_cases.TileMinSize,
			use_cases.MapOffset+block.Position.Y*use_cases.TileMinSize,
		)
		screen.DrawImage(image, op)
	}
}
