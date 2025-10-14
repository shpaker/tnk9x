package states

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/shpaker/gonflict/internal/interfaces"
)

type GameState struct {
}

func (state *GameState) Update() (interfaces.State, error) {
	return nil, nil
}

func (state *GameState) Draw(screen *ebiten.Image) {
	// screen.DrawImage()
	purpleCol := color.RGBA{255, 0, 255, 255}
	for x := 100; x < 200; x++ {
		for y := 100; y < 200; y++ {
			screen.Set(x, y, purpleCol)
		}
	}
}
