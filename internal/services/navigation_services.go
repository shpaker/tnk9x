package services

import (
	"container/heap"
	"math"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

var _ interfaces.INavigationService = (*NavigationService)(nil)

// rayWidth — ширина полосы луча, совпадает с шириной пули
const rayWidth = 4

// NavigationService — поиск пути и проверка линии огня по сетке карты.
// Только отвечает на запросы, решений не принимает.
type NavigationService struct{}

func NewNavigationService() *NavigationService {
	return &NavigationService{}
}

// Поиск пути

type navNode struct {
	col, row int
}

type navQueueItem struct {
	node navNode
	cost int
}

type navQueue []navQueueItem

func (q navQueue) Len() int           { return len(q) }
func (q navQueue) Less(i, j int) bool { return q[i].cost < q[j].cost }
func (q navQueue) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }
func (q *navQueue) Push(item any)     { *q = append(*q, item.(navQueueItem)) }

func (q *navQueue) Pop() any {
	old := *q
	item := old[len(old)-1]
	*q = old[:len(old)-1]
	return item
}

var navDirections = []struct {
	direction types.Direction
	dCol      int
	dRow      int
}{
	{types.DirectionUp, 0, -1},
	{types.DirectionDown, 0, 1},
	{types.DirectionLeft, -1, 0},
	{types.DirectionRight, 1, 0},
}

// FindPath ищет путь танка алгоритмом Дейкстры по позициям с шагом
// в половину танка и возвращает первый шаг. Бетон и вода непроходимы
// (бетон — если не SteelPassable), кирпич проходим с доплатой BrickCost.
// Реализует interfaces.INavigationService.
func (s *NavigationService) FindPath(
	grid *types.NavGridEntity,
	from, to types.Position,
	options types.NavOptions,
) (types.PathStep, bool) {
	step := options.TankSize / 2
	if grid == nil || step <= 0 {
		return types.PathStep{}, false
	}

	sizePx := grid.GetSizePx()
	maxCol := (sizePx.Width - options.TankSize) / step
	maxRow := (sizePx.Height - options.TankSize) / step
	if maxCol < 0 || maxRow < 0 {
		return types.PathStep{}, false
	}

	start := s.toNode(from, step, maxCol, maxRow)
	target := s.toNode(to, step, maxCol, maxRow)
	if start == target {
		return types.PathStep{Length: 0}, true
	}

	width := maxCol + 1
	index := func(n navNode) int { return n.row*width + n.col }
	costs := make([]int, width*(maxRow+1))
	steps := make([]int, len(costs))
	parents := make([]int, len(costs))
	for i := range costs {
		costs[i] = math.MaxInt
		parents[i] = -1
	}
	costs[index(start)] = 0

	queue := &navQueue{{node: start, cost: 0}}
	for queue.Len() > 0 {
		current := heap.Pop(queue).(navQueueItem)
		if current.cost > costs[index(current.node)] {
			continue
		}
		if current.node == target {
			break
		}

		for _, d := range navDirections {
			next := navNode{
				col: current.node.col + d.dCol,
				row: current.node.row + d.dRow,
			}
			if next.col < 0 || next.row < 0 || next.col > maxCol ||
				next.row > maxRow {
				continue
			}

			stepCost, passable := s.stepCost(grid, next, step, options)
			// Клетка цели достижима всегда: цель может стоять вплотную к стене
			if !passable && next != target {
				continue
			}

			cost := current.cost + stepCost
			if cost < costs[index(next)] {
				costs[index(next)] = cost
				steps[index(next)] = steps[index(current.node)] + 1
				parents[index(next)] = index(current.node)
				heap.Push(queue, navQueueItem{node: next, cost: cost})
			}
		}
	}

	if parents[index(target)] < 0 {
		return types.PathStep{}, false
	}

	// Поднимаемся от цели к первому шагу после старта
	first := index(target)
	for parents[first] != index(start) {
		first = parents[first]
	}
	firstNode := navNode{col: first % width, row: first / width}

	return types.PathStep{
		Direction: s.directionBetween(start, firstNode),
		Length:    steps[index(target)],
	}, true
}

func (s *NavigationService) toNode(
	position types.Position,
	step, maxCol, maxRow int,
) navNode {
	col := int(math.Round(position.X / float64(step)))
	row := int(math.Round(position.Y / float64(step)))
	return navNode{
		col: min(max(col, 0), maxCol),
		row: min(max(row, 0), maxRow),
	}
}

// stepCost оценивает шаг в позицию: танк занимает квадрат TankSize,
// проверяются все клетки сетки под ним
func (s *NavigationService) stepCost(
	grid *types.NavGridEntity,
	node navNode,
	step int,
	options types.NavOptions,
) (int, bool) {
	cellSize := grid.GetCellSize()
	fromCol := node.col * step / cellSize
	fromRow := node.row * step / cellSize
	toCol := (node.col*step+options.TankSize)/cellSize - 1
	toRow := (node.row*step+options.TankSize)/cellSize - 1

	hasBrick := false
	for row := fromRow; row <= toRow; row++ {
		for col := fromCol; col <= toCol; col++ {
			switch grid.GetCell(col, row) {
			case types.Water:
				if !options.WaterPassable {
					return 0, false
				}
			case types.Steel:
				if !options.SteelPassable {
					return 0, false
				}
				hasBrick = true
			case types.Brick:
				hasBrick = true
			}
		}
	}

	if hasBrick {
		return 1 + options.BrickCost, true
	}
	return 1, true
}

func (s *NavigationService) directionBetween(
	from, to navNode,
) types.Direction {
	switch {
	case to.row < from.row:
		return types.DirectionUp
	case to.row > from.row:
		return types.DirectionDown
	case to.col < from.col:
		return types.DirectionLeft
	default:
		return types.DirectionRight
	}
}

// Линия огня

// CastRay ведёт полосу шириной с пулю из точки origin в направлении
// direction и возвращает первое препятствие: кирпич, бетон, объект
// из targets или край карты. Вода, лес и лёд пулю не останавливают.
// Реализует interfaces.INavigationService.
func (s *NavigationService) CastRay(
	grid *types.NavGridEntity,
	origin types.Position,
	direction types.Direction,
	targets []types.RayTarget,
) types.RayHit {
	if grid == nil {
		return types.RayHit{Kind: types.RayHitEdge}
	}

	dx, dy := 0.0, 0.0
	switch direction {
	case types.DirectionUp:
		dy = -1
	case types.DirectionDown:
		dy = 1
	case types.DirectionLeft:
		dx = -1
	case types.DirectionRight:
		dx = 1
	}

	sizePx := grid.GetSizePx()
	half := float64(rayWidth) / 2
	for distance := 0; ; distance++ {
		x := origin.X + dx*float64(distance)
		y := origin.Y + dy*float64(distance)
		if x < 0 || y < 0 || x >= float64(sizePx.Width) ||
			y >= float64(sizePx.Height) {
			return types.RayHit{Kind: types.RayHitEdge, Distance: distance}
		}

		// Поперечный отрезок полосы в текущей точке
		left, top, right, bottom := x, y, x+1, y+1
		if dx == 0 {
			left, right = x-half, x+half
		} else {
			top, bottom = y-half, y+half
		}

		if kind, ok := s.blockOnSegment(grid, left, top, right, bottom); ok {
			return types.RayHit{Kind: kind, Distance: distance}
		}
		for _, target := range targets {
			if s.intersects(target, left, top, right, bottom) {
				return types.RayHit{Kind: target.Kind, Distance: distance}
			}
		}
	}
}

func (s *NavigationService) blockOnSegment(
	grid *types.NavGridEntity,
	left, top, right, bottom float64,
) (types.RayHitKind, bool) {
	cellSize := float64(grid.GetCellSize())
	fromCol := int(math.Floor(left / cellSize))
	fromRow := int(math.Floor(top / cellSize))
	toCol := int(math.Ceil(right/cellSize)) - 1
	toRow := int(math.Ceil(bottom/cellSize)) - 1

	hit := types.RayHitKind("")
	for row := fromRow; row <= toRow; row++ {
		for col := fromCol; col <= toCol; col++ {
			switch grid.GetCell(col, row) {
			case types.Steel:
				return types.RayHitSteel, true
			case types.Brick:
				hit = types.RayHitBrick
			}
		}
	}
	return hit, hit != ""
}

func (s *NavigationService) intersects(
	target types.RayTarget,
	left, top, right, bottom float64,
) bool {
	return left < target.Position.X+float64(target.Size.Width) &&
		right > target.Position.X &&
		top < target.Position.Y+float64(target.Size.Height) &&
		bottom > target.Position.Y
}
