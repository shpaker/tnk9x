package processed

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/shpaker/gonflict/internal/constants"
	"github.com/shpaker/gonflict/internal/models"
	"github.com/shpaker/gonflict/internal/repositories/raw"
	"github.com/shpaker/gonflict/internal/types"
)

type MapsDataRepository struct {
	fileRepo    raw.IFileRepository
	spritesRepo ISpritesRepository
}

func NewMapsDataRepository(
	fileRepo raw.IFileRepository,
	spritesRepo ISpritesRepository,
) *MapsDataRepository {
	return &MapsDataRepository{
		fileRepo:    fileRepo,
		spritesRepo: spritesRepo,
	}
}

// getBlockSprite получает спрайт для блока по его типу
func (mdr *MapsDataRepository) getBlockSprite(blockType types.BlockType) (*ebiten.Image, error) {
	return mdr.spritesRepo.GetSprite("blocks", string(blockType))
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
func (mdr *MapsDataRepository) createBlockFromChar(charStr string, x, y int) (*models.Block, error) {
	// Пропускаем пустые места
	if charStr == "." {
		return nil, nil
	}

	// Проверяем, есть ли символ в маппинге
	blockType, exists := constants.MapCharsBlocksMapping[charStr]
	if !exists {
		return nil, fmt.Errorf("неизвестный символ '%s' в позиции (%d, %d)", charStr, x+1, y+1)
	}

	// Получаем спрайт для блока
	sprite, err := mdr.getBlockSprite(blockType)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить спрайт для блока %s: %w", blockType, err)
	}

	// Создаем полный объект блока
	block := &models.Block{
		Image: sprite,
		Data: &models.BlockData{
			Position: types.Position{X: float64(x), Y: float64(y)},
			Name:     blockType,
		},
		Properties:    mdr.createBlockProperties(blockType),
		WorldPosition: types.Position{X: float64(x), Y: float64(y)}, // Пока используем те же координаты
	}

	return block, nil
}

// parseLevelLines парсит строки карты и создает блоки
func (mdr *MapsDataRepository) parseLevelLines(lines []string) ([]models.Block, error) {
	var level []models.Block

	// Проверяем количество строк (должно быть 26)
	if len(lines) != constants.BattleFieldBlocksLength {
		return level, fmt.Errorf("неверное количество строк в уровне: ожидалось %d, получено %d", constants.BattleFieldBlocksLength, len(lines))
	}

	// Парсим каждую строку
	for y, line := range lines {
		line = strings.TrimSpace(line)

		// Проверяем длину строки (должна быть 26)
		if len(line) != constants.BattleFieldBlocksLength {
			return level, fmt.Errorf("неверная длина строки %d: ожидалось %d символов, получено %d", y+1, constants.BattleFieldBlocksLength, len(line))
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
func (mdr *MapsDataRepository) createBlockProperties(blockType types.BlockType) *models.BlockProperties {
	return &models.BlockProperties{
		Collidable: true,
	}
}

func (mdr *MapsDataRepository) GetLevel(levelNumber int) ([]models.Block, error) {
	// Читаем файл уровня
	lines, err := mdr.readFile(levelNumber)
	if err != nil {
		return []models.Block{}, err
	}

	// Парсим строки и создаем блоки
	level, err := mdr.parseLevelLines(lines)
	if err != nil {
		return []models.Block{}, err
	}

	return level, nil
}
