//go:build !js

package platform

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

// Без площадки игра не приостанавливается и рекламы нет: все
// платформенные ветки общего кода на десктопе не срабатывают
func TestPlatformAdapter_Desktop(t *testing.T) {
	adapter := NewPlatformAdapter()

	adapter.Update()
	adapter.Ready()
	adapter.SetGameplayActive(true)
	adapter.RequestIntermission()
	adapter.RequestReward()

	if adapter.IsSuspended() || adapter.IsJustSuspended() {
		t.Error("desktop must never be suspended")
	}
	if adapter.IsRewardAvailable() {
		t.Error("desktop has no rewarded ads")
	}
	if got := adapter.PollReward(); got != types.RewardStatusNone {
		t.Errorf("PollReward %v, want None", got)
	}
}
