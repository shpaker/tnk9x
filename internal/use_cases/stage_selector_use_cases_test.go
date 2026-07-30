package use_cases_test

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

type stageSelectorTestEnv struct {
	selector   *types.StageSelectorEntity
	selectorUC *use_cases.StageSelectorUseCases
}

func newStageSelectorTestEnv(maxStages uint) *stageSelectorTestEnv {
	return &stageSelectorTestEnv{
		selector:   types.NewStageSelector(maxStages),
		selectorUC: use_cases.NewStageSelectorUseCases(),
	}
}

// Next идёт вперёд и с последнего уровня возвращается на первый
func TestStageSelectorUseCases_NextWrapsAround(t *testing.T) {
	env := newStageSelectorTestEnv(3)

	for i, want := range []uint{2, 3, 1, 2} {
		if got := env.selectorUC.Next(env.selector); got != want {
			t.Fatalf("шаг %d: уровень %d, ожидался %d", i, got, want)
		}
		if env.selector.CurrentStage != want {
			t.Fatalf("шаг %d: селектор %d", i, env.selector.CurrentStage)
		}
	}
}

// Previous идёт назад и с первого уровня переходит на последний
func TestStageSelectorUseCases_PreviousWrapsAround(t *testing.T) {
	env := newStageSelectorTestEnv(3)

	for i, want := range []uint{3, 2, 1, 3} {
		if got := env.selectorUC.Previous(env.selector); got != want {
			t.Fatalf("шаг %d: уровень %d, ожидался %d", i, got, want)
		}
		if env.selector.CurrentStage != want {
			t.Fatalf("шаг %d: селектор %d", i, env.selector.CurrentStage)
		}
	}
}

// Единственный уровень: навигация всегда остаётся на нём
func TestStageSelectorUseCases_SingleStage(t *testing.T) {
	env := newStageSelectorTestEnv(1)

	if got := env.selectorUC.Next(env.selector); got != 1 {
		t.Errorf("Next: уровень %d, ожидался 1", got)
	}
	if got := env.selectorUC.Previous(env.selector); got != 1 {
		t.Errorf("Previous: уровень %d, ожидался 1", got)
	}
}

func TestStageSelectorUseCases_Select(t *testing.T) {
	env := newStageSelectorTestEnv(5)
	env.selector.CurrentStage = 4

	if got := env.selectorUC.Select(env.selector); got != 4 {
		t.Errorf("выбран уровень %d, ожидался 4", got)
	}
}

// Номер уровня выводится с ведущим нулём
func TestStageSelectorUseCases_String(t *testing.T) {
	env := newStageSelectorTestEnv(12)

	tests := []struct {
		stage uint
		want  string
	}{
		{1, "STAGE 01"},
		{9, "STAGE 09"},
		{12, "STAGE 12"},
	}
	for _, tt := range tests {
		env.selector.CurrentStage = tt.stage
		if got := env.selectorUC.String(env.selector); got != tt.want {
			t.Errorf("String() = %q, ожидалось %q", got, tt.want)
		}
	}
}

// NextMaxActiveEnemies растёт до 10 и заворачивается на 3;
// значения вне диапазона нормализуются
func TestStageSelectorUseCases_NextMaxActiveEnemies(t *testing.T) {
	env := newStageSelectorTestEnv(1)

	tests := []struct {
		name    string
		current uint
		want    uint
	}{
		{"внутри диапазона", 5, 6},
		{"перед максимумом", 9, 10},
		{"на максимуме", 10, 3},
		{"ниже диапазона", 0, 3},
		{"выше диапазона", 42, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := env.selectorUC.NextMaxActiveEnemies(tt.current)
			if got != tt.want {
				t.Errorf(
					"NextMaxActiveEnemies(%d) = %d, ожидалось %d",
					tt.current,
					got,
					tt.want,
				)
			}
		})
	}
}

// PreviousMaxActiveEnemies убывает до 3 и заворачивается на 10;
// значения вне диапазона нормализуются
func TestStageSelectorUseCases_PreviousMaxActiveEnemies(t *testing.T) {
	env := newStageSelectorTestEnv(1)

	tests := []struct {
		name    string
		current uint
		want    uint
	}{
		{"внутри диапазона", 5, 4},
		{"после минимума", 4, 3},
		{"на минимуме", 3, 10},
		{"ниже диапазона", 0, 10},
		{"выше диапазона", 42, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := env.selectorUC.PreviousMaxActiveEnemies(tt.current)
			if got != tt.want {
				t.Errorf(
					"PreviousMaxActiveEnemies(%d) = %d, ожидалось %d",
					tt.current,
					got,
					tt.want,
				)
			}
		})
	}
}
