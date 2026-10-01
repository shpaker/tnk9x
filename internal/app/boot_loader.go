package app

import "github.com/shpaker/tnk9x/internal/states"

var _ states.Loader = (*bootLoader)(nil)

// bootLoader — загрузка ресурсов игры на сплеше по шагу за вызов:
// шаги достраивают граф App, ошибка ресурса фатальна, как при старте
type bootLoader struct {
	steps []func()
	next  int
}

func newBootLoader(steps ...func()) *bootLoader {
	return &bootLoader{steps: steps}
}

// Step выполняет очередной шаг и возвращает долю готового
func (l *bootLoader) Step() (float64, bool) {
	if l.next < len(l.steps) {
		l.steps[l.next]()
		l.next++
	}
	return float64(l.next) / float64(len(l.steps)), l.next == len(l.steps)
}
