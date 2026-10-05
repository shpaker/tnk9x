package states_test

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/tnk9x/internal/states"
	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

// recordingHelpRenderer запоминает последнюю отрисовку
type recordingHelpRenderer struct {
	view types.HelpViewData
}

func (r *recordingHelpRenderer) Draw(_ *ebiten.Image, view types.HelpViewData) {
	r.view = view
}

func TestHelpOverlay_FirstStageOnce(t *testing.T) {
	env := newMenusEnv()
	renderer := &recordingHelpRenderer{}
	overlay := states.NewHelpOverlay(
		newSettingsUseCases(),
		use_cases.NewControlsUseCases(memoryControlsRepository{}),
		renderer,
		env.input,
		env.settings,
		env.controls,
	)

	if overlay.IsDue(2) {
		t.Error("страница только перед уровнем 1")
	}
	if !overlay.IsDue(1) {
		t.Fatal("перед первым запуском уровня 1 страница показывается")
	}

	// Без нажатий страница остаётся
	env.frame(func() { overlay.Update(2) }, testutil.FakeMenuInput{})
	overlay.Draw(nil)
	if !overlay.IsDue(1) || renderer.view.Players != 2 ||
		len(renderer.view.Rows) == 0 {
		t.Fatalf("страница открыта, вид %+v", renderer.view)
	}

	// Подтверждение закрывает навсегда
	env.frame(
		func() { overlay.Update(2) },
		testutil.FakeMenuInput{Confirm: true},
	)
	if overlay.IsDue(1) || !env.settings.IsHelpShown() {
		t.Error("после закрытия страница больше не показывается")
	}
}
