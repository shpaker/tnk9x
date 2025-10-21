package utils

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/shpaker/gonflict/internal/interfaces"
)

// FormatGreeting форматирует приветственное сообщение
func FormatGreeting(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}

// IsEmpty проверяет, является ли строка пустой
func IsEmpty(s string) bool {
	return s == ""
}

// TruncateString обрезает строку до указанной длины
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

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

// CheckColliders проверяет коллизию между двумя объектами карты
func CheckColliders(
	obj1 interfaces.IMapObject,
	obj2 interfaces.IMapObject,
) bool {
	pos1 := obj1.GetWorldPosition()
	size1 := obj1.GetSize()
	pos2 := obj2.GetWorldPosition()
	size2 := obj2.GetSize()

	// Проверяем пересечение прямоугольников
	return pos1.X < pos2.X+float64(size2.Width) &&
		pos1.X+float64(size1.Width) > pos2.X &&
		pos1.Y < pos2.Y+float64(size2.Height) &&
		pos1.Y+float64(size1.Height) > pos2.Y
}

// CheckColliders проверяет коллизии между объектом и массивом объектов карты
func CheckCollidersWithArray(
	obj interfaces.IMapObject,
	objects []interfaces.IMapObject,
) []interfaces.IMapObject {
	var collidingObjects []interfaces.IMapObject

	for _, mapObj := range objects {
		if CheckColliders(obj, mapObj) {
			collidingObjects = append(collidingObjects, mapObj)
		}
	}

	return collidingObjects
}

// CheckCollidersWithArrayFirst проверяет коллизии между объектом и массивом объектов карты
// Возвращает первый коллидирующий объект или nil, если коллизий нет
func CheckCollidersWithArrayFirst(
	obj interfaces.IMapObject,
	objects []interfaces.IMapObject,
) interfaces.IMapObject {
	for _, mapObj := range objects {
		if CheckColliders(obj, mapObj) {
			return mapObj
		}
	}
	return nil
}
