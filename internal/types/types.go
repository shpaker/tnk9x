package types

type BlockType string
type Direction string

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

// SpriteData содержит данные о спрайте
type SpriteData struct {
	X int `yaml:"x"`
	Y int `yaml:"y"`
	W int `yaml:"w"`
	H int `yaml:"h"`
}

// SpritesConfig представляет конфигурацию спрайтов
type SpritesConfig map[string]map[string]SpriteData

const (
	DirectionUp    Direction = "up"
	DirectionDown  Direction = "down"
	DirectionLeft  Direction = "left"
	DirectionRight Direction = "right"
)

const (
	Brick  BlockType = "brick"
	Steel  BlockType = "steel"
	Forest BlockType = "forest"
	Water  BlockType = "water"
	Ice    BlockType = "ice"
)
