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
	tanksRepo         game.ITanksRepository
	tilesetRepo       processed.ITilesetRepository
	animationUseCases IAnimationUseCases
	spawnAnimation    *types.TileAnimationEntity // Анимация спавна
	spawnDurationMs   uint                       // Длительность спавна в миллисекундах
	playerTankIndex   int                        // Индекс танка игрока в репозитории (всегда 0)
}

// NewTankUseCases создает новый экземпляр TankUseCases
func NewTankUseCases(
	tanksRepo game.ITanksRepository,
	tilesetRepo processed.ITilesetRepository,
	spawnerTilesetRepo processed.ITilesetRepository,
	animationUseCases IAnimationUseCases,
	spawnDurationMs uint,
) *TankUseCases {
	uc := &TankUseCases{
		tanksRepo:         tanksRepo,
		tilesetRepo:       tilesetRepo,
		animationUseCases: animationUseCases,
		spawnDurationMs:   spawnDurationMs,
		playerTankIndex:   0, // Танк игрока всегда первый в репозитории
	}

	// Создаем анимацию спавна
	spawnAnimation, err := uc.makeSpawnAnimation(spawnerTilesetRepo)
	if err != nil {
		panic(err)
	}
	uc.spawnAnimation = spawnAnimation

	return uc
}

// makeTank создает танк с начальными параметрами
func (uc *TankUseCases) makeTank() (types.TankEntity, *types.TileAnimationEntity, error) {
	// Получаем данные анимации для танка
	animationFrames, err := uc.tilesetRepo.GetAnimationData("base_tank")
	if err != nil {
		return types.TankEntity{}, nil, err
	}

	// Создаем TileAnimationEntity для танка
	tankAnimation := types.NewTileAnimationEntity(animationFrames)

	// Добавляем анимацию танка через AnimationUseCases
	uc.animationUseCases.AddAnimation(tankAnimation)

	// Анимация танка начинается остановленной
	// uc.animationUseCases.StartAnimation(tankAnimation) - убираем автоматический запуск

	// Создаем игрока с начальными параметрами
	spawnPosition := types.Position{X: 4 * TankSpriteSize, Y: 12 * TankSpriteSize}

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

	return *player, tankAnimation, nil
}

// makeSpawnAnimation создает анимацию спавна
func (uc *TankUseCases) makeSpawnAnimation(spawnerTilesetRepo processed.ITilesetRepository) (*types.TileAnimationEntity, error) {
	// Получаем данные анимации для спавна
	animationFrames, err := spawnerTilesetRepo.GetAnimationData("spawner")
	if err != nil {
		return nil, err
	}

	// Создаем TileAnimationEntity для спавна
	spawnAnimation := types.NewTileAnimationEntity(animationFrames)

	// Добавляем анимацию спавна через AnimationUseCases
	uc.animationUseCases.AddAnimation(spawnAnimation)

	return spawnAnimation, nil
}

// GetTank возвращает данные танка
func (uc *TankUseCases) GetTank() (*types.TankEntity, error) {
	return uc.tanksRepo.GetTank(uc.playerTankIndex)
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

	// Запускаем анимацию при повороте
	// TODO: Получить tankAnimation из танка
	// uc.animationUseCases.StartAnimation(tankAnimation)

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
	// TODO: Получить tankAnimation из танка
	// uc.animationUseCases.StopAnimation(tankAnimation)

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
	if delta > 0 {
		// TODO: Получить tankAnimation из танка
		// uc.animationUseCases.StartAnimation(tankAnimation)
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
	// Проверяем, есть ли уже танк в репозитории
	tanks := uc.tanksRepo.GetAllTanks()
	if len(tanks) > uc.playerTankIndex {
		tank := tanks[uc.playerTankIndex]
		if tank != nil && !tank.IsSpawned {
			return // Уже спавнится
		}
		if tank != nil && tank.IsSpawned {
			return // Уже заспавнен
		}
	}

	// Создаем танк
	_, _, err := uc.makeTank()
	if err != nil {
		panic(err)
	}
	// Танк уже добавлен в репозиторий в makeTank()

	// Получаем танк из репозитория
	tank, err := uc.GetTank()
	if err != nil {
		panic(err)
	}

	// Инициализируем состояние спавна
	tank.IsSpawned = false
	tank.SpawnedAt = spawnStartTime // Сохраняем время начала спавна

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

	// Проверяем, прошло ли достаточно времени с начала спавна (конвертируем миллисекунды в секунды)
	if currentTime-tank.SpawnedAt >= float64(uc.spawnDurationMs)/1000.0 {
		// Завершаем спавн
		tank.IsSpawned = true
		tank.SpawnedAt = currentTime // Обновляем время завершения спавна

		// Останавливаем анимацию спавна
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
	if !tank.IsSpawned {
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
