package processed

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
	image_providers "github.com/shpaker/gonflict/internal/types/image_providers"
)

var MapCharsBlocksMapping = map[string]types.BlockType{
	"#": types.Brick,
	"@": types.Steel,
	"%": types.Forest,
	"~": types.Water,
	"=": types.Ice,
}

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

	tileEntity := &image_providers.StaticProvider{
		ImageID: string(blockType),
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

func (mdr *MapsDataRepository) parseLevelLines(
	lines []string,
	tileBaseSize int,
) (types.MapBlocks, error) {
	var level types.MapBlocks

	if len(lines) == 0 {
		return level, fmt.Errorf("empty level file")
	}

	mdr.height = uint(len(lines))

	firstLine := strings.TrimSpace(lines[0])
	if len(firstLine) == 0 {
		return level, fmt.Errorf("first line is empty")
	}
	mdr.width = uint(len(firstLine))

	for y, line := range lines {
		line = strings.TrimSpace(line)

		if len(line) != int(mdr.width) {
			return level, fmt.Errorf(
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
				return level, err
			}

			if block != nil {
				level = append(level, block)
			}
		}
	}

	return level, nil
}

func (mdr *MapsDataRepository) GetLevel(
	levelNumber int,
	tileBaseSize int,
) (*types.MapEntity, error) {
	lines, err := mdr.readFile(levelNumber)
	if err != nil {
		return nil, err
	}

	blocks, err := mdr.parseLevelLines(lines, tileBaseSize)
	if err != nil {
		return nil, err
	}

	sizePx := types.Size{
		Width:  int(mdr.width) * tileBaseSize,
		Height: int(mdr.height) * tileBaseSize,
	}

	mapEntity := types.NewMapEntity(sizePx, blocks)

	return mapEntity, nil
}

func (mdr *MapsDataRepository) GetLevelsCount() (int, error) {
	return mdr.fileRepository.CountFiles("levels", "*.bcmap")
}
