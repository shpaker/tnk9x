package use_cases_test

import (
	"errors"
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

// recordingProgressRepository считает сохранения прогресса
type recordingProgressRepository struct {
	saves int
	err   error
}

func (r *recordingProgressRepository) GetProgress() (*types.ProgressEntity, error) {
	return types.NewProgressEntity(), nil
}

func (r *recordingProgressRepository) SaveProgress(
	progress *types.ProgressEntity,
) error {
	r.saves++
	return r.err
}

// testCampaign: пачка 1 — уровни 1-3 по порядку; пачка 2 — уровни 4-5
// сразу целиком после победы на 3 и 5 звёзд
func testCampaign() *types.CampaignEntity {
	return types.NewCampaignEntity("TEST", []types.PackSpec{
		{Name: "ONE", Levels: []int{1, 2, 3}},
		{
			Name:        "TWO",
			Levels:      []int{4, 5},
			Order:       types.PackOrderAny,
			UnlockStars: 5,
			UnlockAfter: 3,
		},
	})
}

func newProgression() (
	*use_cases.ProgressionUseCases,
	*recordingProgressRepository,
) {
	return newProgressionWithInventory(types.NewInventoryEntity())
}

func newProgressionWithInventory(inventory *types.InventoryEntity) (
	*use_cases.ProgressionUseCases,
	*recordingProgressRepository,
) {
	repository := &recordingProgressRepository{}
	return use_cases.NewProgressionUseCases(
		testCampaign(),
		types.NewProgressEntity(),
		inventory,
		repository,
	), repository
}

func TestProgression_CalcStars(t *testing.T) {
	progression, _ := newProgression()
	level := types.NewLevelEntity(1, "L", 4, 600, nil, false, nil)

	tests := []struct {
		name   string
		result types.StageResult
		want   uint
	}{
		{"defeat", types.StageResult{Won: false}, 0},
		{"lives lost", types.StageResult{Won: true, LivesLost: 1}, 1},
		{"too slow", types.StageResult{Won: true, ElapsedTicks: 601}, 2},
		{"perfect", types.StageResult{Won: true, ElapsedTicks: 600}, 3},
		{
			"carried over",
			types.StageResult{Won: true, ElapsedTicks: 600, CarriedOver: true},
			2,
		},
		{
			"carried over lives lost",
			types.StageResult{Won: true, LivesLost: 1, CarriedOver: true},
			1,
		},
	}
	for _, tt := range tests {
		if got := progression.CalcStars(tt.result, level); got != tt.want {
			t.Errorf("%s: %d stars, want %d", tt.name, got, tt.want)
		}
	}
}

// Сохраняется только улучшение результата
func TestProgression_RecordResult(t *testing.T) {
	progression, repository := newProgression()

	if err := progression.RecordResult(1, 2); err != nil {
		t.Fatalf("record: %v", err)
	}
	_ = progression.RecordResult(1, 1)
	_ = progression.RecordResult(1, 3)

	if got := progression.GetLevelStars(1); got != 3 {
		t.Errorf("stars %d, want 3", got)
	}
	if repository.saves != 2 {
		t.Errorf("saves %d, want 2 (only improvements)", repository.saves)
	}

	repository.err = errors.New("disk full")
	if err := progression.RecordResult(2, 1); err == nil {
		t.Error("storage error must be reported")
	}
}

func TestProgression_Unlocking(t *testing.T) {
	progression, _ := newProgression()

	// Новый прогресс: открыт только первый уровень
	for level, want := range map[int]bool{1: true, 2: false, 4: false, 9: false} {
		if got := progression.IsLevelUnlocked(level); got != want {
			t.Errorf(
				"new save: level %d unlocked=%v, want %v",
				level,
				got,
				want,
			)
		}
	}

	_ = progression.RecordResult(1, 1)
	_ = progression.RecordResult(2, 1)
	_ = progression.RecordResult(3, 1)
	if !progression.IsLevelUnlocked(3) {
		t.Error("level 3 must open after level 2")
	}
	// Уровень 3 пройден, но звёзд 3 из 5 — вторая пачка закрыта
	if progression.IsLevelUnlocked(4) {
		t.Error("pack two must need 5 stars")
	}
	status := progression.GetPackStatus(1)
	if status.Unlocked || status.TotalStars != 3 || status.RequiredStars != 5 {
		t.Errorf("pack two status %+v", status)
	}

	_ = progression.RecordResult(1, 3)
	// Порядок any: вся пачка открывается разом
	if !progression.IsLevelUnlocked(4) || !progression.IsLevelUnlocked(5) {
		t.Error("pack two must open whole with 5 stars")
	}
	if status := progression.GetPackStatus(0); status.Stars != 5 ||
		status.MaxStars != 9 {
		t.Errorf("pack one status %+v", status)
	}
	if status := progression.GetPackStatus(7); status.Unlocked {
		t.Error("unknown pack must be locked")
	}
}

func TestProgression_NextLevel(t *testing.T) {
	progression, _ := newProgression()

	if next, unlocked := progression.NextLevel(1); next != 2 || unlocked {
		t.Errorf("next after 1: %d %v, want 2 locked", next, unlocked)
	}
	_ = progression.RecordResult(1, 1)
	if next, unlocked := progression.NextLevel(1); next != 2 || !unlocked {
		t.Errorf("next after won 1: %d %v, want 2 unlocked", next, unlocked)
	}
	if _, unlocked := progression.NextLevel(5); unlocked {
		t.Error("the last level has no next level")
	}
}

// Купленная пачка открыта целиком без звёзд и условий, даже
// с порядком прохождения по одному
func TestProgression_PurchasedPack(t *testing.T) {
	inventory := types.NewInventoryEntity()
	inventory.SetPackOwned(1)
	inventory.SetPackOwned(2)
	progression, _ := newProgressionWithInventory(inventory)

	for _, level := range []int{1, 2, 3, 4, 5} {
		if !progression.IsLevelUnlocked(level) {
			t.Errorf("level %d must be unlocked", level)
		}
	}
	if !progression.GetPackStatus(1).Unlocked {
		t.Error("purchased pack must be unlocked")
	}
}

// Все уровни открыты покупкой всех уровней
func TestProgression_AllLevelsPurchased(t *testing.T) {
	inventory := types.NewInventoryEntity()
	inventory.SetAllLevels(true)
	progression, _ := newProgressionWithInventory(inventory)

	for _, level := range []int{1, 2, 3, 4, 5} {
		if !progression.IsLevelUnlocked(level) {
			t.Errorf("level %d must be unlocked", level)
		}
	}
}
