package types

import "github.com/hajimehoshi/ebiten/v2"

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

// Tank представляет танк (игрока или врага)
type Tank struct {
	Image         *ebiten.Image
	SpawnPosition Position
	WorldPosition Position
	Speed         float64
	Direction     Direction
}

// GetSize возвращает размер танка
func (t *Tank) GetSize() Size {
	return Size{Width: 16, Height: 16} // Стандартный размер танка
}

// GetWorldPosition возвращает позицию танка в мире
func (t *Tank) GetWorldPosition() Position {
	return t.WorldPosition
}

// GetScreenPosition возвращает позицию танка на экране
func (t *Tank) GetScreenPosition() Position {
	return t.WorldPosition
}

// Bullet представляет пулю
type Bullet struct {
	Image         *ebiten.Image
	WorldPosition Position
	Speed         float64
	Direction     Direction
	Owner         *Tank
}

// GetSize возвращает размер пули
func (b *Bullet) GetSize() Size {
	return Size{Width: 4, Height: 4}
}

// GetWorldPosition возвращает позицию пули в мире
func (b *Bullet) GetWorldPosition() Position {
	return b.WorldPosition
}

// GetScreenPosition возвращает позицию пули на экране
func (b *Bullet) GetScreenPosition() Position {
	return b.WorldPosition
}

// Block представляет блок карты
type Block struct {
	Image         *ebiten.Image
	Data          *BlockData
	Properties    *BlockProperties
	WorldPosition Position
}

// BlockData содержит данные блока
type BlockData struct {
	Name     BlockType
	Position Position
}

// BlockProperties содержит свойства блока
type BlockProperties struct {
	Collidable bool
}

// GetSize возвращает размер блока
func (b *Block) GetSize() Size {
	return Size{Width: 8, Height: 8} // Стандартный размер блока (TileMinSize)
}

// GetWorldPosition возвращает позицию блока в мире
func (b *Block) GetWorldPosition() Position {
	return b.WorldPosition
}

// GetScreenPosition возвращает позицию блока на экране
func (b *Block) GetScreenPosition() Position {
	return b.WorldPosition
}

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
