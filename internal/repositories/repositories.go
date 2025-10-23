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
// tilesetRepo, _ := repositories.NewTilesetRepository(fileRepo, "blocks")
// mapsRepo := repositories.NewMapsDataRepository(fileRepo, tilesAdapter)
// bulletsRepo := repositories.NewBulletsRepository()
//
// // Использование интерфейсов:
// var tilesetService repositories.ITilesetRepository = tilesetRepo
// var mapsService repositories.IMapsDataRepository = mapsRepo

// Re-export интерфейсов для удобства
type (
	// ITilesetRepository - интерфейс для работы с тайлсетами
	ITilesetRepository = processed.ITilesetRepository

	// IMapsDataRepository - интерфейс для работы с картами уровней
	IMapsDataRepository = processed.IMapsDataRepository

	// IBulletsRepository - интерфейс для работы с пулями
	IBulletsRepository = game.IBulletsRepository

	// IBlocksRepository - интерфейс для работы с блоками
	IBlocksRepository = game.IBlocksRepository
)

// Re-export структур репозиториев для удобства
type (
	// TilesetRepository - репозиторий для работы с тайлсетами
	TilesetRepository = processed.TilesetDataRepository

	// MapsDataRepository - репозиторий для работы с картами уровней
	MapsDataRepository = processed.MapsDataRepository

	// BulletsRepository - репозиторий для работы с пулями
	BulletsRepository = game.BulletsRepository

	// BlocksRepository - репозиторий для работы с блоками
	BlocksRepository = game.BlocksRepository
)

// Re-export конструкторов для удобства
var (
	NewTilesetRepository  = processed.NewTilesetDataRepository
	NewMapsDataRepository = processed.NewMapsDataRepository
	NewBulletsRepository  = game.NewBulletsRepository
	NewBlocksRepository   = game.NewBlocksRepository
)
