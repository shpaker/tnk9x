package use_cases

import (
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
	image_providers "github.com/shpaker/gonflict/internal/types/image_providers"
)

// HQUseCases отвечает за обработку действий базы и рендеринг
type HQUseCases struct {
	hq                     *types.HQEntity
	bulletUseCases         interfaces.IBulletUseCases
	bulletCollisionService interfaces.IBulletCollisionService
	tilesUseCases          *TilesUseCases
	AnimationGetter        types.IImageProvider // Публичное поле для рендеринга
}

// NewHQUseCases создает новый экземпляр HQUseCases
func NewHQUseCases(
	hq *types.HQEntity,
	bulletUseCases interfaces.IBulletUseCases,
	bulletCollisionService interfaces.IBulletCollisionService,
	tilesUseCases *TilesUseCases,
) *HQUseCases {
	return &HQUseCases{
		hq:                     hq,
		bulletUseCases:         bulletUseCases,
		bulletCollisionService: bulletCollisionService,
		tilesUseCases:          tilesUseCases,
		AnimationGetter:        nil,
	}
}

// HandleBulletHit обрабатывает попадание пули в базу
// Возвращает индексы пуль для удаления и true если база была уничтожена
func (uc *HQUseCases) HandleBulletHit() (bulletIndicesToRemove []int, destroyed bool) {
	if uc.hq == nil || uc.hq.IsDestroyed() ||
		uc.hq.State == types.HQStateExploding {
		return nil, false
	}

	bullets := uc.bulletUseCases.GetBullets()
	bulletIndicesToRemove, destroyed = uc.bulletCollisionService.CheckBulletHQCollision(
		bullets,
		uc.hq,
	)

	// Запускаем анимацию взрыва если нужно
	if destroyed {
		_ = uc.Explode()
	}

	return bulletIndicesToRemove, destroyed
}

// Explode запускает анимацию взрыва базы
func (uc *HQUseCases) Explode() error {
	if uc.hq == nil || uc.hq.State == types.HQStateExploding ||
		uc.hq.IsDestroyed() {
		return nil
	}

	explosionAnim, err := uc.tilesUseCases.CreateExplosionAnimation()
	if err != nil {
		return err
	}

	uc.AnimationGetter = explosionAnim
	uc.hq.State = types.HQStateExploding

	uc.tilesUseCases.StartAnimation(explosionAnim)
	return nil
}

// IsExplosionAnimationFinished возвращает true если анимация взрыва завершена
func (uc *HQUseCases) IsExplosionAnimationFinished() bool {
	if uc.AnimationGetter == nil {
		return true
	}

	// Проверяем, что анимация завершена
	if tileAnim, ok := uc.AnimationGetter.(*image_providers.AnimationProvider); ok {
		return tileAnim.IsFinished()
	}

	return true
}

// IsExplosionFinished проверяет завершение анимации взрыва базы
func (uc *HQUseCases) IsExplosionFinished() {
	if uc.hq != nil && uc.hq.State == types.HQStateExploding {
		if uc.IsExplosionAnimationFinished() {
			uc.hq.State = types.HQStateDestroyed
		}
	}
}
