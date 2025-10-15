package repositories

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shpaker/gonflict/internal/constants"
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/models"
)

type LevelsRepository struct {
	assetsRepository interfaces.IAssetsRepository
}

func NewLevelsRepository(assetsRepository interfaces.IAssetsRepository) interfaces.ILevelsRepository {
	return &LevelsRepository{
		assetsRepository: assetsRepository,
	}
}

func (r *LevelsRepository) GetLevel(num int) (*models.Level, error) {
	levelName := "levels/" + strconv.Itoa(num)

	// Читаем текстовый файл уровня
	data, err := r.assetsRepository.ReadAsset(levelName)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать уровень %d: %w", num, err)
	}

	// Разбиваем на строки
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")

	// Проверяем размер карты
	if len(lines) != constants.BattleFieldBlocksLength {
		return nil, fmt.Errorf("неверная высота карты: ожидалось %d, получено %d", constants.BattleFieldBlocksLength, len(lines))
	}

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

			// Пропускаем пустые места
			if charStr == "." {
				continue
			}

			// Проверяем, есть ли символ в маппинге
			blockType, exists := constants.MapCharsBlocksMapping[charStr]
			if !exists {
				return nil, fmt.Errorf("неизвестный символ '%s' в позиции (%d, %d)", charStr, x+1, y+1)
			}

			// Добавляем блок в уровень
			level = append(level, models.Block{
				Position: models.Position{X: x, Y: y},
				Name:     blockType,
			})
		}
	}

	return &level, nil
}
