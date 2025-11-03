package states

import "github.com/hajimehoshi/ebiten/v2"

// State определяет интерфейс для состояний игры
type State interface {
	Update()
	Draw(screen *ebiten.Image)
}
