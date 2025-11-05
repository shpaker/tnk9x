package types

// MapEntity представляет карту уровня с блоками и размерами
type MapEntity struct {
	sizePx Size
	blocks []BlockEntity
}

// NewMapEntity создает новый MapEntity
func NewMapEntity(sizePx Size, blocks []BlockEntity) *MapEntity {
	return &MapEntity{
		sizePx: sizePx,
		blocks: blocks,
	}
}

// GetSizePx возвращает размер карты в пикселях
func (m *MapEntity) GetSizePx() Size {
	return m.sizePx
}

// GetBlocks возвращает все блоки карты
func (m *MapEntity) GetBlocks() []BlockEntity {
	return m.blocks
}

// SetBlocks устанавливает блоки карты
func (m *MapEntity) SetBlocks(blocks []BlockEntity) {
	m.blocks = blocks
}

// AddBlock добавляет блок в карту
func (m *MapEntity) AddBlock(block BlockEntity) {
	m.blocks = append(m.blocks, block)
}

// RemoveBlock удаляет блок по указателю
func (m *MapEntity) RemoveBlock(block *BlockEntity) error {
	if block == nil {
		return nil
	}

	for i := range m.blocks {
		if &m.blocks[i] == block {
			m.blocks = append(m.blocks[:i], m.blocks[i+1:]...)
			return nil
		}
	}

	return nil
}
