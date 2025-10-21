package repositories

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/shpaker/gonflict/internal/constants"
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/models"
	"github.com/shpaker/gonflict/internal/types"
)

type LevelsRepository struct {
	assetsRepository  interfaces.IAssetsRepository
	spritesRepository interfaces.ISpritesRepository
}

func NewLevelsRepository(
	assetsRepository interfaces.IAssetsRepository,
	spritesRepository interfaces.ISpritesRepository,
) interfaces.ILevelsDataService {
	return &LevelsRepository{
		assetsRepository:  assetsRepository,
		spritesRepository: spritesRepository,
	}
}

// getBlockSprite получает спрайт для блока по его типу
func (s *LevelsRepository) getBlockSprite(blockType types.BlockType) (*ebiten.Image, error) {
	return s.spritesRepository.GetSprite("blocks", string(blockType))
}

// readAndValidateLevelFile читает файл уровня и проверяет его размер
func (s *LevelsRepository) readAndValidateLevelFile(levelNumber int) ([]string, error) {
	levelName := "levels/" + strconv.Itoa(levelNumber)

	// Читаем текстовый файл уровня
	data, err := s.assetsRepository.ReadAsset(levelName)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать уровень %d: %w", levelNumber, err)
	}

	// Разбиваем на строки
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")

	// Проверяем размер карты
	if len(lines) != constants.BattleFieldBlocksLength {
		return nil, fmt.Errorf("неверная высота карты: ожидалось %d, получено %d", constants.BattleFieldBlocksLength, len(lines))
	}

	return lines, nil
}

// createBlockFromChar создает блок из символа карты
func (s *LevelsRepository) createBlockFromChar(charStr string, x, y int) (*models.Block, error) {
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
	sprite, err := s.getBlockSprite(blockType)
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
		Properties:    s.createBlockProperties(blockType),
		WorldPosition: types.Position{X: float64(x), Y: float64(y)}, // Пока используем те же координаты
	}

	return block, nil
}

// parseLevelLines парсит строки карты и создает блоки
func (s *LevelsRepository) parseLevelLines(lines []string) (models.Level, error) {
	var level models.Level

	// Парсим каждую строку
	for y, line := range lines {
		line = strings.TrimSpace(line)

		// Проверяем ширину строки
		if len(line) != constants.BattleFieldBlocksLength {
			return nil, fmt.Errorf("неверная ширина карты в строке %d: ожидалось %d, получено %d", y+1, constants.BattleFieldBlocksLength, len(line))
		}

		// Парсим каждый символ в строке
		for x, char := range line {
			charStr := string(char)

			// Создаем блок из символа
			block, err := s.createBlockFromChar(charStr, x, y)
			if err != nil {
				return nil, err
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
func (s *LevelsRepository) createBlockProperties(blockType types.BlockType) *models.BlockProperties {
	return &models.BlockProperties{
		Collidable: true,
	}
}

func (s *LevelsRepository) GetLevel(levelNumber int) (models.Level, error) {
	// Читаем и валидируем файл уровня
	lines, err := s.readAndValidateLevelFile(levelNumber)
	if err != nil {
		return nil, err
	}

	// Парсим строки и создаем блоки
	level, err := s.parseLevelLines(lines)
	if err != nil {
		return nil, err
	}

	return level, nil
}
