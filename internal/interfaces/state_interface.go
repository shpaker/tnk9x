package interfaces

import "github.com/hajimehoshi/ebiten/v2"

type State interface {
	Update() (State, error)
	Draw(screen *ebiten.Image)
}
