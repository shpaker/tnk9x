package interfaces

import "github.com/shpaker/gonflict/internal/models"

// IPlayerService определяет интерфейс для работы с игроком
type IPlayerService interface {
	// GetPlayer возвращает данные игрока с начальными параметрами
	GetPlayer() (models.Tank, error)
	Update(dt float32)
}
