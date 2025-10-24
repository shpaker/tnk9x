package processed

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/repositories/raw"
	"github.com/shpaker/gonflict/internal/types"
)

// ITilesAdapter определяет интерфейс для работы с тайлами
type ITilesAdapter interface {
	GetTileUseCases() interfaces.ITileUseCases
	GetTilesetRepository() ITilesetRepository
}

type MapsDataRepository struct {
	fileRepo     raw.IFileRepository
	tilesAdapter ITilesAdapter
}

func NewMapsDataRepository(
	fileRepo raw.IFileRepository,
	tilesAdapter ITilesAdapter,
) *MapsDataRepository {
	return &MapsDataRepository{
		fileRepo:     fileRepo,
		tilesAdapter: tilesAdapter,
	}
}

func (mdr *MapsDataRepository) GetFileRepo() raw.IFileRepository {
	return mdr.fileRepo
}

func (mdr *MapsDataRepository) GetTilesetRepository() ITilesetRepository {
	return mdr.tilesAdapter.GetTilesetRepository()
}

func (mdr *MapsDataRepository) GetTilesAdapter() ITilesAdapter {
	return mdr.tilesAdapter
}

// readFile читает файл уровня
func (mdr *MapsDataRepository) readFile(levelNumber int) ([]string, error) {
	levelName := "levels/" + strconv.Itoa(levelNumber)

	// Читаем текстовый файл уровня
	data, err := mdr.fileRepo.ReadFile(levelName)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать уровень %d: %w", levelNumber, err)
	}

	// Разбиваем на строки
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")

	return lines, nil
}

// createBlockFromChar создает блок из символа карты
func (mdr *MapsDataRepository) createBlockFromChar(charStr string, x, y int) (*types.BlockEntity, error) {
	// Пропускаем пустые места
	if charStr == "." {
		return nil, nil
	}

	// Проверяем, есть ли символ в маппинге
	blockType, exists := MapCharsBlocksMapping[charStr]
	if !exists {
		return nil, fmt.Errorf("неизвестный символ '%s' в позиции (%d, %d)", charStr, x+1, y+1)
	}

	// Создаем TileStaticEntity для блока
	tileEntity, err := mdr.tilesAdapter.GetTileUseCases().CreateStaticTile(string(blockType))
	if err != nil {
		return nil, fmt.Errorf("не удалось создать tile entity для блока %s: %w", blockType, err)
	}

	// Создаем полный объект блока используя конструктор
	block := types.NewBlockEntity(string(blockType), float64(x), float64(y), tileEntity)

	return block, nil
}

// parseLevelLines парсит строки карты и создает блоки
func (mdr *MapsDataRepository) parseLevelLines(lines []string) ([]types.BlockEntity, error) {
	var level []types.BlockEntity

	// Проверяем количество строк (должно быть 26)
	if len(lines) != MapBlocksLength {
		return level, fmt.Errorf("неверное количество строк в уровне: ожидалось %d, получено %d", MapBlocksLength, len(lines))
	}

	// Парсим каждую строку
	for y, line := range lines {
		line = strings.TrimSpace(line)

		// Проверяем длину строки (должна быть 26)
		if len(line) != MapBlocksLength {
			return level, fmt.Errorf("неверная длина строки %d: ожидалось %d символов, получено %d", y+1, MapBlocksLength, len(line))
		}

		// Парсим каждый символ в строке
		for x, char := range line {
			charStr := string(char)

			// Создаем блок из символа
			block, err := mdr.createBlockFromChar(charStr, x, y)
			if err != nil {
				return level, err
			}

			// Добавляем блок в уровень (если он не nil)
			if block != nil {
				level = append(level, *block)
			}
		}
	}

	return level, nil
}

// createBlockProperties создает свойства блока на основе его типа
func (mdr *MapsDataRepository) createBlockProperties(blockType types.BlockType) *types.BlockProperties {
	return &types.BlockProperties{
		Collidable: true,
	}
}

func (mdr *MapsDataRepository) GetLevel(levelNumber int) ([]types.BlockEntity, error) {
	// Читаем файл уровня
	lines, err := mdr.readFile(levelNumber)
	if err != nil {
		return []types.BlockEntity{}, err
	}

	// Парсим строки и создаем блоки
	level, err := mdr.parseLevelLines(lines)
	if err != nil {
		return []types.BlockEntity{}, err
	}

	return level, nil
}
