package types

// MaxLevelStars — максимум звёзд за уровень
const MaxLevelStars uint = 3

// MaxCarryOverStars — максимум звёзд за уровень, начатый с переносом
// жизней и прокачки танка
const MaxCarryOverStars uint = 2

// ProgressEntity — прогресс игрока: лучший результат по уровням;
// уровень с ненулевыми звёздами считается пройденным
type ProgressEntity struct {
	stars map[int]uint
}

func NewProgressEntity() *ProgressEntity {
	return &ProgressEntity{
		stars: make(map[int]uint),
	}
}

// NewProgressEntityFromStars восстанавливает прогресс из сохранения;
// значения обрезаются до допустимого диапазона
func NewProgressEntityFromStars(stars map[int]uint) *ProgressEntity {
	progress := NewProgressEntity()
	for level, value := range stars {
		if value == 0 {
			continue
		}
		if value > MaxLevelStars {
			value = MaxLevelStars
		}
		progress.stars[level] = value
	}
	return progress
}

func (p *ProgressEntity) GetStars(level int) uint {
	return p.stars[level]
}

func (p *ProgressEntity) IsCompleted(level int) bool {
	return p.stars[level] > 0
}

// SetStars сохраняет результат, только если он лучше прежнего;
// возвращает true, если прогресс изменился
func (p *ProgressEntity) SetStars(level int, stars uint) bool {
	if stars > MaxLevelStars {
		stars = MaxLevelStars
	}
	if stars <= p.stars[level] {
		return false
	}
	p.stars[level] = stars
	return true
}

func (p *ProgressEntity) TotalStars() uint {
	total := uint(0)
	for _, value := range p.stars {
		total += value
	}
	return total
}

// AllStars — копия результатов для сериализации
func (p *ProgressEntity) AllStars() map[int]uint {
	result := make(map[int]uint, len(p.stars))
	for level, value := range p.stars {
		result[level] = value
	}
	return result
}

// StageResult — итог прохождения уровня для подсчёта звёзд
type StageResult struct {
	Won          bool
	LivesLost    uint
	ElapsedTicks uint
	// CarriedOver — уровень начат с перенесёнными жизнями и прокачкой
	CarriedOver bool
}
