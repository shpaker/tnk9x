package services

import "math"

type CoordinateService struct{}

func NewCoordinateService() *CoordinateService {
	return &CoordinateService{}
}

func (s *CoordinateService) RoundToNearestMultipleOf4(value float64) float64 {
	rounded := math.Round(value)
	nearestMultiple := math.Round(rounded/4) * 4

	if math.Abs(rounded-nearestMultiple) < 0.5 {
		return nearestMultiple
	}

	return rounded
}
