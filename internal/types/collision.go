package types

import "github.com/shpaker/obstenku"

// CollisionResponse — как движущаяся сущность встречает другую:
// не замечает, упирается, скользит вдоль или проходит насквозь
type CollisionResponse uint8

const (
	// CollisionIgnore — сущности не встречаются
	CollisionIgnore CollisionResponse = iota
	// CollisionTouch — сущность останавливается в точке касания
	CollisionTouch
	// CollisionSlide — сущность упирается по оси удара и продолжает
	// движение по другой оси
	CollisionSlide
	// CollisionCross — сущность проходит насквозь, встреча сообщается
	CollisionCross
)

// CollisionFilter решает, как mover встречает other
type CollisionFilter func(mover, other IEntityCollider) CollisionResponse

// CollisionEntity — встреча на пути движущейся сущности
type CollisionEntity struct {
	Other    IEntityCollider
	Response CollisionResponse
	// Time — доля пути до встречи, [0, 1]
	Time float64
	// Touch — позиция движущейся сущности в момент встречи
	Touch Position
	// Overlapped — сущности перекрывались ещё до движения
	Overlapped bool
}

// IMover — сущность, которая движется за тик: мир коллизий проводит её
// от позиции на начало тика к текущей
type IMover interface {
	IEntityCollider
	// GetPrevPosition — позиция на начало тика
	GetPrevPosition() Position
}

// RectOf — прямоугольник сущности для геометрии коллизий
func RectOf(position Position, size Size) obstenku.Rect {
	return obstenku.Rect{
		X: position.X,
		Y: position.Y,
		W: float64(size.Width),
		H: float64(size.Height),
	}
}

// RectsOverlap — пересекаются ли прямоугольники; касание гранью
// пересечением не считается
func RectsOverlap(aPos Position, aSize Size, bPos Position, bSize Size) bool {
	return RectOf(aPos, aSize).Overlaps(RectOf(bPos, bSize))
}

// Overlaps — пересекаются ли сущности на плоскости, без учёта высоты
func Overlaps(a, b IEntityCollider) bool {
	return RectsOverlap(
		a.GetPosition(),
		a.GetSize(),
		b.GetPosition(),
		b.GetSize(),
	)
}

// RectContains — лежит ли точка в прямоугольнике: верхняя и левая
// грани принадлежат ему, нижняя и правая — нет
func RectContains(position Position, size Size, point Position) bool {
	return RectOf(
		position,
		size,
	).ContainsPoint(obstenku.Vec2{X: point.X, Y: point.Y})
}
