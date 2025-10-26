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

// AnimationDataFrame представляет кадр анимации
type AnimationDataFrame struct {
	Image    string `yaml:"image"`
	Duration int    `yaml:"duration"`
}

// AnimationConfig представляет конфигурацию анимации в новом формате
type AnimationConfig struct {
	Duration int      `yaml:"duration"`
	Frames   []string `yaml:"frames"`
}

// AnimationData представляет анимацию (старый формат - массив кадров)
type AnimationData []AnimationDataFrame

// TilesetDataConfig представляет конфигурацию тайлсета
type TilesetDataConfig struct {
	Size       int                        `yaml:"size"`
	Images     map[string][2]int          `yaml:"images"`
	Animations map[string]AnimationConfig `yaml:"animations"`
}
