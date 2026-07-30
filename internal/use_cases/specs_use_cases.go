package use_cases

import (
	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.ISpecsUseCases = (*SpecsUseCases)(nil)

type SpecsUseCases struct {
	playersSpecs [4]*types.SpecsEntity // Уровни 0-3
	enemiesSpecs [4]*types.SpecsEntity // Уровни 0-3
}

func NewSpecsUseCases() *SpecsUseCases {
	uc := &SpecsUseCases{
		playersSpecs: [4]*types.SpecsEntity{},
		enemiesSpecs: [4]*types.SpecsEntity{},
	}

	// Инициализируем спецификации для танков игроков (уровни 0-3)
	// Уровень 0 (базовый): обычная скорость, 1 пуля, не может ломать сталь
	uc.playersSpecs[0] = types.NewSpecsEntity(
		0,     // Уровень 0
		32.0,  // Базовая скорость танка
		false, // Пули не усиленные
		120.0, // Скорость пули
		1,     // Лимит пуль: 1
	)

	// Уровень 1: увеличенная скорость пуль
	uc.playersSpecs[1] = types.NewSpecsEntity(
		1,     // Уровень 1
		32.0,  // Скорость танка без изменений
		false, // Пули не усиленные
		150.0, // Увеличенная скорость пули
		1,     // Лимит пуль: 1
	)

	// Уровень 2: две пули одновременно (как в оригинале)
	uc.playersSpecs[2] = types.NewSpecsEntity(
		2,     // Уровень 2
		32.0,  // Скорость танка без изменений
		false, // Пули не усиленные
		150.0, // Увеличенная скорость пули
		2,     // Лимит пуль: 2
	)

	// Уровень 3: усиленные пули ломают сталь; скорость танка не меняется
	uc.playersSpecs[3] = types.NewSpecsEntity(
		3,     // Уровень 3
		32.0,  // Скорость танка без изменений
		true,  // Пули усиленные (могут ломать сталь)
		150.0, // Увеличенная скорость пули
		2,     // Лимит пуль: 2
	)

	// Инициализируем спецификации для вражеских танков согласно Battle City
	// Уровень 0: Обычный танк - базовая скорость, стандартная скорострельность
	uc.enemiesSpecs[0] = types.NewSpecsEntity(
		0,     // Уровень 0
		32.0,  // Базовая скорость танка
		false, // Пули не усиленные
		120.0, // Скорость пули
		1,     // Лимит пуль: 1
	)

	// Уровень 1: Быстрый танк - повышенная скорость передвижения
	uc.enemiesSpecs[1] = types.NewSpecsEntity(
		1,     // Уровень 1
		48.0,  // Повышенная скорость танка (в 1.5 раза быстрее)
		false, // Пули не усиленные
		120.0, // Скорость пули
		1,     // Лимит пуль: 1
	)

	// Уровень 2: Скорострельный танк - увеличенная скорострельность (через скорость пуль)
	uc.enemiesSpecs[2] = types.NewSpecsEntity(
		2,     // Уровень 2
		32.0,  // Базовая скорость танка
		false, // Пули не усиленные
		180.0, // Увеличенная скорость пули (в 1.5 раза быстрее)
		1,     // Лимит пуль: 1
	)

	// Уровень 3: Тяжёлый танк - повышенная прочность (требует 4 попадания)
	uc.enemiesSpecs[3] = types.NewSpecsEntity(
		3,     // Уровень 3
		32.0,  // Базовая скорость танка
		false, // Пули не усиленные
		120.0, // Скорость пули
		1,     // Лимит пуль: 1
	)

	return uc
}

func (uc *SpecsUseCases) GetTankSpecs(
	isEnemy bool,
	level uint,
) *types.SpecsEntity {
	if level > 3 {
		level = 3
	}

	if isEnemy {
		if int(level) < len(uc.enemiesSpecs) {
			return uc.enemiesSpecs[level]
		}
		return nil
	}

	if int(level) < len(uc.playersSpecs) {
		return uc.playersSpecs[level]
	}
	return nil
}
