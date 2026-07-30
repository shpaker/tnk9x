package types

// EffectEntity — короткоживущий визуальный эффект (взрыв пули):
// проигрывает анимацию в точке и удаляется по её завершении
type EffectEntity struct {
	Position Position
	Size     Size
	Image    IImageProvider
}
