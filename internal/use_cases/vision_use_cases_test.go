package use_cases_test

import (
	"math"
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

// stubLighting отдаёт заданные источники и зрителей
type stubLighting struct {
	lights  []types.LightEntity
	viewers []types.ViewerEntity
}

func (s *stubLighting) GetLights() []types.LightEntity { return s.lights }
func (s *stubLighting) GetMaterial(types.BlockType) types.SurfaceMaterial {
	return types.SurfaceMaterial{}
}
func (s *stubLighting) GetViewers() []types.ViewerEntity { return s.viewers }
func (s *stubLighting) UpdateHeadlights()                {}

type visionTestEnv struct {
	tankCommon *recordingTankCommon
	bullets    *stubBulletList
	mapEntity  *types.MapEntity
	lighting   *stubLighting
	vision     *use_cases.VisionUseCases
}

// Зритель в (104, 150) смотрит вверх; карта 208x208 без блоков
func newVisionTestEnv() *visionTestEnv {
	tankCommon := &recordingTankCommon{}
	bullets := &stubBulletList{}
	mapEntity := types.NewMapEntity(
		types.Size{Width: 208, Height: 208},
		nil,
		nil,
	)
	lighting := &stubLighting{
		viewers: []types.ViewerEntity{{
			Position:  types.Position{X: 104, Y: 150},
			Direction: types.Position{X: 0, Y: -1},
		}},
	}
	return &visionTestEnv{
		tankCommon: tankCommon,
		bullets:    bullets,
		mapEntity:  mapEntity,
		lighting:   lighting,
		vision: use_cases.NewVisionUseCases(
			tankCommon,
			bullets,
			&stubMapUseCases{mapEntity: mapEntity},
			lighting,
		),
	}
}

// enemyAt ставит врага центром в точку (x, y)
func (env *visionTestEnv) enemyAt(x, y float64) *types.TankEntity {
	tank := newEnemyTankInState(types.TankStateStopped)
	tank.Position = types.Position{X: x - 8, Y: y - 8}
	env.tankCommon.tanks = append(env.tankCommon.tanks, tank)
	return tank
}

// hidden — видимость объекта вне взгляда игрока
const hidden = 0.5

// lamp — всенаправленный свет в точке (x, y)
func lamp(x, y float64) types.LightEntity {
	return types.LightEntity{
		Position:  types.Position{X: x, Y: y},
		Radius:    40,
		Intensity: 1.3,
	}
}

// Освещённый враг впереди виден полностью, неосвещённый — тускло,
// враг за спиной — лишь силуэтом
func TestVisionUseCases_SeenAndLit(t *testing.T) {
	env := newVisionTestEnv()
	lit := env.enemyAt(104, 100)
	dark := env.enemyAt(40, 100)
	behind := env.enemyAt(104, 206)
	env.lighting.lights = []types.LightEntity{lamp(104, 100)}

	env.vision.UpdateVisibility()

	if got := lit.GetVisibility(); got < 0.99 {
		t.Errorf("освещённый враг впереди: видимость %v", got)
	}
	if got := dark.GetVisibility(); got <= hidden || got >= 0.99 {
		t.Errorf("враг впереди в темноте: видимость %v, ожидалась тусклая", got)
	}
	if got := behind.GetVisibility(); math.Abs(got-hidden) > 1e-9 {
		t.Errorf("враг за спиной: видимость %v, ожидался силуэт", got)
	}
}

// Здание закрывает взгляд, даже если враг освещён
func TestVisionUseCases_WallBlocksSight(t *testing.T) {
	env := newVisionTestEnv()
	for x := 88.0; x < 120; x += 8 {
		env.mapEntity.AddBlock(brick(x, 120))
	}
	enemy := env.enemyAt(104, 100)
	env.lighting.lights = []types.LightEntity{lamp(104, 100)}

	env.vision.UpdateVisibility()

	if got := enemy.GetVisibility(); math.Abs(got-hidden) > 1e-9 {
		t.Errorf("враг за стеной: видимость %v, ожидался силуэт", got)
	}
}

// Пуля светится сама: видна везде, куда доходит взгляд, и одинаково
// для своих и чужих
func TestVisionUseCases_BulletsAreEmissive(t *testing.T) {
	env := newVisionTestEnv()
	ahead := newBullet(102, 20)
	behind := newBullet(102, 204)
	env.bullets.bullets = []*types.BulletEntity{ahead, behind}

	env.vision.UpdateVisibility()

	if got := ahead.GetVisibility(); got < 0.99 {
		t.Errorf("пуля впереди в темноте: видимость %v", got)
	}
	if got := behind.GetVisibility(); math.Abs(got-hidden) > 1e-9 {
		t.Errorf("пуля за спиной: видимость %v", got)
	}
}

// Враг проявляется и исчезает плавно, а не скачком
func TestVisionUseCases_FadesSmoothly(t *testing.T) {
	env := newVisionTestEnv()
	enemy := env.enemyAt(104, 100)
	env.vision.UpdateVisibility()
	dim := enemy.GetVisibility()

	env.lighting.lights = []types.LightEntity{lamp(104, 100)}
	env.vision.UpdateVisibility()

	if got := enemy.GetVisibility(); got <= dim || got >= 0.99 {
		t.Errorf("после тика видимость %v, ожидалась между %v и 1", got, dim)
	}
}
