package types

// IImageIdGetter определяет интерфейс для получения ID изображения
type IImageIdGetter interface {
	GetImageId() (string, error)
}

// Типы
type BlockType string
type Direction string

// Константы направлений
const (
	DirectionUp    Direction = "up"
	DirectionDown  Direction = "down"
	DirectionLeft  Direction = "left"
	DirectionRight Direction = "right"
)

// Константы типов блоков
const (
	Brick  BlockType = "brick"
	Steel  BlockType = "steel"
	Forest BlockType = "forest"
	Water  BlockType = "water"
	Ice    BlockType = "ice"
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

// AnimationDataFrame представляет кадр анимации
type AnimationDataFrame struct {
	Image    string `yaml:"image"`
	Duration int    `yaml:"duration"`
}

// AnimationData представляет анимацию
type AnimationData []AnimationDataFrame

// TilesetDataConfig представляет конфигурацию тайлсета
type TilesetDataConfig struct {
	Size       int                      `yaml:"size"`
	Images     map[string][2]int        `yaml:"images"`
	Animations map[string]AnimationData `yaml:"animations"`
}
