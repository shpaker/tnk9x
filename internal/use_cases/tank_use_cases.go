package use_cases

import (
	"errors"

	"github.com/shpaker/gonflict/internal/repositories/game"
	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/utils"
)

// TankUseCases реализация интерфейса TankUseCases
type TankUseCases struct {
	tanksRepo          game.ITanksRepository
	tilesetRepo        processed.ITilesetRepository
	spawnerTilesetRepo processed.ITilesetRepository
	animationUseCases  IAnimationUseCases
	spawnAnimation     *types.TileAnimationEntity // Анимация спавна
	tankAnimation      *types.TileAnimationEntity // Анимация танка (движения)
	playerTank         *types.TankEntity          // Указатель на танк игрока
	playerSpawners     [][]int                    // Координаты спавна игроков
}

// NewTankUseCases создает новый экземпляр TankUseCases
func NewTankUseCases(
	tanksRepo game.ITanksRepository,
	tilesetRepo processed.ITilesetRepository,
	spawnerTilesetRepo processed.ITilesetRepository,
	animationUseCases IAnimationUseCases,
	playerSpawners [][]int,
) *TankUseCases {
	uc := &TankUseCases{
		tanksRepo:          tanksRepo,
		tilesetRepo:        tilesetRepo,
		spawnerTilesetRepo: spawnerTilesetRepo,
		animationUseCases:  animationUseCases,
		playerTank:         nil, // Будет установлен при создании танка
		playerSpawners:     playerSpawners,
	}

	return uc
}

// makeTank создает танк с начальными параметрами
func (uc *TankUseCases) makeTank() (types.TankEntity, *types.TileAnimationEntity, error) {
	// Создаем tilesUseCases для создания анимации танка
	tilesUseCases := NewTilesUseCases(uc.tilesetRepo)

	// Создаем анимацию танка через CreateAnimationTile (учитывает duration из конфига)
	tankAnimation, err := tilesUseCases.CreateAnimationTile("base_tank")
	if err != nil {
		return types.TankEntity{}, nil, err
	}

	// Добавляем анимацию танка через AnimationUseCases
	uc.animationUseCases.AddAnimation(tankAnimation)

	// Сохраняем ссылку на анимацию танка
	uc.tankAnimation = tankAnimation

	// Не запускаем анимацию сразу, она запустится при движении
	// uc.animationUseCases.StartAnimation(tankAnimation)

	// Создаем игрока с начальными параметрами
	// Берем координаты спавна первого игрока из конфига
	// Координаты в конфиге указаны в тайлах (8x8), танк имеет размер 16x16 (2x2 тайла)
	var spawnPosition types.Position
	if len(uc.playerSpawners) > 0 {
		firstSpawner := uc.playerSpawners[0]
		if len(firstSpawner) >= 2 {
			// Координаты в тайлах, умножаем на размер тайла
			// Учитываем смещение в 8 пикселей (1 тайл)
			spawnPosition = types.Position{
				X: float64(firstSpawner[0]) * TankSpriteSize,
				Y: float64(firstSpawner[1]) * TankSpriteSize,
			}
		} else {
			// Fallback на старую позицию, если конфиг некорректен
			spawnPosition = types.Position{X: 12 * TankSpriteSize, Y: 24 * TankSpriteSize}
		}
	} else {
		// Fallback на старую позицию, если конфиг пуст
		spawnPosition = types.Position{X: 12 * TankSpriteSize, Y: 24 * TankSpriteSize}
	}

	player := &types.TankEntity{
		AnimationGetter: tankAnimation,
		SpawnPosition:   spawnPosition,
		WorldPosition: types.Position{
			X: spawnPosition.X,
			Y: spawnPosition.Y,
		},
		Speed:     0,
		Direction: types.DirectionUp,
		IsSpawned: false,         // Танк не заспавнен по умолчанию
		SpawnedAt: 0,             // Время спавна будет установлено позже
		Altitude:  types.SURFACE, // Танки на уровне поверхности
	}

	// Добавляем танк в репозиторий
	uc.tanksRepo.AddTank(player)

	// Сохраняем указатель на танк игрока
	uc.playerTank = player

	return *player, tankAnimation, nil
}

// makeSpawnAnimation создает анимацию спавна
func (uc *TankUseCases) makeSpawnAnimation(spawnerTilesetRepo processed.ITilesetRepository) (*types.TileAnimationEntity, error) {
	// Создаем tilesUseCases для создания анимации с правильной конфигурацией
	tilesUseCases := NewTilesUseCases(spawnerTilesetRepo)

	// Создаем анимацию спавна через TilesUseCases (будет применяться repeats: 10 из конфига)
	spawnAnimation, err := tilesUseCases.CreateAnimationTile("spawner")
	if err != nil {
		return nil, err
	}

	// Добавляем анимацию спавна через AnimationUseCases
	uc.animationUseCases.AddAnimation(spawnAnimation)

	return spawnAnimation, nil
}

// GetTank возвращает данные танка игрока
func (uc *TankUseCases) GetTank() (*types.TankEntity, error) {
	if uc.playerTank == nil {
		return nil, errors.New("tank not created yet")
	}
	return uc.playerTank, nil
}

// GetDirection возвращает текущее направление танка
func (uc *TankUseCases) GetDirection() types.Direction {
	tank, err := uc.GetTank()
	if err != nil {
		return types.DirectionUp // Возвращаем направление по умолчанию
	}
	return tank.Direction
}

// RotateTank поворачивает танк в указанном направлении
func (uc *TankUseCases) RotateTank(direction types.Direction) error {
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
func (uc *TankUseCases) StopTank(byCollision bool) error {
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
		tank.WorldPosition.X = float64(int(tank.WorldPosition.X))
		tank.WorldPosition.Y = float64(int(tank.WorldPosition.Y))
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
func (uc *TankUseCases) MoveTank(
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
func (uc *TankUseCases) StartSpawn(spawnStartTime float64) {
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
	// Танк уже добавлен в репозиторий в makeTank()
	// и playerTank уже установлен

	// Инициализируем состояние спавна
	uc.playerTank.IsSpawned = false
	uc.playerTank.SpawnedAt = spawnStartTime // Сохраняем время начала спавна

	// Пересоздаем анимацию спавна для нового цикла
	spawnAnimation, err := uc.makeSpawnAnimation(uc.spawnerTilesetRepo)
	if err != nil {
		panic(err)
	}
	uc.spawnAnimation = spawnAnimation

	// Запускаем анимацию спавна
	uc.animationUseCases.StartAnimation(uc.spawnAnimation)
}

// UpdateSpawn обновляет процесс спавна
func (uc *TankUseCases) UpdateSpawn(currentTime float64) {
	tank, err := uc.GetTank()
	if err != nil {
		return
	}

	// Проверяем, идет ли спавн (IsSpawned == false означает процесс спавна)
	if tank.IsSpawned {
		return
	}

	// Проверяем, завершилась ли анимация спавна
	if uc.spawnAnimation != nil && uc.spawnAnimation.IsFinished() {
		// Завершаем спавн
		tank.IsSpawned = true
		tank.SpawnedAt = currentTime // Обновляем время завершения спавна

		// Анимация уже остановлена (IsFinished вернет true когда циклы закончатся)
		uc.animationUseCases.StopAnimation(uc.spawnAnimation)
	}
}

// IsSpawning возвращает true, если танк в процессе спавна
func (uc *TankUseCases) IsSpawning() bool {
	tank, err := uc.GetTank()
	if err != nil {
		return false
	}
	return !tank.IsSpawned
}

// GetSpawnAnimation возвращает анимацию спавна
func (uc *TankUseCases) GetSpawnAnimation() *types.TileAnimationEntity {
	return uc.spawnAnimation
}

// GetTankImageId возвращает ID изображения танка с учетом состояния спавна
func (uc *TankUseCases) GetTankImageId() (string, error) {
	tank, err := uc.GetTank()
	if err != nil {
		return "", errors.New("tank not created yet")
	}

	// Во время спавна (IsSpawned == false) показываем анимацию спавна
	if !tank.IsSpawned && uc.spawnAnimation != nil {
		return uc.spawnAnimation.GetImageId()
	}

	// После спавна показываем обычный танк
	return tank.GetImageId()
}

// ShouldShowTank возвращает true, если танк должен отображаться
func (uc *TankUseCases) ShouldShowTank() bool {
	tank, err := uc.GetTank()
	if err != nil {
		return false
	}
	return tank.IsSpawned
}
