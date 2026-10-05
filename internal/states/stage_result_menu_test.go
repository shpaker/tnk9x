package states

import (
	"slices"
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

// Набор пунктов меню итогов на десктопе: без рекламы пункты те же,
// что и до появления площадок
func TestStageResultMenu_Items(t *testing.T) {
	next := types.StageResultItemNext
	carry := types.StageResultItemContinue
	retry := types.StageResultItemRetry
	levels := types.StageResultItemLevels

	tests := []struct {
		name string
		menu stageResultMenu
		want []types.StageResultItem
	}{
		{
			name: "победа, следующий открыт",
			menu: stageResultMenu{won: true, nextUnlocked: true},
			want: []types.StageResultItem{next, retry, levels},
		},
		{
			name: "победа с выгодным переносом",
			menu: stageResultMenu{
				won: true, nextUnlocked: true, carryOverAdvantage: true,
			},
			want: []types.StageResultItem{next, carry, retry, levels},
		},
		{
			name: "победа, следующего нет",
			menu: stageResultMenu{won: true, carryOverAdvantage: true},
			want: []types.StageResultItem{retry, levels},
		},
		{
			name: "поражение",
			menu: stageResultMenu{nextUnlocked: true, carryOverAdvantage: true},
			want: []types.StageResultItem{retry, levels},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.menu.items(); !slices.Equal(got, tt.want) {
				t.Errorf("items %v, want %v", got, tt.want)
			}
		})
	}
}

// Пункты за рекламу — только при rewarded у площадки и всегда после
// бесплатного пункта, выбранного по умолчанию
func TestStageResultMenu_RewardItems(t *testing.T) {
	next := types.StageResultItemNext
	carry := types.StageResultItemContinue
	retry := types.StageResultItemRetry
	levels := types.StageResultItemLevels
	revive := types.StageResultItemRevive
	boostNext := types.StageResultItemBoostNext
	boostRetry := types.StageResultItemBoostRetry

	tests := []struct {
		name string
		menu stageResultMenu
		want []types.StageResultItem
	}{
		{
			name: "победа",
			menu: stageResultMenu{
				won: true, nextUnlocked: true, carryOverAdvantage: true,
				rewardAvailable: true,
			},
			want: []types.StageResultItem{
				next,
				carry,
				boostNext,
				retry,
				levels,
			},
		},
		{
			name: "победа, следующего нет",
			menu: stageResultMenu{won: true, rewardAvailable: true},
			want: []types.StageResultItem{retry, boostRetry, levels},
		},
		{
			name: "поражение со вторым шансом",
			menu: stageResultMenu{rewardAvailable: true, canRevive: true},
			want: []types.StageResultItem{retry, revive, boostRetry, levels},
		},
		{
			name: "поражение без второго шанса",
			menu: stageResultMenu{rewardAvailable: true},
			want: []types.StageResultItem{retry, boostRetry, levels},
		},
		{
			name: "без rewarded второй шанс не предлагается",
			menu: stageResultMenu{canRevive: true},
			want: []types.StageResultItem{retry, levels},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.menu.items()
			if !slices.Equal(got, tt.want) {
				t.Errorf("items %v, want %v", got, tt.want)
			}
			if got[0].IsRewarded() {
				t.Error("the default item must be free")
			}
		})
	}
}
