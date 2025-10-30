package adapters

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/services"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// RendererAdapter адаптер для рендеринга игры
type RendererAdapter struct {
	mapUseCases            interfaces.IMapUseCases
	playerTank             *types.TankEntity
	playerRenderUseCases   interfaces.ITankRenderUseCases // Use Cases для графики игрока
	bulletUseCases         interfaces.IBulletUseCases
	enemyTanks             []*types.TankEntity              // Массив врагов
	enemyRenderUseCases    []interfaces.ITankRenderUseCases // Use Cases для графики врагов
	mapTilesUseCases       *use_cases.TilesUseCases
	playerTilesUseCases    *use_cases.TilesUseCases
	bulletTilesUseCases    *use_cases.TilesUseCases
	spawnerTilesUseCases   *use_cases.TilesUseCases
	explosionTilesUseCases *use_cases.TilesUseCases
	hqTilesUseCases        *use_cases.TilesUseCases
	hq                     *types.HQEntity
	hqRenderUseCases       *use_cases.HQRenderUseCases
	imageCache             map[string]*ebiten.Image // Кэш ebiten.Image
	imageService           *services.ImageService   // Сервис для работы с изображениями
}

// NewRendererAdapter создает новый экземпляр RendererAdapter
func NewRendererAdapter(
	mapUseCases interfaces.IMapUseCases,
	playerTank *types.TankEntity,
	playerRenderUseCases interfaces.ITankRenderUseCases,
	bulletUseCases interfaces.IBulletUseCases,
	enemyTanks []*types.TankEntity,
	enemyRenderUseCases []interfaces.ITankRenderUseCases,
	mapTilesUseCases *use_cases.TilesUseCases,
	playerTilesUseCases *use_cases.TilesUseCases,
	bulletTilesUseCases *use_cases.TilesUseCases,
	spawnerTilesUseCases *use_cases.TilesUseCases,
	explosionTilesUseCases *use_cases.TilesUseCases,
	hqTilesUseCases *use_cases.TilesUseCases,
	hq *types.HQEntity,
	hqRenderUseCases *use_cases.HQRenderUseCases,
) *RendererAdapter {
	return &RendererAdapter{
		mapUseCases:            mapUseCases,
		playerTank:             playerTank,
		playerRenderUseCases:   playerRenderUseCases,
		bulletUseCases:         bulletUseCases,
		enemyTanks:             enemyTanks,
		enemyRenderUseCases:    enemyRenderUseCases,
		mapTilesUseCases:       mapTilesUseCases,
		playerTilesUseCases:    playerTilesUseCases,
		bulletTilesUseCases:    bulletTilesUseCases,
		spawnerTilesUseCases:   spawnerTilesUseCases,
		explosionTilesUseCases: explosionTilesUseCases,
		hqTilesUseCases:        hqTilesUseCases,
		hq:                     hq,
		hqRenderUseCases:       hqRenderUseCases,
		imageCache:             make(map[string]*ebiten.Image),
		imageService:           services.NewImageService(),
	}
}

// drawTank отрисовывает танк
func (r *RendererAdapter) drawTank(screen *ebiten.Image) {
	tank := r.playerTank
	if tank == nil {
		return
	}

	// Пропускаем отрисовку взорванного танка
	if tank.State == types.TankStateExploded {
		return
	}

	// Если танк в процессе спавна, отображаем анимацию спавна
	if tank.State == types.TankStateSpawning {
		r.drawSpawnAnimation(screen, tank)
		return
	}

	// Получаем ID изображения танка напрямую из AnimationGetter
	tankRender, ok := r.playerRenderUseCases.(*use_cases.TankRenderUseCases)
	if !ok || tankRender.AnimationGetter == nil {
		return
	}
	imageID, err := tankRender.AnimationGetter.GetImageID()
	if err != nil {
		return
	}

	// Получаем изображение танка через TankTilesUseCases
	imageData, err := r.playerTilesUseCases.GetImage(imageID)
	if err != nil {
		return
	}

	// Конвертируем image.Image в ebiten.Image
	image := ebiten.NewImageFromImage(imageData)

	// Поворачиваем изображение в зависимости от направления танка
	rotationAngle := getRotationAngle(tank.Direction)
	rotatedImg, err := r.imageService.RotateImageByAngle(image, rotationAngle)
	if err != nil {
		return
	}
	rotatedImage, ok := rotatedImg.(*ebiten.Image)
	if !ok {
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
func (r *RendererAdapter) getCachedImage(
	imageID string,
	imageData image.Image,
) *ebiten.Image {
	// Проверяем кэш
	if cachedImage, exists := r.imageCache[imageID]; exists {
		return cachedImage
	}

	// Проверяем размер изображения
	if imageData.Bounds().Dx() == 0 || imageData.Bounds().Dy() == 0 {
		// Возвращаем пустое изображение 1x1 вместо nil
		ebitenImage := ebiten.NewImage(1, 1)
		return ebitenImage
	}

	// Создаем новое изображение и кэшируем его
	ebitenImage := ebiten.NewImageFromImage(imageData)
	r.imageCache[imageID] = ebitenImage
	return ebitenImage
}

// drawEnemiesWithoutExplosions отрисовывает врагов без взрывов (уровень SURFACE)
func (r *RendererAdapter) drawEnemiesWithoutExplosions(screen *ebiten.Image) {
	for i, enemy := range r.enemyTanks {
		// Пропускаем если врага нет
		if enemy == nil {
			continue
		}

		// Пропускаем взрывающихся и взорванных врагов
		if enemy.State == types.TankStateExploding ||
			enemy.State == types.TankStateExploded {
			continue
		}

		// Если враг в процессе спавна, отображаем анимацию спавна
		if enemy.State == types.TankStateSpawning {
			r.drawEnemySpawnAnimation(screen, enemy, i)
			continue
		}

		// Получаем ID изображения врага напрямую из AnimationGetter
		if i >= len(r.enemyRenderUseCases) {
			continue
		}
		enemyRenderUseCases := r.enemyRenderUseCases[i]
		enemyTankRender, ok := enemyRenderUseCases.(*use_cases.TankRenderUseCases)
		if !ok || enemyTankRender.AnimationGetter == nil {
			continue
		}
		imageID, err := enemyTankRender.AnimationGetter.GetImageID()
		if err != nil {
			continue
		}

		// Получаем изображение через TilesUseCases
		imageData, err := r.playerTilesUseCases.GetImage(imageID)
		if err != nil {
			continue
		}

		// Получаем закэшированное изображение
		img := r.getCachedImage(imageID, imageData)

		// Поворачиваем изображение в зависимости от направления
		rotatedImg := r.imageService.RotateImage(img, enemy.Direction)
		rotatedImage, ok := rotatedImg.(*ebiten.Image)
		if !ok {
			rotatedImage = img
		}

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
	for i, enemy := range r.enemyTanks {
		// Пропускаем если врага нет или он не взрывается
		if enemy == nil || enemy.State != types.TankStateExploding {
			continue
		}

		// Получаем ID изображения взрыва напрямую из AnimationGetter
		if i >= len(r.enemyRenderUseCases) {
			continue
		}
		enemyRenderUseCases := r.enemyRenderUseCases[i]
		enemyTankRender, ok := enemyRenderUseCases.(*use_cases.TankRenderUseCases)
		if !ok || enemyTankRender.AnimationGetter == nil {
			continue
		}
		imageID, err := enemyTankRender.AnimationGetter.GetImageID()
		if err != nil {
			continue
		}

		// Получаем изображение через explosion tileset
		imageData, err := r.explosionTilesUseCases.GetImage(imageID)
		if err != nil {
			continue
		}

		// Получаем закэшированное изображение
		img := r.getCachedImage(imageID, imageData)

		op := &ebiten.DrawImageOptions{}

		// Применяем offset если это анимация
		var offsetX, offsetY float64 = 0, 0
		if tileAnim, ok := enemyTankRender.AnimationGetter.(*types.TileAnimationEntity); ok {
			offsetX = tileAnim.Offset[0]
			offsetY = tileAnim.Offset[1]
		}

		op.GeoM.Translate(
			use_cases.MapOffset+enemy.Position.X+offsetX,
			use_cases.MapOffset+enemy.Position.Y+offsetY,
		)

		screen.DrawImage(img, op)
	}
}

// drawEnemySpawnAnimation отрисовывает анимацию спавна врага
func (r *RendererAdapter) drawEnemySpawnAnimation(
	screen *ebiten.Image,
	enemy *types.TankEntity,
	enemyIndex int,
) {
	// Получаем TankRenderUseCases для этого врага
	if enemyIndex >= len(r.enemyRenderUseCases) {
		return
	}
	enemyRenderUseCases := r.enemyRenderUseCases[enemyIndex]

	// Получаем ID изображения анимации спавна напрямую из AnimationGetter
	enemyTankRender, ok := enemyRenderUseCases.(*use_cases.TankRenderUseCases)
	if !ok || enemyTankRender.AnimationGetter == nil {
		return
	}
	imageID, err := enemyTankRender.AnimationGetter.GetImageID()
	if err != nil {
		return
	}

	// Получаем изображение через TilesUseCases
	imageData, err := r.spawnerTilesUseCases.GetImage(imageID)
	if err != nil {
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
func (r *RendererAdapter) drawSpawnAnimation(
	screen *ebiten.Image,
	tank *types.TankEntity,
) {
	// Получаем ID изображения анимации спавна напрямую из AnimationGetter
	playerTankRender, ok := r.playerRenderUseCases.(*use_cases.TankRenderUseCases)
	if !ok || playerTankRender.AnimationGetter == nil {
		return
	}
	imageID, err := playerTankRender.AnimationGetter.GetImageID()
	if err != nil {
		return
	}

	// Получаем изображение анимации спавна через SpawnerTilesUseCases
	imageData, err := r.spawnerTilesUseCases.GetImage(imageID)
	if err != nil {
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

// drawHQ отрисовывает базу
func (r *RendererAdapter) drawHQ(screen *ebiten.Image) {
	if r.hq == nil {
		return
	}

	// Пропускаем отрисовку взорванной базы (она отрисовывается как разрушенная)
	if r.hq.State == types.HQStateDestroyed {
		// Отрисовываем разрушенную базу
		imageID, err := r.hq.GetImageID()
		if err != nil {
			return
		}

		imageData, err := r.hqTilesUseCases.GetImage(imageID)
		if err != nil {
			return
		}

		img := r.getCachedImage(imageID, imageData)
		screenX := use_cases.MapOffset + r.hq.Position.X
		screenY := use_cases.MapOffset + r.hq.Position.Y

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(screenX, screenY)
		screen.DrawImage(img, op)
		return
	}

	// Пропускаем отрисовку базы во время взрыва (взрыв будет отрисован отдельно)
	if r.hq.State == types.HQStateExploding {
		return
	}

	// Отрисовываем целую базу
	imageID, err := r.hq.GetImageID()
	if err != nil {
		return
	}

	imageData, err := r.hqTilesUseCases.GetImage(imageID)
	if err != nil {
		return
	}

	img := r.getCachedImage(imageID, imageData)
	screenX := use_cases.MapOffset + r.hq.Position.X
	screenY := use_cases.MapOffset + r.hq.Position.Y

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(screenX, screenY)
	screen.DrawImage(img, op)
}

// drawHQExplosion отрисовывает взрыв базы (уровень AIR)
func (r *RendererAdapter) drawHQExplosion(screen *ebiten.Image) {
	if r.hq == nil || r.hq.State != types.HQStateExploding ||
		r.hqRenderUseCases == nil {
		return
	}

	// Получаем ID изображения взрыва напрямую из AnimationGetter
	if r.hqRenderUseCases.AnimationGetter == nil {
		return
	}
	imageID, err := r.hqRenderUseCases.AnimationGetter.GetImageID()
	if err != nil {
		return
	}

	// Получаем изображение через explosion tileset
	imageData, err := r.explosionTilesUseCases.GetImage(imageID)
	if err != nil {
		return
	}

	// Получаем закэшированное изображение
	img := r.getCachedImage(imageID, imageData)

	op := &ebiten.DrawImageOptions{}

	// Применяем offset если это анимация
	var offsetX, offsetY float64 = 0, 0
	if tileAnim, ok := r.hqRenderUseCases.AnimationGetter.(*types.TileAnimationEntity); ok {
		offsetX = tileAnim.Offset[0]
		offsetY = tileAnim.Offset[1]
	}

	op.GeoM.Translate(
		use_cases.MapOffset+r.hq.Position.X+offsetX,
		use_cases.MapOffset+r.hq.Position.Y+offsetY,
	)

	screen.DrawImage(img, op)
}

// drawBullets отрисовывает пули
func (r *RendererAdapter) drawBullets(screen *ebiten.Image) {
	bullets := r.bulletUseCases.GetBullets()

	for _, bullet := range bullets {
		if bullet.ImageGetter != nil {
			// Получаем ID изображения пули
			imageID, err := bullet.ImageGetter.GetImageID()
			if err != nil {
				continue
			}

			// Получаем изображение пули через BulletTilesUseCases
			imageData, err := r.bulletTilesUseCases.GetImage(imageID)
			if err != nil {
				continue
			}

			// Конвертируем image.Image в ebiten.Image
			image := ebiten.NewImageFromImage(imageData)

			// Поворачиваем изображение в зависимости от направления пули
			rotationAngle := getRotationAngle(bullet.Direction)
			rotatedImg, err := r.imageService.RotateImageByAngle(
				image,
				rotationAngle,
			)
			if err != nil {
				continue
			}
			rotatedImage, ok := rotatedImg.(*ebiten.Image)
			if !ok {
				continue
			}

			// Вычисляем позицию на экране
			screenX := use_cases.MapOffset + bullet.Position.X
			screenY := use_cases.MapOffset + bullet.Position.Y

			// Создаем опции для отрисовки
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(screenX, screenY)

			screen.DrawImage(rotatedImage, op)
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
	// Затем отрисовываем базу (если на уровне SURFACE)
	r.drawHQ(screen)
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
	// Затем отрисовываем взрыв базы (на уровне AIR)
	r.drawHQExplosion(screen)
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
func (r *RendererAdapter) drawBlocksByAltitude(
	screen *ebiten.Image,
	altitude types.Altitude,
) {
	blocks := r.mapUseCases.GetBlocks()
	for _, block := range blocks {
		// Пропускаем блоки других уровней
		if block.Altitude != altitude {
			continue
		}

		// Получаем ID изображения блока
		imageID, err := block.ImageGetter.GetImageID()
		if err != nil {
			continue
		}

		// Получаем изображение блока через TilesUseCases
		imageData, err := r.mapTilesUseCases.GetImage(imageID)
		if err != nil {
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
