package types_test

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

func TestSettingsEntity_Defaults(t *testing.T) {
	settings := types.NewSettingsEntity()

	if !settings.IsEffectsEnabled() || !settings.IsFullscreen() {
		t.Error("по умолчанию: обычная графика и полный экран")
	}
	if settings.GetVolume() != 0.5 {
		t.Errorf("громкость %v, ожидалась 0.5", settings.GetVolume())
	}
	if settings.GetPlayers() != 1 {
		t.Errorf("игроков %d, ожидался 1", settings.GetPlayers())
	}
}

func TestSettingsEntity_PlayersClamp(t *testing.T) {
	settings := types.NewSettingsEntity()

	settings.SetPlayers(5)
	if settings.GetPlayers() != types.MaxPlayers {
		t.Errorf(
			"игроков %d, ожидалось %d",
			settings.GetPlayers(),
			types.MaxPlayers,
		)
	}
	settings.SetPlayers(0)
	if settings.GetPlayers() != 1 {
		t.Errorf("игроков %d, ожидался 1", settings.GetPlayers())
	}
}

func TestSettingsEntity_VolumeClamp(t *testing.T) {
	settings := types.NewSettingsEntity()

	settings.SetVolumeLevel(types.MaxVolumeLevel + 3)
	if settings.GetVolumeLevel() != types.MaxVolumeLevel ||
		settings.GetVolume() != 1 {
		t.Errorf("шаг %d, ожидался максимум", settings.GetVolumeLevel())
	}

	settings.SetVolumeLevel(-1)
	if settings.GetVolumeLevel() != 0 || settings.GetVolume() != 0 {
		t.Errorf("шаг %d, ожидался 0", settings.GetVolumeLevel())
	}
}
