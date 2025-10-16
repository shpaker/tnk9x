package services

import (
	"github.com/shpaker/gonflict/internal/models"
)

type WorldService struct {
	Level models.Level
}

func NewWorldService(level models.Level) WorldService {
	return WorldService{Level: level}
}
