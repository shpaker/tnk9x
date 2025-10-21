package interfaces

import "github.com/shpaker/gonflict/internal/models"

// ILevelsDataService определяет интерфейс для работы с уровнями
type ILevelsDataService interface {
	// GetLevel загружает уровень по номеру и возвращает его данные
	GetLevel(num int) (models.Level, error)
}
