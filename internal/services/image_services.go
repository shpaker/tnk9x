package services

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/gonflict/internal/types"
)

type ImageService struct{}

func NewImageService() *ImageService {
	return &ImageService{}
}

func (s *ImageService) RotateImage(
	image interface{},
	direction types.Direction,
) interface{} {
	img, ok := image.(*ebiten.Image)
	if !ok {
		return image
	}
	var angle float64
	switch direction {
	case types.DirectionUp:
		angle = 0
	case types.DirectionRight:
		angle = math.Pi / 2
	case types.DirectionDown:
		angle = math.Pi
	case types.DirectionLeft:
		angle = -math.Pi / 2
	default:
		angle = 0
	}

	rotatedImage, err := s.RotateImageByAngle(img, angle)
	if err != nil {
		return img
	}
	return rotatedImage
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
