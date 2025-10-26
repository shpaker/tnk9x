package processed

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shpaker/gonflict/internal/repositories/raw"
	"github.com/shpaker/gonflict/internal/types"
)

type MapsDataRepository struct {
	fileRepo    raw.IFileRepository
	tilesetRepo ITilesetRepository
}

func NewMapsDataRepository(
	fileRepo raw.IFileRepository,
	tilesetRepo ITilesetRepository,
) *MapsDataRepository {
	return &MapsDataRepository{
		fileRepo:    fileRepo,
		tilesetRepo: tilesetRepo,
	}
}

// readFile читает файл уровня
func (mdr *MapsDataRepository) readFile(levelNumber int) ([]string, error) {
	levelName := "levels/" + strconv.Itoa(levelNumber)

	// Читаем текстовый файл уровня
	data, err := mdr.fileRepo.ReadFile(levelName)
	if err != nil {
		return nil, fmt.Errorf("failed to read level %d: %w", levelNumber, err)
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
		return nil, fmt.Errorf("unknown character '%s' at position (%d, %d)", charStr, x+1, y+1)
	}

	// Создаем TileStaticEntity для блока напрямую
	tileEntity := &types.TileStaticEntity{
		ImageId: string(blockType),
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
		return level, fmt.Errorf("invalid row count: expected %d, got %d", MapBlocksLength, len(lines))
	}

	// Парсим каждую строку
	for y, line := range lines {
		line = strings.TrimSpace(line)

		// Проверяем длину строки (должна быть 26)
		if len(line) != MapBlocksLength {
			return level, fmt.Errorf("invalid row %d length: expected %d, got %d", y+1, MapBlocksLength, len(line))
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
