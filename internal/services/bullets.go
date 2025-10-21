package services

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/shpaker/gonflict/internal/constants"
	"github.com/shpaker/gonflict/internal/models"
	"github.com/shpaker/gonflict/internal/types"
)

type BulletsService struct {
	bullets []models.Bullet
}

func NewBulletsService() *BulletsService {
	return &BulletsService{
		bullets: make([]models.Bullet, 0),
	}
}

func (s *BulletsService) AddBullet(tank *models.Tank) {
	// Создаем простую пулю (квадрат 4x4 пикселя)
	bulletImage := ebiten.NewImage(4, 4)
	bulletImage.Fill(color.RGBA{255, 255, 0, 255}) // Желтый цвет

	// Вычисляем позицию пули в зависимости от направления танка
	bulletX := tank.WorldPosition.X + constants.TankSpriteSize/2 - 2
	bulletY := tank.WorldPosition.Y + constants.TankSpriteSize/2 - 2

	// Корректируем позицию в зависимости от направления
	switch tank.Direction {
	case types.DirectionUp:
		bulletY = tank.WorldPosition.Y - 4
	case types.DirectionDown:
		bulletY = tank.WorldPosition.Y + constants.TankSpriteSize
	case types.DirectionLeft:
		bulletX = tank.WorldPosition.X - 4
	case types.DirectionRight:
		bulletX = tank.WorldPosition.X + constants.TankSpriteSize
	}

	bullet := models.Bullet{
		Image: bulletImage,
		WorldPosition: types.Position{
			X: bulletX,
			Y: bulletY,
		},
		Speed:     200.0, // Скорость пули
		Direction: tank.Direction,
		Owner:     tank,
	}

	s.bullets = append(s.bullets, bullet)
}

func (s *BulletsService) Update(dt float64) {
	// Обновляем позиции всех пуль
	for i := len(s.bullets) - 1; i >= 0; i-- {
		bullet := &s.bullets[i]

		// Вычисляем новую позицию
		delta := bullet.Speed * dt
		switch bullet.Direction {
		case types.DirectionUp:
			bullet.WorldPosition.Y -= delta
		case types.DirectionDown:
			bullet.WorldPosition.Y += delta
		case types.DirectionLeft:
			bullet.WorldPosition.X -= delta
		case types.DirectionRight:
			bullet.WorldPosition.X += delta
		}

		// Удаляем пули, которые вышли за границы экрана
		if bullet.WorldPosition.X < 0 || bullet.WorldPosition.X > constants.BattleFieldWidthHeight ||
			bullet.WorldPosition.Y < 0 || bullet.WorldPosition.Y > constants.BattleFieldWidthHeight {
			s.bullets = append(s.bullets[:i], s.bullets[i+1:]...)
		}
	}
}

func (s *BulletsService) Draw(screen *ebiten.Image) {
	for _, bullet := range s.bullets {
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

func (s *BulletsService) GetBullets() []models.Bullet {
	return s.bullets
}

func (s *BulletsService) RemoveBullet(index int) {
	if index >= 0 && index < len(s.bullets) {
		s.bullets = append(s.bullets[:index], s.bullets[index+1:]...)
	}
}
