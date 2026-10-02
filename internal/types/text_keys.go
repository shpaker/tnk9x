package types

import "strconv"

// TextKey — идентификатор текста интерфейса в файлах локалей
// assets/locales/<язык>.yml: путь вложенных ключей через точку
type TextKey string

// TextArgs — именованные параметры шаблона текста ({{.Count}})
type TextArgs map[string]any

// Название языка на нём самом: ENGLISH, РУССКИЙ, TÜRKÇE
const TextLanguageName TextKey = "language.name"

// Главное меню
const (
	TextMainMenuOnePlayer  TextKey = "main_menu.one_player"
	TextMainMenuTwoPlayers TextKey = "main_menu.two_players"
	TextMainMenuSettings   TextKey = "main_menu.settings"
	TextMainMenuQuit       TextKey = "main_menu.quit"
)

// Меню паузы
const (
	TextPauseTitle        TextKey = "pause.title"
	TextPauseContinue     TextKey = "pause.continue"
	TextPauseRestart      TextKey = "pause.restart"
	TextPauseSettings     TextKey = "pause.settings"
	TextPauseExitToLevels TextKey = "pause.exit_to_levels"
)

// Экран настроек; TextSettingsVolumePercent — {{.Percent}}
const (
	TextSettingsTitle           TextKey = "settings.title"
	TextSettingsGraphics        TextKey = "settings.graphics"
	TextSettingsFullscreen      TextKey = "settings.fullscreen"
	TextSettingsVolume          TextKey = "settings.volume"
	TextSettingsLanguage        TextKey = "settings.language"
	TextSettingsControls        TextKey = "settings.controls"
	TextSettingsBack            TextKey = "settings.back"
	TextSettingsGraphicsNormal  TextKey = "settings.graphics_normal"
	TextSettingsGraphicsClassic TextKey = "settings.graphics_classic"
	TextSettingsOn              TextKey = "settings.on"
	TextSettingsOff             TextKey = "settings.off"
	TextSettingsVolumePercent   TextKey = "settings.volume_percent"
	TextSettingsLanguageAuto    TextKey = "settings.language_auto"
)

// Экран раскладки
const (
	TextControlsTitle            TextKey = "controls.title"
	TextControlsPlayer1          TextKey = "controls.player_1"
	TextControlsPlayer2          TextKey = "controls.player_2"
	TextControlsHotkeys          TextKey = "controls.hotkeys"
	TextControlsUp               TextKey = "controls.up"
	TextControlsDown             TextKey = "controls.down"
	TextControlsLeft             TextKey = "controls.left"
	TextControlsRight            TextKey = "controls.right"
	TextControlsFire             TextKey = "controls.fire"
	TextControlsHotkeyGraphics   TextKey = "controls.hotkey_graphics"
	TextControlsHotkeyFullscreen TextKey = "controls.hotkey_fullscreen"
	TextControlsKeyColumn        TextKey = "controls.key_column"
	TextControlsPadColumn        TextKey = "controls.pad_column"
	TextControlsResetAll         TextKey = "controls.reset_all"
	TextControlsBack             TextKey = "controls.back"
	TextControlsPressKey         TextKey = "controls.press_key"
	TextControlsPressButton      TextKey = "controls.press_button"
	TextControlsAlreadyTaken     TextKey = "controls.already_taken"
)

// Подписи клавиш-стрелок в ячейках раскладки
const (
	TextKeysUp    TextKey = "keys.up"
	TextKeysDown  TextKey = "keys.down"
	TextKeysLeft  TextKey = "keys.left"
	TextKeysRight TextKey = "keys.right"
)

// Экран выбора уровня; {{.Stage}} — номер уровня в два знака,
// {{.Button}} — кнопка запуска
const (
	TextLevelSelectLocked        TextKey = "level_select.locked"
	TextLevelSelectCollect       TextKey = "level_select.collect"
	TextLevelSelectWinStage      TextKey = "level_select.win_stage"
	TextLevelSelectEnemies       TextKey = "level_select.enemies"
	TextLevelSelectWinPrevious   TextKey = "level_select.win_previous"
	TextLevelSelectBest          TextKey = "level_select.best"
	TextLevelSelectKeyboardHint  TextKey = "level_select.keyboard_hint"
	TextLevelSelectMainMenu      TextKey = "level_select.main_menu"
	TextLevelSelectStartOne      TextKey = "level_select.start_one"
	TextLevelSelectStartTwo      TextKey = "level_select.start_two"
	TextLevelSelectButtonEnter   TextKey = "level_select.button_enter"
	TextLevelSelectButtonFire    TextKey = "level_select.button_fire"
	TextLevelSelectStageFallback TextKey = "level_select.stage_fallback"
)

// Экран итогов уровня; TextResultStats — {{.Time}} и {{.Lives}},
// TextResultCarryOver — плюрал по {{.Count}}
const (
	TextResultVictory      TextKey = "result.victory"
	TextResultDefeat       TextKey = "result.defeat"
	TextResultNextStage    TextKey = "result.next_stage"
	TextResultContinue     TextKey = "result.continue"
	TextResultRetry        TextKey = "result.retry"
	TextResultStages       TextKey = "result.stages"
	TextResultRevive       TextKey = "result.revive"
	TextResultBoostNext    TextKey = "result.boost_next"
	TextResultBoostRetry   TextKey = "result.boost_retry"
	TextResultKeepCaption  TextKey = "result.keep_caption"
	TextResultBoostCaption TextKey = "result.boost_caption"
	TextResultAd           TextKey = "result.ad"
	TextResultStats        TextKey = "result.stats"
	TextResultNewBest      TextKey = "result.new_best"
	TextResultCarryOver    TextKey = "result.carry_over"
)

// TextKeys — все постоянные ключи: каждый файл локали содержит их все
var TextKeys = []TextKey{
	TextLanguageName,

	TextMainMenuOnePlayer,
	TextMainMenuTwoPlayers,
	TextMainMenuSettings,
	TextMainMenuQuit,

	TextPauseTitle,
	TextPauseContinue,
	TextPauseRestart,
	TextPauseSettings,
	TextPauseExitToLevels,

	TextSettingsTitle,
	TextSettingsGraphics,
	TextSettingsFullscreen,
	TextSettingsVolume,
	TextSettingsLanguage,
	TextSettingsControls,
	TextSettingsBack,
	TextSettingsGraphicsNormal,
	TextSettingsGraphicsClassic,
	TextSettingsOn,
	TextSettingsOff,
	TextSettingsVolumePercent,
	TextSettingsLanguageAuto,

	TextControlsTitle,
	TextControlsPlayer1,
	TextControlsPlayer2,
	TextControlsHotkeys,
	TextControlsUp,
	TextControlsDown,
	TextControlsLeft,
	TextControlsRight,
	TextControlsFire,
	TextControlsHotkeyGraphics,
	TextControlsHotkeyFullscreen,
	TextControlsKeyColumn,
	TextControlsPadColumn,
	TextControlsResetAll,
	TextControlsBack,
	TextControlsPressKey,
	TextControlsPressButton,
	TextControlsAlreadyTaken,

	TextKeysUp,
	TextKeysDown,
	TextKeysLeft,
	TextKeysRight,

	TextLevelSelectLocked,
	TextLevelSelectCollect,
	TextLevelSelectWinStage,
	TextLevelSelectEnemies,
	TextLevelSelectWinPrevious,
	TextLevelSelectBest,
	TextLevelSelectKeyboardHint,
	TextLevelSelectMainMenu,
	TextLevelSelectStartOne,
	TextLevelSelectStartTwo,
	TextLevelSelectButtonEnter,
	TextLevelSelectButtonFire,
	TextLevelSelectStageFallback,

	TextResultVictory,
	TextResultDefeat,
	TextResultNextStage,
	TextResultContinue,
	TextResultRetry,
	TextResultStages,
	TextResultRevive,
	TextResultBoostNext,
	TextResultBoostRetry,
	TextResultKeepCaption,
	TextResultBoostCaption,
	TextResultAd,
	TextResultStats,
	TextResultNewBest,
	TextResultCarryOver,
}

// LevelNameTextKey — перевод названия уровня кампании: levels.<номер>;
// без перевода показывается название из файла карты
func LevelNameTextKey(number int) TextKey {
	return TextKey("levels." + strconv.Itoa(number))
}

// PackNameTextKey — перевод названия пачки кампании по её порядковому
// номеру с 1: packs.<номер>; без перевода — название из файла кампании
func PackNameTextKey(number int) TextKey {
	return TextKey("packs." + strconv.Itoa(number))
}
