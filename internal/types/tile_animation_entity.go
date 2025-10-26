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

// GetImageId возвращает ID изображения текущего кадра (реализует IImageIdGetter)
func (tae *TileAnimationEntity) GetImageId() (string, error) {
	if len(tae.AnimationFrames) == 0 {
		return "", errors.New("no animation frames available")
	}

	if int(tae.CurrentFrame) >= len(tae.AnimationFrames) {
		return "", errors.New("current frame index out of bounds")
	}

	return tae.AnimationFrames[tae.CurrentFrame].Image, nil
}

// StartAnimation запускает анимацию
func (tae *TileAnimationEntity) StartAnimation() {
	tae.IsAnimating = true
	// Если у анимации есть repeats, сбрасываем счетчик при каждом запуске
	if tae.LoopCount != nil && *tae.LoopCount <= 0 {
		// Восстанавливаем оригинальное значение repeats из конфигурации
		// Но мы не можем это сделать здесь, так как не храним оригинальное значение
		// Это будет обработано на уровне Use Cases при пересоздании анимации
	}
}

// StopAnimation останавливает анимацию
func (tae *TileAnimationEntity) StopAnimation() {
	tae.IsAnimating = false
	tae.CurrentFrame = 0 // Сбрасываем на первый кадр
	tae.CurrentTick = 0
}

// UpdateAnimation обновляет анимацию на основе тиков
func (tae *TileAnimationEntity) UpdateAnimation() {
	if len(tae.AnimationFrames) == 0 || !tae.IsAnimating {
		return
	}

	tae.CurrentTick++

	// Проверяем, нужно ли переключить кадр
	if int(tae.CurrentFrame) < len(tae.AnimationFrames) {
		currentFrameDuration := tae.AnimationFrames[tae.CurrentFrame].Duration
		if tae.CurrentTick >= uint(currentFrameDuration) {
			// Переключаем на следующий кадр
			nextFrame := (tae.CurrentFrame + 1) % uint(len(tae.AnimationFrames))

			// Если закончили один цикл (вернулись к началу)
			if nextFrame == 0 && tae.LoopCount != nil {
				loopsLeft := *tae.LoopCount
				loopsLeft--
				tae.LoopCount = &loopsLeft

				if loopsLeft <= 0 {
					// Анимация закончилась
					tae.IsAnimating = false
					return
				}
			}

			tae.CurrentFrame = nextFrame
			tae.CurrentTick = 0
		}
	}
}

// NewTileAnimationEntity создает новый экземпляр TileAnimationEntity с бесконечными циклами
func NewTileAnimationEntity(animationFrames AnimationData) *TileAnimationEntity {
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
func NewTileAnimationEntityWithLoops(animationFrames AnimationData, loops int) *TileAnimationEntity {
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
func NewTileAnimationEntityWithOffset(animationFrames AnimationData, offset [2]float64) *TileAnimationEntity {
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
func NewTileAnimationEntityWithLoopsAndOffset(animationFrames AnimationData, loops int, offset [2]float64) *TileAnimationEntity {
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
