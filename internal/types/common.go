package types

import "math"

type (
	BlockType string
	Direction int
)

const (
	DirectionUp    Direction = 0
	DirectionDown  Direction = 1
	DirectionLeft  Direction = 2
	DirectionRight Direction = 3
)

// Vector возвращает единичный вектор направления в экранных
// координатах (ось Y вниз)
func (d Direction) Vector() Position {
	switch d {
	case DirectionDown:
		return Position{X: 0, Y: 1}
	case DirectionLeft:
		return Position{X: -1, Y: 0}
	case DirectionRight:
		return Position{X: 1, Y: 0}
	default:
		return Position{X: 0, Y: -1}
	}
}

// Angle возвращает угол направления в радианах: atan2 вектора Vector
func (d Direction) Angle() float64 {
	v := d.Vector()
	return math.Atan2(v.Y, v.X)
}

const (
	Brick  BlockType = "brick"
	Steel  BlockType = "steel"
	Forest BlockType = "forest"
	Water  BlockType = "water"
	Ice    BlockType = "ice"
)

type Altitude int

const (
	GROUND  Altitude = 0
	SURFACE Altitude = 1
	AIR     Altitude = 2
)

type Position struct {
	X float64
	Y float64
}

type Size struct {
	Width  int
	Height int
}
