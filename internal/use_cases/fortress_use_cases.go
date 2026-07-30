package use_cases

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
)

// Тайминги бонуса «лопата», как в оригинале: стальное кольцо ~20 секунд,
// в конце мигает кирпич/сталь и возвращается к кирпичу
const (
	shovelDurationTicks = 1200
	shovelFlashTicks    = 180
	shovelFlashPeriod   = 30
)

// FortressUseCases управляет кольцом укреплений вокруг штаба:
// бонус «лопата» временно заменяет кирпич сталью
type FortressUseCases struct {
	mapUseCases  interfaces.IMapUseCases
	stageSession *session_entities.StageSessionEntity

	// ring — px-координаты 8px-тайлов кольца; неизменны для этапа
	ring         []types.Position
	tileBaseSize int
}

func NewFortressUseCases(
	mapUseCases interfaces.IMapUseCases,
	stageSession *session_entities.StageSessionEntity,
	ring []types.Position,
	tileBaseSize int,
) *FortressUseCases {
	return &FortressUseCases{
		mapUseCases:  mapUseCases,
		stageSession: stageSession,
		ring:         ring,
		tileBaseSize: tileBaseSize,
	}
}

// Apply укрепляет кольцо вокруг штаба сталью на время действия лопаты
func (uc *FortressUseCases) Apply() {
	uc.stageSession.SetShovelTicks(shovelDurationTicks)
	uc.rebuildRing(types.Steel)
}

// Update ведёт отсчёт лопаты: перед откатом кольцо мигает
// кирпич/сталь, по истечении восстанавливается кирпич
func (uc *FortressUseCases) Update() {
	ticks := uc.stageSession.GetShovelTicks()
	if ticks == 0 {
		return
	}

	ticks--
	uc.stageSession.SetShovelTicks(ticks)

	if ticks == 0 {
		uc.rebuildRing(types.Brick)
		return
	}

	if ticks <= shovelFlashTicks && ticks%shovelFlashPeriod == 0 {
		if (ticks/shovelFlashPeriod)%2 == 0 {
			uc.rebuildRing(types.Steel)
		} else {
			uc.rebuildRing(types.Brick)
		}
	}
}

// rebuildRing заменяет тайлы кольца блоками указанного типа;
// как в NES, кольцо каждый раз строится заново независимо от повреждений
func (uc *FortressUseCases) rebuildRing(blockType types.BlockType) {
	for _, position := range uc.ring {
		uc.removeBlockAt(position)
		block := types.NewBlockEntity(
			string(blockType),
			position.X,
			position.Y,
			uc.tileBaseSize,
			&image_providers.StaticProvider{ImageID: string(blockType)},
		)
		uc.mapUseCases.AddBlock(block)
	}
}

func (uc *FortressUseCases) removeBlockAt(position types.Position) {
	for _, block := range uc.mapUseCases.GetBlocks() {
		if block == nil {
			continue
		}
		blockPosition := block.GetPosition()
		if blockPosition.X == position.X && blockPosition.Y == position.Y {
			_ = uc.mapUseCases.RemoveBlock(block)
			return
		}
	}
}
