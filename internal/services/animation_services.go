package services

import (
	image_providers "github.com/shpaker/gonflict/internal/types/image_providers"
)

// AnimationService предоставляет логику обновления анимаций
type AnimationService struct{}

// NewAnimationService создает новый сервис анимаций
func NewAnimationService() *AnimationService {
	return &AnimationService{}
}

// UpdateAnimation обновляет анимацию на основе тиков
func (s *AnimationService) UpdateAnimation(
	animation *image_providers.AnimationProvider,
) {
	if animation == nil {
		return
	}

	if len(animation.AnimationFrames) == 0 || !animation.IsAnimating {
		return
	}

	animation.CurrentTick++

	// Проверяем, нужно ли переключить кадр
	if !s.ShouldAdvanceFrame(animation) {
		return
	}

	nextFrame := s.CalculateNextFrame(animation)

	// Проверяем завершение циклов и останавливаем анимацию если нужно
	if s.CheckAndHandleLoopCompletion(animation, nextFrame) {
		return
	}

	animation.CurrentFrame = nextFrame
	animation.CurrentTick = 0
}

// ShouldAdvanceFrame проверяет, нужно ли переключать кадр
func (s *AnimationService) ShouldAdvanceFrame(
	animation *image_providers.AnimationProvider,
) bool {
	if int(animation.CurrentFrame) >= len(animation.AnimationFrames) {
		return false
	}
	currentFrameDuration := animation.AnimationFrames[animation.CurrentFrame].Duration
	return animation.CurrentTick >= uint(currentFrameDuration)
}

// CalculateNextFrame вычисляет следующий кадр анимации
func (s *AnimationService) CalculateNextFrame(
	animation *image_providers.AnimationProvider,
) uint {
	return (animation.CurrentFrame + 1) % uint(len(animation.AnimationFrames))
}

// CheckAndHandleLoopCompletion проверяет завершение цикла и останавливает анимацию если нужно
// Возвращает true, если анимация была остановлена
func (s *AnimationService) CheckAndHandleLoopCompletion(
	animation *image_providers.AnimationProvider,
	nextFrame uint,
) bool {
	if nextFrame == 0 && animation.LoopCount != nil {
		loopsLeft := *animation.LoopCount
		loopsLeft--
		animation.LoopCount = &loopsLeft

		if loopsLeft <= 0 {
			animation.IsAnimating = false
			return true
		}
	}
	return false
}
