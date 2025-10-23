package utils

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// RotateImage создает DrawImageOptions для поворота изображения вокруг его центра в указанной позиции
func RotateImage(
	image *ebiten.Image,
	angle float64,
	x, y float64,
) *ebiten.DrawImageOptions {
	op := &ebiten.DrawImageOptions{}

	// Получаем размеры изображения
	bounds := image.Bounds()
	centerX := float64(bounds.Dx()) / 2
	centerY := float64(bounds.Dy()) / 2

	// Перемещаем центр изображения в (0,0)
	op.GeoM.Translate(-centerX, -centerY)
	// op.GeoM.Translate(0, 0)

	// Поворачиваем изображение
	op.GeoM.Rotate(angle)

	// Перемещаем в нужную позицию
	op.GeoM.Translate(x+centerX, y+centerY)

	return op
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
		return int(number) // Возвращаем исходное число, если делитель некорректный
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
