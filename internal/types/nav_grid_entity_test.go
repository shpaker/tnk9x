package types

import "testing"

func TestNavGridEntityFillsCellsFromBlocks(t *testing.T) {
	blocks := MapBlocks{
		NewBlockEntity(string(Brick), 8, 0, 8, nil),
		// Обломок кирпича 8x4 занимает одну строку клеток
		{
			Position: Position{X: 0, Y: 12},
			Size:     Size{Width: 8, Height: 4},
			Data:     &BlockData{Name: Brick},
		},
		NewBlockEntity(string(Forest), 16, 16, 8, nil),
		// Бетон поверх леса важнее
		NewBlockEntity(string(Steel), 16, 16, 8, nil),
	}
	grid := NewNavGridEntity(Size{Width: 32, Height: 32}, 4, blocks)

	if grid.GetCols() != 8 || grid.GetRows() != 8 {
		t.Fatalf("unexpected grid size: %dx%d", grid.GetCols(), grid.GetRows())
	}
	if grid.GetCell(2, 0) != Brick || grid.GetCell(3, 1) != Brick {
		t.Fatal("full brick must fill 2x2 cells")
	}
	if grid.GetCell(0, 3) != Brick || grid.GetCell(0, 2) != "" {
		t.Fatal("brick slab must fill exactly one row")
	}
	if grid.GetBlockAt(18, 18) != Steel {
		t.Fatal("steel must override forest")
	}
	if grid.GetCell(-1, 0) != "" || grid.Contains(8, 0) {
		t.Fatal("cells outside the map must be empty")
	}
}
