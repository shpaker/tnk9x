package types

// IImageProvider определяет интерфейс для получения ID изображения
type IImageProvider interface {
	GetImageID() (string, error)
}

// IMapObject определяет интерфейс для объектов карты
type IMapObject interface {
	GetSize() Size
	GetPosition() Position
	GetAltitude() Altitude
}
