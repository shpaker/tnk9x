package services

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type ImageService struct{}

func NewImageService() *ImageService {
	return &ImageService{}
}

func (s *ImageService) RotateImageByAngle(
	image interface{},
	angle float64,
) (interface{}, error) {
	img, ok := image.(*ebiten.Image)
	if !ok {
		return image, nil
	}

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

	return rotatedImage, nil
}
