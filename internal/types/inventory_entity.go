package types

// InventoryEntity — купленное игроком: разовые покупки и жетоны;
// одно на приложение, общее для всех режимов. credited — зачисленные
// покупки жетонов: покупка, которую площадка не успела списать или
// отдала снова, повторно не зачисляется. Список не чистится — пустой
// ответ площадки (нет сети) не отличить от списания
type InventoryEntity struct {
	noAds     bool
	allLevels bool
	packs     map[int]bool
	tokens    uint
	credited  map[string]bool
}

func NewInventoryEntity() *InventoryEntity {
	return &InventoryEntity{
		packs:    make(map[int]bool),
		credited: make(map[string]bool),
	}
}

// Разовые покупки

func (e *InventoryEntity) HasNoAds() bool { return e.noAds }

func (e *InventoryEntity) SetNoAds(owned bool) { e.noAds = owned }

func (e *InventoryEntity) HasAllLevels() bool { return e.allLevels }

func (e *InventoryEntity) SetAllLevels(owned bool) { e.allLevels = owned }

// IsPackOwned — пачка с порядковым номером pack (с 1) куплена
func (e *InventoryEntity) IsPackOwned(pack int) bool { return e.packs[pack] }

func (e *InventoryEntity) SetPackOwned(pack int) { e.packs[pack] = true }

// GetPacks — номера купленных пачек
func (e *InventoryEntity) GetPacks() []int {
	packs := make([]int, 0, len(e.packs))
	for pack := range e.packs {
		packs = append(packs, pack)
	}
	return packs
}

// Жетоны

func (e *InventoryEntity) GetTokens() uint { return e.tokens }

func (e *InventoryEntity) AddTokens(amount uint) { e.tokens += amount }

// SpendToken списывает жетон; false — жетонов нет
func (e *InventoryEntity) SpendToken() bool {
	if e.tokens == 0 {
		return false
	}
	e.tokens--
	return true
}

// Зачисленные покупки жетонов

func (e *InventoryEntity) IsCredited(token string) bool {
	return e.credited[token]
}

func (e *InventoryEntity) MarkCredited(token string) {
	e.credited[token] = true
}

// GetCredited — зачисленные покупки жетонов
func (e *InventoryEntity) GetCredited() []string {
	tokens := make([]string, 0, len(e.credited))
	for token := range e.credited {
		tokens = append(tokens, token)
	}
	return tokens
}
