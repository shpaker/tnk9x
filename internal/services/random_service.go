package services

import (
	"math/rand/v2"

	"github.com/shpaker/tnk9x/internal/interfaces"
)

var _ interfaces.IRandomService = (*RandomService)(nil)

// RandomService — случайные числа из глобального генератора math/rand/v2
type RandomService struct{}

func NewRandomService() *RandomService {
	return &RandomService{}
}

func (s *RandomService) Float64() float64 {
	return rand.Float64()
}
