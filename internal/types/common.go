package types

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
