package adapters

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/opentype"

	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/services"
	"github.com/shpaker/gonflict/internal/types"
	image_providers "github.com/shpaker/gonflict/internal/types/image_providers"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// StageRendererAdapter адаптер для рендеринга уровня игры
type StageRendererAdapter struct {
	mapUseCases            interfaces.IMapUseCases
	tankCommonUseCases     interfaces.ITankCommonUseCases // Для получения всех танков
	tankRenderUseCases     interfaces.ITankRenderUseCases // Общий use case для графики всех танков
	bulletUseCases         interfaces.IBulletUseCases
	mapTilesUseCases       *use_cases.TilesUseCases
	playerTilesUseCases    *use_cases.TilesUseCases
	bulletTilesUseCases    *use_cases.TilesUseCases
	spawnerTilesUseCases   *use_cases.TilesUseCases
	explosionTilesUseCases *use_cases.TilesUseCases
	hqTilesUseCases        *use_cases.TilesUseCases
	hqUseCases             interfaces.IHQUseCases
	imageCache             map[string]*ebiten.Image // Кэш ebiten.Image
	imageService           *services.ImageService   // Сервис для работы с изображениями
	fontUseCases           interfaces.IFontUseCases
	pauseFontFace          text.Face
	tileMinSize            int
	mapOffsetX             int
	mapOffsetY             int
	mapWidthHeight         int
}

// NewStageRendererAdapter создает новый экземпляр StageRendererAdapter
func NewStageRendererAdapter(
	mapUseCases interfaces.IMapUseCases,
	tankCommonUseCases interfaces.ITankCommonUseCases,
	tankRenderUseCases interfaces.ITankRenderUseCases,
	bulletUseCases interfaces.IBulletUseCases,
	mapTilesUseCases *use_cases.TilesUseCases,
	playerTilesUseCases *use_cases.TilesUseCases,
	bulletTilesUseCases *use_cases.TilesUseCases,
	spawnerTilesUseCases *use_cases.TilesUseCases,
	explosionTilesUseCases *use_cases.TilesUseCases,
	hqTilesUseCases *use_cases.TilesUseCases,
	hqUseCases interfaces.IHQUseCases,
	fontUseCases interfaces.IFontUseCases,
	tileMinSize int,
	mapOffsetX int,
	mapOffsetY int,
	mapWidthHeight int,
) *StageRendererAdapter {
	return &StageRendererAdapter{
		mapUseCases:            mapUseCases,
		tankCommonUseCases:     tankCommonUseCases,
		tankRenderUseCases:     tankRenderUseCases,
		bulletUseCases:         bulletUseCases,
		mapTilesUseCases:       mapTilesUseCases,
		playerTilesUseCases:    playerTilesUseCases,
		bulletTilesUseCases:    bulletTilesUseCases,
		spawnerTilesUseCases:   spawnerTilesUseCases,
		explosionTilesUseCases: explosionTilesUseCases,
		hqTilesUseCases:        hqTilesUseCases,
		hqUseCases:             hqUseCases,
		fontUseCases:           fontUseCases,
		imageCache:             make(map[string]*ebiten.Image),
		imageService:           services.NewImageService(),
		tileMinSize:            tileMinSize,
		mapOffsetX:             mapOffsetX,
		mapOffsetY:             mapOffsetY,
		mapWidthHeight:         mapWidthHeight,
	}
}

// drawTanks отрисовывает все танки без взрывов (уровень SURFACE)
func (r *StageRendererAdapter) drawTanks(screen *ebiten.Image) {
	allTanks := r.tankCommonUseCases.GetAllTanks()
	for _, tank := range allTanks {
		if tank == nil {
			continue
		}

		// Пропускаем взрывающихся и взорванных танков
		if tank.State == types.TankStateExploding ||
			tank.State == types.TankStateExploded {
			continue
		}

		// Если танк в процессе спавна, отображаем анимацию спавна
		if tank.State == types.TankStateSpawning {
			r.drawSpawnAnimation(screen, tank)
			continue
		}

		// Получаем ID изображения танка напрямую из Image
		if tank.Image == nil {
			continue
		}
		imageID, err := tank.Image.GetImageID()
		if err != nil {
			continue
		}

		// Получаем изображение танка через TankTilesUseCases
		imageData, err := r.playerTilesUseCases.GetImage(imageID)
		if err != nil {
			continue
		}

		// Получаем закэшированное изображение
		img := r.getCachedImage(imageID, imageData)

		// Поворачиваем изображение в зависимости от направления
		rotatedImg := r.imageService.RotateImage(img, tank.Direction)
		rotatedImage, ok := rotatedImg.(*ebiten.Image)
		if !ok {
			rotatedImage = img
		}

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(
			float64(r.mapOffsetX)+tank.Position.X,
			float64(r.mapOffsetY)+tank.Position.Y,
		)

		screen.DrawImage(rotatedImage, op)
	}
}

// getCachedImage возвращает закэшированное ebiten.Image или создает новое
func (r *StageRendererAdapter) getCachedImage(
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

// drawSpawnAnimation отрисовывает анимацию спавна
func (r *StageRendererAdapter) drawSpawnAnimation(
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
	img := ebiten.NewImageFromImage(imageData)

	// Вычисляем позицию на экране (в центре позиции танка)
	screenX := float64(r.mapOffsetX) + tank.Position.X
	screenY := float64(r.mapOffsetY) + tank.Position.Y

	// Создаем опции для отрисовки
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(screenX, screenY)

	screen.DrawImage(img, op)
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

// drawHeadquarters отрисовывает базу
func (r *StageRendererAdapter) drawHeadquarters(screen *ebiten.Image) {
	hq := r.hqUseCases.GetHQ()
	if hq == nil {
		return
	}

	// Пропускаем отрисовку взорванной базы (она отрисовывается как разрушенная)
	if hq.State == types.HQStateDestroyed {
		// Отрисовываем разрушенную базу
		if hq.Image == nil {
			return
		}
		imageID, err := hq.Image.GetImageID()
		if err != nil {
			return
		}

		imageData, err := r.hqTilesUseCases.GetImage(imageID)
		if err != nil {
			return
		}

		img := r.getCachedImage(imageID, imageData)
		screenX := float64(r.mapOffsetX) + hq.Position.X
		screenY := float64(r.mapOffsetY) + hq.Position.Y

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(screenX, screenY)
		screen.DrawImage(img, op)
		return
	}

	// Пропускаем отрисовку базы во время взрыва (взрыв будет отрисован отдельно)
	if hq.State == types.HQStateExploding {
		return
	}

	// Отрисовываем целую базу
	if hq.Image == nil {
		return
	}
	imageID, err := hq.Image.GetImageID()
	if err != nil {
		return
	}

	imageData, err := r.hqTilesUseCases.GetImage(imageID)
	if err != nil {
		return
	}

	img := r.getCachedImage(imageID, imageData)
	screenX := float64(r.mapOffsetX) + hq.Position.X
	screenY := float64(r.mapOffsetY) + hq.Position.Y

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(screenX, screenY)
	screen.DrawImage(img, op)
}

// drawExplosions отрисовывает взрывы всех сущностей (танки и HQ, уровень AIR)
func (r *StageRendererAdapter) drawExplosions(screen *ebiten.Image) {
	// Отрисовываем взрывы танков
	allTanks := r.tankCommonUseCases.GetAllTanks()
	for _, tank := range allTanks {
		if tank == nil || tank.State != types.TankStateExploding {
			continue
		}

		// Получаем ID изображения взрыва напрямую из Image
		if tank.Image == nil {
			continue
		}
		imageID, err := tank.Image.GetImageID()
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
		if tileAnim, ok := tank.Image.(*image_providers.AnimationProvider); ok {
			offsetX = tileAnim.Offset[0]
			offsetY = tileAnim.Offset[1]
		}

		op.GeoM.Translate(
			float64(r.mapOffsetX)+tank.Position.X+offsetX,
			float64(r.mapOffsetY)+tank.Position.Y+offsetY,
		)

		screen.DrawImage(img, op)
	}

	// Отрисовываем взрыв HQ
	hq := r.hqUseCases.GetHQ()
	if hq != nil && hq.State == types.HQStateExploding && hq.Image != nil {
		// Получаем Image из HQEntity (анимация взрыва хранится в entity)
		imageID, err := hq.Image.GetImageID()
		if err == nil {
			// Получаем изображение через explosion tileset
			imageData, err := r.explosionTilesUseCases.GetImage(imageID)
			if err == nil {
				// Получаем закэшированное изображение
				img := r.getCachedImage(imageID, imageData)

				op := &ebiten.DrawImageOptions{}

				// Применяем offset если это анимация
				var offsetX, offsetY float64 = 0, 0
				if tileAnim, ok := hq.Image.(*image_providers.AnimationProvider); ok {
					offsetX = tileAnim.Offset[0]
					offsetY = tileAnim.Offset[1]
				}

				op.GeoM.Translate(
					float64(r.mapOffsetX)+hq.Position.X+offsetX,
					float64(r.mapOffsetY)+hq.Position.Y+offsetY,
				)

				screen.DrawImage(img, op)
			}
		}
	}
}

// drawBullets отрисовывает пули
func (r *StageRendererAdapter) drawBullets(screen *ebiten.Image) {
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
			img := ebiten.NewImageFromImage(imageData)

			// Поворачиваем изображение в зависимости от направления пули
			rotationAngle := getRotationAngle(bullet.Direction)
			rotatedImg, err := r.imageService.RotateImageByAngle(
				img,
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
func (r *StageRendererAdapter) DrawAll(screen *ebiten.Image) {
	// Сначала отрисовываем серый фон экрана
	r.drawScreenBackground(screen)
	// Затем отрисовываем черный фон карты
	r.drawMapBackground(screen)
	// Затем отрисовываем блоки уровня GROUND
	r.drawBlocksByAltitude(screen, types.GROUND)
	// Затем отрисовываем базу (если на уровне SURFACE)
	r.drawHeadquarters(screen)
	// Затем отрисовываем все танки без взрывов (если на уровне SURFACE)
	r.drawTanks(screen)
	// Затем отрисовываем пули (если на уровне SURFACE)
	r.drawBullets(screen)
	// Затем отрисовываем блоки уровня SURFACE (если есть)
	r.drawBlocksByAltitude(screen, types.SURFACE)
	// Затем отрисовываем взрывы всех сущностей (танки и HQ, на уровне AIR)
	r.drawExplosions(screen)
	// В конце отрисовываем блоки уровня AIR (деревья)
	r.drawBlocksByAltitude(screen, types.AIR)
}

// DrawPauseOverlay отрисовывает экранную заставку паузы
func (r *StageRendererAdapter) DrawPauseOverlay(screen *ebiten.Image) {
	bounds := screen.Bounds()
	width := float32(bounds.Dx())
	height := float32(bounds.Dy())

	vector.FillRect(
		screen,
		0,
		0,
		width,
		height,
		color.NRGBA{R: 0, G: 0, B: 0, A: 180},
		false,
	)

	pauseFace := r.ensurePauseFontFace()
	if pauseFace == nil {
		return
	}

	message := "PAUSED"
	textWidth, textHeight := text.Measure(message, pauseFace, 0)
	x := (float64(bounds.Dx()) - textWidth) / 2
	y := float64(bounds.Dy())/2 - textHeight/2

	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(color.White)

	text.Draw(screen, message, pauseFace, op)
}

func (r *StageRendererAdapter) ensurePauseFontFace() text.Face {
	if r.pauseFontFace != nil {
		return r.pauseFontFace
	}
	if r.fontUseCases == nil {
		return nil
	}

	baseFont, err := r.fontUseCases.GetFont()
	if err != nil || baseFont == nil {
		return nil
	}

	face, err := opentype.NewFace(baseFont, &opentype.FaceOptions{
		Size: 32,
		DPI:  72,
	})
	if err != nil {
		return nil
	}

	r.pauseFontFace = text.NewGoXFace(face)
	return r.pauseFontFace
}

// drawScreenBackground отрисовывает серый фон экрана
func (r *StageRendererAdapter) drawScreenBackground(screen *ebiten.Image) {
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
func (r *StageRendererAdapter) drawMapBackground(screen *ebiten.Image) {
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
func (r *StageRendererAdapter) drawBlocksByAltitude(
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
		img := ebiten.NewImageFromImage(imageData)

		op := &ebiten.DrawImageOptions{}
		// Блоки уже хранят позиции в пикселях, используем их напрямую
		op.GeoM.Translate(
			float64(r.mapOffsetX)+block.Position.X,
			float64(r.mapOffsetY)+block.Position.Y,
		)
		screen.DrawImage(img, op)
	}
}
