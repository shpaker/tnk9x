package services

import "math"

type CoordinateService struct{}

func NewCoordinateService() *CoordinateService {
	return &CoordinateService{}
}

func (s *CoordinateService) RoundToNearestMultipleOf4(value float64) float64 {
	return math.Round(value/4) * 4
}
