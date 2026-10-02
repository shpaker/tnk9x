package use_cases

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.IControlsUseCases = (*ControlsUseCases)(nil)

// menuKeys — клавиши меню: хоткеи на них сработали бы при каждом
// шаге по меню, поэтому хоткею их назначить нельзя
var menuKeys = []string{
	"Enter", "Space",
	"ArrowUp", "ArrowDown", "ArrowLeft", "ArrowRight",
	"W", "A", "S", "D",
}

// ControlsUseCases — назначение клавиш и кнопок с запретом занятых
// имён и сохранением раскладки после каждого изменения
type ControlsUseCases struct {
	// Repositories
	controlsRepository interfaces.IControlsRepository
}

func NewControlsUseCases(
	controlsRepository interfaces.IControlsRepository,
) *ControlsUseCases {
	return &ControlsUseCases{
		controlsRepository: controlsRepository,
	}
}

// Rows: заголовок страницы, действия игрока или хоткеи, сброс, назад
func (uc *ControlsUseCases) Rows(page types.ControlsPage) []types.ControlsRow {
	rows := []types.ControlsRow{{Kind: types.ControlsRowPage}}
	if _, isPlayer := page.Player(); isPlayer {
		for action := range types.InputActionsCount {
			rows = append(rows, types.ControlsRow{
				Kind:   types.ControlsRowAction,
				Action: action,
			})
		}
	} else {
		for hotkey := range types.HotkeysCount {
			rows = append(rows, types.ControlsRow{
				Kind:   types.ControlsRowHotkey,
				Hotkey: hotkey,
			})
		}
	}
	return append(
		rows,
		types.ControlsRow{Kind: types.ControlsRowReset},
		types.ControlsRow{Kind: types.ControlsRowBack},
	)
}

// Bind: клавиатура общая для обоих игроков и хоткеев, поэтому
// клавиша проверяется по всем ячейкам; у каждого игрока свой
// геймпад, кнопка проверяется только среди его кнопок
func (uc *ControlsUseCases) Bind(
	controls *types.ControlsEntity,
	slot types.ControlsSlot,
	name string,
) error {
	if name == "" || uc.isTaken(controls, slot, name) {
		return types.ErrBindingTaken
	}
	if controls.GetName(slot) == name {
		return nil
	}

	controls.SetName(slot, name)
	return uc.controlsRepository.SaveControls(controls)
}

func (uc *ControlsUseCases) ResetAll(controls *types.ControlsEntity) error {
	controls.Reset()
	return uc.controlsRepository.SaveControls(controls)
}

func (uc *ControlsUseCases) BuildView(
	controls *types.ControlsEntity,
	cursor types.ControlsCursor,
) types.ControlsViewData {
	rows := uc.Rows(cursor.Page)
	player, _ := cursor.Page.Player()
	for i, row := range rows {
		switch row.Kind {
		case types.ControlsRowAction:
			rows[i].Key = controls.GetKey(player, row.Action)
			rows[i].Button = controls.GetButton(player, row.Action)
		case types.ControlsRowHotkey:
			rows[i].Key = controls.GetHotkey(row.Hotkey)
		}
	}

	return types.ControlsViewData{
		Cursor: cursor,
		Rows:   rows,
	}
}

// HelpRows: движение и огонь по раскладке игроков, геймпад — кнопки
// первого игрока, пауза — зарезервированные Esc и Start
func (uc *ControlsUseCases) HelpRows(
	controls *types.ControlsEntity,
) []types.HelpRow {
	rows := make([]types.HelpRow, 0, types.InputActionsCount+1)
	for action := range types.InputActionsCount {
		rows = append(rows, types.HelpRow{
			Kind:   types.HelpRowAction,
			Action: action,
			Keys: [types.MaxPlayers]string{
				controls.GetKey(types.PlayerTankNumPlayer1, action),
				controls.GetKey(types.PlayerTankNumPlayer2, action),
			},
			Button: controls.GetButton(types.PlayerTankNumPlayer1, action),
		})
	}
	return append(rows, types.HelpRow{
		Kind:   types.HelpRowPause,
		Keys:   [types.MaxPlayers]string{types.ReservedKey, types.ReservedKey},
		Button: types.ReservedButton,
	})
}

// isTaken — имя зарезервировано или назначено другой ячейке
func (uc *ControlsUseCases) isTaken(
	controls *types.ControlsEntity,
	slot types.ControlsSlot,
	name string,
) bool {
	if slot.Device == types.ControlsDeviceGamepad {
		if name == types.ReservedButton {
			return true
		}
		player, _ := slot.Page.Player()
		for action := range types.InputActionsCount {
			if action != slot.Action &&
				controls.GetButton(player, action) == name {
				return true
			}
		}
		return false
	}

	if name == types.ReservedKey {
		return true
	}
	if slot.Page == types.ControlsPageHotkeys {
		for _, menuKey := range menuKeys {
			if name == menuKey {
				return true
			}
		}
	}
	for _, other := range controls.KeySlots() {
		if other != slot && controls.GetName(other) == name {
			return true
		}
	}
	return false
}
