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
	image_providers "github.com/shpaker/gonflict/internal/types/image_providers"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// GameStateRendererAdapter адаптер для рендеринга игры
type GameStateRendererAdapter struct {
	mapUseCases            interfaces.IMapUseCases
	playerTank             *types.TankEntity
	tankRenderUseCases     interfaces.ITankRenderUseCases // Общий use case для графики всех танков
	bulletUseCases         interfaces.IBulletUseCases
	enemyTanks             []*types.TankEntity // Массив врагов
	mapTilesUseCases       *use_cases.TilesUseCases
	playerTilesUseCases    *use_cases.TilesUseCases
	bulletTilesUseCases    *use_cases.TilesUseCases
	spawnerTilesUseCases   *use_cases.TilesUseCases
	explosionTilesUseCases *use_cases.TilesUseCases
	hqTilesUseCases        *use_cases.TilesUseCases
	hq                     *types.HQEntity
	hqUseCases             interfaces.IHQUseCases
	imageCache             map[string]*ebiten.Image // Кэш ebiten.Image
	imageService           *services.ImageService   // Сервис для работы с изображениями
	tileMinSize            int
	mapOffsetX             int
	mapOffsetY             int
	mapWidthHeight         int
}

// NewGameStateRendererAdapter создает новый экземпляр GameStateRendererAdapter
func NewGameStateRendererAdapter(
	mapUseCases interfaces.IMapUseCases,
	playerTank *types.TankEntity,
	tankRenderUseCases interfaces.ITankRenderUseCases,
	bulletUseCases interfaces.IBulletUseCases,
	enemyTanks []*types.TankEntity,
	mapTilesUseCases *use_cases.TilesUseCases,
	playerTilesUseCases *use_cases.TilesUseCases,
	bulletTilesUseCases *use_cases.TilesUseCases,
	spawnerTilesUseCases *use_cases.TilesUseCases,
	explosionTilesUseCases *use_cases.TilesUseCases,
	hqTilesUseCases *use_cases.TilesUseCases,
	hq *types.HQEntity,
	hqUseCases interfaces.IHQUseCases,
	tileMinSize int,
	mapOffsetX int,
	mapOffsetY int,
	mapWidthHeight int,
) *GameStateRendererAdapter {
	return &GameStateRendererAdapter{
		mapUseCases:            mapUseCases,
		playerTank:             playerTank,
		tankRenderUseCases:     tankRenderUseCases,
		bulletUseCases:         bulletUseCases,
		enemyTanks:             enemyTanks,
		mapTilesUseCases:       mapTilesUseCases,
		playerTilesUseCases:    playerTilesUseCases,
		bulletTilesUseCases:    bulletTilesUseCases,
		spawnerTilesUseCases:   spawnerTilesUseCases,
		explosionTilesUseCases: explosionTilesUseCases,
		hqTilesUseCases:        hqTilesUseCases,
		hq:                     hq,
		hqUseCases:             hqUseCases,
		imageCache:             make(map[string]*ebiten.Image),
		imageService:           services.NewImageService(),
		tileMinSize:            tileMinSize,
		mapOffsetX:             mapOffsetX,
		mapOffsetY:             mapOffsetY,
		mapWidthHeight:         mapWidthHeight,
	}
}

// drawTank отрисовывает танк
func (r *GameStateRendererAdapter) drawTank(screen *ebiten.Image) {
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

	// Получаем ID изображения танка напрямую из Image
	if tank.Image == nil {
		return
	}
	imageID, err := tank.Image.GetImageID()
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
	screenX := float64(r.mapOffsetX) + tank.Position.X
	screenY := float64(r.mapOffsetY) + tank.Position.Y

	// Создаем опции для отрисовки
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(screenX, screenY)

	screen.DrawImage(rotatedImage, op)
}

// getCachedImage возвращает закэшированное ebiten.Image или создает новое
func (r *GameStateRendererAdapter) getCachedImage(
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
func (r *GameStateRendererAdapter) drawEnemiesWithoutExplosions(
	screen *ebiten.Image,
) {
	for _, enemy := range r.enemyTanks {
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
			r.drawEnemySpawnAnimation(screen, enemy)
			continue
		}

		// Получаем ID изображения врага напрямую из Image
		if enemy.Image == nil {
			continue
		}
		imageID, err := enemy.Image.GetImageID()
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
			float64(r.mapOffsetX)+enemy.Position.X,
			float64(r.mapOffsetY)+enemy.Position.Y,
		)

		screen.DrawImage(rotatedImage, op)
	}
}

// drawEnemiesExplosions отрисовывает взрывы врагов (уровень AIR)
func (r *GameStateRendererAdapter) drawEnemiesExplosions(screen *ebiten.Image) {
	for _, enemy := range r.enemyTanks {
		// Пропускаем если врага нет или он не взрывается
		if enemy == nil || enemy.State != types.TankStateExploding {
			continue
		}

		// Получаем ID изображения взрыва напрямую из Image
		if enemy.Image == nil {
			continue
		}
		imageID, err := enemy.Image.GetImageID()
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
		if tileAnim, ok := enemy.Image.(*image_providers.AnimationProvider); ok {
			offsetX = tileAnim.Offset[0]
			offsetY = tileAnim.Offset[1]
		}

		op.GeoM.Translate(
			float64(r.mapOffsetX)+enemy.Position.X+offsetX,
			float64(r.mapOffsetY)+enemy.Position.Y+offsetY,
		)

		screen.DrawImage(img, op)
	}
}

// drawEnemySpawnAnimation отрисовывает анимацию спавна врага
func (r *GameStateRendererAdapter) drawEnemySpawnAnimation(
	screen *ebiten.Image,
	enemy *types.TankEntity,
) {
	// Получаем ID изображения анимации спавна напрямую из Image
	if enemy.Image == nil {
		return
	}
	imageID, err := enemy.Image.GetImageID()
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
		float64(r.mapOffsetX)+enemy.Position.X,
		float64(r.mapOffsetY)+enemy.Position.Y,
	)

	screen.DrawImage(image, op)
}

// drawSpawnAnimation отрисовывает анимацию спавна
func (r *GameStateRendererAdapter) drawSpawnAnimation(
	screen *ebiten.Image,
	tank *types.TankEntity,
) {
	// Получаем ID изображения анимации спавна напрямую из Image
	if tank.Image == nil {
		return
	}
	imageID, err := tank.Image.GetImageID()
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
	screenX := float64(r.mapOffsetX) + tank.Position.X
	screenY := float64(r.mapOffsetY) + tank.Position.Y

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
func (r *GameStateRendererAdapter) drawHQ(screen *ebiten.Image) {
	if r.hq == nil {
		return
	}

	// Пропускаем отрисовку взорванной базы (она отрисовывается как разрушенная)
	if r.hq.State == types.HQStateDestroyed {
		// Отрисовываем разрушенную базу
		if r.hq.Image == nil {
			return
		}
		imageID, err := r.hq.Image.GetImageID()
		if err != nil {
			return
		}

		imageData, err := r.hqTilesUseCases.GetImage(imageID)
		if err != nil {
			return
		}

		img := r.getCachedImage(imageID, imageData)
		screenX := float64(r.mapOffsetX) + r.hq.Position.X
		screenY := float64(r.mapOffsetY) + r.hq.Position.Y

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
	if r.hq.Image == nil {
		return
	}
	imageID, err := r.hq.Image.GetImageID()
	if err != nil {
		return
	}

	imageData, err := r.hqTilesUseCases.GetImage(imageID)
	if err != nil {
		return
	}

	img := r.getCachedImage(imageID, imageData)
	screenX := float64(r.mapOffsetX) + r.hq.Position.X
	screenY := float64(r.mapOffsetY) + r.hq.Position.Y

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(screenX, screenY)
	screen.DrawImage(img, op)
}

// drawHQExplosion отрисовывает взрыв базы (уровень AIR)
func (r *GameStateRendererAdapter) drawHQExplosion(screen *ebiten.Image) {
	if r.hq == nil || r.hq.State != types.HQStateExploding ||
		r.hqUseCases == nil {
		return
	}

	// Получаем HQUseCases как конкретный тип для доступа к AnimationGetter
	hqUseCases, ok := r.hqUseCases.(*use_cases.HQUseCases)
	if !ok || hqUseCases.AnimationGetter == nil {
		return
	}
	imageID, err := hqUseCases.AnimationGetter.GetImageID()
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
	if tileAnim, ok := hqUseCases.AnimationGetter.(*image_providers.AnimationProvider); ok {
		offsetX = tileAnim.Offset[0]
		offsetY = tileAnim.Offset[1]
	}

	op.GeoM.Translate(
		float64(r.mapOffsetX)+r.hq.Position.X+offsetX,
		float64(r.mapOffsetY)+r.hq.Position.Y+offsetY,
	)

	screen.DrawImage(img, op)
}

// drawBullets отрисовывает пули
func (r *GameStateRendererAdapter) drawBullets(screen *ebiten.Image) {
	bullets := r.bulletUseCases.GetBullets()

	for _, bullet := range bullets {
		if bullet.Image != nil {
			// Получаем ID изображения пули
			imageID, err := bullet.Image.GetImageID()
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
			screenX := float64(r.mapOffsetX) + bullet.Position.X
			screenY := float64(r.mapOffsetY) + bullet.Position.Y

			// Создаем опции для отрисовки
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(screenX, screenY)

			screen.DrawImage(rotatedImage, op)
		}
	}
}

// DrawAll отрисовывает все элементы игры
func (r *GameStateRendererAdapter) DrawAll(screen *ebiten.Image) {
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
func (r *GameStateRendererAdapter) drawScreenBackground(screen *ebiten.Image) {
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
func (r *GameStateRendererAdapter) drawMapBackground(screen *ebiten.Image) {
	vector.FillRect(
		screen,
		float32(r.mapOffsetX),
		float32(r.mapOffsetY),
		float32(r.mapWidthHeight),
		float32(r.mapWidthHeight),
		color.Black,
		false,
	)
}

// drawBlocksByAltitude отрисовывает блоки на определенном уровне высоты
func (r *GameStateRendererAdapter) drawBlocksByAltitude(
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
		imageID, err := block.Image.GetImageID()
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
			float64(r.mapOffsetX)+block.Position.X*float64(r.tileMinSize),
			float64(r.mapOffsetY)+block.Position.Y*float64(r.tileMinSize),
		)
		screen.DrawImage(image, op)
	}
}
