package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/shpaker/gonflict/internal"
)

func main() {
	// Загружаем конфигурацию
	cfg, err := internal.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Создаем приложение
	application := internal.New(cfg)

	// Создаем контекст с возможностью отмены
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	// Запускаем приложение
	if err := application.Run(ctx); err != nil {
		os.Exit(1)
	}
}
