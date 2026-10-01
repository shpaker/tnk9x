package use_cases

import (
	"strconv"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.ISettingsUseCases = (*SettingsUseCases)(nil)

// SettingsUseCases — изменение пользовательских настроек
// с сохранением после каждого изменения
type SettingsUseCases struct {
	// Repositories
	settingsRepository interfaces.ISettingsRepository
	// items — строки экрана настроек; FULLSCREEN — только там, где
	// полный экран можно включить при старте (десктоп)
	items []types.SettingsItem
}

func NewSettingsUseCases(
	settingsRepository interfaces.ISettingsRepository,
	fullscreenAvailable bool,
) *SettingsUseCases {
	items := []types.SettingsItem{types.SettingsItemGraphics}
	if fullscreenAvailable {
		items = append(items, types.SettingsItemFullscreen)
	}
	items = append(
		items,
		types.SettingsItemVolume,
		types.SettingsItemControls,
		types.SettingsItemBack,
	)

	return &SettingsUseCases{
		settingsRepository: settingsRepository,
		items:              items,
	}
}

func (uc *SettingsUseCases) Items() []types.SettingsItem {
	return uc.items
}

// SetPlayers: число игроков ограничивается 1..MaxPlayers; без смены
// режима хранилище не трогается
func (uc *SettingsUseCases) SetPlayers(
	settings *types.SettingsEntity,
	players uint,
) error {
	previous := settings.GetPlayers()
	settings.SetPlayers(players)
	if settings.GetPlayers() == previous {
		return nil
	}
	return uc.settingsRepository.SaveSettings(settings)
}

// Change: флаги переключаются при любом направлении, громкость
// сдвигается на step шагов в пределах шкалы; без изменений
// настройки не сохраняются
func (uc *SettingsUseCases) Change(
	settings *types.SettingsEntity,
	item types.SettingsItem,
	step int,
) error {
	switch item {
	case types.SettingsItemGraphics:
		settings.SetEffectsEnabled(!settings.IsEffectsEnabled())
	case types.SettingsItemFullscreen:
		settings.SetFullscreen(!settings.IsFullscreen())
	case types.SettingsItemVolume:
		level := settings.GetVolumeLevel()
		settings.SetVolumeLevel(level + step)
		if settings.GetVolumeLevel() == level {
			return nil
		}
	default:
		return nil
	}

	return uc.settingsRepository.SaveSettings(settings)
}

func (uc *SettingsUseCases) BuildView(
	settings *types.SettingsEntity,
	activeIndex int,
) types.SettingsViewData {
	rows := make([]types.SettingsRow, len(uc.items))
	for i, item := range uc.items {
		rows[i] = types.SettingsRow{
			Item:  item,
			Value: settingsValue(settings, item),
		}
	}

	return types.SettingsViewData{
		Rows:        rows,
		ActiveIndex: activeIndex,
	}
}

// settingsValue — текущее значение пункта для экрана настроек
func settingsValue(
	settings *types.SettingsEntity,
	item types.SettingsItem,
) string {
	switch item {
	case types.SettingsItemGraphics:
		if settings.IsEffectsEnabled() {
			return "NORMAL"
		}
		return "CLASSIC"
	case types.SettingsItemFullscreen:
		if settings.IsFullscreen() {
			return "ON"
		}
		return "OFF"
	case types.SettingsItemVolume:
		percent := settings.GetVolumeLevel() * 100 / types.MaxVolumeLevel
		return strconv.Itoa(percent) + "%"
	default:
		return ""
	}
}
