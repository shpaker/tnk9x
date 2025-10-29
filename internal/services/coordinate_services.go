package services

import "math"

// CoordinateService предоставляет логику работы с координатами
type CoordinateService struct{}

// NewCoordinateService создает новый сервис координат
func NewCoordinateService() *CoordinateService {
	return &CoordinateService{}
}

// RoundToNearestMultipleOf4 округляет координату до ближайшего кратного 4
func (s *CoordinateService) RoundToNearestMultipleOf4(value float64) float64 {
	rounded := math.Round(value)
	nearestMultiple := math.Round(rounded/4) * 4

	// Если разница меньше 0.5, округляем до кратного 4
	if math.Abs(rounded-nearestMultiple) < 0.5 {
		return nearestMultiple
	}

	return rounded
}
