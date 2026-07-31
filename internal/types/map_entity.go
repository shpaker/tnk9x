package types

import "math/rand"

type MapBlocks []*BlockEntity

type MapEntity struct {
	sizePx              Size
	blocks              MapBlocks
	bonusSpawnPositions []Position

	hqFortifyTicks uint      // Оставшиеся тики укрепления штаба лопатой
	hqSavedBlocks  MapBlocks // Блоки, заменённые бетоном при укреплении
	hqSteelBlocks  MapBlocks // Установленные при укреплении бетонные блоки
}

func NewMapEntity(
	sizePx Size,
	blocks MapBlocks,
	bonusSpawnPositions []Position,
) *MapEntity {
	return &MapEntity{
		sizePx:              sizePx,
		blocks:              blocks,
		bonusSpawnPositions: bonusSpawnPositions,
	}
}

func (m *MapEntity) GetSizePx() Size {
	return m.sizePx
}

func (m *MapEntity) GetBlocks() MapBlocks {
	return m.blocks
}

func (m *MapEntity) SetBlocks(blocks MapBlocks) {
	m.blocks = blocks
}

func (m *MapEntity) AddBlock(block *BlockEntity) {
	m.blocks = append(m.blocks, block)
}

func (m *MapEntity) RemoveBlock(block *BlockEntity) error {
	if block == nil {
		return nil
	}

	for i, b := range m.blocks {
		if b == block {
			m.blocks = append(m.blocks[:i], m.blocks[i+1:]...)
			return nil
		}
	}

	return nil
}

func (m *MapEntity) IsHQFortified() bool {
	return m.hqFortifyTicks > 0
}

// SetHQFortification сохраняет заменённые и установленные блоки укрепления
// и запускает его отсчёт
func (m *MapEntity) SetHQFortification(
	savedBlocks MapBlocks,
	steelBlocks MapBlocks,
	ticks uint,
) {
	m.hqSavedBlocks = savedBlocks
	m.hqSteelBlocks = steelBlocks
	m.hqFortifyTicks = ticks
}

// ResetHQFortifyCountdown продлевает действующее укрепление
func (m *MapEntity) ResetHQFortifyCountdown(ticks uint) {
	if m.hqFortifyTicks > 0 {
		m.hqFortifyTicks = ticks
	}
}

// UpdateHQFortifyCountdown уменьшает отсчёт укрепления;
// true — укрепление истекло в этом тике
func (m *MapEntity) UpdateHQFortifyCountdown() bool {
	if m.hqFortifyTicks == 0 {
		return false
	}
	m.hqFortifyTicks--
	return m.hqFortifyTicks == 0
}

// TakeHQFortification возвращает блоки укрепления и очищает его состояние
func (m *MapEntity) TakeHQFortification() (MapBlocks, MapBlocks) {
	savedBlocks := m.hqSavedBlocks
	steelBlocks := m.hqSteelBlocks
	m.hqSavedBlocks = nil
	m.hqSteelBlocks = nil
	m.hqFortifyTicks = 0
	return savedBlocks, steelBlocks
}

func (m *MapEntity) GetRandomBonusSpawnPosition() Position {
	if len(m.bonusSpawnPositions) == 0 {
		return Position{X: 0, Y: 0}
	}
	return m.bonusSpawnPositions[rand.Intn(len(m.bonusSpawnPositions))]
}
