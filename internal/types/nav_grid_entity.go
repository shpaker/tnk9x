package types

import "math"

// NavGridEntity — сетка карты для AI: в каждой клетке тип блока
// или пустая строка. Клетка соответствует минимальному обломку кирпича.
type NavGridEntity struct {
	sizePx   Size
	cellSize int
	cols     int
	rows     int
	cells    []BlockType
}

// Приоритет типов при наложении блоков в одной клетке:
// непроходимое важнее проходимого
var navBlockPriority = map[BlockType]int{
	Ice:    1,
	Forest: 2,
	Water:  3,
	Brick:  4,
	Steel:  5,
}

// NewNavGridEntity строит сетку по блокам карты
func NewNavGridEntity(
	sizePx Size,
	cellSize int,
	blocks MapBlocks,
) *NavGridEntity {
	cols := sizePx.Width / cellSize
	rows := sizePx.Height / cellSize
	grid := &NavGridEntity{
		sizePx:   sizePx,
		cellSize: cellSize,
		cols:     cols,
		rows:     rows,
		cells:    make([]BlockType, cols*rows),
	}

	for _, block := range blocks {
		if block == nil || block.Data == nil {
			continue
		}
		grid.fill(block)
	}

	return grid
}

func (g *NavGridEntity) fill(block *BlockEntity) {
	size := block.GetSize()
	cellSize := float64(g.cellSize)
	fromCol := int(math.Floor(block.Position.X / cellSize))
	fromRow := int(math.Floor(block.Position.Y / cellSize))
	toCol := int(math.Ceil((block.Position.X+float64(size.Width))/cellSize)) - 1
	toRow := int(
		math.Ceil((block.Position.Y+float64(size.Height))/cellSize),
	) - 1

	for row := fromRow; row <= toRow; row++ {
		for col := fromCol; col <= toCol; col++ {
			if !g.Contains(col, row) {
				continue
			}
			index := row*g.cols + col
			current := g.cells[index]
			if navBlockPriority[block.Data.Name] > navBlockPriority[current] {
				g.cells[index] = block.Data.Name
			}
		}
	}
}

func (g *NavGridEntity) GetSizePx() Size {
	return g.sizePx
}

func (g *NavGridEntity) GetCellSize() int {
	return g.cellSize
}

func (g *NavGridEntity) GetCols() int {
	return g.cols
}

func (g *NavGridEntity) GetRows() int {
	return g.rows
}

// Contains проверяет, что клетка лежит в пределах карты
func (g *NavGridEntity) Contains(col, row int) bool {
	return col >= 0 && row >= 0 && col < g.cols && row < g.rows
}

// GetCell возвращает тип блока в клетке; вне карты — пустая строка
func (g *NavGridEntity) GetCell(col, row int) BlockType {
	if !g.Contains(col, row) {
		return ""
	}
	return g.cells[row*g.cols+col]
}

// GetBlockAt возвращает тип блока в точке карты в пикселях
func (g *NavGridEntity) GetBlockAt(x, y float64) BlockType {
	cellSize := float64(g.cellSize)
	return g.GetCell(
		int(math.Floor(x/cellSize)),
		int(math.Floor(y/cellSize)),
	)
}
