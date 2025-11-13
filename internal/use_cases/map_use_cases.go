package use_cases

import (
	"errors"
	"fmt"
	"math/rand"

	"github.com/shpaker/tnk25/internal/interfaces"
	"github.com/shpaker/tnk25/internal/services/collision_services"
	"github.com/shpaker/tnk25/internal/types"
)

type MapUseCases struct {
	mapEntity                *types.MapEntity
	entitiesCollisionService *collision_services.EntitiesCollisionService
	tankCommonUseCases       interfaces.ITankCommonUseCases
	tilesUseCases            interfaces.ITilesUseCases
}

func NewMapUseCases(
	mapEntity *types.MapEntity,
) *MapUseCases {
	return &MapUseCases{
		mapEntity: mapEntity,
	}
}

func NewMapUseCasesWithCollision(
	mapEntity *types.MapEntity,
	entitiesCollisionService *collision_services.EntitiesCollisionService,
	tankCommonUseCases interfaces.ITankCommonUseCases,
	tilesUseCases interfaces.ITilesUseCases,
) *MapUseCases {
	return &MapUseCases{
		mapEntity:                mapEntity,
		entitiesCollisionService: entitiesCollisionService,
		tankCommonUseCases:       tankCommonUseCases,
		tilesUseCases:            tilesUseCases,
	}
}

func (uc *MapUseCases) GetBlocks() types.MapBlocks {
	if uc.mapEntity == nil {
		return types.MapBlocks{}
	}
	return uc.mapEntity.GetBlocks()
}

func (uc *MapUseCases) RemoveBlock(block *types.BlockEntity) error {
	if uc.mapEntity == nil {
		return nil
	}
	return uc.mapEntity.RemoveBlock(block)
}

func (uc *MapUseCases) RemoveBlocks(blocks []*types.BlockEntity) error {
	if uc.mapEntity == nil {
		return nil
	}
	for _, block := range blocks {
		_ = uc.mapEntity.RemoveBlock(block)
	}
	return nil
}

func (uc *MapUseCases) GetSizePx() types.Size {
	if uc.mapEntity == nil {
		return types.Size{}
	}
	return uc.mapEntity.GetSizePx()
}

func (uc *MapUseCases) GetRandomBonusSpawnPosition() types.Position {
	if uc.mapEntity == nil {
		return types.Position{X: 0, Y: 0}
	}
	return uc.mapEntity.GetRandomBonusSpawnPosition()
}

func (uc *MapUseCases) SpawnBonus(baseSizePx uint) (*types.BonusEntity, error) {
	if uc.mapEntity == nil {
		return nil, errors.New("map entity is nil")
	}

	if uc.entitiesCollisionService == nil || uc.tankCommonUseCases == nil ||
		uc.tilesUseCases == nil {
		return nil, errors.New(
			"collision service, tank use cases or tiles use cases not initialized",
		)
	}

	bonusSize := types.Size{
		Width:  int(baseSizePx),
		Height: int(baseSizePx),
	}

	const maxAttempts = 3
	for attempt := 0; attempt < maxAttempts; attempt++ {
		position := uc.mapEntity.GetRandomBonusSpawnPosition()

		// Создаем временную сущность бонуса для проверки коллизий
		// Используем конструктор с минимальными значениями, так как для проверки коллизий
		// нужны только Position, Size и Altitude
		bonusCandidate := types.NewBonusEntity(
			types.BonusTypeGrenade, // Тип не важен для проверки коллизий
			position,
			bonusSize,
			nil, // Изображение не нужно для проверки коллизий
		)

		// Проверяем коллизии с танками
		allTanks := uc.tankCommonUseCases.GetAllTanks()
		hasTankCollision := false
		for _, tank := range allTanks {
			if tank == nil || !tank.IsActive() {
				continue
			}
			if uc.entitiesCollisionService.CheckColliders(
				bonusCandidate,
				tank,
			) {
				hasTankCollision = true
				break
			}
		}

		if hasTankCollision {
			continue
		}

		// Проверяем коллизии с блоками
		blocks := uc.mapEntity.GetBlocks()
		hasBlockCollision := false
		for _, block := range blocks {
			if block == nil {
				continue
			}
			if uc.entitiesCollisionService.CheckColliders(
				bonusCandidate,
				block,
			) {
				hasBlockCollision = true
				break
			}
		}

		if !hasBlockCollision {
			// Выбираем случайный тип бонуса
			bonusTypes := []types.BonusType{
				types.BonusTypeHelmet,
				types.BonusTypeTimer,
				types.BonusTypeTimer,
				types.BonusTypeShovel,
				types.BonusTypeStar,
				types.BonusTypeGrenade,
				types.BonusTypeGrenade,
				types.BonusTypeTank,
			}
			randomType := bonusTypes[rand.Intn(len(bonusTypes))]

			// Создаем изображение для бонуса по его типу
			imageProvider, err := uc.tilesUseCases.CreateStaticTile(
				string(randomType),
			)
			if err != nil {
				// Если не удалось создать изображение, продолжаем попытки
				continue
			}

			// Создаем бонус с выбранным типом, соответствующим изображению
			bonus := types.NewBonusEntity(
				randomType,
				position,
				bonusSize,
				imageProvider,
			)

			return bonus, nil
		}
	}

	return nil, fmt.Errorf(
		"не удалось найти свободную позицию для спавна бонуса после %d попыток",
		maxAttempts,
	)
}
