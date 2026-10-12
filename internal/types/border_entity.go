package types

type BorderEntity struct {
	Position Position
	Size     Size
}

func (b *BorderEntity) GetSize() Size         { return b.Size }
func (b *BorderEntity) GetPosition() Position { return b.Position }

func (b *BorderEntity) GetAltitude() Altitude { return SURFACE }

// Borders are as thick as the map itself, so no speed can tunnel through them.
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
