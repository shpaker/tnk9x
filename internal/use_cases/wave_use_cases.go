package use_cases

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
)

var _ interfaces.IWaveUseCases = (*WaveUseCases)(nil)

// WaveUseCases выдаёт врагов по волнам сценария уровня;
// позиция в сценарии хранится в сессии уровня
type WaveUseCases struct{}

func NewWaveUseCases() *WaveUseCases {
	return &WaveUseCases{}
}

// NextTank — следующий танк сценария, если его волна уже началась
func (uc *WaveUseCases) NextTank(
	session *session_entities.StageSessionEntity,
) (types.WaveTank, bool) {
	waves, waveIndex, tankIndex, ok := uc.resolveCursor(session)
	if !ok {
		return types.WaveTank{}, false
	}

	tank := waves[waveIndex].Tanks[tankIndex]
	if !session.GetLevel().HasExplicitBonuses() {
		tank.HasBonus = isClassicBonusNumber(session.GetSpawnedEnemies() + 1)
	}
	return tank, true
}

// CommitSpawn сдвигает курсор на следующий танк и запускает паузу
// волны до следующего спауна
func (uc *WaveUseCases) CommitSpawn(
	session *session_entities.StageSessionEntity,
	spawnerIndex int,
) {
	waves, waveIndex, tankIndex, ok := uc.resolveCursor(session)
	if !ok {
		return
	}

	session.SetWaveCursor(waveIndex, tankIndex+1)
	session.RegisterEnemySpawned(spawnerIndex, waves[waveIndex].DelayTicks)
}

// resolveCursor — позиция следующего танка: когда волна исчерпана,
// курсор переходит на следующую, только если выполнено её условие старта
func (uc *WaveUseCases) resolveCursor(
	session *session_entities.StageSessionEntity,
) ([]types.WaveSpec, int, int, bool) {
	level := session.GetLevel()
	if level == nil {
		return nil, 0, 0, false
	}

	waves := level.GetWaves()
	waveIndex, tankIndex := session.GetWaveCursor()
	for waveIndex < len(waves) && tankIndex >= len(waves[waveIndex].Tanks) {
		next := waveIndex + 1
		if next >= len(waves) || !isWaveStarted(session, waves[next]) {
			return nil, 0, 0, false
		}
		waveIndex, tankIndex = next, 0
	}
	if waveIndex >= len(waves) {
		return nil, 0, 0, false
	}

	return waves, waveIndex, tankIndex, true
}

// isWaveStarted проверяет условие старта волны по живым врагам
// предыдущих волн: все они к этому моменту уже вышли на поле
func isWaveStarted(
	session *session_entities.StageSessionEntity,
	wave types.WaveSpec,
) bool {
	alive := uint(0)
	if session.GetSpawnedEnemies() > session.GetDestroyedEnemies() {
		alive = session.GetSpawnedEnemies() - session.GetDestroyedEnemies()
	}

	switch wave.Start.Kind {
	case types.WaveStartLeft:
		return alive <= wave.Start.Left
	case types.WaveStartClear:
		return alive == 0
	default:
		return true
	}
}

// isClassicBonusNumber — классическая нумерация носителей бонусов
// для уровней без явной разметки: 4, 9, 15, 22, …
// (каждый следующий интервал на единицу длиннее предыдущего)
func isClassicBonusNumber(enemyNumber uint) bool {
	bonusNumber := uint(4)
	step := uint(5)
	for bonusNumber < enemyNumber {
		bonusNumber += step
		step++
	}
	return bonusNumber == enemyNumber
}
