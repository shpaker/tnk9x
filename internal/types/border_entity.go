package types

// BorderEntity — невидимая стена за краем карты: танки упираются
// в неё, пули гаснут
type BorderEntity struct {
	Position Position
	Size     Size
}

func (b *BorderEntity) GetSize() Size         { return b.Size }
func (b *BorderEntity) GetPosition() Position { return b.Position }

// GetAltitude — край карты преграждает путь на любой высоте
func (b *BorderEntity) GetAltitude() Altitude { return SURFACE }

// NewMapBorders окружает карту размера sizePx четырьмя стенами
// толщиной в саму карту: их не перелететь ни на какой скорости
func NewMapBorders(sizePx Size) []*BorderEntity {
	w, h := sizePx.Width, sizePx.Height
	thick := max(w, h)
	return []*BorderEntity{
		{
			Position: Position{X: float64(-thick), Y: float64(-thick)},
			Size:     Size{Width: w + 2*thick, Height: thick},
		},
		{
			Position: Position{X: float64(-thick), Y: float64(h)},
			Size:     Size{Width: w + 2*thick, Height: thick},
		},
		{
			Position: Position{X: float64(-thick), Y: 0},
			Size:     Size{Width: thick, Height: h},
		},
		{
			Position: Position{X: float64(w), Y: 0},
			Size:     Size{Width: thick, Height: h},
		},
	}
}
