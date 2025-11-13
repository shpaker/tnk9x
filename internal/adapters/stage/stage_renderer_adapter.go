package adapters

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/opentype"

	"github.com/shpaker/tnk25/internal/interfaces"
	"github.com/shpaker/tnk25/internal/services"
	"github.com/shpaker/tnk25/internal/types"
	image_providers "github.com/shpaker/tnk25/internal/types/image_providers"
	"github.com/shpaker/tnk25/internal/use_cases"
)

type StageRendererAdapter struct {
	mapUseCases            interfaces.IMapUseCases
	tankCommonUseCases     interfaces.ITankCommonUseCases
	tankRenderUseCases     interfaces.ITankRenderUseCases
	bulletUseCases         interfaces.IBulletUseCases
	mapTilesUseCases       *use_cases.TilesUseCases
	tankTilesUseCases      *use_cases.TilesUseCases
	bulletTilesUseCases    *use_cases.TilesUseCases
	spawnerTilesUseCases   *use_cases.TilesUseCases
	explosionTilesUseCases *use_cases.TilesUseCases
	hqTilesUseCases        *use_cases.TilesUseCases
	hqUseCases             interfaces.IHQUseCases
	imageCache             map[string]*ebiten.Image
	imageService           *services.ImageService
	fontUseCases           interfaces.IFontUseCases
	fontFace               text.Face
	titleFontSize          int
	subtitleFontSize       int
	regularFontSize        int
	tileMinSize            int
	mapOffsetX             int
	mapOffsetY             int
	mapWidthHeight         int
}

func NewStageRendererAdapter(
	mapUseCases interfaces.IMapUseCases,
	tankCommonUseCases interfaces.ITankCommonUseCases,
	tankRenderUseCases interfaces.ITankRenderUseCases,
	bulletUseCases interfaces.IBulletUseCases,
	mapTilesUseCases *use_cases.TilesUseCases,
	tankTilesUseCases *use_cases.TilesUseCases,
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
	titleFontSize int,
	subtitleFontSize int,
	regularFontSize int,
) *StageRendererAdapter {
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
		mapUseCases:            mapUseCases,
		tankCommonUseCases:     tankCommonUseCases,
		tankRenderUseCases:     tankRenderUseCases,
		bulletUseCases:         bulletUseCases,
		mapTilesUseCases:       mapTilesUseCases,
		tankTilesUseCases:      tankTilesUseCases,
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
		titleFontSize:          titleFontSize,
		subtitleFontSize:       subtitleFontSize,
		regularFontSize:        regularFontSize,
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

		screenX := float64(r.mapOffsetX) + bullet.Position.X
		screenY := float64(r.mapOffsetY) + bullet.Position.Y

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(screenX, screenY)

		screen.DrawImage(rotatedImage, op)
	}
}

func (r *StageRendererAdapter) DrawAll(screen *ebiten.Image) {
	r.drawScreenBackground(screen)

	r.drawMapBackground(screen)

	r.drawBlocksByAltitude(screen, types.GROUND)

	r.drawHeadquarters(screen)

	r.drawTanks(screen)

	r.drawBullets(screen)

	r.drawBlocksByAltitude(screen, types.SURFACE)

	r.drawExplosions(screen)

	r.drawBlocksByAltitude(screen, types.AIR)
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

	face := r.ensureFontFace()
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

func (r *StageRendererAdapter) ensureFontFace() text.Face {
	if r.fontFace != nil {
		return r.fontFace
	}
	if r.fontUseCases == nil {
		return nil
	}

	baseFont, err := r.fontUseCases.GetFont()
	if err != nil || baseFont == nil {
		return nil
	}

	face, err := opentype.NewFace(baseFont, &opentype.FaceOptions{
		Size: float64(r.titleFontSize),
		DPI:  72,
	})
	if err != nil {
		return nil
	}

	r.fontFace = text.NewGoXFace(face)
	return r.fontFace
}

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
