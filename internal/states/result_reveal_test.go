package states

import "testing"

func TestResultReveal_Stages(t *testing.T) {
	const stars = 3

	start := resultReveal(0, stars)
	if start.Backdrop != 0 || start.Title != 0 || start.Stars != 0 || start.Menu {
		t.Errorf("на старте ничего не видно: %+v", start)
	}

	afterTitle := resultReveal(revealBackdropTicks+revealTitleTicks, stars)
	if afterTitle.Backdrop != 1 || afterTitle.Title != 1 || afterTitle.Stars != 1 {
		t.Errorf("после заголовка загорается первая звезда: %+v", afterTitle)
	}

	beforeMenu := resultReveal(resultRevealLength(stars)-1, stars)
	if beforeMenu.Stars != stars || beforeMenu.Menu {
		t.Errorf("перед меню горят все звёзды, меню ещё нет: %+v", beforeMenu)
	}

	done := resultReveal(resultRevealLength(stars), stars)
	if !done.Menu || done.Stats != 1 {
		t.Errorf("в конце видны статистика и меню: %+v", done)
	}
}

// При поражении звёзд нет — меню появляется раньше
func TestResultReveal_NoStars(t *testing.T) {
	length := resultRevealLength(0)
	if length >= resultRevealLength(1) {
		t.Fatal("без звёзд появление короче")
	}
	if reveal := resultReveal(length, 0); reveal.Stars != 0 || !reveal.Menu {
		t.Errorf("без звёзд: %+v", reveal)
	}
}
