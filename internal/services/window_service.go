package services

import (
	"fmt"
	"log"

	"github.com/shpaker/gonflict/internal/config"
)

// WindowService представляет сервисный слой приложения
type WindowService struct {
	// Здесь будут храниться зависимости сервиса
	// например: repository, external APIs, etc.
	cfg *config.Config
}

// New создает новый сервис
func NewWindowService() *WindowService {
	return &WindowService{}
}

// PrintHelloWorld выполняет основную логику приложения - печатает Hello World
func (s *WindowService) PrintHelloWorld() {
	message := "Hello, World!"
	log.Printf("Service layer: %s", message)
	fmt.Println(message)
}

// func (s *WindowService) DrawWindow() {
// ebiten.SetWindowSize(s.cfg.App.ScreenWidth*2, s.cfg.App.ScreenWidth)
// ebiten.SetWindowTitle(s.cfg.App.Name)

// if err := ebiten.RunGame(&s); err != nil {
// panic(err)
// }

// }
