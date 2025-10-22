package services

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type RendererService struct {
	// Здесь можно добавить поля для кэширования изображений, шрифтов и т.д.
}

func NewRendererService() *RendererService {
	return &RendererService{}
}

// DrawDebugInfo отрисовывает отладочную информацию
func (s *RendererService) DrawDebugInfo(screen *ebiten.Image, info string) {
	// Базовая реализация для отрисовки отладочной информации
	// В будущем здесь можно добавить более сложную логику отрисовки
}

// DrawUI отрисовывает пользовательский интерфейс
func (s *RendererService) DrawUI(screen *ebiten.Image) {
	// Базовая реализация для отрисовки UI
	// В будущем здесь можно добавить отрисовку меню, счетчиков и т.д.
}
