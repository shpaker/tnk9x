package services

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/gonflict/internal/types"
)

// ImageService предоставляет логику работы с изображениями
type ImageService struct{}

// NewImageService создает новый сервис изображений
func NewImageService() *ImageService {
	return &ImageService{}
}

// RotateImage поворачивает изображение в зависимости от направления
func (s *ImageService) RotateImage(
	image *ebiten.Image,
	direction types.Direction,
) *ebiten.Image {
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

	rotatedImage, err := s.RotateImageByAngle(image, angle)
	if err != nil {
		return image
	}
	return rotatedImage
}

// RotateImageByAngle поворачивает изображение на указанный угол
func (s *ImageService) RotateImageByAngle(
	image *ebiten.Image,
	angle float64,
) (*ebiten.Image, error) {
	// Получаем размеры изображения
	bounds := image.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Создаем новое изображение с теми же размерами
	rotatedImage := ebiten.NewImage(width, height)

	// Создаем опции для поворота
	op := &ebiten.DrawImageOptions{}

	// Перемещаем центр изображения в (0,0)
	centerX := float64(width) / 2
	centerY := float64(height) / 2
	op.GeoM.Translate(-centerX, -centerY)

	// Поворачиваем изображение
	op.GeoM.Rotate(angle)

	// Перемещаем обратно в центр
	op.GeoM.Translate(centerX, centerY)

	// Отрисовываем повернутое изображение
	rotatedImage.DrawImage(image, op)

	return rotatedImage, nil
}
