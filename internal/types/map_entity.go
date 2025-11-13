package types

import "math/rand"

type MapBlocks []*BlockEntity

type MapEntity struct {
	sizePx              Size
	blocks              MapBlocks
	bonusSpawnPositions []Position
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

func (m *MapEntity) GetRandomBonusSpawnPosition() Position {
	if len(m.bonusSpawnPositions) == 0 {
		return Position{X: 0, Y: 0}
	}
	return m.bonusSpawnPositions[rand.Intn(len(m.bonusSpawnPositions))]
}
