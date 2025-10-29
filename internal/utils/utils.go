package utils

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/gonflict/internal/types"
)

// RotateImage поворачивает изображение в зависимости от направления
func RotateImage(image *ebiten.Image, direction types.Direction) *ebiten.Image {
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

	rotatedImage, err := RotateImageByAngle(image, angle)
	if err != nil {
		return image
	}
	return rotatedImage
}

// RotateImageByAngle поворачивает изображение на указанный угол
func RotateImageByAngle(
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

func RoundToEven(
	number float64,
	up bool,
) int {
	return RoundToDivisible(number, 2, up)
}

// RoundToDivisible округляет число до ближайшего числа, делимого на divisor
// up: true - округление в большую сторону, false - в меньшую
func RoundToDivisible(
	number float64,
	divisor int,
	up bool,
) int {
	if divisor <= 0 {
		return int(
			number,
		) // Возвращаем исходное число, если делитель некорректный
	}

	// Сначала округляем до ближайшего целого
	rounded := int(math.Round(number))

	// Вычисляем остаток от деления
	remainder := rounded % divisor

	// Если остаток равен 0, число уже делится нацело
	if remainder == 0 {
		return rounded
	}

	// Вычисляем ближайшие числа, делимые на divisor
	lower := rounded - remainder
	upper := rounded + (divisor - remainder)

	// Выбираем направление округления
	if up {
		return upper
	}
	return lower
}

// RoundToNearestMultipleOf4 округляет координату до ближайшего кратного 4
func RoundToNearestMultipleOf4(value float64) float64 {
	rounded := math.Round(value)
	nearestMultiple := math.Round(rounded/4) * 4

	// Если разница меньше 0.5, округляем до кратного 4
	if math.Abs(rounded-nearestMultiple) < 0.5 {
		return nearestMultiple
	}

	return rounded
}
