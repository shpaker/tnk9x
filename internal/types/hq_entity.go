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
	State    HQState
}

// GetImageID возвращает ID изображения базы
func (h *HQEntity) GetImageID() (string, error) {
	if h.State == HQStateDestroyed {
		return "hq_destroyed", nil
	}
	if h.State == HQStateExploding {
		// Во время взрыва используем изображение целой базы (анимация взрыва будет поверх)
		return "hq_intact", nil
	}
	return "hq_intact", nil
}

// GetSize возвращает размер базы
func (h *HQEntity) GetSize() Size {
	return Size{Width: 16, Height: 16}
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
	return SURFACE
}

// IsDestroyed возвращает true если база разрушена
func (h *HQEntity) IsDestroyed() bool {
	return h.State == HQStateDestroyed
}
