package services

import (
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
)

const testTankSize = 16

// testGrid строит сетку 64x64 px с клетками 4px из блоков 8x8
func testGrid(blocks ...*types.BlockEntity) *types.NavGridEntity {
	return types.NewNavGridEntity(types.Size{Width: 64, Height: 64}, 4, blocks)
}

func block(kind types.BlockType, x, y float64) *types.BlockEntity {
	return types.NewBlockEntity(string(kind), x, y, 8, nil)
}

// wall ставит вертикальную стену блоков 8x8 в колонке x по строкам ys
func wall(kind types.BlockType, x float64, ys ...float64) []*types.BlockEntity {
	blocks := make([]*types.BlockEntity, 0, len(ys))
	for _, y := range ys {
		blocks = append(blocks, block(kind, x, y))
	}
	return blocks
}

func TestFindPathOpenField(t *testing.T) {
	service := NewNavigationService()

	step, found := service.FindPath(
		testGrid(),
		types.Position{X: 0, Y: 0},
		types.Position{X: 48, Y: 0},
		types.NavOptions{TankSize: testTankSize},
	)
	if !found {
		t.Fatal("path must be found")
	}
	if step.Direction != types.DirectionRight || step.Length != 6 {
		t.Fatalf("unexpected step: %+v", step)
	}
}

func TestFindPathAlreadyAtTarget(t *testing.T) {
	service := NewNavigationService()

	step, found := service.FindPath(
		testGrid(),
		types.Position{X: 16, Y: 16},
		types.Position{X: 17, Y: 15},
		types.NavOptions{TankSize: testTankSize},
	)
	if !found || step.Length != 0 {
		t.Fatalf("expected zero-length path, got %+v, %v", step, found)
	}
}

func TestFindPathGoesAroundSteel(t *testing.T) {
	service := NewNavigationService()
	// Бетон в колонке x=24 на всю высоту, кроме нижнего прохода 16px
	grid := testGrid(wall(types.Steel, 24, 0, 8, 16, 24, 32)...)

	step, found := service.FindPath(
		grid,
		types.Position{X: 0, Y: 0},
		types.Position{X: 48, Y: 0},
		types.NavOptions{TankSize: testTankSize},
	)
	if !found {
		t.Fatal("path around steel must be found")
	}
	if step.Direction != types.DirectionDown {
		t.Fatalf("expected to go down around steel, got %+v", step)
	}
}

func TestFindPathNoPathThroughSteelAndWater(t *testing.T) {
	service := NewNavigationService()
	blocks := wall(types.Steel, 24, 0, 8, 16, 24)
	blocks = append(blocks, wall(types.Water, 24, 32, 40, 48, 56)...)

	_, found := service.FindPath(
		testGrid(blocks...),
		types.Position{X: 0, Y: 0},
		types.Position{X: 48, Y: 0},
		types.NavOptions{TankSize: testTankSize},
	)
	if found {
		t.Fatal("steel and water wall must block the path")
	}
}

func TestFindPathSteelPassableForReinforced(t *testing.T) {
	service := NewNavigationService()
	grid := testGrid(wall(types.Steel, 24, 0, 8, 16, 24, 32, 40, 48, 56)...)

	step, found := service.FindPath(
		grid,
		types.Position{X: 0, Y: 0},
		types.Position{X: 48, Y: 0},
		types.NavOptions{TankSize: testTankSize, SteelPassable: true},
	)
	if !found || step.Direction != types.DirectionRight {
		t.Fatalf("reinforced tank must go through steel: %+v, %v", step, found)
	}
}

func TestFindPathBrickCostChoosesDetour(t *testing.T) {
	service := NewNavigationService()
	// Кирпичная стена с проходом внизу
	grid := testGrid(wall(types.Brick, 24, 0, 8, 16, 24, 32)...)
	from := types.Position{X: 0, Y: 0}
	to := types.Position{X: 48, Y: 0}

	cheap, _ := service.FindPath(grid, from, to, types.NavOptions{
		TankSize: testTankSize, BrickCost: 0,
	})
	if cheap.Direction != types.DirectionRight {
		t.Fatalf("free bricks must be crossed straight: %+v", cheap)
	}

	expensive, _ := service.FindPath(grid, from, to, types.NavOptions{
		TankSize: testTankSize, BrickCost: 10,
	})
	if expensive.Direction != types.DirectionDown {
		t.Fatalf("expensive bricks must be avoided: %+v", expensive)
	}
}

func TestFindPathNilGrid(t *testing.T) {
	service := NewNavigationService()

	if _, found := service.FindPath(
		nil, types.Position{}, types.Position{X: 8},
		types.NavOptions{TankSize: testTankSize},
	); found {
		t.Fatal("nil grid must not produce a path")
	}
}

func TestCastRayHitsTargetThroughForestAndWater(t *testing.T) {
	service := NewNavigationService()
	grid := testGrid(block(types.Forest, 28, 16), block(types.Water, 28, 24))
	targets := []types.RayTarget{{
		Kind:     types.RayHitPlayer,
		Position: types.Position{X: 24, Y: 40},
		Size:     types.Size{Width: 16, Height: 16},
	}}

	hit := service.CastRay(grid, types.Position{X: 32, Y: 8}, types.DirectionDown, targets)
	if hit.Kind != types.RayHitPlayer || hit.Distance != 32 {
		t.Fatalf("unexpected hit: %+v", hit)
	}
}

func TestCastRayBlockedBySteel(t *testing.T) {
	service := NewNavigationService()
	grid := testGrid(block(types.Steel, 32, 24))
	targets := []types.RayTarget{{
		Kind:     types.RayHitHQ,
		Position: types.Position{X: 48, Y: 24},
		Size:     types.Size{Width: 16, Height: 16},
	}}

	hit := service.CastRay(grid, types.Position{X: 8, Y: 28}, types.DirectionRight, targets)
	if hit.Kind != types.RayHitSteel || hit.Distance != 24 {
		t.Fatalf("unexpected hit: %+v", hit)
	}
}

func TestCastRayHitsBrickAndEdge(t *testing.T) {
	service := NewNavigationService()
	grid := testGrid(block(types.Brick, 8, 0))

	if hit := service.CastRay(grid, types.Position{X: 12, Y: 32}, types.DirectionUp, nil); hit.Kind != types.RayHitBrick {
		t.Fatalf("expected brick, got %+v", hit)
	}
	if hit := service.CastRay(grid, types.Position{X: 32, Y: 32}, types.DirectionLeft, nil); hit.Kind != types.RayHitEdge || hit.Distance != 33 {
		t.Fatalf("expected edge, got %+v", hit)
	}
}
