package use_cases

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
)

var _ interfaces.IWaveUseCases = (*WaveUseCases)(nil)

// coopSpawnDelayReductionTicks — насколько короче пауза между спаунами
// при игре вдвоём, как в оригинале
const coopSpawnDelayReductionTicks = 20

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
		if uc.wavesExhausted(session) {
			return session.PeekExtraEnemy()
		}
		return types.WaveTank{}, false
	}

	tank := waves[waveIndex].Tanks[tankIndex]
	if !session.GetLevel().HasExplicitBonuses() {
		tank.HasBonus = isBonusCarrierNumber(session.GetSpawnedEnemies() + 1)
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
		if _, extra := session.PeekExtraEnemy(); extra &&
			uc.wavesExhausted(session) {
			session.PopExtraEnemy()
			session.RegisterEnemySpawned(
				spawnerIndex,
				coopSpawnDelay(session, lastWaveDelay(session)),
			)
		}
		return
	}

	session.SetWaveCursor(waveIndex, tankIndex+1)
	session.RegisterEnemySpawned(
		spawnerIndex,
		coopSpawnDelay(session, waves[waveIndex].DelayTicks),
	)
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

// wavesExhausted — все танки сценария уже вышли на поле:
// дальше идут только дополнительные враги
func (uc *WaveUseCases) wavesExhausted(
	session *session_entities.StageSessionEntity,
) bool {
	level := session.GetLevel()
	if level == nil {
		return false
	}
	waves := level.GetWaves()
	waveIndex, tankIndex := session.GetWaveCursor()
	return waveIndex >= len(waves) ||
		(waveIndex == len(waves)-1 && tankIndex >= len(waves[waveIndex].Tanks))
}

// lastWaveDelay — пауза между спаунами последней волны: с ней выходят
// дополнительные враги
func lastWaveDelay(session *session_entities.StageSessionEntity) uint {
	waves := session.GetLevel().GetWaves()
	if len(waves) == 0 {
		return 0
	}
	return waves[len(waves)-1].DelayTicks
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

// isBonusCarrierNumber — нумерация носителей бонусов по умолчанию
// для уровней без явной разметки: 4-й, 11-й и 18-й танки, как в оригинале
func isBonusCarrierNumber(enemyNumber uint) bool {
	return enemyNumber == 4 || enemyNumber == 11 || enemyNumber == 18
}

// coopSpawnDelay — пауза волны с учётом кооператива:
// во втором игроке враги выходят чаще
func coopSpawnDelay(
	session *session_entities.StageSessionEntity,
	delay uint,
) uint {
	if session.GetPlayerCount() <= 1 {
		return delay
	}
	return max(
		delay,
		coopSpawnDelayReductionTicks,
	) - coopSpawnDelayReductionTicks
}
