package use_cases

import (
	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/types"
)

// EnemyUseCases управляет одним конкретным врагом
type EnemyUseCases struct {
	tankUseCases         ITankUseCasesRef
	explosionTilesetRepo processed.ITilesetRepository
	animationUseCases    IAnimationUseCases
	enemyTank            *types.TankEntity          // Ссылка на танк этого врага
	spawnAnimation       *types.TileAnimationEntity // Анимация спавна
}

func NewEnemyUseCases(
	tankUseCases ITankUseCasesRef,
	explosionTilesetRepo processed.ITilesetRepository,
	animationUseCases IAnimationUseCases,
) *EnemyUseCases {
	return &EnemyUseCases{
		tankUseCases:         tankUseCases,
		explosionTilesetRepo: explosionTilesetRepo,
		animationUseCases:    animationUseCases,
		enemyTank:            nil,
	}
}

// GetEnemy возвращает танк этого врага
func (uc *EnemyUseCases) GetEnemy() *types.TankEntity {
	return uc.enemyTank
}

// GetEnemies возвращает массив с одним врагом (для совместимости с интерфейсом)
func (uc *EnemyUseCases) GetEnemies() []*types.TankEntity {
	if uc.enemyTank == nil {
		return []*types.TankEntity{}
	}
	return []*types.TankEntity{uc.enemyTank}
}

// RemoveEnemyByRef удаляет врага по ссылке (запускает анимацию взрыва)
func (uc *EnemyUseCases) RemoveEnemyByRef(enemy *types.TankEntity) {
	if enemy == nil || enemy != uc.enemyTank || enemy.State == types.TankStateExploding {
		return
	}

	// Создаем анимацию взрыва
	tilesUseCases := NewTilesUseCases(uc.explosionTilesetRepo)
	explosionAnim, err := tilesUseCases.CreateAnimationTile("explosion")
	if err == nil {
		// Заменяем AnimationGetter на анимацию взрыва
		enemy.AnimationGetter = explosionAnim
		enemy.State = types.TankStateExploding
		uc.animationUseCases.AddAnimation(explosionAnim)
		uc.animationUseCases.StartAnimation(explosionAnim)
	}
}

// GetEnemyRealIndex возвращает 0 (так как у нас один враг)
func (uc *EnemyUseCases) GetEnemyRealIndex(index int) (int, error) {
	return 0, nil
}

// InitEnemies вызывает StartEnemySpawn для обратной совместимости
func (uc *EnemyUseCases) InitEnemies(enemySpawners [][]int) error {
	if len(enemySpawners) == 0 {
		return nil
	}

	spawner := enemySpawners[0]
	if len(spawner) != 2 {
		return nil
	}

	position := types.Position{
		X: float64(spawner[0]) * TankSpriteSize,
		Y: float64(spawner[1]) * TankSpriteSize,
	}

	return uc.StartEnemySpawn(position)
}

// StartEnemySpawn начинает процесс спавна врага с анимацией
func (uc *EnemyUseCases) StartEnemySpawn(position types.Position) error {
	// Создаем танк через базовый use case
	enemy, spawnAnimation, _, err := uc.tankUseCases.CreateTankWithSpawn(
		position,
		types.DirectionDown,
	)
	if err != nil {
		return err
	}

	// Сохраняем ссылку на танк
	uc.enemyTank = enemy

	// Сохраняем ссылку на анимацию спавна
	uc.spawnAnimation = spawnAnimation

	// Запускаем анимацию спавна
	uc.animationUseCases.StartAnimation(spawnAnimation)

	return nil
}

// UpdateEnemiesSpawn обновляет процесс спавна врага
func (uc *EnemyUseCases) UpdateEnemiesSpawn(currentTime float64) {
	if uc.enemyTank == nil {
		return
	}

	// Если танк взрывается, проверяем завершение анимации взрыва
	if uc.enemyTank.State == types.TankStateExploding {
		// Проверяем, является ли AnimationGetter анимацией взрыва
		if anim, ok := uc.enemyTank.AnimationGetter.(*types.TileAnimationEntity); ok {
			if anim.IsFinished() {
				// Анимация взрыва закончилась, удаляем танк
				uc.enemyTank = nil
			}
		}
		return
	}

	// Если танк еще не заспавнен, проверяем анимацию спавна
	if uc.enemyTank.State == types.TankStateSpawning {
		// Проверяем, завершилась ли анимация спавна
		if uc.spawnAnimation != nil && uc.spawnAnimation.IsFinished() {
			// Завершаем спавн
			uc.enemyTank.State = types.TankStateMoving
			uc.enemyTank.SpawnedAt = currentTime

			// Останавливаем анимацию спавна
			uc.animationUseCases.StopAnimation(uc.spawnAnimation)
			uc.spawnAnimation = nil
		}
	}
}

// UpdateEnemiesAnimations обновляет анимации врага в зависимости от его движения
func (uc *EnemyUseCases) UpdateEnemiesAnimations() {
	if uc.enemyTank == nil {
		return
	}

	// Пропускаем взрывающихся врагов
	if uc.enemyTank.State == types.TankStateExploding {
		return
	}

	// Пропускаем врагов в процессе спавна
	if uc.enemyTank.State == types.TankStateSpawning {
		return
	}

	// Получаем текущую анимацию из AnimationGetter
	animGetter := uc.enemyTank.AnimationGetter
	if animGetter == nil {
		return
	}

	anim, ok := animGetter.(*types.TileAnimationEntity)
	if !ok {
		return
	}

	// Управляем анимацией в зависимости от скорости
	if uc.enemyTank.Speed > 0 {
		// Если танк движется, запускаем анимацию
		if !anim.IsAnimating {
			uc.animationUseCases.StartAnimation(anim)
		}
	} else {
		// Если танк стоит, останавливаем анимацию
		if anim.IsAnimating {
			uc.animationUseCases.StopAnimation(anim)
		}
	}
}
