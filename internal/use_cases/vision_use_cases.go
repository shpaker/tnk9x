package use_cases

import (
	"math"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// Зрение как в жизни, одинаковое для своих и чужих: объект виден, если
// он в поле зрения игрока (передняя полусфера или вплотную) и взгляд
// до него не закрыт зданием; полностью — если он ещё и освещён: фарой,
// пулей, вспышкой, взрывом. Пуля светится сама и видна везде, куда
// доходит взгляд. Кирпич и сталь — здания, лес низкий и лишь приглушает
const (
	// visibilityHidden — вне взгляда объект приглушён, но различим:
	// полностью скрытые танки делают поле боя пустым
	visibilityHidden = 0.5
	// visionAmbient — освещённость без источников в долях полной:
	// в темноте объект в поле зрения виден тусклее освещённого
	visionAmbient = 0.5
	// visionLitLevel — освещённость, при которой объект виден полностью
	visionLitLevel   = 0.6
	forestVisibility = 0.5 // множитель, если между — лес

	// Скорость проявления и исчезновения — доля разницы за тик
	visibilityRise = 0.35
	visibilityFall = 0.1

	// Точки обзора танка: центр и углы, отступившие от края на
	// visionTankInset, — танк, выглянувший из-за угла, уже виден
	visionTankInset = 5.0
)

// Сетка линии видимости: карта в клетках visionCell пикселей
// (кратно слою скола кирпича), луч проверяется с шагом visionStep;
// у концов луча лес не учитывается — танк под деревьями видит
// и виден, как на открытом месте
const (
	visionCell        = 4
	visionStep        = 2.0
	visionTargetGap   = 3.0
	visionForestReach = 8.0
)

const (
	visionCellEmpty uint8 = iota
	visionCellForest
	visionCellWall
)

var _ interfaces.IVisionUseCases = (*VisionUseCases)(nil)

// VisionUseCases реализует IVisionUseCases: считает, насколько игроки
// видят врагов и пули
type VisionUseCases struct {
	// Use Cases
	tankCommonUseCases interfaces.ITankCommonUseCases
	bulletUseCases     interfaces.IBulletUseCases
	mapUseCases        interfaces.IMapUseCases
	lightingUseCases   interfaces.ILightingUseCases
}

func NewVisionUseCases(
	tankCommonUseCases interfaces.ITankCommonUseCases,
	bulletUseCases interfaces.IBulletUseCases,
	mapUseCases interfaces.IMapUseCases,
	lightingUseCases interfaces.ILightingUseCases,
) *VisionUseCases {
	return &VisionUseCases{
		tankCommonUseCases: tankCommonUseCases,
		bulletUseCases:     bulletUseCases,
		mapUseCases:        mapUseCases,
		lightingUseCases:   lightingUseCases,
	}
}

// UpdateVisibility реализует IVisionUseCases
func (uc *VisionUseCases) UpdateVisibility() {
	sight := vision{
		viewers: uc.lightingUseCases.GetViewers(),
		lights:  uc.lightingUseCases.GetLights(),
		grid: newVisionGrid(
			uc.mapUseCases.GetBlocks(),
			uc.mapUseCases.GetSizePx(),
		),
	}

	for _, tank := range uc.tankCommonUseCases.GetAllTanks() {
		if tank == nil || !tank.IsEnemy() || !tank.IsActive() {
			continue
		}
		target := 0.0
		for _, point := range tankViewPoints(tank) {
			target = math.Max(target, sight.at(point, false))
		}
		tank.FadeVisibility(target, visibilityRise, visibilityFall)
	}

	for _, bullet := range uc.bulletUseCases.GetBullets() {
		if bullet != nil {
			bullet.SetVisibility(sight.at(bulletCenter(bullet), true))
		}
	}
}

// vision — что видят игроки в этом тике: зрители, свет кадра
// и препятствия взгляду и свету
type vision struct {
	viewers []types.ViewerEntity
	lights  []types.LightEntity
	grid    visionGrid
}

// at — видимость точки: насколько она в поле зрения хотя бы одного
// игрока, умноженное на её освещённость; светящаяся сама (emissive)
// точка освещена всегда
func (v vision) at(point types.Position, emissive bool) float64 {
	seen := 0.0
	for _, viewer := range v.viewers {
		field := viewer.FieldAt(point)
		if field > seen {
			seen = math.Max(seen, field*v.grid.sight(viewer.Position, point))
		}
	}
	if seen == 0 {
		return visibilityHidden
	}
	lit := 1.0
	if !emissive {
		lit = math.Min(1, visionAmbient+v.illumination(point)/visionLitLevel)
	}
	return visibilityHidden + (1-visibilityHidden)*seen*lit
}

// illumination — освещённость точки всеми источниками кадра;
// здание между источником и точкой свет не пропускает
func (v vision) illumination(point types.Position) float64 {
	total := 0.0
	for _, light := range v.lights {
		if amount := light.IlluminationAt(point); amount > 0 {
			total += amount * v.grid.sight(light.Position, point)
		}
	}
	return total
}

func bulletCenter(bullet *types.BulletEntity) types.Position {
	size := bullet.GetSize()
	return types.Position{
		X: bullet.Position.X + float64(size.Width)/2,
		Y: bullet.Position.Y + float64(size.Height)/2,
	}
}

// tankViewPoints — центр танка и его углы с отступом от края
func tankViewPoints(tank *types.TankEntity) [5]types.Position {
	center := tankCenter(tank)
	halfWidth := float64(tank.Size.Width)/2 - visionTankInset
	halfHeight := float64(tank.Size.Height)/2 - visionTankInset
	return [5]types.Position{
		center,
		{X: center.X - halfWidth, Y: center.Y - halfHeight},
		{X: center.X + halfWidth, Y: center.Y - halfHeight},
		{X: center.X - halfWidth, Y: center.Y + halfHeight},
		{X: center.X + halfWidth, Y: center.Y + halfHeight},
	}
}

// visionGrid — карта препятствий взгляду в клетках visionCell
type visionGrid struct {
	cells   []uint8
	columns int
	rows    int
}

func newVisionGrid(blocks types.MapBlocks, size types.Size) visionGrid {
	grid := visionGrid{
		columns: (size.Width + visionCell - 1) / visionCell,
		rows:    (size.Height + visionCell - 1) / visionCell,
	}
	grid.cells = make([]uint8, grid.columns*grid.rows)
	for _, block := range blocks {
		if block == nil || block.Data == nil {
			continue
		}
		var cell uint8
		switch block.Data.Name {
		case types.Brick, types.Steel:
			cell = visionCellWall
		case types.Forest:
			cell = visionCellForest
		default:
			continue
		}
		blockSize := block.GetSize()
		left := int(block.Position.X) / visionCell
		top := int(block.Position.Y) / visionCell
		right := (int(block.Position.X) + blockSize.Width + visionCell - 1) / visionCell
		bottom := (int(block.Position.Y) + blockSize.Height + visionCell - 1) / visionCell
		for y := max(top, 0); y < min(bottom, grid.rows); y++ {
			for x := max(left, 0); x < min(right, grid.columns); x++ {
				// Здание важнее леса в той же клетке
				grid.cells[y*grid.columns+x] = max(
					grid.cells[y*grid.columns+x],
					cell,
				)
			}
		}
	}
	return grid
}

func (g visionGrid) at(point types.Position) uint8 {
	x, y := int(point.X)/visionCell, int(point.Y)/visionCell
	if point.X < 0 || point.Y < 0 || x >= g.columns || y >= g.rows {
		return visionCellEmpty
	}
	return g.cells[y*g.columns+x]
}

// sight — доля взгляда, дошедшая от eye до target: здание закрывает
// целиком, лес приглушает; у самой цели луч не проверяется — пуля
// в момент попадания касается стены
func (g visionGrid) sight(eye, target types.Position) float64 {
	distance := math.Hypot(target.X-eye.X, target.Y-eye.Y)
	forest := false
	for travelled := visionStep; travelled < distance-visionTargetGap; travelled += visionStep {
		t := travelled / distance
		point := types.Position{
			X: eye.X + (target.X-eye.X)*t,
			Y: eye.Y + (target.Y-eye.Y)*t,
		}
		switch g.at(point) {
		case visionCellWall:
			return 0
		case visionCellForest:
			if travelled > visionForestReach &&
				distance-travelled > visionForestReach {
				forest = true
			}
		}
	}
	if forest {
		return forestVisibility
	}
	return 1
}
