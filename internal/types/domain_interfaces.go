package types

type IImageProvider interface {
	GetImageID() (string, error)
}

type IEntityCollider interface {
	GetSize() Size
	GetPosition() Position
	GetAltitude() Altitude
}
