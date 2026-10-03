package states

import (
	"log"

	"github.com/shpaker/tnk9x/internal/types"
)

// Награды экрана итогов: пункты REVIVE и BOOST оплачиваются жетоном
// игрока или рекламой площадки, а сами механики доменные — второй
// шанс и усиленный перенос. Пока реклама показывается, площадка
// приостанавливает игру; исход разбирается в первом кадре после неё

// spendToken оплачивает пункт жетоном; false — жетонов нет.
// Ошибка сохранения награду не отменяет
func (state *StageState) spendToken() bool {
	spent, err := state.inventoryUseCases.SpendToken()
	if err != nil {
		log.Printf("save inventory: %v", err)
	}
	return spent
}

// requestReward показывает рекламу за пункт item
func (state *StageState) requestReward(item types.StageResultItem) {
	state.rewardItem = item
	state.rewardPending = true
	state.rewardAdapter.RequestReward()
}

// pollReward ждёт исхода рекламы, ввод меню на это время не
// принимается: награда применяется, без неё меню итогов остаётся
func (state *StageState) pollReward() types.StateTransition {
	switch state.rewardAdapter.PollReward() {
	case types.RewardStatusPending:
		return types.StateTransition{}
	case types.RewardStatusGranted:
		state.rewardPending = false
		return state.grantReward(state.rewardItem)
	default:
		state.rewardPending = false
		return types.StateTransition{}
	}
}

// grantReward применяет награду за пункт item
func (state *StageState) grantReward(
	item types.StageResultItem,
) types.StateTransition {
	state.soundUseCases.RequestStopAll()

	switch item {
	case types.StageResultItemRevive:
		state.revive()
		return types.StateTransition{}
	case types.StageResultItemBoostNext:
		return state.boostTransition(uint(state.nextLevel))
	default:
		return state.boostTransition(state.stageSession.GetStageNumber())
	}
}

// revive — второй шанс: экран итогов закрывается, уровень
// продолжается с того же места, танки вернёт штатный респаун
func (state *StageState) revive() {
	state.stageUseCases.RevivePlayers()
	state.result = nil
	state.resultTicks = 0
	state.endSoundHandled = false
	state.pauseMenuWasOpen = false
	state.stageUseCases.ResumeStageState()
}

// boostTransition — запуск уровня с усиленным переносом; реклама
// уже показана, логическая пауза с ещё одной не нужна
func (state *StageState) boostTransition(level uint) types.StateTransition {
	state.stageUseCases.BoostCarryOver()
	return types.StateTransition{
		Target:    types.TransitionToStage,
		Level:     level,
		CarryOver: true,
	}
}
