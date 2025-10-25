package use_cases

import (
	"errors"

	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/utils"
)

// TankUseCases реализация интерфейса TankUseCases
type TankUseCases struct {
	tilesetRepo       processed.ITilesetRepository
	animationUseCases IAnimationUseCases
	tank              *types.TankEntity // Теперь указатель, может быть nil
	tankAnimation     *types.TileAnimationEntity
	spawnAnimation    *types.TileAnimationEntity // Анимация спавна
	spawnStartTime    float64                    // Время начала спавна
	isSpawning        bool                       // Флаг процесса спавна
}

// NewTankUseCases создает новый экземпляр TankUseCases
func NewTankUseCases(
	tilesetRepo processed.ITilesetRepository,
	spawnerTilesetRepo processed.ITilesetRepository,
	animationUseCases IAnimationUseCases,
) *TankUseCases {
	uc := &TankUseCases{
		tilesetRepo:       tilesetRepo,
		animationUseCases: animationUseCases,
		tank:              nil, // Танк не создается при инициализации
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

	player := types.TankEntity{
		AnimationGetter: tankAnimation,
		SpawnPosition:   spawnPosition,
		WorldPosition: types.Position{
			X: spawnPosition.X,
			Y: spawnPosition.Y,
		},
		Speed:     0,
		Direction: types.DirectionUp,
		IsSpawned: false, // Танк не заспавнен по умолчанию
		SpawnedAt: 0,     // Время спавна будет установлено позже
	}

	return player, tankAnimation, nil
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
	if uc.tank == nil {
		return nil, errors.New("tank not created yet")
	}
	return uc.tank, nil
}

// GetDirection возвращает текущее направление танка
func (uc *TankUseCases) GetDirection() types.Direction {
	if uc.tank == nil {
		return types.DirectionUp // Возвращаем направление по умолчанию
	}
	return uc.tank.Direction
}

// RotateTank поворачивает танк в указанном направлении
func (uc *TankUseCases) RotateTank(direction types.Direction) error {
	if uc.tank == nil {
		return errors.New("tank not created yet")
	}
	if !uc.tank.IsSpawned {
		return errors.New("tank is not spawned yet")
	}

	if uc.tank.Speed != 0 {
		return errors.New("cannot rotate while moving")
	}

	uc.tank.Speed = 32.0
	if uc.tank.Direction == direction {
		return errors.New("already facing this direction")
	}

	uc.tank.Direction = direction

	// Запускаем анимацию при повороте
	uc.animationUseCases.StartAnimation(uc.tankAnimation)

	return nil
}

// StopTank останавливает танк
func (uc *TankUseCases) StopTank(byCollision bool) error {
	if uc.tank == nil {
		return errors.New("tank not created yet")
	}
	if !uc.tank.IsSpawned {
		return errors.New("tank is not spawned yet")
	}

	uc.tank.Speed = 0

	// Останавливаем анимацию танка
	uc.animationUseCases.StopAnimation(uc.tankAnimation)

	if byCollision {
		uc.tank.WorldPosition.X = float64(int(uc.tank.WorldPosition.X))
		uc.tank.WorldPosition.Y = float64(int(uc.tank.WorldPosition.Y))
		return nil
	}

	// Выравниваем позицию по сетке
	switch uc.tank.Direction {
	case types.DirectionUp:
		uc.tank.WorldPosition.Y = float64(utils.RoundToEven(uc.tank.WorldPosition.Y, false))
	case types.DirectionDown:
		uc.tank.WorldPosition.Y = float64(utils.RoundToEven(uc.tank.WorldPosition.Y, true))
	case types.DirectionLeft:
		uc.tank.WorldPosition.X = float64(utils.RoundToEven(uc.tank.WorldPosition.X, false))
	case types.DirectionRight:
		uc.tank.WorldPosition.X = float64(utils.RoundToEven(uc.tank.WorldPosition.X, true))
	}

	return nil
}

// MoveTank перемещает танк в указанном направлении
func (uc *TankUseCases) MoveTank(
	direction types.Direction,
	dt float64,
) error {
	if uc.tank == nil {
		return errors.New("tank not created yet")
	}
	if !uc.tank.IsSpawned {
		return errors.New("tank is not spawned yet")
	}

	delta := uc.tank.Speed * dt

	// Если танк движется, запускаем анимацию
	if delta > 0 {
		uc.animationUseCases.StartAnimation(uc.tankAnimation)
	}

	switch uc.tank.Direction {
	case types.DirectionUp:
		uc.tank.WorldPosition.Y -= delta
	case types.DirectionDown:
		uc.tank.WorldPosition.Y += delta
	case types.DirectionLeft:
		uc.tank.WorldPosition.X -= delta
	case types.DirectionRight:
		uc.tank.WorldPosition.X += delta
	}

	return nil
}

// StartSpawn начинает процесс спавна танка
func (uc *TankUseCases) StartSpawn() {
	if uc.isSpawning || (uc.tank != nil && uc.tank.IsSpawned) {
		return // Уже спавнится или уже заспавнен
	}

	// Создаем танк
	tank, tankAnimation, err := uc.makeTank()
	if err != nil {
		panic(err)
	}
	uc.tank = &tank
	uc.tankAnimation = tankAnimation

	uc.isSpawning = true
	uc.spawnStartTime = 0 // Начинаем с 0
	uc.tank.IsSpawned = false
	uc.tank.SpawnedAt = 0 // Время спавна

	// Запускаем анимацию спавна
	uc.animationUseCases.StartAnimation(uc.spawnAnimation)
}

// UpdateSpawn обновляет процесс спавна
func (uc *TankUseCases) UpdateSpawn(currentTime float64) {
	if !uc.isSpawning {
		return
	}

	// Проверяем, прошло ли 4 секунды
	if currentTime-uc.spawnStartTime >= 4.0 {
		// Завершаем спавн
		uc.isSpawning = false
		uc.tank.IsSpawned = true
		uc.tank.SpawnedAt = currentTime // Устанавливаем время спавна

		// Останавливаем анимацию спавна
		uc.animationUseCases.StopAnimation(uc.spawnAnimation)
	}
}

// IsSpawning возвращает true, если танк в процессе спавна
func (uc *TankUseCases) IsSpawning() bool {
	return uc.isSpawning
}

// GetSpawnAnimation возвращает анимацию спавна
func (uc *TankUseCases) GetSpawnAnimation() *types.TileAnimationEntity {
	return uc.spawnAnimation
}

// GetTankImageId возвращает ID изображения танка с учетом состояния спавна
func (uc *TankUseCases) GetTankImageId() (string, error) {
	if uc.tank == nil {
		return "", errors.New("tank not created yet")
	}

	if uc.isSpawning {
		// Во время спавна показываем анимацию спавна
		return uc.spawnAnimation.GetImageId()
	}

	// После спавна показываем обычный танк
	return uc.tank.GetImageId()
}

// ShouldShowTank возвращает true, если танк должен отображаться
func (uc *TankUseCases) ShouldShowTank() bool {
	return uc.tank != nil && uc.tank.IsSpawned
}
