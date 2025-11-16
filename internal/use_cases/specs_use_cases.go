package use_cases

import (
	"math/rand"

	"github.com/shpaker/tnk25/internal/types"
)

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

	// Уровень 2: может ломать сталь, увеличенная скорость пуль
	uc.playersSpecs[2] = types.NewSpecsEntity(
		2,     // Уровень 2
		32.0,  // Скорость танка без изменений
		true,  // Пули усиленные (могут ломать сталь)
		150.0, // Увеличенная скорость пули
		1,     // Лимит пуль: 1
	)

	// Уровень 3: максимальные характеристики - быстрее, может ломать сталь, больше пуль
	uc.playersSpecs[3] = types.NewSpecsEntity(
		3,     // Уровень 3
		40.0,  // Увеличенная скорость танка
		true,  // Пули усиленные
		150.0, // Увеличенная скорость пули
		2,     // Лимит пуль: 2 (может стрелять двумя пулями одновременно)
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

// GetEnemyLevelByRemainingCount определяет уровень вражеского танка
// на основе количества оставшихся врагов
func (uc *SpecsUseCases) GetEnemyLevelByRemainingCount(
	remainingEnemies uint,
) uint {
	// Первые три танка всегда 0 уровня (оставшиеся 18-20)
	if remainingEnemies > 17 {
		return 0
	}

	// Танки 4-5 (оставшиеся 16-17): всегда уровень 1
	if remainingEnemies > 15 {
		return 1
	}

	// Танки 6-10 (оставшиеся 11-15): либо 0 ур либо 1 ур (вероятность 50 процентов)
	if remainingEnemies > 10 {
		if rand.Intn(2) == 0 {
			return 1
		}
		return 2
	}

	// Танки 11-15 (оставшиеся 6-10): 0 ур - вероятность 4 к 10, 1 ур - 4 к 10, 2 ур - 2 к 10
	if remainingEnemies > 5 {
		roll := rand.Intn(10)
		if roll < 4 {
			return 1
		} else if roll < 8 {
			return 2
		}
		return 3
	}

	// Танки 16-20 (оставшиеся 1-5): 1 ур - вероятность 2 к 10, 2 ур - 2 к 10, 3 ур - 6 к 10
	roll := rand.Intn(10)
	if roll < 2 {
		return 1
	} else if roll < 4 {
		return 2
	}
	return 3
}
