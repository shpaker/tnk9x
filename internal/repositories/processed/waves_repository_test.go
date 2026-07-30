package processed

import (
	"os"
	"testing"

	"github.com/shpaker/tnk9x/internal/repositories/raw"
)

const testWavesYML = `---
stages:
  1:
    - {tier: basic, count: 18}
    - {tier: fast, count: 2}
  2:
    - {tier: fast, count: 6}
    - {tier: power, count: 4}
    - {tier: armor, count: 10}
`

func newWavesTestRepo(data string) *WavesRepository {
	fileRepo := NewMockFileRepository()
	fileRepo.AddFile("levels/waves.yml", []byte(data))
	return NewWavesRepository(fileRepo)
}

// Группы разворачиваются в упорядоченный список уровней
func TestWavesRepository_GetWave(t *testing.T) {
	repo := newWavesTestRepo(testWavesYML)

	wave, err := repo.GetWave(1)
	if err != nil {
		t.Fatalf("волна этапа 1: %v", err)
	}
	if len(wave.Tiers) != 20 {
		t.Fatalf("размер волны %d, ожидалось 20", len(wave.Tiers))
	}
	for i := 0; i < 18; i++ {
		if wave.Tiers[i] != 0 {
			t.Fatalf("враг %d уровня %d, ожидался 0", i, wave.Tiers[i])
		}
	}
	for i := 18; i < 20; i++ {
		if wave.Tiers[i] != 1 {
			t.Fatalf("враг %d уровня %d, ожидался 1", i, wave.Tiers[i])
		}
	}

	wave, err = repo.GetWave(2)
	if err != nil {
		t.Fatalf("волна этапа 2: %v", err)
	}
	if wave.Tiers[0] != 1 || wave.Tiers[6] != 2 || wave.Tiers[10] != 3 {
		t.Errorf("порядок уровней волны 2: %v", wave.Tiers)
	}
}

func TestWavesRepository_GetWave_UnknownStage(t *testing.T) {
	repo := newWavesTestRepo(testWavesYML)

	if _, err := repo.GetWave(3); err == nil {
		t.Error("ожидалась ошибка для этапа без волны")
	}
}

// Волна неверного размера — ошибка загрузки
func TestWavesRepository_InvalidWaveSize(t *testing.T) {
	repo := newWavesTestRepo(`---
stages:
  1:
    - {tier: basic, count: 5}
`)

	if _, err := repo.GetWave(1); err == nil {
		t.Error("ожидалась ошибка для волны из 5 врагов")
	}
}

func TestWavesRepository_UnknownTier(t *testing.T) {
	repo := newWavesTestRepo(`---
stages:
  1:
    - {tier: mystery, count: 20}
`)

	if _, err := repo.GetWave(1); err == nil {
		t.Error("ожидалась ошибка для неизвестного типа врага")
	}
}

func TestWavesRepository_GetStages(t *testing.T) {
	repo := newWavesTestRepo(testWavesYML)

	stages, err := repo.GetStages()
	if err != nil {
		t.Fatalf("список этапов: %v", err)
	}
	if len(stages) != 2 || stages[0] != 1 || stages[1] != 2 {
		t.Errorf("этапы %v, ожидались [1 2]", stages)
	}
}

// Реальный waves.yml: волны заданы для всех 35 этапов
func TestWavesRepository_Integration(t *testing.T) {
	assetsPath := "assets"
	if _, err := os.Stat(assetsPath + "/levels/waves.yml"); os.IsNotExist(err) {
		assetsPath = "../../../assets"
	}
	if _, err := os.Stat(assetsPath + "/levels/waves.yml"); os.IsNotExist(err) {
		t.Skip("Пропуск интеграционного теста: waves.yml не найден")
	}

	repo := NewWavesRepository(raw.NewFileRepository(assetsPath))

	stages, err := repo.GetStages()
	if err != nil {
		t.Fatalf("загрузка waves.yml: %v", err)
	}
	if len(stages) != 35 {
		t.Fatalf("этапов с волнами %d, ожидалось 35", len(stages))
	}
	for stage := 1; stage <= 35; stage++ {
		if _, err := repo.GetWave(stage); err != nil {
			t.Errorf("этап %d: %v", stage, err)
		}
	}
}
