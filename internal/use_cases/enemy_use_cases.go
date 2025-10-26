package use_cases

import (
	"github.com/shpaker/gonflict/internal/repositories/game"
	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/types"
)

type EnemyUseCases struct {
	tanksRepo            game.ITanksRepository
	tilesetRepo          processed.ITilesetRepository
	spawnerTilesetRepo   processed.ITilesetRepository
	explosionTilesetRepo processed.ITilesetRepository
	animationUseCases    IAnimationUseCases
	spawnAnimations      map[int]*types.TileAnimationEntity // Анимации спавна для каждого врага
	tankAnimations       map[int]*types.TileAnimationEntity // Анимации танков (движения) для каждого врага
}

func NewEnemyUseCases(
	tanksRepo game.ITanksRepository,
	tilesetRepo processed.ITilesetRepository,
	spawnerTilesetRepo processed.ITilesetRepository,
	explosionTilesetRepo processed.ITilesetRepository,
	animationUseCases IAnimationUseCases,
) *EnemyUseCases {
	return &EnemyUseCases{
		tanksRepo:            tanksRepo,
		tilesetRepo:          tilesetRepo,
		spawnerTilesetRepo:   spawnerTilesetRepo,
		explosionTilesetRepo: explosionTilesetRepo,
		animationUseCases:    animationUseCases,
		spawnAnimations:      make(map[int]*types.TileAnimationEntity),
		tankAnimations:       make(map[int]*types.TileAnimationEntity),
	}
}

// GetEnemies возвращает всех врагов
func (uc *EnemyUseCases) GetEnemies() []*types.TankEntity {
	return uc.tanksRepo.GetAllTanks()
}

// RemoveEnemy удаляет врага по индексу
func (uc *EnemyUseCases) RemoveEnemy(index int) error {
	// Вместо удаления танка, запускаем анимацию взрыва
	tanks := uc.tanksRepo.GetAllTanks()
	if index >= 0 && index < len(tanks) {
		enemy := tanks[index]
		if enemy != nil && !enemy.IsExploding {
			// Создаем анимацию взрыва
			tilesUseCases := NewTilesUseCases(uc.explosionTilesetRepo)
			explosionAnim, err := tilesUseCases.CreateAnimationTile("explosion")
			if err == nil {
				// Заменяем AnimationGetter на анимацию взрыва
				enemy.AnimationGetter = explosionAnim
				enemy.IsExploding = true
				uc.animationUseCases.AddAnimation(explosionAnim)
				uc.animationUseCases.StartAnimation(explosionAnim)
			}
		}
	}
	return nil
}

// InitEnemies создает начальное количество врагов на карте
func (uc *EnemyUseCases) InitEnemies(enemySpawners [][]int) error {
	// Создаем врагов на позициях из конфигурации
	for i, spawner := range enemySpawners {
		if len(spawner) != 2 {
			continue // Пропускаем некорректные записи
		}

		// Конвертируем координаты в пиксели
		// enemySpawners содержит координаты в танках (1 танк = 16 пикселей)
		position := types.Position{
			X: float64(spawner[0]) * TankSpriteSize,
			Y: float64(spawner[1]) * TankSpriteSize,
		}

		if err := uc.StartEnemySpawn(position, i); err != nil {
			return err
		}
	}

	return nil
}

// StartEnemySpawn начинает процесс спавна врага с анимацией
func (uc *EnemyUseCases) StartEnemySpawn(position types.Position, enemyIndex int) error {
	// Создаем tilesUseCases для создания анимации с правильной конфигурацией
	tilesUseCases := NewTilesUseCases(uc.spawnerTilesetRepo)

	// Создаем анимацию спавна через TilesUseCases (будет применяться repeats: 10 из конфига)
	spawnAnimation, err := tilesUseCases.CreateAnimationTile("spawner")
	if err != nil {
		return err
	}

	uc.animationUseCases.AddAnimation(spawnAnimation)
	uc.animationUseCases.StartAnimation(spawnAnimation)
	uc.spawnAnimations[enemyIndex] = spawnAnimation

	// Создаем анимацию танка через CreateAnimationTile (учитывает duration из конфига)
	tankTilesUseCases := NewTilesUseCases(uc.tilesetRepo)
	tankAnimation, err := tankTilesUseCases.CreateAnimationTile("base_tank")
	if err != nil {
		return err
	}
	uc.animationUseCases.AddAnimation(tankAnimation)

	// Сохраняем ссылку на анимацию танка
	uc.tankAnimations[enemyIndex] = tankAnimation

	// Не запускаем анимацию сразу, она запустится при движении
	// uc.animationUseCases.StartAnimation(tankAnimation)

	// Создаем врага
	enemy := &types.TankEntity{
		AnimationGetter: tankAnimation,
		SpawnPosition:   position,
		WorldPosition:   position,
		Speed:           0,
		Direction:       types.DirectionDown,
		IsSpawned:       false, // Враг не заспавнен, идет процесс спавна
		SpawnedAt:       0,
		Altitude:        types.SURFACE,
	}

	uc.tanksRepo.AddTank(enemy)
	return nil
}

// UpdateEnemiesSpawn обновляет процесс спавна врагов
func (uc *EnemyUseCases) UpdateEnemiesSpawn(currentTime float64) {
	enemies := uc.tanksRepo.GetAllTanks()

	for i, enemy := range enemies {
		if enemy == nil {
			continue
		}

		// Если танк взрывается, проверяем завершение анимации взрыва
		if enemy.IsExploding {
			// Проверяем, является ли AnimationGetter анимацией взрыва
			if anim, ok := enemy.AnimationGetter.(*types.TileAnimationEntity); ok {
				if anim.IsFinished() {
					// Анимация взрыва закончилась, удаляем танк
					uc.tanksRepo.RemoveTank(i)
				}
			}
			continue
		}

		// Если танк еще не заспавнен, проверяем анимацию спавна
		if !enemy.IsSpawned {
			// Проверяем, завершилась ли анимация спавна
			if spawnAnim, exists := uc.spawnAnimations[i]; exists {
				if spawnAnim.IsFinished() {
					// Завершаем спавн
					enemy.IsSpawned = true
					enemy.SpawnedAt = currentTime

					// Останавливаем анимацию спавна
					uc.animationUseCases.StopAnimation(spawnAnim)
					delete(uc.spawnAnimations, i)
				}
			}
		}
	}
}

// GetEnemySpawnAnimation возвращает анимацию спавна для врага
func (uc *EnemyUseCases) GetEnemySpawnAnimation(enemyIndex int) *types.TileAnimationEntity {
	if anim, exists := uc.spawnAnimations[enemyIndex]; exists {
		return anim
	}
	return nil
}

// UpdateEnemiesAnimations обновляет анимации врагов в зависимости от их движения
func (uc *EnemyUseCases) UpdateEnemiesAnimations() {
	enemies := uc.tanksRepo.GetAllTanks()

	for i, enemy := range enemies {
		if enemy == nil {
			continue
		}

		// Пропускаем взрывающихся врагов
		if enemy.IsExploding {
			continue
		}

		// Пропускаем врагов в процессе спавна
		if !enemy.IsSpawned {
			continue
		}

		// Получаем анимацию танка для этого врага
		tankAnimation, exists := uc.tankAnimations[i]
		if !exists {
			continue
		}

		// Управляем анимацией в зависимости от скорости
		if enemy.Speed > 0 {
			// Если танк движется, запускаем анимацию
			if !tankAnimation.IsAnimating {
				uc.animationUseCases.StartAnimation(tankAnimation)
			}
		} else {
			// Если танк стоит, останавливаем анимацию
			if tankAnimation.IsAnimating {
				uc.animationUseCases.StopAnimation(tankAnimation)
			}
		}
	}
}
