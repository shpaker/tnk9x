package types

// PackOrder — порядок открытия уровней внутри пачки
type PackOrder int

const (
	// PackOrderSequential — следующий уровень открывается победой
	// на предыдущем
	PackOrderSequential PackOrder = iota
	// PackOrderAny — все уровни открываются вместе с пачкой
	PackOrderAny
)

// PackSpec — пачка уровней кампании и условия её открытия
type PackSpec struct {
	Name        string
	Levels      []int
	Order       PackOrder
	UnlockStars uint
	// UnlockAfter — уровень, который нужно пройти до открытия;
	// 0 — без условия
	UnlockAfter int
}

// CampaignEntity — кампания: упорядоченные пачки уровней
type CampaignEntity struct {
	name  string
	packs []PackSpec
}

func NewCampaignEntity(name string, packs []PackSpec) *CampaignEntity {
	return &CampaignEntity{
		name:  name,
		packs: packs,
	}
}

func (c *CampaignEntity) GetName() string {
	return c.name
}

func (c *CampaignEntity) GetPacks() []PackSpec {
	return c.packs
}

// FindLevel — индекс пачки и позиция уровня в ней
func (c *CampaignEntity) FindLevel(level int) (int, int, bool) {
	for packIndex, pack := range c.packs {
		for position, number := range pack.Levels {
			if number == level {
				return packIndex, position, true
			}
		}
	}
	return 0, 0, false
}

// NextLevel — следующий уровень кампании после заданного
func (c *CampaignEntity) NextLevel(level int) (int, bool) {
	packIndex, position, ok := c.FindLevel(level)
	if !ok {
		return 0, false
	}
	pack := c.packs[packIndex]
	if position+1 < len(pack.Levels) {
		return pack.Levels[position+1], true
	}
	for next := packIndex + 1; next < len(c.packs); next++ {
		if len(c.packs[next].Levels) > 0 {
			return c.packs[next].Levels[0], true
		}
	}
	return 0, false
}

// MaxStars — максимум звёзд в кампании
func (c *CampaignEntity) MaxStars() uint {
	total := uint(0)
	for _, pack := range c.packs {
		total += uint(len(pack.Levels)) * MaxLevelStars
	}
	return total
}

// PackStatus — состояние пачки для экрана выбора уровня
type PackStatus struct {
	Unlocked bool
	// Stars — звёзды, набранные в пачке, из MaxStars возможных
	Stars    uint
	MaxStars uint
	// RequiredStars — порог открытия, TotalStars — звёзды кампании
	RequiredStars uint
	TotalStars    uint
	// UnlockAfter — уровень, который нужно пройти; 0 — без условия
	UnlockAfter int
}
