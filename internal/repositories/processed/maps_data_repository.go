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
	fileRepository    interfaces.IFileRepository
	tilesetRepository interfaces.ITilesetRepository
	width             uint
	height            uint
}

func NewMapsDataRepository(
	fileRepository interfaces.IFileRepository,
	tilesetRepository interfaces.ITilesetRepository,
) *MapsDataRepository {
	return &MapsDataRepository{
		fileRepository:    fileRepository,
		tilesetRepository: tilesetRepository,
		width:             0,
		height:            0,
	}
}

// readFile читает файл уровня
func (mdr *MapsDataRepository) readFile(levelNumber int) ([]string, error) {
	levelName := "levels/" + strconv.Itoa(levelNumber) + ".bcmap"

	// Читаем текстовый файл уровня
	data, err := mdr.fileRepository.ReadFile(levelName)
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
// Вычисляет width (длина первой строки) и height (количество строк)
func (mdr *MapsDataRepository) parseLevelLines(
	lines []string,
) ([]types.BlockEntity, error) {
	var level []types.BlockEntity

	if len(lines) == 0 {
		return level, fmt.Errorf("empty level file")
	}

	// Вычисляем высоту (количество строк)
	mdr.height = uint(len(lines))

	// Вычисляем ширину (длина первой строки)
	firstLine := strings.TrimSpace(lines[0])
	if len(firstLine) == 0 {
		return level, fmt.Errorf("first line is empty")
	}
	mdr.width = uint(len(firstLine))

	// Парсим каждую строку
	for y, line := range lines {
		line = strings.TrimSpace(line)

		// Проверяем длину строки (должна совпадать с первой строкой)
		if len(line) != int(mdr.width) {
			return level, fmt.Errorf(
				"invalid row %d length: expected %d, got %d",
				y+1,
				mdr.width,
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

// GetLevelsCount возвращает количество доступных карт (файлы вида *.bcmap)
func (mdr *MapsDataRepository) GetLevelsCount() (int, error) {
	return mdr.fileRepository.CountFiles("levels", "*.bcmap")
}

// GetSize возвращает размеры карты [width, height]
// Размеры вычисляются при загрузке уровня через GetLevel
func (mdr *MapsDataRepository) GetSize() [2]uint {
	return [2]uint{mdr.width, mdr.height}
}
