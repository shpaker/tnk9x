package use_cases

import (
	"errors"

	"github.com/shpaker/gonflict/internal/types"
)

// PlayerUseCases управляет танком игрока
type PlayerUseCases struct {
	tankUseCases      ITankUseCasesRef
	animationUseCases IAnimationUseCases
	spawnAnimation    *types.TileAnimationEntity // Анимация спавна
	tankAnimation     *types.TileAnimationEntity // Анимация танка (движения)
	playerTank        *types.TankEntity          // Указатель на танк игрока
	playerSpawners    [][]int                    // Координаты спавна игроков
}

// NewPlayerUseCases создает новый экземпляр PlayerUseCases
func NewPlayerUseCases(
	tankUseCases ITankUseCasesRef,
	animationUseCases IAnimationUseCases,
	playerSpawners [][]int,
) *PlayerUseCases {
	uc := &PlayerUseCases{
		tankUseCases:      tankUseCases,
		animationUseCases: animationUseCases,
		playerTank:        nil,
		playerSpawners:    playerSpawners,
	}

	return uc
}

// makeTank создает танк с начальными параметрами
func (uc *PlayerUseCases) makeTank() (types.TankEntity, *types.TileAnimationEntity, error) {
	// Берем координаты спавна первого игрока из конфига
	var spawnPosition types.Position
	if len(uc.playerSpawners) > 0 {
		firstSpawner := uc.playerSpawners[0]
		if len(firstSpawner) >= 2 {
			spawnPosition = types.Position{
				X: float64(firstSpawner[0]) * TankSpriteSize,
				Y: float64(firstSpawner[1]) * TankSpriteSize,
			}
		} else {
			spawnPosition = types.Position{X: 12 * TankSpriteSize, Y: 24 * TankSpriteSize}
		}
	} else {
		spawnPosition = types.Position{X: 12 * TankSpriteSize, Y: 24 * TankSpriteSize}
	}

	// Создаем танк через базовый use case
	player, spawnAnim, tankAnim, err := uc.tankUseCases.CreateTankWithSpawn(
		spawnPosition,
		types.DirectionUp,
	)
	if err != nil {
		return types.TankEntity{}, nil, err
	}

	// Сохраняем ссылки
	uc.spawnAnimation = spawnAnim
	uc.tankAnimation = tankAnim
	uc.playerTank = player

	return *player, tankAnim, nil
}

// GetTank возвращает данные танка игрока
func (uc *PlayerUseCases) GetTank() (*types.TankEntity, error) {
	if uc.playerTank == nil {
		return nil, errors.New("tank not created yet")
	}
	return uc.playerTank, nil
}

// StartSpawn начинает процесс спавна танка
func (uc *PlayerUseCases) StartSpawn(spawnStartTime float64) {
	// Проверяем, есть ли уже танк
	if uc.playerTank != nil && uc.playerTank.State == types.TankStateSpawning {
		return // Уже спавнится
	}
	if uc.playerTank != nil && uc.playerTank.State != types.TankStateSpawning {
		return // Уже заспавнен
	}

	// Создаем танк только если его еще нет
	_, _, err := uc.makeTank()
	if err != nil {
		panic(err)
	}

	// Инициализируем состояние спавна
	uc.playerTank.State = types.TankStateSpawning
	uc.playerTank.SpawnedAt = spawnStartTime

	// Запускаем анимацию спавна
	uc.animationUseCases.StartAnimation(uc.spawnAnimation)
}

// UpdateSpawn обновляет процесс спавна
func (uc *PlayerUseCases) UpdateSpawn(currentTime float64) {
	tank, err := uc.GetTank()
	if err != nil {
		return
	}

	// Проверяем, идет ли спавн
	if tank.State != types.TankStateSpawning {
		return
	}

	// Проверяем, завершилась ли анимация спавна
	if uc.spawnAnimation != nil && uc.spawnAnimation.IsFinished() {
		// Завершаем спавн
		tank.State = types.TankStateMoving
		tank.SpawnedAt = currentTime

		uc.animationUseCases.StopAnimation(uc.spawnAnimation)
	}
}

// GetSpawnAnimation возвращает анимацию спавна
func (uc *PlayerUseCases) GetSpawnAnimation() *types.TileAnimationEntity {
	return uc.spawnAnimation
}

// GetTankImageId возвращает ID изображения танка с учетом состояния спавна
func (uc *PlayerUseCases) GetTankImageId() (string, error) {
	tank, err := uc.GetTank()
	if err != nil {
		return "", errors.New("tank not created yet")
	}

	// Во время спавна показываем анимацию спавна
	if tank.State == types.TankStateSpawning && uc.spawnAnimation != nil {
		return uc.spawnAnimation.GetImageId()
	}

	// После спавна показываем обычный танк
	return tank.GetImageId()
}
