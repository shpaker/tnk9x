package processed

import (
	"fmt"
	"image"
	"sort"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.ITilesetRepositoryRegistry = (*TilesetRepositoryRegistry)(nil)

// tilesetSources сопоставляет тип тайлсета с путём до его ассетов
var tilesetSources = map[types.TilesetType]string{
	types.TilesetTypeBlocks:          "tiles/blocks",
	types.TilesetTypePlayer:          "tiles/tanks_players",
	types.TilesetTypeEnemy:           "tiles/tanks_enemies",
	types.TilesetTypeBullet:          "tiles/bullet",
	types.TilesetTypeSpawner:         "tiles/spawner",
	types.TilesetTypeExplosion:       "tiles/explosion_tank",
	types.TilesetTypeBulletExplosion: "tiles/bullet_explosion",
	types.TilesetTypeHQ:              "tiles/hq",
	types.TilesetTypeBonuses:         "tiles/bonuses",
	types.TilesetTypeHUD:             "tiles/hud",
}

type TilesetRepositoryRegistry struct {
	tilesets map[types.TilesetType]*TilesetDataRepository
}

func NewTilesetRepositoryRegistry(
	fileRepo interfaces.IFileRepository,
) (*TilesetRepositoryRegistry, error) {
	tilesetTypes := make([]types.TilesetType, 0, len(tilesetSources))
	for tilesetType := range tilesetSources {
		tilesetTypes = append(tilesetTypes, tilesetType)
	}
	sort.Slice(tilesetTypes, func(i, j int) bool {
		return tilesetTypes[i] < tilesetTypes[j]
	})

	tilesets := make(
		map[types.TilesetType]*TilesetDataRepository,
		len(tilesetSources),
	)
	for _, tilesetType := range tilesetTypes {
		repo, err := NewTilesetDataRepository(
			fileRepo,
			tilesetSources[tilesetType],
		)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to create %s tileset: %w",
				tilesetType,
				err,
			)
		}
		tilesets[tilesetType] = repo
	}

	return &TilesetRepositoryRegistry{tilesets: tilesets}, nil
}

func (tr *TilesetRepositoryRegistry) tileset(
	tilesetType types.TilesetType,
) (*TilesetDataRepository, error) {
	repo, exists := tr.tilesets[tilesetType]
	if !exists {
		return nil, fmt.Errorf("unknown tileset type: %s", tilesetType)
	}
	return repo, nil
}

func (tr *TilesetRepositoryRegistry) GetImageData(
	tilesetType types.TilesetType,
	id string,
) (image.Image, error) {
	repo, err := tr.tileset(tilesetType)
	if err != nil {
		return nil, err
	}
	return repo.getImage(id)
}

func (tr *TilesetRepositoryRegistry) GetAnimationData(
	tilesetType types.TilesetType,
	id string,
) (types.AnimationData, error) {
	repo, err := tr.tileset(tilesetType)
	if err != nil {
		return nil, err
	}
	return repo.getAnimationData(id)
}

func (tr *TilesetRepositoryRegistry) GetAnimationConfig(
	tilesetType types.TilesetType,
	id string,
) (types.AnimationConfig, error) {
	repo, err := tr.tileset(tilesetType)
	if err != nil {
		return types.AnimationConfig{}, err
	}
	return repo.getAnimationConfig(id)
}

func (tr *TilesetRepositoryRegistry) GetImageIDs(
	tilesetType types.TilesetType,
) []string {
	repo, err := tr.tileset(tilesetType)
	if err != nil {
		return nil
	}
	return repo.imageIDs()
}
