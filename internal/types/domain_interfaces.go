package types

// IImageProvider определяет интерфейс для получения ID изображения
type IImageProvider interface {
	GetImageID() (string, error)
}

// IEntityCollider определяет интерфейс для сущностей, которые могут участвовать в коллизиях
type IEntityCollider interface {
	GetSize() Size
	GetPosition() Position
	GetAltitude() Altitude
}
