package image_providers

import (
	"errors"

	"github.com/shpaker/gonflict/internal/types"
)

// AnimationProvider представляет провайдер анимированных изображений
type AnimationProvider struct {
	CurrentFrame    uint
	AnimationFrames types.AnimationData
	CurrentTick     uint
	IsAnimating     bool
	LoopCount       *int
	Offset          [2]float64 // Смещение анимации относительно сущности [x, y]
}

// GetImageID возвращает ID изображения текущего кадра (реализует IImageProvider)
func (ap *AnimationProvider) GetImageID() (string, error) {
	if len(ap.AnimationFrames) == 0 {
		return "", errors.New("no animation frames available")
	}

	if int(ap.CurrentFrame) >= len(ap.AnimationFrames) {
		return "", errors.New("current frame index out of bounds")
	}

	return ap.AnimationFrames[ap.CurrentFrame].Image, nil
}

// NewAnimationProvider создает новый экземпляр AnimationProvider с бесконечными циклами
func NewAnimationProvider(
	animationFrames types.AnimationData,
) *AnimationProvider {
	return &AnimationProvider{
		CurrentFrame:    0,
		AnimationFrames: animationFrames,
		CurrentTick:     0,
		IsAnimating:     false,
		LoopCount:       nil,              // Бесконечно
		Offset:          [2]float64{0, 0}, // Без смещения
	}
}

// NewAnimationProviderWithLoops создает анимацию с заданным количеством циклов
func NewAnimationProviderWithLoops(
	animationFrames types.AnimationData,
	loops int,
) *AnimationProvider {
	return &AnimationProvider{
		CurrentFrame:    0,
		AnimationFrames: animationFrames,
		CurrentTick:     0,
		IsAnimating:     false,
		LoopCount:       &loops,
		Offset:          [2]float64{0, 0}, // Без смещения
	}
}

// NewAnimationProviderWithOffset создает анимацию с offset
func NewAnimationProviderWithOffset(
	animationFrames types.AnimationData,
	offset [2]float64,
) *AnimationProvider {
	return &AnimationProvider{
		CurrentFrame:    0,
		AnimationFrames: animationFrames,
		CurrentTick:     0,
		IsAnimating:     false,
		LoopCount:       nil, // Бесконечно
		Offset:          offset,
	}
}

// NewAnimationProviderWithLoopsAndOffset создает анимацию с количеством циклов и offset
func NewAnimationProviderWithLoopsAndOffset(
	animationFrames types.AnimationData,
	loops int,
	offset [2]float64,
) *AnimationProvider {
	return &AnimationProvider{
		CurrentFrame:    0,
		AnimationFrames: animationFrames,
		CurrentTick:     0,
		IsAnimating:     false,
		LoopCount:       &loops,
		Offset:          offset,
	}
}

// IsFinished проверяет, завершилась ли анимация
func (ap *AnimationProvider) IsFinished() bool {
	return !ap.IsAnimating
}
