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
	mapUseCases        interfaces.IMapUseCases
	tankCommonUseCases interfaces.ITankCommonUseCases
	bulletUseCases     interfaces.IBulletUseCases
	hqUseCases         interfaces.IHQUseCases
	hudUseCases        interfaces.IHUDUseCases
	renderUseCases     interfaces.IRenderUseCases
	bonusUseCases      interfaces.IBonusUseCases
	spriteCache        *SpriteCache
	fontFace           text.Face
	hudFontFace        text.Face
	titleFontSize      int
	subtitleFontSize   int
	regularFontSize    int
	mapOffsetX         int
	mapOffsetY         int
	mapWidthHeight     int

	// Последняя отрисовка меню паузы — для хит-тестов тапов;
	// до первого кадра меню зоны неизвестны
	lastWidth      float64
	lastHeight     float64
	pauseMenuItems []types.PauseMenuItem
}

// StageRendererDependencies — готовый граф зависимостей рендера уровня;
// собирается composition root'ом, все поля обязательны
type StageRendererDependencies struct {
	// Use Cases
	MapUseCases        interfaces.IMapUseCases
	TankCommonUseCases interfaces.ITankCommonUseCases
	BulletUseCases     interfaces.IBulletUseCases
	HQUseCases         interfaces.IHQUseCases
	HUDUseCases        interfaces.IHUDUseCases
	RenderUseCases     interfaces.IRenderUseCases
	BonusUseCases      interfaces.IBonusUseCases

	// Кэш GPU-спрайтов, общий для всех уровней
	SpriteCache *SpriteCache

	// Шрифты и раскладка
	FontFace         text.Face
	HUDFontFace      text.Face
	MapOffsetX       int
	MapOffsetY       int
	MapWidthHeight   int
	TitleFontSize    int
	SubtitleFontSize int
	RegularFontSize  int
}

func NewStageRendererAdapter(
	deps StageRendererDependencies,
) *StageRendererAdapter {
	return &StageRendererAdapter{
		mapUseCases:        deps.MapUseCases,
		tankCommonUseCases: deps.TankCommonUseCases,
		bulletUseCases:     deps.BulletUseCases,
		hqUseCases:         deps.HQUseCases,
		hudUseCases:        deps.HUDUseCases,
		renderUseCases:     deps.RenderUseCases,
		bonusUseCases:      deps.BonusUseCases,
		spriteCache:        deps.SpriteCache,
		fontFace:           deps.FontFace,
		hudFontFace:        deps.HUDFontFace,
		mapOffsetX:         deps.MapOffsetX,
		mapOffsetY:         deps.MapOffsetY,
		mapWidthHeight:     deps.MapWidthHeight,
		titleFontSize:      deps.TitleFontSize,
		subtitleFontSize:   deps.SubtitleFontSize,
		regularFontSize:    deps.RegularFontSize,
	}
}

// drawSprite отрисовывает спрайт тайлсета в экранных координатах
func (r *StageRendererAdapter) drawSprite(
	screen *ebiten.Image,
	tilesetType types.TilesetType,
	imageID string,
	x float64,
	y float64,
) {
	img, err := r.spriteCache.Image(tilesetType, imageID)
	if err != nil {
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	screen.DrawImage(img, op)
}

// drawEntitySprite отрисовывает спрайт сущности со смещением игрового поля
func (r *StageRendererAdapter) drawEntitySprite(
	screen *ebiten.Image,
	tilesetType types.TilesetType,
	imageID string,
	position types.Position,
) {
	r.drawSprite(
		screen,
		tilesetType,
		imageID,
		float64(r.mapOffsetX)+position.X,
		float64(r.mapOffsetY)+position.Y,
	)
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

		if !r.renderUseCases.IsTankVisible(tank) {
			continue
		}

		if tank.Image == nil {
			continue
		}
		imageID, err := tank.Image.GetImageID()
		if err != nil {
			continue
		}

		r.drawEntitySprite(
			screen,
			types.TankTilesetType(tank.IsEnemy()),
			imageID,
			tank.Position,
		)

		if overlayColor, ok := r.renderUseCases.TankHealthOverlay(tank); ok {
			r.drawTankHealthOverlay(screen, tank, overlayColor)
		}
	}
}

// drawTankHealthOverlay отрисовывает полупрозрачный слой поверх танка
// цветом, выбранным use case'ом по его здоровью
func (r *StageRendererAdapter) drawTankHealthOverlay(
	screen *ebiten.Image,
	tank *types.TankEntity,
	overlayColor color.NRGBA,
) {
	vector.FillRect(
		screen,
		float32(r.mapOffsetX)+float32(tank.Position.X),
		float32(r.mapOffsetY)+float32(tank.Position.Y),
		float32(tank.Size.Width),
		float32(tank.Size.Height),
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

	r.drawEntitySprite(
		screen,
		types.TilesetTypeSpawner,
		imageID,
		tank.Position,
	)
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
	if hq == nil || hq.State == types.HQStateExploding || hq.Image == nil {
		return
	}

	imageID, err := hq.Image.GetImageID()
	if err != nil {
		return
	}

	r.drawEntitySprite(screen, types.TilesetTypeHQ, imageID, hq.Position)
}

// drawExplosion отрисовывает кадр взрыва с учётом смещения анимации
func (r *StageRendererAdapter) drawExplosion(
	screen *ebiten.Image,
	provider types.IImageProvider,
	position types.Position,
) {
	imageID, err := provider.GetImageID()
	if err != nil {
		return
	}

	if tileAnim, ok := provider.(*image_providers.AnimationProvider); ok {
		position.X += tileAnim.Offset[0]
		position.Y += tileAnim.Offset[1]
	}

	r.drawEntitySprite(screen, types.TilesetTypeExplosion, imageID, position)
}

func (r *StageRendererAdapter) drawExplosions(screen *ebiten.Image) {
	allTanks := r.tankCommonUseCases.GetAllTanks()
	for _, tank := range allTanks {
		if tank == nil || tank.State != types.TankStateExploding ||
			tank.Image == nil {
			continue
		}
		r.drawExplosion(screen, tank.Image, tank.Position)
	}

	hq := r.hqUseCases.GetHQ()
	if hq != nil && hq.State == types.HQStateExploding && hq.Image != nil {
		r.drawExplosion(screen, hq.Image, hq.Position)
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

		img, err := r.spriteCache.RotatedImage(
			types.TilesetTypeBullet,
			imageID,
			getRotationAngle(bullet.Direction),
		)
		if err != nil {
			continue
		}

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(
			float64(r.mapOffsetX)+bullet.Position.X,
			float64(r.mapOffsetY)+bullet.Position.Y,
		)

		screen.DrawImage(img, op)
	}
}

func (r *StageRendererAdapter) drawBonuses(screen *ebiten.Image) {
	for _, bonus := range r.bonusUseCases.VisibleBonuses() {
		if bonus.GetImage() == nil {
			continue
		}

		imageID, err := bonus.GetImage().GetImageID()
		if err != nil {
			continue
		}

		r.drawEntitySprite(
			screen,
			types.TilesetTypeBonuses,
			imageID,
			bonus.GetPosition(),
		)
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

	r.drawSidebarStageFlag(screen, panelX, hudFlagY, hud.StageNumber)
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
	stageNumber uint,
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
		fmt.Sprintf("%d", stageNumber),
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
	r.drawSprite(screen, types.TilesetTypeHUD, imageID, x, y)
}

func (r *StageRendererAdapter) drawHUDText(
	screen *ebiten.Image,
	message string,
	x float64,
	y float64,
) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(color.Black)
	text.Draw(screen, message, r.hudFontFace, op)
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
		overlayBackdropColor,
		false,
	)

	face := r.fontFace

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

// Цвета экрана уровня
var (
	// nesBorderGray — серый рамки NES, совпадает с фоном спрайтов HUD
	nesBorderGray = color.RGBA{R: 109, G: 109, B: 109, A: 255}
	// overlayBackdropColor — полупрозрачная подложка оверлеев паузы
	// и конца уровня
	overlayBackdropColor = color.NRGBA{R: 40, G: 40, B: 40, A: 240}
)

func (r *StageRendererAdapter) drawScreenBackground(screen *ebiten.Image) {
	vector.FillRect(
		screen,
		0,
		0,
		float32(screen.Bounds().Dx()),
		float32(screen.Bounds().Dy()),
		nesBorderGray,
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

		r.drawBlockSprite(screen, imageID, block)
	}
}

// drawBlockSprite отрисовывает блок; у сколотого пулями остатка
// кирпича берётся соответствующий фрагмент спрайта исходного тайла
func (r *StageRendererAdapter) drawBlockSprite(
	screen *ebiten.Image,
	imageID string,
	block *types.BlockEntity,
) {
	img, err := r.spriteCache.Image(types.TilesetTypeBlocks, imageID)
	if err != nil {
		return
	}

	size := block.GetSize()
	bounds := img.Bounds()
	if block.Data != nil &&
		(size.Width != bounds.Dx() || size.Height != bounds.Dy()) {
		// Data.Position хранит origin исходного тайла: смещение
		// остатка относительно него задаёт фрагмент спрайта
		offsetX := int(block.Position.X - block.Data.Position.X)
		offsetY := int(block.Position.Y - block.Data.Position.Y)
		srcRect := image.Rect(
			bounds.Min.X+offsetX,
			bounds.Min.Y+offsetY,
			bounds.Min.X+offsetX+size.Width,
			bounds.Min.Y+offsetY+size.Height,
		)
		img, _ = img.SubImage(srcRect).(*ebiten.Image)
		if img == nil {
			return
		}
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(
		float64(r.mapOffsetX)+block.Position.X,
		float64(r.mapOffsetY)+block.Position.Y,
	)
	screen.DrawImage(img, op)
}
