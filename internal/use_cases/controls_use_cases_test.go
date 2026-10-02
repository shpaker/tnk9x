package use_cases_test

import (
	"errors"
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

// recordingControlsRepository считает сохранения раскладки
type recordingControlsRepository struct {
	saves int
}

func (r *recordingControlsRepository) GetControls() (
	*types.ControlsEntity,
	error,
) {
	return types.NewControlsEntity(), nil
}

func (r *recordingControlsRepository) SaveControls(
	*types.ControlsEntity,
) error {
	r.saves++
	return nil
}

func newControls() (
	*use_cases.ControlsUseCases,
	*recordingControlsRepository,
	*types.ControlsEntity,
) {
	repository := &recordingControlsRepository{}
	return use_cases.NewControlsUseCases(repository),
		repository,
		types.NewControlsEntity()
}

func keySlot(
	page types.ControlsPage,
	action types.InputAction,
) types.ControlsSlot {
	return types.ControlsSlot{Page: page, Action: action}
}

func TestControls_BindFreeKey(t *testing.T) {
	controlsUseCases, repository, controls := newControls()
	slot := keySlot(types.ControlsPagePlayer1, types.InputActionFire)

	if err := controlsUseCases.Bind(controls, slot, "Space"); err != nil {
		t.Fatalf("свободная клавиша: %v", err)
	}
	if controls.GetKey(
		types.PlayerTankNumPlayer1,
		types.InputActionFire,
	) != "Space" ||
		repository.saves != 1 {
		t.Error("клавиша должна назначиться и сохраниться")
	}

	// Повторное назначение той же клавиши ничего не пишет
	_ = controlsUseCases.Bind(controls, slot, "Space")
	if repository.saves != 1 {
		t.Errorf("сохранений %d, ожидалось 1", repository.saves)
	}
}

func TestControls_BindTakenKey(t *testing.T) {
	controlsUseCases, repository, controls := newControls()
	hotkey := types.ControlsSlot{
		Page: types.ControlsPageHotkeys, Hotkey: types.HotkeyGraphics,
	}

	cases := []struct {
		name string
		slot types.ControlsSlot
		key  string
	}{
		{
			"своя клавиша",
			keySlot(types.ControlsPagePlayer1, types.InputActionUp),
			"S",
		},
		{
			"клавиша P2",
			keySlot(types.ControlsPagePlayer1, types.InputActionUp),
			"I",
		},
		{
			"хоткей",
			keySlot(types.ControlsPagePlayer2, types.InputActionFire),
			"F2",
		},
		{
			"Esc",
			keySlot(types.ControlsPagePlayer1, types.InputActionFire),
			"Escape",
		},
		{"клавиша меню у хоткея", hotkey, "Enter"},
	}
	for _, tc := range cases {
		err := controlsUseCases.Bind(controls, tc.slot, tc.key)
		if !errors.Is(err, types.ErrBindingTaken) {
			t.Errorf("%s: ошибка %v, ожидалась ErrBindingTaken", tc.name, err)
		}
	}
	if repository.saves != 0 {
		t.Errorf(
			"отказы не должны сохраняться, сохранений %d",
			repository.saves,
		)
	}
}

// У каждого игрока свой геймпад: одинаковые кнопки у P1 и P2 разрешены
func TestControls_BindPadButton(t *testing.T) {
	controlsUseCases, _, controls := newControls()
	fire := types.ControlsSlot{
		Page:   types.ControlsPagePlayer2,
		Action: types.InputActionFire,
		Device: types.ControlsDeviceGamepad,
	}

	if err := controlsUseCases.Bind(controls, fire, "B"); err != nil {
		t.Fatalf("свободная кнопка: %v", err)
	}
	if err := controlsUseCases.Bind(controls, fire, "D-UP"); !errors.Is(
		err, types.ErrBindingTaken,
	) {
		t.Error("кнопка движения того же геймпада занята")
	}
	if err := controlsUseCases.Bind(controls, fire, "START"); !errors.Is(
		err, types.ErrBindingTaken,
	) {
		t.Error("Start зарезервирован")
	}

	p1Fire := fire
	p1Fire.Page = types.ControlsPagePlayer1
	if err := controlsUseCases.Bind(controls, p1Fire, "B"); err != nil {
		t.Errorf("B на геймпаде P1 свободна: %v", err)
	}
}

func TestControls_ResetAll(t *testing.T) {
	controlsUseCases, repository, controls := newControls()
	_ = controlsUseCases.Bind(
		controls,
		keySlot(types.ControlsPagePlayer1, types.InputActionFire),
		"Space",
	)

	if err := controlsUseCases.ResetAll(controls); err != nil {
		t.Fatal(err)
	}
	if controls.GetKey(
		types.PlayerTankNumPlayer1,
		types.InputActionFire,
	) != "G" ||
		repository.saves != 2 {
		t.Error("сброс должен вернуть умолчания и сохраниться")
	}
}

func TestControls_RowsAndView(t *testing.T) {
	controlsUseCases, _, controls := newControls()

	player := controlsUseCases.BuildView(controls, types.ControlsCursor{
		Page: types.ControlsPagePlayer2, Row: 3,
	})
	// Заголовок, пять действий, сброс, назад
	if len(player.Rows) != 8 || player.Cursor.Row != 3 {
		t.Fatalf("строк %d, курсор %d", len(player.Rows), player.Cursor.Row)
	}
	up := player.Rows[1]
	if up.Kind != types.ControlsRowAction || up.Key != "I" ||
		up.Button != "D-UP" {
		t.Errorf("строка UP P2: %+v", up)
	}

	hotkeys := controlsUseCases.BuildView(controls, types.ControlsCursor{
		Page: types.ControlsPageHotkeys,
	})
	if len(hotkeys.Rows) != 5 || hotkeys.Rows[2].Key != "F11" {
		t.Errorf("страница хоткеев: %+v", hotkeys.Rows)
	}
}

func TestControls_HelpRows(t *testing.T) {
	useCases, _, controls := newControls()

	rows := useCases.HelpRows(controls)
	if len(rows) != int(types.InputActionsCount)+1 {
		t.Fatalf("строк %d", len(rows))
	}
	fire := rows[types.InputActionFire]
	if fire.Kind != types.HelpRowAction ||
		fire.Keys != [types.MaxPlayers]string{"G", "Quote"} ||
		fire.Button != "A" {
		t.Errorf("огонь %+v", fire)
	}
	pause := rows[len(rows)-1]
	if pause.Kind != types.HelpRowPause ||
		pause.Keys[0] != types.ReservedKey ||
		pause.Button != types.ReservedButton {
		t.Errorf("пауза %+v", pause)
	}
}
