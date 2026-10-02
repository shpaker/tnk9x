package use_cases

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.ISettingsUseCases = (*SettingsUseCases)(nil)

// SettingsUseCases — изменение пользовательских настроек
// с сохранением после каждого изменения
type SettingsUseCases struct {
	// Repositories
	settingsRepository interfaces.ISettingsRepository
	// Adapters
	texts interfaces.ITextsAdapter
	// items — строки экрана настроек; FULLSCREEN — только там, где
	// полный экран можно включить при старте (десктоп)
	items []types.SettingsItem
	// languages — значения пункта LANGUAGE по кругу: AUTO и языки
	// конфигурации
	languages []types.Language
}

func NewSettingsUseCases(
	settingsRepository interfaces.ISettingsRepository,
	texts interfaces.ITextsAdapter,
	fullscreenAvailable bool,
	languages []types.Language,
) *SettingsUseCases {
	items := []types.SettingsItem{types.SettingsItemGraphics}
	if fullscreenAvailable {
		items = append(items, types.SettingsItemFullscreen)
	}
	items = append(
		items,
		types.SettingsItemVolume,
		types.SettingsItemLanguage,
		types.SettingsItemControls,
		types.SettingsItemBack,
	)

	return &SettingsUseCases{
		settingsRepository: settingsRepository,
		texts:              texts,
		items:              items,
		languages: append(
			[]types.Language{types.LanguageAuto}, languages...,
		),
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

// MarkHelpShown: повторно хранилище не трогается
func (uc *SettingsUseCases) MarkHelpShown(
	settings *types.SettingsEntity,
) error {
	if settings.IsHelpShown() {
		return nil
	}
	settings.SetHelpShown(true)
	return uc.settingsRepository.SaveSettings(settings)
}

// Change: флаги переключаются при любом направлении, громкость
// сдвигается на step шагов в пределах шкалы, язык — по кругу
// AUTO и языков конфигурации; без изменений настройки
// не сохраняются
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
	case types.SettingsItemLanguage:
		settings.SetLanguage(uc.nextLanguage(settings.GetLanguage(), step))
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
			Value: uc.settingsValue(settings, item),
		}
	}

	return types.SettingsViewData{
		Rows:        rows,
		ActiveIndex: activeIndex,
	}
}

// settingsValue — текущее значение пункта для экрана настроек
// на активном языке; язык показывается своим названием
func (uc *SettingsUseCases) settingsValue(
	settings *types.SettingsEntity,
	item types.SettingsItem,
) string {
	switch item {
	case types.SettingsItemGraphics:
		if settings.IsEffectsEnabled() {
			return uc.texts.Get(types.TextSettingsGraphicsNormal)
		}
		return uc.texts.Get(types.TextSettingsGraphicsClassic)
	case types.SettingsItemFullscreen:
		if settings.IsFullscreen() {
			return uc.texts.Get(types.TextSettingsOn)
		}
		return uc.texts.Get(types.TextSettingsOff)
	case types.SettingsItemVolume:
		percent := settings.GetVolumeLevel() * 100 / types.MaxVolumeLevel
		return uc.texts.Format(
			types.TextSettingsVolumePercent,
			types.TextArgs{"Percent": percent},
		)
	case types.SettingsItemLanguage:
		language := settings.GetLanguage()
		if uc.languageIndex(language) == 0 {
			return uc.texts.Get(types.TextSettingsLanguageAuto)
		}
		return uc.texts.GetLanguageName(language)
	default:
		return ""
	}
}

// nextLanguage — язык через step шагов по кругу AUTO и языков
// конфигурации
func (uc *SettingsUseCases) nextLanguage(
	language types.Language,
	step int,
) types.Language {
	count := len(uc.languages)
	index := ((uc.languageIndex(language)+step)%count + count) % count
	return uc.languages[index]
}

// languageIndex — позиция языка в круге значений; выбор, которого
// больше нет среди языков, — AUTO
func (uc *SettingsUseCases) languageIndex(language types.Language) int {
	for index, candidate := range uc.languages {
		if candidate == language {
			return index
		}
	}
	return 0
}
