package use_cases

import (
	"errors"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
)

var _ interfaces.IAIUseCases = (*AIUseCases)(nil)

// navGridCellSize — размер клетки сетки AI: минимальный обломок кирпича
const navGridCellSize = 4

// AIUseCases собирает снимок мира для вражеского танка и передаёт его
// AI-скрипту; вся логика поведения — в скрипте
type AIUseCases struct {
	// Entities
	stageSession *session_entities.StageSessionEntity
	mapEntity    *types.MapEntity
	hq           *types.HQEntity

	// Repositories
	tanksRepository   interfaces.ITanksRepository
	bulletsRepository interfaces.IBulletsRepository
	bonusesRepository interfaces.IBonusesRepository

	// bonusHunting — враги видят бонусы и едут за ними
	// (game.enemy_bonus_pickup)
	bonusHunting bool

	// Adapters
	scriptEngine interfaces.IAIScriptEngine
}

func NewAIUseCases(
	stageSession *session_entities.StageSessionEntity,
	mapEntity *types.MapEntity,
	hq *types.HQEntity,
	tanksRepository interfaces.ITanksRepository,
	bulletsRepository interfaces.IBulletsRepository,
	bonusesRepository interfaces.IBonusesRepository,
	bonusHunting bool,
	scriptEngine interfaces.IAIScriptEngine,
) *AIUseCases {
	return &AIUseCases{
		stageSession:      stageSession,
		mapEntity:         mapEntity,
		hq:                hq,
		tanksRepository:   tanksRepository,
		bulletsRepository: bulletsRepository,
		bonusesRepository: bonusesRepository,
		bonusHunting:      bonusHunting,
		scriptEngine:      scriptEngine,
	}
}

// ExecuteAI возвращает решение скрипта для танка на тике tick.
// Реализует interfaces.IAIUseCases.
func (uc *AIUseCases) ExecuteAI(
	tank *types.TankEntity,
	tick int,
) (types.EnemyAIDecision, error) {
	if tank == nil {
		return types.EnemyAIDecision{}, errors.New("tank is nil")
	}

	return uc.scriptEngine.UpdateEnemyAI(uc.buildContext(tank, tick))
}

func (uc *AIUseCases) buildContext(
	tank *types.TankEntity,
	tick int,
) types.EnemyAIContext {
	context := types.EnemyAIContext{
		Self:        uc.tankInfo(tank),
		StageNumber: uc.stageSession.GetStageNumber(),
		Tick:        tick,
		Grid: types.NewNavGridEntity(
			uc.mapEntity.GetSizePx(),
			navGridCellSize,
			uc.mapEntity.GetBlocks(),
		),
		HQ: types.AIHQInfo{
			Position: uc.hq.Position,
			Size:     uc.hq.GetSize(),
			Intact:   uc.hq.State == types.HQStateIntact,
		},
	}

	for _, player := range uc.tanksRepository.GetActivePlayerTanks() {
		context.Players = append(context.Players, uc.tankInfo(player))
	}

	for _, enemy := range uc.tanksRepository.GetAllEnemies() {
		if enemy == nil || enemy == tank || !enemy.IsActive() {
			continue
		}
		context.Enemies = append(context.Enemies, uc.tankInfo(enemy))
	}

	for _, bullet := range uc.bulletsRepository.GetAllBullets() {
		if bullet == nil {
			continue
		}
		context.Bullets = append(context.Bullets, types.AIBulletInfo{
			Position:  bullet.Position,
			Size:      bullet.GetSize(),
			Direction: bullet.Direction,
			FromEnemy: bullet.GetOwner().IsEnemy(),
		})
	}

	if uc.bonusHunting {
		for _, bonus := range uc.bonusesRepository.GetAllBonuses() {
			if bonus == nil {
				continue
			}
			context.Bonuses = append(context.Bonuses, types.AIBonusInfo{
				Position: bonus.GetPosition(),
				Size:     bonus.GetSize(),
				Type:     bonus.GetType(),
			})
		}
	}

	return context
}

func (uc *AIUseCases) tankInfo(tank *types.TankEntity) types.AITankInfo {
	specs := tank.GetSpecs()
	return types.AITankInfo{
		ID:         tank.GetID(),
		Position:   tank.Position,
		Size:       tank.GetSize(),
		Direction:  tank.Direction,
		Level:      specs.GetLevel(),
		HitPoints:  tank.GetHitPoints(),
		WithBonus:  tank.GetWithBonus(),
		Reinforced: specs.GetBulletsReinforced(),
		Boat:       tank.CanSail(),
	}
}
