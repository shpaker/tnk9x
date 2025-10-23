package repositories

import (
	"github.com/shpaker/gonflict/internal/repositories/game"
	"github.com/shpaker/gonflict/internal/repositories/processed"
)

// Этот файл экспортирует интерфейсы, структуры и конструкторы из подпакетов для удобства импорта
//
// Использование:
// import "github.com/shpaker/gonflict/internal/repositories"
//
// // Создание репозиториев:
// fileRepo := raw.NewFileRepository("assets")  // raw импортируется отдельно
// spritesRepo, _ := repositories.NewSpritesRepository(fileRepo)
// mapsRepo := repositories.NewMapsDataRepository(fileRepo, spritesRepo)
// bulletsRepo := repositories.NewBulletsRepository()
//
// // Использование интерфейсов:
// var spritesService repositories.ISpritesRepository = spritesRepo
// var mapsService repositories.IMapsDataRepository = mapsRepo

// Re-export интерфейсов для удобства
type (
	// ISpritesRepository - интерфейс для работы со спрайтами
	ISpritesRepository = processed.ISpritesRepository

	// IMapsDataRepository - интерфейс для работы с картами уровней
	IMapsDataRepository = processed.IMapsDataRepository

	// IBulletsRepository - интерфейс для работы с пулями
	IBulletsRepository = game.IBulletsRepository

	// IBlocksRepository - интерфейс для работы с блоками
	IBlocksRepository = game.IBlocksRepository
)

// Re-export структур репозиториев для удобства
type (
	// SpritesRepository - репозиторий для работы со спрайтами
	SpritesRepository = processed.SpritesRepository

	// MapsDataRepository - репозиторий для работы с картами уровней
	MapsDataRepository = processed.MapsDataRepository

	// BulletsRepository - репозиторий для работы с пулями
	BulletsRepository = game.BulletsRepository

	// BlocksRepository - репозиторий для работы с блоками
	BlocksRepository = game.BlocksRepository
)

// Re-export конструкторов для удобства
var (
	NewSpritesRepository  = processed.NewSpritesRepository
	NewMapsDataRepository = processed.NewMapsDataRepository
	NewBulletsRepository  = game.NewBulletsRepository
	NewBlocksRepository   = game.NewBlocksRepository
)
