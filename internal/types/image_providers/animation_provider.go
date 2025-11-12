package image_providers

import (
	"errors"

	"github.com/shpaker/gonflict/internal/types"
)

type AnimationProvider struct {
	CurrentFrame    uint
	AnimationFrames types.AnimationData
	CurrentTick     uint
	IsAnimating     bool
	LoopCount       *int
	Offset          [2]float64
}

func (ap *AnimationProvider) GetImageID() (string, error) {
	if len(ap.AnimationFrames) == 0 {
		return "", errors.New("no animation frames available")
	}

	if int(ap.CurrentFrame) >= len(ap.AnimationFrames) {
		return "", errors.New("current frame index out of bounds")
	}

	return ap.AnimationFrames[ap.CurrentFrame].Image, nil
}

func NewAnimationProvider(
	animationFrames types.AnimationData,
) *AnimationProvider {
	return &AnimationProvider{
		CurrentFrame:    0,
		AnimationFrames: animationFrames,
		CurrentTick:     0,
		IsAnimating:     false,
		LoopCount:       nil,
		Offset:          [2]float64{0, 0},
	}
}

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
		Offset:          [2]float64{0, 0},
	}
}

func NewAnimationProviderWithOffset(
	animationFrames types.AnimationData,
	offset [2]float64,
) *AnimationProvider {
	return &AnimationProvider{
		CurrentFrame:    0,
		AnimationFrames: animationFrames,
		CurrentTick:     0,
		IsAnimating:     false,
		LoopCount:       nil,
		Offset:          offset,
	}
}

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

func (ap *AnimationProvider) IsFinished() bool {
	return !ap.IsAnimating
}
