package states

import "github.com/hajimehoshi/ebiten/v2"

// State определяет интерфейс для состояний игры
type State interface {
	Update() (State, error)
	Draw(screen *ebiten.Image)
}
