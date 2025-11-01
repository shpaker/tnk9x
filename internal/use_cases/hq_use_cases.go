package use_cases

import (
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
)

// HQRenderUseCases отвечает за графику и рендеринг базы
type HQRenderUseCases struct {
	AnimationGetter types.IImageIDGetter
}

// NewHQRenderUseCases создает новый экземпляр HQRenderUseCases
func NewHQRenderUseCases() *HQRenderUseCases {
	return &HQRenderUseCases{
		AnimationGetter: nil,
	}
}

// IsExplosionAnimationFinished возвращает true если анимация взрыва завершена
func (r *HQRenderUseCases) IsExplosionAnimationFinished() bool {
	if r.AnimationGetter == nil {
		return true
	}

	// Проверяем, что анимация завершена
	if tileAnim, ok := r.AnimationGetter.(*types.TileAnimationEntity); ok {
		return tileAnim.IsFinished()
	}

	return true
}

// HQUseCases отвечает за обработку действий базы
type HQUseCases struct {
	hq                     *types.HQEntity
	bulletUseCases         interfaces.IBulletUseCases
	bulletCollisionService interfaces.IBulletCollisionService
	tilesUseCases          *TilesUseCases
	renderUseCases         *HQRenderUseCases
}

// NewHQUseCases создает новый экземпляр HQUseCases
func NewHQUseCases(
	hq *types.HQEntity,
	bulletUseCases interfaces.IBulletUseCases,
	bulletCollisionService interfaces.IBulletCollisionService,
	tilesUseCases *TilesUseCases,
	renderUseCases *HQRenderUseCases,
) *HQUseCases {
	return &HQUseCases{
		hq:                     hq,
		bulletUseCases:         bulletUseCases,
		bulletCollisionService: bulletCollisionService,
		tilesUseCases:          tilesUseCases,
		renderUseCases:         renderUseCases,
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
		uc.Explode()
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

	uc.renderUseCases.AnimationGetter = explosionAnim
	uc.hq.State = types.HQStateExploding

	uc.tilesUseCases.StartAnimation(explosionAnim)
	return nil
}

// IsExplosionFinished проверяет завершение анимации взрыва базы
func (uc *HQUseCases) IsExplosionFinished() {
	if uc.hq != nil && uc.hq.State == types.HQStateExploding {
		if uc.renderUseCases.IsExplosionAnimationFinished() {
			uc.hq.State = types.HQStateDestroyed
		}
	}
}
