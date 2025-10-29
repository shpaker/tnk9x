package types

import "errors"

// TileAnimationEntity представляет анимированный тайл
type TileAnimationEntity struct {
	CurrentFrame    uint
	AnimationFrames AnimationData
	CurrentTick     uint
	IsAnimating     bool
	LoopCount       *int
	Offset          [2]float64 // Смещение анимации относительно сущности [x, y]
}

// GetImageID возвращает ID изображения текущего кадра (реализует IImageIDGetter)
func (tae *TileAnimationEntity) GetImageID() (string, error) {
	if len(tae.AnimationFrames) == 0 {
		return "", errors.New("no animation frames available")
	}

	if int(tae.CurrentFrame) >= len(tae.AnimationFrames) {
		return "", errors.New("current frame index out of bounds")
	}

	return tae.AnimationFrames[tae.CurrentFrame].Image, nil
}

// NewTileAnimationEntity создает новый экземпляр TileAnimationEntity с бесконечными циклами
func NewTileAnimationEntity(
	animationFrames AnimationData,
) *TileAnimationEntity {
	return &TileAnimationEntity{
		CurrentFrame:    0,
		AnimationFrames: animationFrames,
		CurrentTick:     0,
		IsAnimating:     false,
		LoopCount:       nil,              // Бесконечно
		Offset:          [2]float64{0, 0}, // Без смещения
	}
}

// NewTileAnimationEntityWithLoops создает анимацию с заданным количеством циклов
func NewTileAnimationEntityWithLoops(
	animationFrames AnimationData,
	loops int,
) *TileAnimationEntity {
	return &TileAnimationEntity{
		CurrentFrame:    0,
		AnimationFrames: animationFrames,
		CurrentTick:     0,
		IsAnimating:     false,
		LoopCount:       &loops,
		Offset:          [2]float64{0, 0}, // Без смещения
	}
}

// NewTileAnimationEntityWithOffset создает анимацию с offset
func NewTileAnimationEntityWithOffset(
	animationFrames AnimationData,
	offset [2]float64,
) *TileAnimationEntity {
	return &TileAnimationEntity{
		CurrentFrame:    0,
		AnimationFrames: animationFrames,
		CurrentTick:     0,
		IsAnimating:     false,
		LoopCount:       nil, // Бесконечно
		Offset:          offset,
	}
}

// NewTileAnimationEntityWithLoopsAndOffset создает анимацию с количеством циклов и offset
func NewTileAnimationEntityWithLoopsAndOffset(
	animationFrames AnimationData,
	loops int,
	offset [2]float64,
) *TileAnimationEntity {
	return &TileAnimationEntity{
		CurrentFrame:    0,
		AnimationFrames: animationFrames,
		CurrentTick:     0,
		IsAnimating:     false,
		LoopCount:       &loops,
		Offset:          offset,
	}
}

// IsFinished проверяет, завершилась ли анимация
func (tae *TileAnimationEntity) IsFinished() bool {
	return !tae.IsAnimating
}
