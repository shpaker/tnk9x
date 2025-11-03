package types

// Типы
type (
	BlockType string
	Direction int
)

// Константы направлений
const (
	DirectionUp    Direction = 0
	DirectionDown  Direction = 1
	DirectionLeft  Direction = 2
	DirectionRight Direction = 3
)

// Константы типов блоков
const (
	Brick  BlockType = "brick"
	Steel  BlockType = "steel"
	Forest BlockType = "forest"
	Water  BlockType = "water"
	Ice    BlockType = "ice"
)

// Altitude представляет высоту объекта
type Altitude int

// Константы уровней высоты
const (
	GROUND  Altitude = 0
	SURFACE Altitude = 1
	AIR     Altitude = 2
)

// Position представляет координаты в 2D пространстве
type Position struct {
	X float64
	Y float64
}

// Size представляет размеры объекта
type Size struct {
	Width  int
	Height int
}
