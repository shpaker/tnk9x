package processed

import (
	"encoding/json"
	"fmt"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// controlsKey — ключ сохранения раскладки в хранилище
const controlsKey = "controls"

// controlsFormatVersion — версия формата сохранения
const controlsFormatVersion = 1

var _ interfaces.IControlsRepository = (*ControlsRepository)(nil)

// ControlsRepository сериализует раскладку управления в JSON
// поверх пользовательского хранилища
type ControlsRepository struct {
	storage interfaces.IStorageRepository
}

func NewControlsRepository(
	storage interfaces.IStorageRepository,
) *ControlsRepository {
	return &ControlsRepository{
		storage: storage,
	}
}

// Имена действий и хоткеев в JSON
var (
	actionKeys = [types.InputActionsCount]string{
		"up", "down", "left", "right", "fire",
	}
	hotkeyKeys = [types.HotkeysCount]string{"graphics", "fullscreen"}
)

// controlsSchema: игрок — клавиши и кнопки по действиям; пустое
// или отсутствующее имя оставляет значение по умолчанию
type controlsSchema struct {
	Version int                    `json:"version"`
	Players []playerControlsSchema `json:"players"`
	Hotkeys map[string]string      `json:"hotkeys"`
}

type playerControlsSchema struct {
	Keys    map[string]string `json:"keys"`
	Buttons map[string]string `json:"buttons"`
}

// GetControls читает раскладку; отсутствие сохранения — умолчания
func (cr *ControlsRepository) GetControls() (*types.ControlsEntity, error) {
	controls := types.NewControlsEntity()

	data, err := cr.storage.Load(controlsKey)
	if err != nil {
		return controls, err
	}
	if len(data) == 0 {
		return controls, nil
	}

	var schema controlsSchema
	if err := json.Unmarshal(data, &schema); err != nil {
		return controls, fmt.Errorf("malformed controls: %w", err)
	}
	if schema.Version != controlsFormatVersion {
		return controls, fmt.Errorf(
			"unsupported controls version %d", schema.Version,
		)
	}

	pages := []types.ControlsPage{
		types.ControlsPagePlayer1, types.ControlsPagePlayer2,
	}
	for i, player := range schema.Players {
		if i >= len(pages) {
			break
		}
		for action, key := range actionKeys {
			slot := types.ControlsSlot{
				Page: pages[i], Action: types.InputAction(action),
			}
			setIfPresent(controls, slot, player.Keys[key])
			slot.Device = types.ControlsDeviceGamepad
			setIfPresent(controls, slot, player.Buttons[key])
		}
	}
	for hotkey, key := range hotkeyKeys {
		setIfPresent(controls, types.ControlsSlot{
			Page:   types.ControlsPageHotkeys,
			Hotkey: types.HotkeyAction(hotkey),
		}, schema.Hotkeys[key])
	}
	return controls, nil
}

func (cr *ControlsRepository) SaveControls(
	controls *types.ControlsEntity,
) error {
	schema := controlsSchema{
		Version: controlsFormatVersion,
		Hotkeys: map[string]string{},
	}
	for _, player := range []types.PlayerTankNum{
		types.PlayerTankNumPlayer1, types.PlayerTankNumPlayer2,
	} {
		playerSchema := playerControlsSchema{
			Keys:    map[string]string{},
			Buttons: map[string]string{},
		}
		for action, key := range actionKeys {
			playerSchema.Keys[key] = controls.GetKey(
				player, types.InputAction(action),
			)
			playerSchema.Buttons[key] = controls.GetButton(
				player, types.InputAction(action),
			)
		}
		schema.Players = append(schema.Players, playerSchema)
	}
	for hotkey, key := range hotkeyKeys {
		schema.Hotkeys[key] = controls.GetHotkey(types.HotkeyAction(hotkey))
	}

	data, err := json.Marshal(schema)
	if err != nil {
		return err
	}
	return cr.storage.Save(controlsKey, data)
}

func setIfPresent(
	controls *types.ControlsEntity,
	slot types.ControlsSlot,
	name string,
) {
	if name != "" {
		controls.SetName(slot, name)
	}
}
