package use_cases

import (
	"errors"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.IAIUseCases = (*AIUseCases)(nil)

type AIUseCases struct {
	scriptEngine interfaces.IAIScriptEngine
}

func NewAIUseCases(
	scriptEngine interfaces.IAIScriptEngine,
) *AIUseCases {
	return &AIUseCases{
		scriptEngine: scriptEngine,
	}
}

func (uc *AIUseCases) ExecuteAI(
	tank *types.TankEntity,
	context types.EnemyAIContext,
) (types.EnemyAIDecision, error) {
	if tank == nil {
		return types.EnemyAIDecision{}, errors.New("tank is nil")
	}

	return uc.scriptEngine.UpdateEnemyAI(
		tank.Position.X,
		tank.Position.Y,
		int(tank.Direction),
		int(tank.State),
		context,
	)
}
