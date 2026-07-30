package use_cases

// Формула интервала спавна врагов из NES-версии:
// 190 - 4*этап - 20*(игроки-1) кадров, но не меньше минимума
const (
	waveSpawnDelayBase           = 190
	waveSpawnDelayPerStage       = 4
	waveSpawnDelayPerExtraPlayer = 20
	waveSpawnDelayMin            = 30
)

// WaveUseCases — расчёт параметров вражеских волн
type WaveUseCases struct{}

func NewWaveUseCases() *WaveUseCases {
	return &WaveUseCases{}
}

// SpawnDelayTicks возвращает интервал между спавнами врагов в тиках:
// чем дальше этап и чем больше игроков, тем быстрее подкрепление
func (uc *WaveUseCases) SpawnDelayTicks(stage, playerCount uint) uint {
	if stage == 0 {
		stage = 1
	}
	if playerCount == 0 {
		playerCount = 1
	}

	delay := waveSpawnDelayBase -
		waveSpawnDelayPerStage*int(stage) -
		waveSpawnDelayPerExtraPlayer*(int(playerCount)-1)
	if delay < waveSpawnDelayMin {
		delay = waveSpawnDelayMin
	}
	return uint(delay)
}
