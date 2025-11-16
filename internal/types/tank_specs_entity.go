package types

// SpecsEntity содержит спецификации танка и пуль
type SpecsEntity struct {
	level             uint    // Уровень танка (0-3)
	speed             float64 // Скорость танка
	bulletsReinforced bool    // Могут ли пули ломать сталь
	bulletsSpeed      float64 // Скорость пуль
	bulletsLimit      uint    // Лимит одновременно выпущенных пуль
}

func NewSpecsEntity(
	level uint,
	speed float64,
	bulletsReinforced bool,
	bulletsSpeed float64,
	bulletsLimit uint,
) *SpecsEntity {
	return &SpecsEntity{
		level:             level,
		speed:             speed,
		bulletsReinforced: bulletsReinforced,
		bulletsSpeed:      bulletsSpeed,
		bulletsLimit:      bulletsLimit,
	}
}

func (t *SpecsEntity) GetLevel() uint {
	if t == nil {
		return 0
	}
	return t.level
}

func (t *SpecsEntity) SetLevel(level uint) {
	if t == nil {
		return
	}
	if level > 3 {
		level = 3
	}
	t.level = level
}

func (t *SpecsEntity) GetSpeed() float64 {
	if t == nil {
		return 0
	}
	return t.speed
}

func (t *SpecsEntity) GetBulletsReinforced() bool {
	if t == nil {
		return false
	}
	return t.bulletsReinforced
}

func (t *SpecsEntity) GetBulletsSpeed() float64 {
	if t == nil {
		return 0
	}
	return t.bulletsSpeed
}

func (t *SpecsEntity) GetBulletsLimit() uint {
	if t == nil {
		return 0
	}
	return t.bulletsLimit
}
