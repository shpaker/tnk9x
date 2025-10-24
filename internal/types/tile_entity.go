package types

// TileStaticEntity представляет статический тайл с изображением
type TileStaticEntity struct {
	ImageId string
}

// GetImageId возвращает ID изображения тайла
func (tse *TileStaticEntity) GetImageId() string {
	return tse.ImageId
}

// TileAnimationEntity представляет анимированный тайл
type TileAnimationEntity struct {
	CurrentFrame    uint
	AnimationFrames AnimationData
	CurrentTick     uint // Текущий тик для анимации
	IsAnimating     bool // Флаг активности анимации
}

// GetImageId возвращает ID изображения текущего кадра (реализует IImageIdGetter)
func (tae *TileAnimationEntity) GetImageId() string {
	if len(tae.AnimationFrames) == 0 {
		return ""
	}

	if int(tae.CurrentFrame) >= len(tae.AnimationFrames) {
		return ""
	}

	return tae.AnimationFrames[tae.CurrentFrame].Image
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
			tae.CurrentFrame = (tae.CurrentFrame + 1) % uint(len(tae.AnimationFrames))
			tae.CurrentTick = 0
		}
	}
}

// StartAnimation запускает анимацию
func (tae *TileAnimationEntity) StartAnimation() {
	tae.IsAnimating = true
}

// StopAnimation останавливает анимацию
func (tae *TileAnimationEntity) StopAnimation() {
	tae.IsAnimating = false
	tae.CurrentFrame = 0 // Сбрасываем на первый кадр
	tae.CurrentTick = 0
}

// NewTileAnimationEntity создает новый экземпляр TileAnimationEntity
func NewTileAnimationEntity(animationFrames AnimationData) *TileAnimationEntity {
	return &TileAnimationEntity{
		CurrentFrame:    0,
		AnimationFrames: animationFrames,
		CurrentTick:     0,
		IsAnimating:     false, // По умолчанию анимация остановлена
	}
}
