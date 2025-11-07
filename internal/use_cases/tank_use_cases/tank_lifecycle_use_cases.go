package tank_use_cases

import (
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// TankLifecycleUseCases отвечает за жизненный цикл танка (спавн, взрыв, анимации)
type TankLifecycleUseCases struct {
	tilesUseCases      *use_cases.TilesUseCases
	renderUseCases     interfaces.ITankRenderUseCases
	tankCommonUseCases interfaces.ITankCommonUseCases
}

// NewTankLifecycleUseCases создает новый экземпляр TankLifecycleUseCases
func NewTankLifecycleUseCases(
	tilesUseCases *use_cases.TilesUseCases,
	renderUseCases interfaces.ITankRenderUseCases,
	tankCommonUseCases interfaces.ITankCommonUseCases,
) *TankLifecycleUseCases {
	return &TankLifecycleUseCases{
		tilesUseCases:      tilesUseCases,
		renderUseCases:     renderUseCases,
		tankCommonUseCases: tankCommonUseCases,
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
			uc.finishSpawnAnimation(tank)
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
}

// UpdateAllTanksLifecycle обновляет жизненный цикл всех танков (спавн и взрыв)
func (uc *TankLifecycleUseCases) UpdateAllTanksLifecycle() error {
	if uc.tankCommonUseCases == nil {
		return nil
	}

	allTanks := uc.tankCommonUseCases.GetAllTanks()
	for _, tank := range allTanks {
		if tank != nil {
			uc.updateTankSpawn(tank)
			uc.updateTankExplosion(tank)
		}
	}
	return nil
}

// updateTankSpawn проверяет и обновляет процесс спавна танка
func (uc *TankLifecycleUseCases) updateTankSpawn(tank *types.TankEntity) {
	if tank.State != types.TankStateSpawning {
		return
	}

	if uc.renderUseCases == nil {
		return
	}

	if uc.renderUseCases.IsSpawnAnimationFinished(tank) {
		uc.finishSpawnAnimation(tank)
	}
}

// updateTankExplosion проверяет завершение анимации взрыва танка
func (uc *TankLifecycleUseCases) updateTankExplosion(tank *types.TankEntity) {
	if tank.State != types.TankStateExploding {
		return
	}

	if uc.renderUseCases == nil {
		return
	}

	if uc.renderUseCases.IsExplosionAnimationFinished(tank) {
		tank.State = types.TankStateExploded
	}
}
