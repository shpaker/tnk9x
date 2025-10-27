package use_cases

import (
	"errors"

	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/utils"
)

// PlayerUseCases управляет танком игрока
type PlayerUseCases struct {
	tankUseCases      ITankUseCasesRef
	animationUseCases IAnimationUseCases
	spawnAnimation    *types.TileAnimationEntity // Анимация спавна
	tankAnimation     *types.TileAnimationEntity // Анимация танка (движения)
	playerTank        *types.TankEntity          // Указатель на танк игрока
	playerTankIndex   int                        // Индекс танка игрока
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
		playerTankIndex:   -1,
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

	// Находим индекс танка
	tanks := uc.tankUseCases.GetAllTanks()
	for i, t := range tanks {
		if t == player {
			uc.playerTankIndex = i
			break
		}
	}

	return *player, tankAnim, nil
}

// makeSpawnAnimation создает анимацию спавна
func (uc *PlayerUseCases) makeSpawnAnimation() (*types.TileAnimationEntity, error) {
	return uc.tankUseCases.CreateSpawnAnimation()
}

// GetTank возвращает данные танка игрока
func (uc *PlayerUseCases) GetTank() (*types.TankEntity, error) {
	if uc.playerTank == nil {
		return nil, errors.New("tank not created yet")
	}
	return uc.playerTank, nil
}

// GetDirection возвращает текущее направление танка
func (uc *PlayerUseCases) GetDirection() types.Direction {
	tank, err := uc.GetTank()
	if err != nil {
		return types.DirectionUp
	}
	return tank.Direction
}

// RotateTank поворачивает танк в указанном направлении
func (uc *PlayerUseCases) RotateTank(direction types.Direction) error {
	tank, err := uc.GetTank()
	if err != nil {
		return err
	}
	if !tank.IsSpawned {
		return errors.New("tank is not spawned yet")
	}

	if tank.Speed != 0 {
		return errors.New("cannot rotate while moving")
	}

	tank.Speed = 32.0
	if tank.Direction == direction {
		return errors.New("already facing this direction")
	}

	tank.Direction = direction

	// Запускаем анимацию при повороте (и последующем движении)
	if uc.tankAnimation != nil && !uc.tankAnimation.IsAnimating {
		uc.animationUseCases.StartAnimation(uc.tankAnimation)
	}

	return nil
}

// StopTank останавливает танк
func (uc *PlayerUseCases) StopTank(byCollision bool) error {
	tank, err := uc.GetTank()
	if err != nil {
		return err
	}
	if !tank.IsSpawned {
		return errors.New("tank is not spawned yet")
	}

	tank.Speed = 0

	// Останавливаем анимацию танка
	if uc.tankAnimation != nil {
		uc.animationUseCases.StopAnimation(uc.tankAnimation)
	}

	if byCollision {
		// Округляем координаты до ближайшего кратного 4
		tank.WorldPosition.X = utils.RoundToNearestMultipleOf4(tank.WorldPosition.X)
		tank.WorldPosition.Y = utils.RoundToNearestMultipleOf4(tank.WorldPosition.Y)
		return nil
	}

	// Выравниваем позицию по сетке
	switch tank.Direction {
	case types.DirectionUp:
		tank.WorldPosition.Y = float64(utils.RoundToEven(tank.WorldPosition.Y, false))
	case types.DirectionDown:
		tank.WorldPosition.Y = float64(utils.RoundToEven(tank.WorldPosition.Y, true))
	case types.DirectionLeft:
		tank.WorldPosition.X = float64(utils.RoundToEven(tank.WorldPosition.X, false))
	case types.DirectionRight:
		tank.WorldPosition.X = float64(utils.RoundToEven(tank.WorldPosition.X, true))
	}

	return nil
}

// MoveTank перемещает танк в указанном направлении
func (uc *PlayerUseCases) MoveTank(
	direction types.Direction,
	dt float64,
) error {
	tank, err := uc.GetTank()
	if err != nil {
		return err
	}
	if !tank.IsSpawned {
		return errors.New("tank is not spawned yet")
	}

	delta := tank.Speed * dt

	// Если танк движется, запускаем анимацию
	if delta > 0 && uc.tankAnimation != nil {
		// Запускаем анимацию только если она не идет
		if !uc.tankAnimation.IsAnimating {
			uc.animationUseCases.StartAnimation(uc.tankAnimation)
		}
	}

	switch tank.Direction {
	case types.DirectionUp:
		tank.WorldPosition.Y -= delta
	case types.DirectionDown:
		tank.WorldPosition.Y += delta
	case types.DirectionLeft:
		tank.WorldPosition.X -= delta
	case types.DirectionRight:
		tank.WorldPosition.X += delta
	}

	return nil
}

// StartSpawn начинает процесс спавна танка
func (uc *PlayerUseCases) StartSpawn(spawnStartTime float64) {
	// Проверяем, есть ли уже танк
	if uc.playerTank != nil && !uc.playerTank.IsSpawned {
		return // Уже спавнится
	}
	if uc.playerTank != nil && uc.playerTank.IsSpawned {
		return // Уже заспавнен
	}

	// Создаем танк только если его еще нет
	_, _, err := uc.makeTank()
	if err != nil {
		panic(err)
	}

	// Инициализируем состояние спавна
	uc.playerTank.IsSpawned = false
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
	if tank.IsSpawned {
		return
	}

	// Проверяем, завершилась ли анимация спавна
	if uc.spawnAnimation != nil && uc.spawnAnimation.IsFinished() {
		// Завершаем спавн
		tank.IsSpawned = true
		tank.SpawnedAt = currentTime

		uc.animationUseCases.StopAnimation(uc.spawnAnimation)
	}
}

// IsSpawning возвращает true, если танк в процессе спавна
func (uc *PlayerUseCases) IsSpawning() bool {
	tank, err := uc.GetTank()
	if err != nil {
		return false
	}
	return !tank.IsSpawned
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
	if !tank.IsSpawned && uc.spawnAnimation != nil {
		return uc.spawnAnimation.GetImageId()
	}

	// После спавна показываем обычный танк
	return tank.GetImageId()
}

// ShouldShowTank возвращает true, если танк должен отображаться
func (uc *PlayerUseCases) ShouldShowTank() bool {
	tank, err := uc.GetTank()
	if err != nil {
		return false
	}
	return tank.IsSpawned
}
