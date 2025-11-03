package use_cases

import (
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
)

// TankLifecycleUseCases отвечает за жизненный цикл танка (спавн, взрыв, анимации)
type TankLifecycleUseCases struct {
	tilesUseCases  *TilesUseCases
	renderUseCases interfaces.ITankRenderUseCases
}

// NewTankLifecycleUseCases создает новый экземпляр TankLifecycleUseCases
func NewTankLifecycleUseCases(
	tilesUseCases *TilesUseCases,
	renderUseCases interfaces.ITankRenderUseCases,
) *TankLifecycleUseCases {
	return &TankLifecycleUseCases{
		tilesUseCases:  tilesUseCases,
		renderUseCases: renderUseCases,
	}
}

// Spawn создает танк и запускает процесс спавна с анимацией
func (uc *TankLifecycleUseCases) Spawn(tank *types.TankEntity) error {
	spawnAnimation, err := uc.tilesUseCases.CreateSpawnAnimation()
	if err != nil {
		return err
	}

	tank.Image = spawnAnimation
	tank.State = types.TankStateSpawning
	tank.Altitude = types.SURFACE

	uc.tilesUseCases.StartAnimation(spawnAnimation)
	return nil
}

// Explode устанавливает и запускает анимацию взрыва для танка
func (uc *TankLifecycleUseCases) Explode(tank *types.TankEntity) error {
	explosionAnim, err := uc.tilesUseCases.CreateExplosionAnimation()
	if err != nil {
		return err
	}

	tank.Image = explosionAnim
	tank.State = types.TankStateExploding
	tank.Altitude = types.AIR

	uc.tilesUseCases.StartAnimation(explosionAnim)
	return nil
}

// IsSpawnFinished проверяет и обновляет процесс спавна танка
func (uc *TankLifecycleUseCases) IsSpawnFinished(
	tank *types.TankEntity,
	currentTime float64,
) {
	if tank.State == types.TankStateSpawning {
		if uc.renderUseCases.IsSpawnAnimationFinished(tank) {
			uc.finishSpawnAnimation(tank, currentTime)
		}
	}
}

// IsExplosionFinished проверяет завершение анимации взрыва танка
func (uc *TankLifecycleUseCases) IsExplosionFinished(tank *types.TankEntity) {
	if tank.State == types.TankStateExploding {
		if uc.renderUseCases.IsExplosionAnimationFinished(tank) {
			tank.State = types.TankStateExploded
		}
	}
}

// finishSpawnAnimation завершает анимацию спавна и устанавливает анимацию танка
func (uc *TankLifecycleUseCases) finishSpawnAnimation(
	tank *types.TankEntity,
	currentTime float64,
) {
	tankAnimation, err := uc.tilesUseCases.CreateAnimationTile("base_tank")
	if err == nil {
		tank.Image = tankAnimation
		uc.tilesUseCases.AddAnimation(tankAnimation)
		// Анимация будет запущена автоматически когда танк начнет двигаться
		// Сейчас танк стоит, поэтому анимация должна быть остановлена
		uc.tilesUseCases.StopAnimation(tankAnimation)
	}

	tank.State = types.TankStateStopped
	tank.SpawnedAt = currentTime
}
