package services

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/shpaker/gonflict/internal/constants"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/utils"
)

type RendererService struct {
	battleFieldService MapService
	playerOneService   *PlayerService
	bulletsService     *BulletsService
	collidersService   *CollidersService
}

func NewRendererService(
	battleFieldService MapService,
	playerOneService *PlayerService,
	bulletsService *BulletsService,
	collidersService *CollidersService,
) *RendererService {
	return &RendererService{
		battleFieldService: battleFieldService,
		playerOneService:   playerOneService,
		bulletsService:     bulletsService,
		collidersService:   collidersService,
	}
}

func (s *RendererService) drawBattlefield(screen *ebiten.Image) {
	vector.FillRect(
		screen,
		float32(constants.BattleFieldOffset),
		float32(constants.BattleFieldOffset),
		float32(constants.BattleFieldWidthHeight),
		float32(constants.BattleFieldWidthHeight),
		color.Black,
		false,
	)

	// Draw blocks on the battle field
	for _, block := range s.battleFieldService.Level {
		if block.Image == nil {
			continue
		}
		op := &ebiten.DrawImageOptions{}
		// Предполагаем, что блоки имеют координаты X, Y в WorldPosition
		op.GeoM.Translate(
			constants.BattleFieldOffset+block.WorldPosition.X*constants.TileMinSize,
			constants.BattleFieldOffset+block.WorldPosition.Y*constants.TileMinSize,
		)
		screen.DrawImage(block.Image, op)
	}
}

func (s *RendererService) drawPlayerOne(screen *ebiten.Image) {
	tank, err := s.playerOneService.GetPlayer()
	if err != nil || tank.Image == nil {
		return
	}

	// Определяем угол поворота в зависимости от направления
	var rotationAngle float64
	switch tank.Direction {
	case types.DirectionUp:
		rotationAngle = 0
	case types.DirectionRight:
		rotationAngle = math.Pi / 2
	case types.DirectionDown:
		rotationAngle = math.Pi
	case types.DirectionLeft:
		rotationAngle = 3 * math.Pi / 2
	}

	// Вычисляем позицию на экране
	screenX := constants.BattleFieldOffset + tank.WorldPosition.X
	screenY := constants.BattleFieldOffset + tank.WorldPosition.Y

	// Используем функцию для поворота изображения вокруг центра
	op := utils.RotateImage(tank.Image, rotationAngle, screenX, screenY)

	screen.DrawImage(tank.Image, op)
}

func (s *RendererService) drawBullets(screen *ebiten.Image) {
	bullets := s.bulletsService.GetBullets()
	for _, bullet := range bullets {
		if bullet.Image != nil {
			// Вычисляем позицию на экране
			screenX := constants.BattleFieldOffset + bullet.WorldPosition.X
			screenY := constants.BattleFieldOffset + bullet.WorldPosition.Y

			// Создаем опции для отрисовки
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Translate(screenX, screenY)

			screen.DrawImage(bullet.Image, op)
		}
	}
}

func (s *RendererService) DrawAll(screen *ebiten.Image) {
	s.drawBattlefield(screen)
	s.drawPlayerOne(screen)
	s.drawBullets(screen)
}
