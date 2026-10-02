package types

type IImageProvider interface {
	GetImageID() (string, error)
}

type IEntityCollider interface {
	GetSize() Size
	GetPosition() Position
	GetAltitude() Altitude
}

// BlinkPhaseTicks — длительность одной фазы мигания в тиках
const BlinkPhaseTicks = 10

// IBlink — мигающая сущность: фаза переключается каждые BlinkPhaseTicks
type IBlink interface {
	UpdateBlink()
	GetBlinkFlag() bool
	// GetBlinkProgress — пройденная доля текущей фазы мигания, [0, 1)
	GetBlinkProgress() float64
}
