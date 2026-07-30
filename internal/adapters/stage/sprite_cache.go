package stage

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

type spriteKey struct {
	tilesetType types.TilesetType
	imageID     string
}

type rotatedSpriteKey struct {
	key   spriteKey
	angle float64
}

// SpriteCache хранит спрайты, преобразованные в GPU-формат; живёт всё
// время работы приложения и переиспользуется между уровнями
type SpriteCache struct {
	spriteUseCases interfaces.ISpriteUseCases
	images         map[spriteKey]*ebiten.Image
	rotated        map[rotatedSpriteKey]*ebiten.Image
}

func NewSpriteCache(
	spriteUseCases interfaces.ISpriteUseCases,
) *SpriteCache {
	return &SpriteCache{
		spriteUseCases: spriteUseCases,
		images:         make(map[spriteKey]*ebiten.Image),
		rotated:        make(map[rotatedSpriteKey]*ebiten.Image),
	}
}

// Image возвращает GPU-изображение спрайта, лениво пополняя кэш
func (c *SpriteCache) Image(
	tilesetType types.TilesetType,
	imageID string,
) (*ebiten.Image, error) {
	key := spriteKey{tilesetType: tilesetType, imageID: imageID}
	if img, exists := c.images[key]; exists {
		return img, nil
	}

	imageData, err := c.spriteUseCases.GetImage(tilesetType, imageID)
	if err != nil {
		return nil, err
	}

	var img *ebiten.Image
	if imageData.Bounds().Dx() == 0 || imageData.Bounds().Dy() == 0 {
		img = ebiten.NewImage(1, 1)
	} else {
		img = ebiten.NewImageFromImage(imageData)
	}
	c.images[key] = img
	return img, nil
}

// RotatedImage возвращает спрайт, повёрнутый вокруг центра на angle
func (c *SpriteCache) RotatedImage(
	tilesetType types.TilesetType,
	imageID string,
	angle float64,
) (*ebiten.Image, error) {
	if angle == 0 {
		return c.Image(tilesetType, imageID)
	}

	key := rotatedSpriteKey{
		key:   spriteKey{tilesetType: tilesetType, imageID: imageID},
		angle: angle,
	}
	if img, exists := c.rotated[key]; exists {
		return img, nil
	}

	baseImage, err := c.Image(tilesetType, imageID)
	if err != nil {
		return nil, err
	}

	img := rotateImageByAngle(baseImage, angle)
	c.rotated[key] = img
	return img, nil
}

// Preload прогревает кэш всеми спрайтами перечисленных тайлсетов
func (c *SpriteCache) Preload(tilesetTypes []types.TilesetType) {
	for _, tilesetType := range tilesetTypes {
		for _, imageID := range c.spriteUseCases.GetImageIDs(tilesetType) {
			_, _ = c.Image(tilesetType, imageID)
		}
	}
}

// rotateImageByAngle возвращает копию изображения, повёрнутую вокруг центра
func rotateImageByAngle(img *ebiten.Image, angle float64) *ebiten.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	rotatedImage := ebiten.NewImage(width, height)

	op := &ebiten.DrawImageOptions{}

	centerX := float64(width) / 2
	centerY := float64(height) / 2
	op.GeoM.Translate(-centerX, -centerY)

	op.GeoM.Rotate(angle)

	op.GeoM.Translate(centerX, centerY)

	rotatedImage.DrawImage(img, op)

	return rotatedImage
}
