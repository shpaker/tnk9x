package states

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/shpaker/gonflict/internal/interfaces"
)

type GameState struct {
	spritesRepository interfaces.ISpritesRepository
}

// NewGameState создает новое состояние игры с переданным репозиторием спрайтов
func NewGameState(spritesRepository interfaces.ISpritesRepository) *GameState {
	return &GameState{
		spritesRepository: spritesRepository,
	}
}

func (state *GameState) Update() (interfaces.State, error) {
	return nil, nil
}

func (state *GameState) Draw(screen *ebiten.Image) {
	// Заливаем фон черным
	screen.Fill(color.RGBA{0, 0, 0, 255})

	// Рисуем фиолетовый квадрат для тестирования
	purpleCol := color.RGBA{255, 0, 255, 255}
	for x := 100; x < 200; x++ {
		for y := 100; y < 200; y++ {
			screen.Set(x, y, purpleCol)
		}
	}

	if sprite, err := state.spritesRepository.GetSprite("blocks", "brick"); err == nil {
		// Получаем размер спрайта для отладки
		// bounds := sprite.Bounds()
		// fmt.Printf("Спрайт brick загружен, размер: %dx%d\n", bounds.Dx(), bounds.Dy())

		// Создаем опции для отрисовки
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(8, 8)
		op.GeoM.Translate(100, 100)
		screen.DrawImage(sprite, op)
	}
}
