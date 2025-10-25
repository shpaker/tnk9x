package states

import (
	"errors"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/shpaker/gonflict/internal/adapters"
	"github.com/shpaker/gonflict/internal/repositories/game"
	"github.com/shpaker/gonflict/internal/repositories/processed"
	"github.com/shpaker/gonflict/internal/use_cases"
)

type GameState struct {
	tankUseCases      *use_cases.TankUseCases
	bulletUseCases    *use_cases.BulletUseCases
	mapUseCases       *use_cases.MapUseCases
	collisionUseCases *use_cases.CollisionUseCases
	animationUseCases *use_cases.AnimationUseCases
	inputAdapter      *adapters.InputAdapter
	rendererAdapter   *adapters.RendererAdapter
	startTime         time.Time // Время начала игры для отслеживания спавна
}

// NewGameState создает новое состояние игры с переданным репозиторием карт
func NewGameState(
	mapsRepo processed.IMapsDataRepository,
	mapTilesetRepo processed.ITilesetRepository, // Репозиторий для блоков карты
	playerTilesetRepo processed.ITilesetRepository, // Репозиторий для игрока
	bulletTilesetRepo processed.ITilesetRepository, // Репозиторий для пуль
	spawnerTilesetRepo processed.ITilesetRepository, // Репозиторий для спавна
) (GameState, error) {
	level, err := mapsRepo.GetLevel(13)
	if err != nil {
		return GameState{}, err
	}

	// Создаем репозитории
	blocksRepo := game.NewBlocksRepository()
	bulletsRepo := game.NewBulletsRepository()
	animationsRepo := game.NewAnimationsRepository()

	// Заполняем репозиторий блоков данными уровня
	for _, block := range level {
		blocksRepo.AddBlock(block)
	}

	// Создаем Use Cases
	animationUseCases := use_cases.NewAnimationUseCases(animationsRepo)
	tankUseCases := use_cases.NewTankUseCases(playerTilesetRepo, spawnerTilesetRepo, animationUseCases)
	bulletTilesUseCases := use_cases.NewTilesUseCases(bulletTilesetRepo)
	bulletUseCases := use_cases.NewBulletUseCases(bulletsRepo, bulletTilesUseCases)
	mapUseCases := use_cases.NewMapUseCases(blocksRepo)
	collisionUseCases := use_cases.NewCollisionUseCases(
		bulletUseCases,
		tankUseCases,
		mapUseCases,
	)

	// Создаем TilesUseCases для рендерера
	// Для карты используем репозиторий блоков
	mapTilesUseCases := use_cases.NewTilesUseCases(mapTilesetRepo)
	playerTilesUseCases := use_cases.NewTilesUseCases(playerTilesetRepo)
	bulletTilesUseCasesForRenderer := use_cases.NewTilesUseCases(bulletTilesetRepo)
	spawnerTilesUseCases := use_cases.NewTilesUseCases(spawnerTilesetRepo)

	rendererAdapter := adapters.NewRendererAdapter(
		mapUseCases,
		tankUseCases,
		bulletUseCases,
		mapTilesUseCases,
		playerTilesUseCases,
		bulletTilesUseCasesForRenderer,
		spawnerTilesUseCases,
	)
	inputAdapter := adapters.NewInputAdapter(
		tankUseCases,
		bulletUseCases,
		ebiten.KeyW,     // up
		ebiten.KeyS,     // down
		ebiten.KeyA,     // left
		ebiten.KeyD,     // right
		ebiten.KeySpace, // shoot
	)

	gameState := GameState{
		tankUseCases:      tankUseCases,
		bulletUseCases:    bulletUseCases,
		mapUseCases:       mapUseCases,
		collisionUseCases: collisionUseCases,
		animationUseCases: animationUseCases,
		inputAdapter:      inputAdapter,
		rendererAdapter:   rendererAdapter,
		startTime:         time.Now(),
	}

	// Запускаем спавн танка на старте
	gameState.StartTankSpawn()

	return gameState, nil
}

// StartTankSpawn запускает спавн танка
func (state GameState) StartTankSpawn() {
	state.tankUseCases.StartSpawn()
}

// UpdateTankSpawn обновляет процесс спавна танка
func (state GameState) UpdateTankSpawn(currentTime float64) {
	state.tankUseCases.UpdateSpawn(currentTime)
}

func (state GameState) Update() (State, error) {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return nil, errors.New("exit application")
	}

	// Обновляем спавн танка
	elapsedTime := time.Since(state.startTime).Seconds()
	state.UpdateTankSpawn(elapsedTime)

	state.inputAdapter.Update()
	state.tankUseCases.MoveTank(state.tankUseCases.GetDirection(), use_cases.DT)
	state.animationUseCases.UpdateAnimations() // Централизованное обновление всех анимаций
	state.bulletUseCases.UpdateBullets(use_cases.DT)
	state.collisionUseCases.UpdateCollisions()
	return nil, nil
}

func (state GameState) Draw(screen *ebiten.Image) {
	vector.FillRect(
		screen,
		0,
		0,
		float32(screen.Bounds().Dx()),
		float32(screen.Bounds().Dy()),
		color.Gray{Y: 128},
		false,
	)
	state.rendererAdapter.DrawAll(screen)
}
