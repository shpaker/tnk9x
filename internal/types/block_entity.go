package types

import "errors"

type BlockEntity struct {
	Position Position
	Size     Size
	Altitude Altitude
	Image    IImageProvider
	Data     *BlockData
}

type BlockData struct {
	Name     BlockType
	Position Position
}

func (b *BlockEntity) GetImageID() (string, error) {
	if b.Image == nil {
		return "", errors.New("image is nil")
	}
	return b.Image.GetImageID()
}

func (b *BlockEntity) GetSize() Size {
	if b.Size.Width == 0 && b.Size.Height == 0 {
		return Size{
			Width:  8,
			Height: 8,
		}
	}
	return b.Size
}

func (b *BlockEntity) GetPosition() Position {
	return b.Position
}

func (b *BlockEntity) GetAltitude() Altitude {
	return b.Altitude
}

func NewBlockEntity(
	blockType string,
	positionX,
	positionY float64,
	size int,
	imageGetter IImageProvider,
) *BlockEntity {
	altitude := SURFACE

	if blockType == string(Forest) {
		altitude = AIR
	}

	return &BlockEntity{
		Position: Position{X: positionX, Y: positionY},
		Size:     Size{Width: size, Height: size},
		Altitude: altitude,
		Image:    imageGetter,
		Data: &BlockData{
			Name:     BlockType(blockType),
			Position: Position{X: positionX, Y: positionY},
		},
	}
}
