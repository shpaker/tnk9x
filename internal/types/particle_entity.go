package types

import (
	"image/color"
	"math"
)

// ParticleEntity — частица эффекта (искра, осколок, дым, пыль)
// в координатах игрового поля
type ParticleEntity struct {
	Position Position
	Velocity Position // Пикселей за тик
	Color    color.NRGBA
	Size     float64 // Сторона квадрата в пикселях
	Life     uint    // Оставшиеся тики
	MaxLife  uint
	Drag     float64 // Доля скорости, сохраняемая за тик
}

// Fade возвращает долю оставшейся жизни частицы от 1 до 0
func (p *ParticleEntity) Fade() float64 {
	if p.MaxLife == 0 {
		return 0
	}
	return float64(p.Life) / float64(p.MaxLife)
}

// FlashEntity — короткая вспышка света (выстрел, удар по стене)
type FlashEntity struct {
	Position  Position
	Radius    float64
	Color     color.NRGBA
	Intensity float64
	Life      uint // Оставшиеся тики
	MaxLife   uint
}

// Fade возвращает долю оставшейся жизни вспышки от 1 до 0
func (f *FlashEntity) Fade() float64 {
	if f.MaxLife == 0 {
		return 0
	}
	return float64(f.Life) / float64(f.MaxLife)
}

// ScreenShakeEntity — тряска экрана: «травма» копится от взрывов
// и затухает, смещение растёт как квадрат травмы
type ScreenShakeEntity struct {
	trauma float64
	ticks  uint
}

// AddTrauma добавляет силу тряски, не выше 1
func (s *ScreenShakeEntity) AddTrauma(amount float64) {
	s.trauma = math.Min(1, s.trauma+amount)
}

// Decay ослабляет тряску за один тик и продвигает её фазу
func (s *ScreenShakeEntity) Decay(amount float64) {
	s.trauma = math.Max(0, s.trauma-amount)
	s.ticks++
}

func (s *ScreenShakeEntity) GetTrauma() float64 {
	return s.trauma
}

// GetTicks возвращает фазу тряски — число прошедших тиков
func (s *ScreenShakeEntity) GetTicks() uint {
	return s.ticks
}
