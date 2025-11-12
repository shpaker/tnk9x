package services

import (
	image_providers "github.com/shpaker/gonflict/internal/types/image_providers"
)

type AnimationService struct{}

func NewAnimationService() *AnimationService {
	return &AnimationService{}
}

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

	if !s.ShouldAdvanceFrame(animation) {
		return
	}

	nextFrame := s.CalculateNextFrame(animation)

	if s.CheckAndHandleLoopCompletion(animation, nextFrame) {
		return
	}

	animation.CurrentFrame = nextFrame
	animation.CurrentTick = 0
}

func (s *AnimationService) ShouldAdvanceFrame(
	animation *image_providers.AnimationProvider,
) bool {
	if int(animation.CurrentFrame) >= len(animation.AnimationFrames) {
		return false
	}
	currentFrameDuration := animation.AnimationFrames[animation.CurrentFrame].Duration
	return animation.CurrentTick >= uint(currentFrameDuration)
}

func (s *AnimationService) CalculateNextFrame(
	animation *image_providers.AnimationProvider,
) uint {
	return (animation.CurrentFrame + 1) % uint(len(animation.AnimationFrames))
}

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
