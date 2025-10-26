package use_cases

import (
	"github.com/shpaker/gonflict/internal/repositories/game"
	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/types"
)

type EnemyUseCases struct {
	tanksRepo          game.ITanksRepository
	tilesetRepo        processed.ITilesetRepository
	spawnerTilesetRepo processed.ITilesetRepository
	animationUseCases  IAnimationUseCases
	spawnAnimations    map[int]*types.TileAnimationEntity // Анимации спавна для каждого врага
}

func NewEnemyUseCases(
	tanksRepo game.ITanksRepository,
	tilesetRepo processed.ITilesetRepository,
	spawnerTilesetRepo processed.ITilesetRepository,
	animationUseCases IAnimationUseCases,
) *EnemyUseCases {
	return &EnemyUseCases{
		tanksRepo:          tanksRepo,
		tilesetRepo:        tilesetRepo,
		spawnerTilesetRepo: spawnerTilesetRepo,
		animationUseCases:  animationUseCases,
		spawnAnimations:    make(map[int]*types.TileAnimationEntity),
	}
}

// SpawnEnemy создает нового врага в указанной позиции
func (uc *EnemyUseCases) SpawnEnemy(position types.Position) error {
	// Получаем данные анимации для врага
	animationFrames, err := uc.tilesetRepo.GetAnimationData("base_tank")
	if err != nil {
		return err
	}

	// Создаем TileAnimationEntity для врага
	tankAnimation := types.NewTileAnimationEntity(animationFrames)

	// Добавляем анимацию врага через AnimationUseCases
	uc.animationUseCases.AddAnimation(tankAnimation)

	// Создаем врага
	enemy := &types.TankEntity{
		AnimationGetter: tankAnimation,
		SpawnPosition:   position,
		WorldPosition:   position,
		Speed:           0,
		Direction:       types.DirectionDown,
		IsSpawned:       true, // Враг сразу заспавнен
		SpawnedAt:       0,
		Altitude:        types.SURFACE,
	}

	uc.tanksRepo.AddTank(enemy)
	return nil
}

// GetEnemies возвращает всех врагов
func (uc *EnemyUseCases) GetEnemies() []*types.TankEntity {
	return uc.tanksRepo.GetAllTanks()
}

// RemoveEnemy удаляет врага по индексу
func (uc *EnemyUseCases) RemoveEnemy(index int) error {
	return uc.tanksRepo.RemoveTank(index)
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
	// Получаем данные анимации для спавна
	animationFrames, err := uc.spawnerTilesetRepo.GetAnimationData("spawner")
	if err != nil {
		return err
	}

	// Создаем анимацию спавна
	spawnAnimation := types.NewTileAnimationEntity(animationFrames)
	uc.animationUseCases.AddAnimation(spawnAnimation)
	uc.animationUseCases.StartAnimation(spawnAnimation)
	uc.spawnAnimations[enemyIndex] = spawnAnimation

	// Получаем данные анимации для танка
	tankAnimationFrames, err := uc.tilesetRepo.GetAnimationData("base_tank")
	if err != nil {
		return err
	}

	// Создаем TileAnimationEntity для танка
	tankAnimation := types.NewTileAnimationEntity(tankAnimationFrames)
	uc.animationUseCases.AddAnimation(tankAnimation)

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
		if enemy.IsSpawned {
			continue
		}

		// Если время спавна еще не установлено, устанавливаем его
		if enemy.SpawnedAt == 0 {
			enemy.SpawnedAt = currentTime
		}

		// Проверяем, прошло ли 2 секунды с начала спавна
		if currentTime-enemy.SpawnedAt >= 2.0 {
			// Завершаем спавн
			enemy.IsSpawned = true
			enemy.SpawnedAt = currentTime

			// Останавливаем анимацию спавна
			if spawnAnim, exists := uc.spawnAnimations[i]; exists {
				uc.animationUseCases.StopAnimation(spawnAnim)
				delete(uc.spawnAnimations, i)
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
