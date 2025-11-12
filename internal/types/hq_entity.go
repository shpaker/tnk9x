package types

type HQState int

const (
	HQStateIntact HQState = iota
	HQStateExploding
	HQStateDestroyed
)

type HQEntity struct {
	Position Position
	Size     Size
	Altitude Altitude
	Image    IImageProvider
	State    HQState
}

func (h *HQEntity) GetSize() Size {
	if h.Size.Width == 0 && h.Size.Height == 0 {
		return Size{Width: 16, Height: 16}
	}
	return h.Size
}

func (h *HQEntity) GetPosition() Position {
	return h.Position
}

func (h *HQEntity) GetAltitude() Altitude {
	if h.State == HQStateExploding {
		return AIR
	}

	if h.State == HQStateDestroyed {
		return SURFACE
	}
	if h.Altitude == 0 {
		return SURFACE
	}
	return h.Altitude
}

func (h *HQEntity) IsDestroyed() bool {
	return h.State == HQStateDestroyed
}
