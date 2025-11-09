package types

// HQState представляет состояние базы
type HQState int

const (
	HQStateIntact    HQState = iota // База цела
	HQStateExploding                // База взрывается
	HQStateDestroyed                // База разрушена
)

// HQEntity представляет базу (headquarters)
type HQEntity struct {
	Position Position
	Size     Size
	Altitude Altitude
	Image    IImageProvider
	State    HQState
}

// GetSize возвращает размер базы
func (h *HQEntity) GetSize() Size {
	if h.Size.Width == 0 && h.Size.Height == 0 {
		return Size{Width: 16, Height: 16}
	}
	return h.Size
}

// GetPosition возвращает позицию базы в мире
func (h *HQEntity) GetPosition() Position {
	return h.Position
}

// GetAltitude возвращает высоту базы
func (h *HQEntity) GetAltitude() Altitude {
	// Если база взрывается, показываем выше всего
	if h.State == HQStateExploding {
		return AIR
	}
	// Разрушенная база на уровне поверхности (как танки)
	if h.State == HQStateDestroyed {
		return SURFACE
	}
	if h.Altitude == 0 {
		return SURFACE
	}
	return h.Altitude
}

// IsDestroyed возвращает true если база разрушена
func (h *HQEntity) IsDestroyed() bool {
	return h.State == HQStateDestroyed
}
