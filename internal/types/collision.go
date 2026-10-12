package types

import "github.com/shpaker/obstenku"

type CollisionResponse uint8

const (
	CollisionIgnore CollisionResponse = iota
	CollisionTouch
	CollisionSlide
	CollisionCross
)

type CollisionFilter func(mover, other IEntityCollider) CollisionResponse

type CollisionEntity struct {
	Other      IEntityCollider
	Response   CollisionResponse
	Time       float64
	Touch      Position
	Overlapped bool
}

type IMover interface {
	IEntityCollider
	GetPrevPosition() Position
}

func RectOf(position Position, size Size) obstenku.Rect {
	return obstenku.Rect{
		X: position.X,
		Y: position.Y,
		W: float64(size.Width),
		H: float64(size.Height),
	}
}

// RectsOverlap reports whether rectangles intersect; touching edges do not count.
func RectsOverlap(aPos Position, aSize Size, bPos Position, bSize Size) bool {
	return RectOf(aPos, aSize).Overlaps(RectOf(bPos, bSize))
}

func Overlaps(a, b IEntityCollider) bool {
	return RectsOverlap(
		a.GetPosition(),
		a.GetSize(),
		b.GetPosition(),
		b.GetSize(),
	)
}

// RectContains: top and left edges belong to the rectangle, bottom and right do not.
func RectContains(position Position, size Size, point Position) bool {
	return RectOf(
		position,
		size,
	).ContainsPoint(obstenku.Vec2{X: point.X, Y: point.Y})
}
