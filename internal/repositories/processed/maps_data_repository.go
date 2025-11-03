package processed

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
	image_providers "github.com/shpaker/gonflict/internal/types/image_providers"
)

// MapCharsBlocksMapping содержит соответствие символов карты типам блоков
var MapCharsBlocksMapping = map[string]types.BlockType{
	"#": types.Brick,
	"@": types.Steel,
	"%": types.Forest,
	"~": types.Water,
	"=": types.Ice,
}

type MapsDataRepository struct {
	fileRepo        interfaces.IFileRepository
	tilesetRepo     interfaces.ITilesetRepository
	mapBlocksLength int
}

func NewMapsDataRepository(
	fileRepo interfaces.IFileRepository,
	tilesetRepo interfaces.ITilesetRepository,
	mapBlocksLength int,
) *MapsDataRepository {
	return &MapsDataRepository{
		fileRepo:        fileRepo,
		tilesetRepo:     tilesetRepo,
		mapBlocksLength: mapBlocksLength,
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
func (mdr *MapsDataRepository) createBlockFromChar(
	charStr string,
	x, y int,
) (*types.BlockEntity, error) {
	// Пропускаем пустые места
	if charStr == "." {
		return nil, nil
	}

	// Проверяем, есть ли символ в маппинге
	blockType, exists := MapCharsBlocksMapping[charStr]
	if !exists {
		return nil, fmt.Errorf(
			"unknown character '%s' at position (%d, %d)",
			charStr,
			x+1,
			y+1,
		)
	}

	// Создаем StaticProvider для блока напрямую
	tileEntity := &image_providers.StaticProvider{
		ImageID: string(blockType),
	}

	// Создаем полный объект блока используя конструктор
	block := types.NewBlockEntity(
		string(blockType),
		float64(x),
		float64(y),
		tileEntity,
	)

	return block, nil
}

// parseLevelLines парсит строки карты и создает блоки
func (mdr *MapsDataRepository) parseLevelLines(
	lines []string,
) ([]types.BlockEntity, error) {
	var level []types.BlockEntity

	// Проверяем количество строк (должно быть 26)
	if len(lines) != mdr.mapBlocksLength {
		return level, fmt.Errorf(
			"invalid row count: expected %d, got %d",
			mdr.mapBlocksLength,
			len(lines),
		)
	}

	// Парсим каждую строку
	for y, line := range lines {
		line = strings.TrimSpace(line)

		// Проверяем длину строки (должна быть 26)
		if len(line) != mdr.mapBlocksLength {
			return level, fmt.Errorf(
				"invalid row %d length: expected %d, got %d",
				y+1,
				mdr.mapBlocksLength,
				len(line),
			)
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

func (mdr *MapsDataRepository) GetLevel(
	levelNumber int,
) ([]types.BlockEntity, error) {
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
