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

// Первая значимая переменная локали; C и POSIX языка не задают
func TestLanguageFromEnv(t *testing.T) {
	cases := []struct {
		env  map[string]string
		want string
	}{
		{map[string]string{}, ""},
		{map[string]string{"LANG": "ru_RU.UTF-8"}, "ru_RU.UTF-8"},
		{map[string]string{"LANG": "C"}, ""},
		{
			map[string]string{"LC_ALL": "tr_TR.UTF-8", "LANG": "en_US"},
			"tr_TR.UTF-8",
		},
		{
			map[string]string{"LC_ALL": "POSIX", "LC_MESSAGES": "kk_KZ"},
			"kk_KZ",
		},
	}
	for _, c := range cases {
		lookup := func(name string) string { return c.env[name] }
		if got := languageFromEnv(lookup); got != c.want {
			t.Errorf("%v: %q, want %q", c.env, got, c.want)
		}
	}
}
