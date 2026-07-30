package processed

import (
	"fmt"
	"image"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
)

var _ interfaces.ITilesetRepositoryRegistry = (*TilesetRepositoryRegistry)(nil)

type TilesetRepositoryRegistry struct {
	blocks          *TilesetDataRepository
	player          *TilesetDataRepository
	enemy           *TilesetDataRepository
	bullet          *TilesetDataRepository
	spawner         *TilesetDataRepository
	explosion       *TilesetDataRepository
	bulletExplosion *TilesetDataRepository
	hq              *TilesetDataRepository
	bonuses         *TilesetDataRepository
	hud             *TilesetDataRepository
}

func NewTilesetRepositoryRegistry(
	fileRepo interfaces.IFileRepository,
) (*TilesetRepositoryRegistry, error) {
	blocksRepo, err := NewTilesetDataRepository(fileRepo, "tiles/blocks")
	if err != nil {
		return nil, fmt.Errorf("failed to create blocks tileset: %w", err)
	}

	playerRepo, err := NewTilesetDataRepository(fileRepo, "tiles/tanks_players")
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create tanks_players tileset: %w",
			err,
		)
	}

	enemyRepo, err := NewTilesetDataRepository(fileRepo, "tiles/tanks_enemies")
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create tanks_enemies tileset: %w",
			err,
		)
	}

	bulletRepo, err := NewTilesetDataRepository(fileRepo, "tiles/bullet")
	if err != nil {
		return nil, fmt.Errorf("failed to create bullet tileset: %w", err)
	}

	spawnerRepo, err := NewTilesetDataRepository(fileRepo, "tiles/spawner")
	if err != nil {
		return nil, fmt.Errorf("failed to create spawner tileset: %w", err)
	}

	explosionRepo, err := NewTilesetDataRepository(
		fileRepo,
		"tiles/explosion_tank",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create explosion tileset: %w", err)
	}

	bulletExplosionRepo, err := NewTilesetDataRepository(
		fileRepo,
		"tiles/bullet_explosion",
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create bullet explosion tileset: %w",
			err,
		)
	}

	hqRepo, err := NewTilesetDataRepository(fileRepo, "tiles/hq")
	if err != nil {
		return nil, fmt.Errorf("failed to create hq tileset: %w", err)
	}

	bonusesRepo, err := NewTilesetDataRepository(fileRepo, "tiles/bonuses")
	if err != nil {
		return nil, fmt.Errorf("failed to create bonuses tileset: %w", err)
	}

	hudRepo, err := NewTilesetDataRepository(fileRepo, "tiles/hud")
	if err != nil {
		return nil, fmt.Errorf("failed to create hud tileset: %w", err)
	}

	return &TilesetRepositoryRegistry{
		blocks:          blocksRepo,
		player:          playerRepo,
		enemy:           enemyRepo,
		bullet:          bulletRepo,
		spawner:         spawnerRepo,
		explosion:       explosionRepo,
		bulletExplosion: bulletExplosionRepo,
		hq:              hqRepo,
		bonuses:         bonusesRepo,
		hud:             hudRepo,
	}, nil
}

func (tr *TilesetRepositoryRegistry) getImageData(
	tilesetType types.TilesetType,
	id string,
) (image.Image, error) {
	switch tilesetType {
	case types.TilesetTypeBlocks:
		if tr.blocks == nil {
			return nil, fmt.Errorf("blocks repository not initialized")
		}
		return tr.blocks.getImage(id)
	case types.TilesetTypePlayer:
		if tr.player == nil {
			return nil, fmt.Errorf("player repository not initialized")
		}
		return tr.player.getImage(id)
	case types.TilesetTypeEnemy:
		if tr.enemy == nil {
			return nil, fmt.Errorf("enemy repository not initialized")
		}
		return tr.enemy.getImage(id)
	case types.TilesetTypeBullet:
		if tr.bullet == nil {
			return nil, fmt.Errorf("bullet repository not initialized")
		}
		return tr.bullet.getImage(id)
	case types.TilesetTypeSpawner:
		if tr.spawner == nil {
			return nil, fmt.Errorf("spawner repository not initialized")
		}
		return tr.spawner.getImage(id)
	case types.TilesetTypeExplosion:
		if tr.explosion == nil {
			return nil, fmt.Errorf("explosion repository not initialized")
		}
		return tr.explosion.getImage(id)
	case types.TilesetTypeHQ:
		if tr.hq == nil {
			return nil, fmt.Errorf("hq repository not initialized")
		}
		return tr.hq.getImage(id)
	case types.TilesetTypeBonuses:
		if tr.bonuses == nil {
			return nil, fmt.Errorf("bonuses repository not initialized")
		}
		return tr.bonuses.getImage(id)
	case types.TilesetTypeHUD:
		if tr.hud == nil {
			return nil, fmt.Errorf("hud repository not initialized")
		}
		return tr.hud.getImage(id)
	default:
		return nil, fmt.Errorf("unknown tileset type: %s", tilesetType)
	}
}

func (tr *TilesetRepositoryRegistry) GetImageData(
	tilesetType string,
	id string,
) (image.Image, error) {
	return tr.getImageData(types.TilesetType(tilesetType), id)
}

func (tr *TilesetRepositoryRegistry) GetBlocksImage(
	id string,
) (types.IImageProvider, error) {
	if tr.blocks == nil {
		return nil, fmt.Errorf("blocks repository not initialized")
	}

	_, err := tr.blocks.getImage(id)
	if err != nil {
		return nil, err
	}
	return &image_providers.StaticProvider{
		ImageID: id,
	}, nil
}

func (tr *TilesetRepositoryRegistry) GetBlocksAnimationData(
	id string,
) (types.AnimationData, error) {
	if tr.blocks == nil {
		return nil, fmt.Errorf("blocks repository not initialized")
	}
	return tr.blocks.getAnimationData(id)
}

func (tr *TilesetRepositoryRegistry) GetBlocksAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	if tr.blocks == nil {
		return types.AnimationConfig{}, fmt.Errorf(
			"blocks repository not initialized",
		)
	}
	return tr.blocks.getAnimationConfig(id)
}

func (tr *TilesetRepositoryRegistry) GetPlayerImage(
	id string,
) (types.IImageProvider, error) {
	if tr.player == nil {
		return nil, fmt.Errorf("player repository not initialized")
	}

	_, err := tr.player.getImage(id)
	if err != nil {
		return nil, fmt.Errorf("image '%s' not found: %w", id, err)
	}
	return &image_providers.StaticProvider{
		ImageID: id,
	}, nil
}

func (tr *TilesetRepositoryRegistry) GetPlayerAnimationData(
	id string,
) (types.AnimationData, error) {
	if tr.player == nil {
		return nil, fmt.Errorf("player repository not initialized")
	}
	return tr.player.getAnimationData(id)
}

func (tr *TilesetRepositoryRegistry) GetPlayerAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	if tr.player == nil {
		return types.AnimationConfig{}, fmt.Errorf(
			"player repository not initialized",
		)
	}
	return tr.player.getAnimationConfig(id)
}

func (tr *TilesetRepositoryRegistry) GetEnemyImage(
	id string,
) (types.IImageProvider, error) {
	if tr.enemy == nil {
		return nil, fmt.Errorf("enemy repository not initialized")
	}

	_, err := tr.enemy.getImage(id)
	if err != nil {
		return nil, fmt.Errorf("image '%s' not found: %w", id, err)
	}
	return &image_providers.StaticProvider{
		ImageID: id,
	}, nil
}

func (tr *TilesetRepositoryRegistry) GetEnemyAnimationData(
	id string,
) (types.AnimationData, error) {
	if tr.enemy == nil {
		return nil, fmt.Errorf("enemy repository not initialized")
	}
	return tr.enemy.getAnimationData(id)
}

func (tr *TilesetRepositoryRegistry) GetEnemyAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	if tr.enemy == nil {
		return types.AnimationConfig{}, fmt.Errorf(
			"enemy repository not initialized",
		)
	}
	return tr.enemy.getAnimationConfig(id)
}

func (tr *TilesetRepositoryRegistry) GetBulletImage(
	id string,
) (types.IImageProvider, error) {
	if tr.bullet == nil {
		return nil, fmt.Errorf("bullet repository not initialized")
	}

	_, err := tr.bullet.getImage(id)
	if err != nil {
		return nil, fmt.Errorf("image '%s' not found: %w", id, err)
	}
	return &image_providers.StaticProvider{
		ImageID: id,
	}, nil
}

func (tr *TilesetRepositoryRegistry) GetBulletAnimationData(
	id string,
) (types.AnimationData, error) {
	if tr.bullet == nil {
		return nil, fmt.Errorf("bullet repository not initialized")
	}
	return tr.bullet.getAnimationData(id)
}

func (tr *TilesetRepositoryRegistry) GetBulletAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	if tr.bullet == nil {
		return types.AnimationConfig{}, fmt.Errorf(
			"bullet repository not initialized",
		)
	}
	return tr.bullet.getAnimationConfig(id)
}

func (tr *TilesetRepositoryRegistry) GetSpawnerImage(
	id string,
) (types.IImageProvider, error) {
	if tr.spawner == nil {
		return nil, fmt.Errorf("spawner repository not initialized")
	}

	_, err := tr.spawner.getImage(id)
	if err != nil {
		return nil, fmt.Errorf("image '%s' not found: %w", id, err)
	}
	return &image_providers.StaticProvider{
		ImageID: id,
	}, nil
}

func (tr *TilesetRepositoryRegistry) GetSpawnerAnimationData(
	id string,
) (types.AnimationData, error) {
	if tr.spawner == nil {
		return nil, fmt.Errorf("spawner repository not initialized")
	}
	return tr.spawner.getAnimationData(id)
}

func (tr *TilesetRepositoryRegistry) GetSpawnerAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	if tr.spawner == nil {
		return types.AnimationConfig{}, fmt.Errorf(
			"spawner repository not initialized",
		)
	}
	return tr.spawner.getAnimationConfig(id)
}

func (tr *TilesetRepositoryRegistry) GetExplosionImage(
	id string,
) (types.IImageProvider, error) {
	if tr.explosion == nil {
		return nil, fmt.Errorf("explosion repository not initialized")
	}

	_, err := tr.explosion.getImage(id)
	if err != nil {
		return nil, fmt.Errorf("image '%s' not found: %w", id, err)
	}
	return &image_providers.StaticProvider{
		ImageID: id,
	}, nil
}

func (tr *TilesetRepositoryRegistry) GetExplosionAnimationData(
	id string,
) (types.AnimationData, error) {
	if tr.explosion == nil {
		return nil, fmt.Errorf("explosion repository not initialized")
	}
	return tr.explosion.getAnimationData(id)
}

func (tr *TilesetRepositoryRegistry) GetExplosionAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	if tr.explosion == nil {
		return types.AnimationConfig{}, fmt.Errorf(
			"explosion repository not initialized",
		)
	}
	return tr.explosion.getAnimationConfig(id)
}

func (tr *TilesetRepositoryRegistry) GetExplosionTankImage(
	id string,
) (types.IImageProvider, error) {
	if tr.explosion == nil {
		return nil, fmt.Errorf("explosion repository not initialized")
	}

	_, err := tr.explosion.getImage(id)
	if err != nil {
		return nil, fmt.Errorf("image '%s' not found: %w", id, err)
	}
	return &image_providers.StaticProvider{
		ImageID: id,
	}, nil
}

func (tr *TilesetRepositoryRegistry) GetExplosionTankAnimationData(
	id string,
) (types.AnimationData, error) {
	if tr.explosion == nil {
		return nil, fmt.Errorf("explosion repository not initialized")
	}
	return tr.explosion.getAnimationData(id)
}

func (tr *TilesetRepositoryRegistry) GetExplosionTankAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	if tr.explosion == nil {
		return types.AnimationConfig{}, fmt.Errorf(
			"explosion repository not initialized",
		)
	}
	return tr.explosion.getAnimationConfig(id)
}

func (tr *TilesetRepositoryRegistry) GetBulletExplosionImage(
	id string,
) (types.IImageProvider, error) {
	if tr.bulletExplosion == nil {
		return nil, fmt.Errorf("bullet explosion repository not initialized")
	}

	_, err := tr.bulletExplosion.getImage(id)
	if err != nil {
		return nil, fmt.Errorf("image '%s' not found: %w", id, err)
	}
	return &image_providers.StaticProvider{
		ImageID: id,
	}, nil
}

func (tr *TilesetRepositoryRegistry) GetBulletExplosionAnimationData(
	id string,
) (types.AnimationData, error) {
	if tr.bulletExplosion == nil {
		return nil, fmt.Errorf("bullet explosion repository not initialized")
	}
	return tr.bulletExplosion.getAnimationData(id)
}

func (tr *TilesetRepositoryRegistry) GetBulletExplosionAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	if tr.bulletExplosion == nil {
		return types.AnimationConfig{}, fmt.Errorf(
			"bullet explosion repository not initialized",
		)
	}
	return tr.bulletExplosion.getAnimationConfig(id)
}

func (tr *TilesetRepositoryRegistry) GetHQImage(
	id string,
) (types.IImageProvider, error) {
	if tr.hq == nil {
		return nil, fmt.Errorf("hq repository not initialized")
	}

	_, err := tr.hq.getImage(id)
	if err != nil {
		return nil, fmt.Errorf("image '%s' not found: %w", id, err)
	}
	return &image_providers.StaticProvider{
		ImageID: id,
	}, nil
}

func (tr *TilesetRepositoryRegistry) GetHUDImage(
	id string,
) (types.IImageProvider, error) {
	if tr.hud == nil {
		return nil, fmt.Errorf("hud repository not initialized")
	}

	_, err := tr.hud.getImage(id)
	if err != nil {
		return nil, fmt.Errorf("image '%s' not found: %w", id, err)
	}
	return &image_providers.StaticProvider{
		ImageID: id,
	}, nil
}

func (tr *TilesetRepositoryRegistry) GetBonusesImage(
	id string,
) (types.IImageProvider, error) {
	if tr.bonuses == nil {
		return nil, fmt.Errorf("bonuses repository not initialized")
	}

	_, err := tr.bonuses.getImage(id)
	if err != nil {
		return nil, fmt.Errorf("image '%s' not found: %w", id, err)
	}
	return &image_providers.StaticProvider{
		ImageID: id,
	}, nil
}

func (tr *TilesetRepositoryRegistry) GetBonusesAnimationData(
	id string,
) (types.AnimationData, error) {
	if tr.bonuses == nil {
		return nil, fmt.Errorf("bonuses repository not initialized")
	}
	return tr.bonuses.getAnimationData(id)
}

func (tr *TilesetRepositoryRegistry) GetBonusesAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	if tr.bonuses == nil {
		return types.AnimationConfig{}, fmt.Errorf(
			"bonuses repository not initialized",
		)
	}
	return tr.bonuses.getAnimationConfig(id)
}

func (tr *TilesetRepositoryRegistry) GetHQAnimationData(
	id string,
) (types.AnimationData, error) {
	if tr.hq == nil {
		return nil, fmt.Errorf("hq repository not initialized")
	}
	return tr.hq.getAnimationData(id)
}

func (tr *TilesetRepositoryRegistry) GetHQAnimationConfig(
	id string,
) (types.AnimationConfig, error) {
	if tr.hq == nil {
		return types.AnimationConfig{}, fmt.Errorf(
			"hq repository not initialized",
		)
	}
	return tr.hq.getAnimationConfig(id)
}
