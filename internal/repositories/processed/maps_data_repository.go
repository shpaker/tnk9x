package processed

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
)

var MapCharsBlocksMapping = map[string]types.BlockType{
	"#": types.Brick,
	"@": types.Steel,
	"%": types.Forest,
	"~": types.Water,
	"-": types.Ice,
}

var _ interfaces.IMapsDataRepository = (*MapsDataRepository)(nil)

type MapsDataRepository struct {
	fileRepository  interfaces.IFileRepository
	tilesetRegistry interfaces.ITilesetRepositoryRegistry
	defaults        types.LevelDefaults
	width           uint
	height          uint
}

func NewMapsDataRepository(
	fileRepository interfaces.IFileRepository,
	tilesetRegistry interfaces.ITilesetRepositoryRegistry,
	defaults types.LevelDefaults,
) *MapsDataRepository {
	return &MapsDataRepository{
		fileRepository:  fileRepository,
		tilesetRegistry: tilesetRegistry,
		defaults:        defaults,
		width:           0,
		height:          0,
	}
}

// levelPath — файл уровня levels/<name>.bcmap
func levelPath(name string) string {
	return "levels/" + name + ".bcmap"
}

func (mdr *MapsDataRepository) createBlockFromChar(
	charStr string,
	x, y int,
	tileBaseSize int,
) (*types.BlockEntity, error) {
	if charStr == "." {
		return nil, nil
	}

	blockType, exists := MapCharsBlocksMapping[charStr]
	if !exists {
		return nil, fmt.Errorf(
			"unknown character '%s' at position (%d, %d)",
			charStr,
			x+1,
			y+1,
		)
	}

	tileEntity, err := mdr.createImageProvider(blockType)
	if err != nil {
		return nil, err
	}

	positionX := float64(x) * float64(tileBaseSize)
	positionY := float64(y) * float64(tileBaseSize)

	block := types.NewBlockEntity(
		string(blockType),
		positionX,
		positionY,
		tileBaseSize,
		tileEntity,
	)

	return block, nil
}

// createImageProvider возвращает анимированный провайдер для воды
// и статичный для остальных блоков
func (mdr *MapsDataRepository) createImageProvider(
	blockType types.BlockType,
) (types.IImageProvider, error) {
	if blockType != types.Water {
		return &image_providers.StaticProvider{
			ImageID: string(blockType),
		}, nil
	}

	animationData, err := mdr.tilesetRegistry.GetAnimationData(
		types.TilesetTypeBlocks,
		string(types.Water),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get water animation: %w", err)
	}

	provider := image_providers.NewAnimationProvider(animationData)
	provider.IsAnimating = true

	return provider, nil
}

func (mdr *MapsDataRepository) parseLevelLines(
	lines []string,
	tileBaseSize int,
) (types.MapBlocks, []types.Position, error) {
	var level types.MapBlocks
	var bonusSpawnPositions []types.Position

	if len(lines) == 0 {
		return level, bonusSpawnPositions, fmt.Errorf("empty level file")
	}

	mdr.height = uint(len(lines))

	firstLine := strings.TrimSpace(lines[0])
	if len(firstLine) == 0 {
		return level, bonusSpawnPositions, fmt.Errorf("first line is empty")
	}
	mdr.width = uint(len(firstLine))

	for y, line := range lines {
		line = strings.TrimSpace(line)

		if len(line) != int(mdr.width) {
			return level, bonusSpawnPositions, fmt.Errorf(
				"invalid row %d length: expected %d, got %d",
				y+1,
				mdr.width,
				len(line),
			)
		}

		for x, char := range line {
			charStr := string(char)

			block, err := mdr.createBlockFromChar(charStr, x, y, tileBaseSize)
			if err != nil {
				return level, bonusSpawnPositions, err
			}

			if block != nil {
				level = append(level, block)
			} else if charStr == "." {
				// Позиция не занята блоком - можно спавнить бонус
				// Добавляем только координаты четных блоков
				if x%2 == 0 && y%2 == 0 {
					positionX := float64(x) * float64(tileBaseSize)
					positionY := float64(y) * float64(tileBaseSize)
					bonusSpawnPositions = append(bonusSpawnPositions, types.Position{
						X: positionX,
						Y: positionY,
					})
				}
			}
		}
	}

	return level, bonusSpawnPositions, nil
}

// GetLevel читает и разбирает уровень; каждый вызов создаёт новую
// карту, потому что блоки разрушаются по ходу игры
func (mdr *MapsDataRepository) GetLevel(
	levelNumber int,
	tileBaseSize int,
) (*types.LevelEntity, error) {
	return mdr.loadLevel(
		levelPath(strconv.Itoa(levelNumber)),
		levelNumber,
		tileBaseSize,
	)
}

// GetSceneLevel читает карту вне кампании (например, демо-сцену
// главного меню) из levels/<name>.bcmap; номер уровня у неё 0
func (mdr *MapsDataRepository) GetSceneLevel(
	name string,
	tileBaseSize int,
) (*types.LevelEntity, error) {
	return mdr.loadLevel(levelPath(name), 0, tileBaseSize)
}

// loadLevel читает и разбирает файл уровня path
func (mdr *MapsDataRepository) loadLevel(
	path string,
	levelNumber int,
	tileBaseSize int,
) (*types.LevelEntity, error) {
	data, err := mdr.fileRepository.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read level %s: %w", path, err)
	}

	file, err := parseLevelFile(string(data), mdr.defaults)
	if err != nil {
		return nil, fmt.Errorf("level %s: %w", path, err)
	}

	blocks, bonusSpawnPositions, err := mdr.parseLevelLines(
		file.mapLines,
		tileBaseSize,
	)
	if err != nil {
		return nil, fmt.Errorf("level %s: %w", path, err)
	}

	sizePx := types.Size{
		Width:  int(mdr.width) * tileBaseSize,
		Height: int(mdr.height) * tileBaseSize,
	}

	mapEntity := types.NewMapEntity(sizePx, blocks, bonusSpawnPositions)

	// Карта без названия: рендер подпишет её номером на языке
	// интерфейса
	return types.NewLevelEntity(
		levelNumber,
		file.name,
		file.maxActive,
		file.time3StarTicks,
		file.waves,
		file.explicitBonuses,
		mapEntity,
	), nil
}

// HasLevel — существует ли файл уровня
func (mdr *MapsDataRepository) HasLevel(levelNumber int) bool {
	_, err := mdr.fileRepository.ReadFile(levelPath(strconv.Itoa(levelNumber)))
	return err == nil
}

func (mdr *MapsDataRepository) GetLevelsCount() (int, error) {
	return mdr.fileRepository.CountFiles("levels", "*.bcmap")
}
