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
	width           uint
	height          uint
}

func NewMapsDataRepository(
	fileRepository interfaces.IFileRepository,
	tilesetRegistry interfaces.ITilesetRepositoryRegistry,
) *MapsDataRepository {
	return &MapsDataRepository{
		fileRepository:  fileRepository,
		tilesetRegistry: tilesetRegistry,
		width:           0,
		height:          0,
	}
}

func (mdr *MapsDataRepository) readFile(levelNumber int) ([]string, error) {
	levelName := "levels/" + strconv.Itoa(levelNumber) + ".bcmap"

	data, err := mdr.fileRepository.ReadFile(levelName)
	if err != nil {
		return nil, fmt.Errorf("failed to read level %d: %w", levelNumber, err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")

	return lines, nil
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

	animationData, err := mdr.tilesetRegistry.GetBlocksAnimationData(
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

func (mdr *MapsDataRepository) GetLevel(
	levelNumber int,
	tileBaseSize int,
) (*types.MapEntity, error) {
	lines, err := mdr.readFile(levelNumber)
	if err != nil {
		return nil, err
	}

	blocks, bonusSpawnPositions, err := mdr.parseLevelLines(lines, tileBaseSize)
	if err != nil {
		return nil, err
	}

	sizePx := types.Size{
		Width:  int(mdr.width) * tileBaseSize,
		Height: int(mdr.height) * tileBaseSize,
	}

	mapEntity := types.NewMapEntity(sizePx, blocks, bonusSpawnPositions)

	return mapEntity, nil
}

func (mdr *MapsDataRepository) GetLevelsCount() (int, error) {
	return mdr.fileRepository.CountFiles("levels", "*.bcmap")
}
