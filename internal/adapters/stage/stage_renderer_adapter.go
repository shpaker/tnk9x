package stage

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
)

type StageRendererAdapter struct {
	mapUseCases            interfaces.IMapUseCases
	tankCommonUseCases     interfaces.ITankCommonUseCases
	bulletUseCases         interfaces.IBulletUseCases
	mapTilesUseCases       interfaces.ITilesUseCases
	tankTilesUseCases      interfaces.ITilesUseCases
	bulletTilesUseCases    interfaces.ITilesUseCases
	spawnerTilesUseCases   interfaces.ITilesUseCases
	explosionTilesUseCases interfaces.ITilesUseCases
	hqTilesUseCases        interfaces.ITilesUseCases
	hqUseCases             interfaces.IHQUseCases
	hudUseCases            interfaces.IHUDUseCases
	bonusesRepository      interfaces.IBonusesRepository
	bonusTilesUseCases     interfaces.ITilesUseCases
	hudTilesUseCases       interfaces.ITilesUseCases
	imageCache             map[string]*ebiten.Image
	fontFace               text.Face
	hudFontFace            text.Face
	titleFontSize          int
	subtitleFontSize       int
	regularFontSize        int
	tileMinSize            int
	mapOffsetX             int
	mapOffsetY             int
	mapWidthHeight         int
	stageNumber            int
}

// StageRendererDependencies — готовый граф зависимостей рендера уровня;
// собирается composition root'ом, все поля обязательны
type StageRendererDependencies struct {
	// Use Cases
	MapUseCases            interfaces.IMapUseCases
	TankCommonUseCases     interfaces.ITankCommonUseCases
	BulletUseCases         interfaces.IBulletUseCases
	HQUseCases             interfaces.IHQUseCases
	HUDUseCases            interfaces.IHUDUseCases
	MapTilesUseCases       interfaces.ITilesUseCases
	TankTilesUseCases      interfaces.ITilesUseCases
	BulletTilesUseCases    interfaces.ITilesUseCases
	SpawnerTilesUseCases   interfaces.ITilesUseCases
	ExplosionTilesUseCases interfaces.ITilesUseCases
	HQTilesUseCases        interfaces.ITilesUseCases
	BonusTilesUseCases     interfaces.ITilesUseCases
	HUDTilesUseCases       interfaces.ITilesUseCases

	// Repositories
	BonusesRepository interfaces.IBonusesRepository

	// Шрифты и раскладка
	FontFace         text.Face
	HUDFontFace      text.Face
	TileMinSize      int
	MapOffsetX       int
	MapOffsetY       int
	MapWidthHeight   int
	TitleFontSize    int
	SubtitleFontSize int
	RegularFontSize  int
	StageNumber      int
}

func NewStageRendererAdapter(
	deps StageRendererDependencies,
) *StageRendererAdapter {
	titleFontSize := deps.TitleFontSize
	subtitleFontSize := deps.SubtitleFontSize
	regularFontSize := deps.RegularFontSize
	if titleFontSize <= 0 {
		titleFontSize = 32
	}
	if subtitleFontSize <= 0 {
		subtitleFontSize = titleFontSize / 2
		if subtitleFontSize == 0 {
			subtitleFontSize = 16
		}
	}
	if regularFontSize <= 0 {
		regularFontSize = subtitleFontSize
	}
	return &StageRendererAdapter{
		mapUseCases:            deps.MapUseCases,
		tankCommonUseCases:     deps.TankCommonUseCases,
		bulletUseCases:         deps.BulletUseCases,
		mapTilesUseCases:       deps.MapTilesUseCases,
		tankTilesUseCases:      deps.TankTilesUseCases,
		bulletTilesUseCases:    deps.BulletTilesUseCases,
		spawnerTilesUseCases:   deps.SpawnerTilesUseCases,
		explosionTilesUseCases: deps.ExplosionTilesUseCases,
		hqTilesUseCases:        deps.HQTilesUseCases,
		hqUseCases:             deps.HQUseCases,
		bonusesRepository:      deps.BonusesRepository,
		bonusTilesUseCases:     deps.BonusTilesUseCases,
		hudTilesUseCases:       deps.HUDTilesUseCases,
		hudUseCases:            deps.HUDUseCases,
		fontFace:               deps.FontFace,
		hudFontFace:            deps.HUDFontFace,
		imageCache:             make(map[string]*ebiten.Image),
		tileMinSize:            deps.TileMinSize,
		mapOffsetX:             deps.MapOffsetX,
		mapOffsetY:             deps.MapOffsetY,
		mapWidthHeight:         deps.MapWidthHeight,
		titleFontSize:          titleFontSize,
		subtitleFontSize:       subtitleFontSize,
		regularFontSize:        regularFontSize,
		stageNumber:            deps.StageNumber,
	}
}

func (r *StageRendererAdapter) drawTanks(screen *ebiten.Image) {
	allTanks := r.tankCommonUseCases.GetAllTanks()
	for _, tank := range allTanks {
		if tank == nil {
			continue
		}

		if tank.State == types.TankStateExploding ||
			tank.State == types.TankStateExploded {
			continue
		}

		if tank.State == types.TankStateSpawning {
			r.drawSpawnAnimation(screen, tank)
			continue
		}

		// Проверяем мигание для танков с бонусом
		if tank.IsEnemy() && tank.GetWithBonus() && !tank.GetBlinkFlag() {
			continue
		}

		if tank.Image == nil {
			continue
		}
		imageID, err := tank.Image.GetImageID()
		if err != nil {
			continue
		}

		imageData, err := r.tankTilesUseCases.GetTankImage(
			imageID,
			tank.IsEnemy(),
		)
		if err != nil {
			continue
		}

		img := r.getCachedImage(imageID, imageData)

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(
			float64(r.mapOffsetX)+tank.Position.X,
			float64(r.mapOffsetY)+tank.Position.Y,
		)

		screen.DrawImage(img, op)

		// Отрисовываем цветной слой здоровья для тяжёлого танка
		if tank.IsEnemy() && tank.GetSpecs() != nil &&
			tank.GetSpecs().GetLevel() == 3 {
			r.drawTankHealthOverlay(screen, tank)
		}
	}
}

func (r *StageRendererAdapter) getCachedImage(
	imageID string,
	imageData image.Image,
) *ebiten.Image {
	if cachedImage, exists := r.imageCache[imageID]; exists {
		return cachedImage
	}

	if imageData.Bounds().Dx() == 0 || imageData.Bounds().Dy() == 0 {

		ebitenImage := ebiten.NewImage(1, 1)
		return ebitenImage
	}

	ebitenImage := ebiten.NewImageFromImage(imageData)
	r.imageCache[imageID] = ebitenImage
	return ebitenImage
}

// drawTankHealthOverlay отрисовывает цветной полупрозрачный слой поверх тяжёлого танка
// для визуализации его здоровья
func (r *StageRendererAdapter) drawTankHealthOverlay(
	screen *ebiten.Image,
	tank *types.TankEntity,
) {
	if tank == nil {
		return
	}

	hitPoints := tank.GetHitPoints()
	// Если здоровье 1 или меньше, слой не отрисовывается
	if hitPoints <= 1 {
		return
	}

	// Определяем цвет в зависимости от здоровья
	var overlayColor color.NRGBA
	switch hitPoints {
	case 4:
		// Красный для 4 HP
		overlayColor = color.NRGBA{R: 255, G: 0, B: 0, A: 128}
	case 3:
		// Жёлтый для 3 HP
		overlayColor = color.NRGBA{R: 255, G: 255, B: 0, A: 128}
	case 2:
		// Зелёный для 2 HP
		overlayColor = color.NRGBA{R: 0, G: 255, B: 0, A: 128}
	default:
		return
	}

	// Отрисовываем полупрозрачный прямоугольник поверх танка
	screenX := float32(r.mapOffsetX) + float32(tank.Position.X)
	screenY := float32(r.mapOffsetY) + float32(tank.Position.Y)
	tankWidth := float32(tank.Size.Width)
	tankHeight := float32(tank.Size.Height)

	vector.FillRect(
		screen,
		screenX,
		screenY,
		tankWidth,
		tankHeight,
		overlayColor,
		false,
	)
}

func (r *StageRendererAdapter) drawSpawnAnimation(
	screen *ebiten.Image,
	tank *types.TankEntity,
) {
	if tank.Image == nil {
		return
	}
	imageID, err := tank.Image.GetImageID()
	if err != nil {
		return
	}

	imageData, err := r.spawnerTilesUseCases.GetImage(imageID)
	if err != nil {
		return
	}

	img := ebiten.NewImageFromImage(imageData)

	screenX := float64(r.mapOffsetX) + tank.Position.X
	screenY := float64(r.mapOffsetY) + tank.Position.Y

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(screenX, screenY)

	screen.DrawImage(img, op)
}

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

func (r *StageRendererAdapter) drawHeadquarters(screen *ebiten.Image) {
	hq := r.hqUseCases.GetHQ()
	if hq == nil {
		return
	}

	if hq.State == types.HQStateDestroyed {

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

	if hq.State == types.HQStateExploding {
		return
	}

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

func (r *StageRendererAdapter) drawExplosions(screen *ebiten.Image) {
	allTanks := r.tankCommonUseCases.GetAllTanks()
	for _, tank := range allTanks {
		if tank == nil || tank.State != types.TankStateExploding {
			continue
		}

		if tank.Image == nil {
			continue
		}
		imageID, err := tank.Image.GetImageID()
		if err != nil {
			continue
		}

		imageData, err := r.explosionTilesUseCases.GetImage(imageID)
		if err != nil {
			continue
		}

		img := r.getCachedImage(imageID, imageData)

		op := &ebiten.DrawImageOptions{}

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

	hq := r.hqUseCases.GetHQ()
	if hq != nil && hq.State == types.HQStateExploding && hq.Image != nil {

		imageID, err := hq.Image.GetImageID()
		if err == nil {

			imageData, err := r.explosionTilesUseCases.GetImage(imageID)
			if err == nil {

				img := r.getCachedImage(imageID, imageData)

				op := &ebiten.DrawImageOptions{}

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

func (r *StageRendererAdapter) drawBullets(screen *ebiten.Image) {
	bullets := r.bulletUseCases.GetBullets()

	for _, bullet := range bullets {
		if bullet == nil || bullet.Image == nil {
			continue
		}

		imageID, err := bullet.Image.GetImageID()
		if err != nil {
			continue
		}

		imageData, err := r.bulletTilesUseCases.GetImage(imageID)
		if err != nil {
			continue
		}

		img := ebiten.NewImageFromImage(imageData)

		rotationAngle := getRotationAngle(bullet.Direction)
		rotatedImage := rotateImageByAngle(img, rotationAngle)

		screenX := float64(r.mapOffsetX) + bullet.Position.X
		screenY := float64(r.mapOffsetY) + bullet.Position.Y

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(screenX, screenY)

		screen.DrawImage(rotatedImage, op)
	}
}

func (r *StageRendererAdapter) drawBonuses(screen *ebiten.Image) {
	if r.bonusesRepository == nil || r.bonusTilesUseCases == nil {
		return
	}

	bonuses := r.bonusesRepository.GetAllBonuses()
	for _, bonus := range bonuses {
		if bonus == nil || bonus.GetImage() == nil {
			continue
		}

		// Проверяем видимость перед отрисовкой
		if !bonus.GetBlinkFlag() {
			continue
		}

		imageID, err := bonus.GetImage().GetImageID()
		if err != nil {
			continue
		}

		imageData, err := r.bonusTilesUseCases.GetImage(imageID)
		if err != nil {
			continue
		}

		img := ebiten.NewImageFromImage(imageData)

		position := bonus.GetPosition()
		screenX := float64(r.mapOffsetX) + position.X
		screenY := float64(r.mapOffsetY) + position.Y

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(screenX, screenY)

		screen.DrawImage(img, op)
	}
}

func (r *StageRendererAdapter) DrawAll(screen *ebiten.Image) {
	r.drawScreenBackground(screen)

	r.drawMapBackground(screen)

	// Блоки-подложки рисуются до танков и пуль: пуля летит НАД водой,
	// а не под ней; поверх всех только лес (AIR)
	r.drawBlocksByAltitude(screen, types.GROUND)

	r.drawBlocksByAltitude(screen, types.SURFACE)

	r.drawHeadquarters(screen)

	r.drawTanks(screen)

	r.drawBullets(screen)

	r.drawBonuses(screen)

	r.drawExplosions(screen)

	r.drawBlocksByAltitude(screen, types.AIR)
}

// Раскладка боковой панели HUD (как в NES Battle City): колонка справа
// от поля, все координаты на сетке 8px
const (
	hudEnemyIconSize    = 8
	hudEnemyGridColumns = 2
	hudEnemyGridRows    = 10
	hudPanelMargin      = 8 // отступ панели от правого края поля
	hudEnemyGridY       = 16
	hudLivesP1LabelY    = 128
	hudLivesP2LabelY    = 152
	hudFlagY            = 176
)

// DrawSidebar рисует боковую панель: резерв врагов, жизни игроков
// и флаг с номером уровня. Блок второго игрока — только в режиме
// двух игроков
func (r *StageRendererAdapter) DrawSidebar(
	screen *ebiten.Image,
	hud types.StageHUDData,
) {
	panelX := r.mapOffsetX + r.mapWidthHeight + hudPanelMargin

	r.drawSidebarEnemyReserve(screen, panelX, hud.EnemiesForSpawn)

	r.drawSidebarLivesBlock(
		screen,
		panelX,
		hudLivesP1LabelY,
		"roman_one",
		hud.Player1Lives,
	)
	if hud.PlayerCount > 1 {
		r.drawSidebarLivesBlock(
			screen,
			panelX,
			hudLivesP2LabelY,
			"roman_two",
			hud.Player2Lives,
		)
	}

	r.drawSidebarStageFlag(screen, panelX, hudFlagY)
}

// drawSidebarEnemyReserve рисует сетку иконок ещё не заспавнившихся врагов
func (r *StageRendererAdapter) drawSidebarEnemyReserve(
	screen *ebiten.Image,
	panelX int,
	enemiesForSpawn uint,
) {
	offsets := r.hudUseCases.EnemyIconOffsets(
		enemiesForSpawn,
		hudEnemyGridColumns,
		hudEnemyGridRows,
		hudEnemyIconSize,
	)
	for _, offset := range offsets {
		r.drawHUDImage(
			screen,
			"enemy_icon",
			float64(panelX)+offset.X,
			hudEnemyGridY+offset.Y,
		)
	}
}

// drawSidebarLivesBlock рисует блок жизней игрока: метку (номер + P),
// иконку танка и число оставшихся жизней
func (r *StageRendererAdapter) drawSidebarLivesBlock(
	screen *ebiten.Image,
	panelX int,
	labelY int,
	romanImageID string,
	lives uint,
) {
	r.drawHUDImage(screen, romanImageID, float64(panelX), float64(labelY))
	r.drawHUDImage(
		screen,
		"letter_p",
		float64(panelX+hudEnemyIconSize),
		float64(labelY),
	)
	r.drawHUDImage(
		screen,
		"life_icon",
		float64(panelX),
		float64(labelY+hudEnemyIconSize),
	)
	r.drawHUDText(
		screen,
		fmt.Sprintf("%d", lives),
		float64(panelX+hudEnemyIconSize),
		float64(labelY+hudEnemyIconSize),
	)
}

// drawSidebarStageFlag рисует флаг уровня (четыре четверти 8x8)
// и номер уровня под ним
func (r *StageRendererAdapter) drawSidebarStageFlag(
	screen *ebiten.Image,
	panelX int,
	flagY int,
) {
	quadrants := []struct {
		imageID string
		dx      int
		dy      int
	}{
		{"flag_tl", 0, 0},
		{"flag_tr", hudEnemyIconSize, 0},
		{"flag_bl", 0, hudEnemyIconSize},
		{"flag_br", hudEnemyIconSize, hudEnemyIconSize},
	}
	for _, q := range quadrants {
		r.drawHUDImage(
			screen,
			q.imageID,
			float64(panelX+q.dx),
			float64(flagY+q.dy),
		)
	}

	r.drawHUDText(
		screen,
		fmt.Sprintf("%d", r.stageNumber),
		float64(panelX+hudEnemyIconSize),
		float64(flagY+2*hudEnemyIconSize),
	)
}

func (r *StageRendererAdapter) drawHUDImage(
	screen *ebiten.Image,
	imageID string,
	x float64,
	y float64,
) {
	imageData, err := r.hudTilesUseCases.GetImage(imageID)
	if err != nil {
		return
	}

	img := r.getCachedImage(imageID, imageData)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	screen.DrawImage(img, op)
}

func (r *StageRendererAdapter) drawHUDText(
	screen *ebiten.Image,
	message string,
	x float64,
	y float64,
) {
	if r.hudFontFace == nil {
		return
	}

	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(color.Black)
	text.Draw(screen, message, r.hudFontFace, op)
}

func (r *StageRendererAdapter) DrawPauseOverlay(screen *ebiten.Image) {
	r.drawOverlayMessage(screen, "PAUSED", "")
}

func (r *StageRendererAdapter) DrawStageEndOverlay(
	screen *ebiten.Image,
	message string,
) {
	r.drawOverlayMessage(screen, message, "press any key to continue")
}

func (r *StageRendererAdapter) drawOverlayMessage(
	screen *ebiten.Image,
	message string,
	subtitle string,
) {
	bounds := screen.Bounds()
	width := float32(bounds.Dx())
	height := float32(bounds.Dy())

	vector.FillRect(
		screen,
		0,
		0,
		width,
		height,
		color.NRGBA{R: 40, G: 40, B: 40, A: 240},
		false,
	)

	face := r.fontFace
	if face == nil {
		return
	}

	textWidth, textHeight := text.Measure(message, face, 0)
	x := (float64(bounds.Dx()) - textWidth) / 2
	y := (float64(bounds.Dy()) - textHeight) / 2

	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(color.White)

	text.Draw(screen, message, face, op)

	if subtitle == "" {
		return
	}

	scale := float64(r.subtitleFontSize) / float64(r.titleFontSize)
	if scale <= 0 {
		return
	}

	subtitleWidth, _ := text.Measure(subtitle, face, 0)
	scaledWidth := subtitleWidth * scale
	subtitleX := (float64(bounds.Dx()) - scaledWidth) / 2
	subtitleY := float64(bounds.Dy()) - float64(r.titleFontSize)/2

	subtitleOp := &text.DrawOptions{}
	subtitleOp.GeoM.Scale(scale, scale)
	subtitleOp.GeoM.Translate(subtitleX, subtitleY)
	subtitleOp.ColorScale.ScaleWithColor(color.White)

	text.Draw(screen, subtitle, face, subtitleOp)
}

func (r *StageRendererAdapter) drawScreenBackground(screen *ebiten.Image) {
	// Серый рамки NES — совпадает с фоном, выключенным в спрайтах HUD
	vector.FillRect(
		screen,
		0,
		0,
		float32(screen.Bounds().Dx()),
		float32(screen.Bounds().Dy()),
		color.RGBA{R: 109, G: 109, B: 109, A: 255},
		false,
	)
}

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

func (r *StageRendererAdapter) drawBlocksByAltitude(
	screen *ebiten.Image,
	altitude types.Altitude,
) {
	blocks := r.mapUseCases.GetBlocks()
	for _, block := range blocks {

		if block.Altitude != altitude {
			continue
		}

		imageID, err := block.Image.GetImageID()
		if err != nil {
			continue
		}

		imageData, err := r.mapTilesUseCases.GetImage(imageID)
		if err != nil {
			continue
		}

		img := ebiten.NewImageFromImage(imageData)

		op := &ebiten.DrawImageOptions{}

		op.GeoM.Translate(
			float64(r.mapOffsetX)+block.Position.X,
			float64(r.mapOffsetY)+block.Position.Y,
		)
		screen.DrawImage(img, op)
	}
}
