package use_cases

import (
	"github.com/shpaker/gonflict/internal/types"
	image_providers "github.com/shpaker/gonflict/internal/types/image_providers"
)

// HQUseCases отвечает за обработку действий базы и рендеринг
type HQUseCases struct {
	tilesUseCases *TilesUseCases
}

// NewHQUseCases создает новый экземпляр HQUseCases
func NewHQUseCases(
	tilesUseCases *TilesUseCases,
) *HQUseCases {
	return &HQUseCases{
		tilesUseCases: tilesUseCases,
	}
}

// Explode запускает анимацию взрыва базы
func (uc *HQUseCases) Explode(hq *types.HQEntity) error {
	if hq == nil || hq.State == types.HQStateExploding ||
		hq.IsDestroyed() {
		return nil
	}

	explosionAnim, err := uc.tilesUseCases.CreateExplosionAnimation()
	if err != nil {
		return err
	}

	// Сохраняем анимацию в entity
	hq.Image = explosionAnim
	hq.State = types.HQStateExploding

	uc.tilesUseCases.StartAnimation(explosionAnim)
	return nil
}

// IsExplosionAnimationFinished возвращает true если анимация взрыва завершена
func (uc *HQUseCases) IsExplosionAnimationFinished(hq *types.HQEntity) bool {
	if hq == nil || hq.Image == nil {
		return true
	}

	// Проверяем, что анимация завершена
	if tileAnim, ok := hq.Image.(*image_providers.AnimationProvider); ok {
		return tileAnim.IsFinished()
	}

	return true
}

// IsExplosionFinished проверяет завершение анимации взрыва базы
func (uc *HQUseCases) IsExplosionFinished(hq *types.HQEntity) {
	if hq != nil && hq.State == types.HQStateExploding {
		if uc.IsExplosionAnimationFinished(hq) {
			// Создаем статическое изображение разрушенной базы
			destroyedImage, err := uc.tilesUseCases.CreateStaticTile(
				"hq_destroyed",
			)
			if err == nil {
				hq.Image = destroyedImage
			}
			// Устанавливаем состояние на разрушенное и высоту на SURFACE (как у танков)
			hq.State = types.HQStateDestroyed
			hq.Altitude = types.SURFACE
		}
	}
}
