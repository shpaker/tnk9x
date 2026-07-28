package stage

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// rotateImageByAngle возвращает копию изображения, повёрнутую вокруг центра.
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
