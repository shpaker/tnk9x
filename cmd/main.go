package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/shpaker/gonflict/internal/app"
	"github.com/shpaker/gonflict/internal/config"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Создаем приложение
	application := app.New(cfg)

	// Создаем контекст с возможностью отмены
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Запускаем приложение
	if err := application.Run(ctx); err != nil {
		log.Fatalf("Application failed: %v", err)
	}

	log.Println("Application stopped gracefully")
}
