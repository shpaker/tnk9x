package use_cases_test

import (
	"math"
	"testing"

	"github.com/shpaker/tnk9x/internal/testutil"
	"github.com/shpaker/tnk9x/internal/types"
	image_providers "github.com/shpaker/tnk9x/internal/types/image_providers"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

// stubBulletList отдаёт заданный список пуль
type stubBulletList struct {
	bullets []*types.BulletEntity
}

func (s *stubBulletList) ShootBullet(
	tank *types.TankEntity,
) (bool, error) {
	return false, nil
}

func (s *stubBulletList) UpdateBullets(
	dt float64,
) error {
	return nil
}

func (s *stubBulletList) GetBullets() []*types.BulletEntity { return s.bullets }
func (s *stubBulletList) RemoveBullet(bullet *types.BulletEntity) error {
	return nil
}

// stubVisibleBonuses отдаёт заданный список видимых бонусов
type stubVisibleBonuses struct {
	bonuses []*types.BonusEntity
}

func (s *stubVisibleBonuses) Apply(
	bonus *types.BonusEntity,
	tank *types.TankEntity,
) {
}
func (s *stubVisibleBonuses) UpdateEffects() {}
func (s *stubVisibleBonuses) SpawnRandomBonusEntity(
	position types.Position,
) *types.BonusEntity {
	return nil
}

func (s *stubVisibleBonuses) VisibleBonuses() []*types.BonusEntity {
	return s.bonuses
}

type lightingTestEnv struct {
	tankCommon *recordingTankCommon
	bullets    *stubBulletList
	bonuses    *stubVisibleBonuses
	hq         *types.HQEntity
	effects    *testutil.FakeVisualEffectsUseCases
	lighting   *use_cases.LightingUseCases
}

func newLightingTestEnv() *lightingTestEnv {
	tankCommon := &recordingTankCommon{}
	bullets := &stubBulletList{}
	bonuses := &stubVisibleBonuses{}
	hq := &types.HQEntity{
		Position: types.Position{X: 96, Y: 192},
		Size:     types.Size{Width: 16, Height: 16},
		State:    types.HQStateDestroyed,
	}

	effects := &testutil.FakeVisualEffectsUseCases{}

	return &lightingTestEnv{
		tankCommon: tankCommon,
		bullets:    bullets,
		bonuses:    bonuses,
		hq:         hq,
		effects:    effects,
		lighting: use_cases.NewLightingUseCases(
			tankCommon,
			bullets,
			&stubHQUseCases{hq: hq},
			bonuses,
			effects,
		),
	}
}

func newBullet(x, y float64) *types.BulletEntity {
	return &types.BulletEntity{
		Position: types.Position{X: x, Y: y},
		Size:     types.Size{Width: 4, Height: 4},
	}
}

// Источник света ставится в центр сущности
func TestLightingUseCases_GetLights_CenteredOnEntity(t *testing.T) {
	env := newLightingTestEnv()
	env.bullets.bullets = []*types.BulletEntity{newBullet(10, 20)}

	lights := env.lighting.GetLights()

	if len(lights) != 1 {
		t.Fatalf("источников %d, ожидался 1", len(lights))
	}
	want := types.Position{X: 12, Y: 22}
	if lights[0].Position != want {
		t.Errorf("позиция %v, ожидалась %v", lights[0].Position, want)
	}
}

// Уничтоженные танк и штаб не светят, целый штаб светит
func TestLightingUseCases_GetLights_SkipsDestroyed(t *testing.T) {
	env := newLightingTestEnv()
	env.tankCommon.tanks = []*types.TankEntity{
		newEnemyTankInState(types.TankStateExploded),
	}

	if lights := env.lighting.GetLights(); len(lights) != 0 {
		t.Fatalf("источников %d, ожидалось 0", len(lights))
	}

	env.hq.State = types.HQStateIntact
	if lights := env.lighting.GetLights(); len(lights) != 1 {
		t.Fatalf("источников %d, ожидался свет целого штаба", len(lights))
	}
}

// Игрок светит фарой-конусом вперёд по стволу и аурой вокруг себя
func TestLightingUseCases_GetLights_PlayerHeadlightAndAura(t *testing.T) {
	env := newLightingTestEnv()
	env.tankCommon.tanks = []*types.TankEntity{
		newPlayerTank(types.TankRolePlayer1),
	}

	lights := env.lighting.GetLights()

	if len(lights) != 2 {
		t.Fatalf("источников %d, ожидались фара и аура", len(lights))
	}
	headlight, aura := lights[0], lights[1]
	if !headlight.IsCone() || aura.IsCone() {
		t.Fatal("первой ожидалась фара-конус, второй — всенаправленная аура")
	}
	if headlight.Radius <= aura.Radius {
		t.Error("фара должна светить дальше ауры")
	}
	// Танк 16x16 в (0,0) смотрит вверх: фара у среза ствола, ось вверх
	if math.Abs(headlight.Direction.X) > 1e-9 || headlight.Direction.Y != -1 {
		t.Errorf("ось фары %v, ожидалась вверх", headlight.Direction)
	}
	if headlight.Position.Y >= aura.Position.Y {
		t.Error("фара должна быть вынесена вперёд от центра танка")
	}
}

// Щит заменяет ауру танка, фара остаётся
func TestLightingUseCases_GetLights_ShieldReplacesAura(t *testing.T) {
	env := newLightingTestEnv()
	plain := newPlayerTank(types.TankRolePlayer1)
	shielded := newPlayerTank(types.TankRolePlayer1)
	shielded.ActivateShield(60)
	env.tankCommon.tanks = []*types.TankEntity{plain, shielded}

	lights := env.lighting.GetLights()

	if len(lights) != 4 {
		t.Fatalf("источников %d, ожидалось 4", len(lights))
	}
	var omni []types.LightEntity
	for _, light := range lights {
		if !light.IsCone() {
			omni = append(omni, light)
		}
	}
	if len(omni) != 2 || omni[0].Color == omni[1].Color {
		t.Error("свет щита должен отличаться от ауры танка")
	}
}

// Фара доворачивается за стволом плавно, за несколько тиков
func TestLightingUseCases_UpdateHeadlights_TurnsSmoothly(t *testing.T) {
	env := newLightingTestEnv()
	tank := newPlayerTank(types.TankRolePlayer1)
	env.tankCommon.tanks = []*types.TankEntity{tank}
	env.lighting.UpdateHeadlights()

	tank.Direction = types.DirectionRight
	env.lighting.UpdateHeadlights()

	axis := env.lighting.GetLights()[0].Direction
	if axis.X <= 0 || axis.Y >= 0 {
		t.Errorf("после тика ось %v, ожидалась между верхом и правом", axis)
	}

	for range 30 {
		env.lighting.UpdateHeadlights()
	}
	axis = env.lighting.GetLights()[0].Direction
	if math.Abs(axis.X-1) > 1e-9 || math.Abs(axis.Y) > 1e-9 {
		t.Errorf("ось %v, ожидалась вправо", axis)
	}
}

// Вспышки эффектов светят наравне с остальными источниками
func TestLightingUseCases_GetLights_IncludesFlashes(t *testing.T) {
	env := newLightingTestEnv()
	env.effects.FlashLights = []types.LightEntity{{Radius: 40, Intensity: 1}}

	if lights := env.lighting.GetLights(); len(lights) != 1 ||
		lights[0].Radius != 40 {
		t.Errorf("источники %v, ожидалась вспышка", lights)
	}
}

// Вспышка взрыва растёт к концу анимации
func TestLightingUseCases_GetLights_ExplosionGrows(t *testing.T) {
	env := newLightingTestEnv()
	tank := newEnemyTankInState(types.TankStateExploding)
	anim := image_providers.NewAnimationProvider(types.AnimationData{
		{Image: "explosion_1"},
		{Image: "explosion_2"},
		{Image: "explosion_3"},
	})
	tank.Image = anim
	env.tankCommon.tanks = []*types.TankEntity{tank}

	first := env.lighting.GetLights()[0].Radius
	anim.CurrentFrame = 2
	last := env.lighting.GetLights()[0].Radius

	if last <= first {
		t.Errorf(
			"радиус %v на последнем кадре, ожидался больше %v",
			last,
			first,
		)
	}
}

// При переполнении остаются самые важные источники: взрывы и пули
// вытесняют танки врагов
func TestLightingUseCases_GetLights_LimitKeepsPriority(t *testing.T) {
	env := newLightingTestEnv()
	for range types.MaxLights {
		env.tankCommon.tanks = append(
			env.tankCommon.tanks,
			newEnemyTankInState(types.TankStateMoving),
		)
	}
	env.tankCommon.tanks = append(
		env.tankCommon.tanks,
		newEnemyTankInState(types.TankStateExploding),
	)
	env.bullets.bullets = []*types.BulletEntity{newBullet(0, 0)}
	env.bonuses.bonuses = []*types.BonusEntity{
		types.NewBonusEntity(
			types.BonusTypeStar,
			types.Position{},
			types.Size{Width: 16, Height: 16},
			nil,
		),
	}

	lights := env.lighting.GetLights()

	if len(lights) != types.MaxLights {
		t.Fatalf("источников %d, ожидалось %d", len(lights), types.MaxLights)
	}
	explosion, bullet := lights[0], lights[1]
	if explosion.Intensity <= bullet.Intensity {
		t.Error("первым должен идти взрыв")
	}
	if bullet.Position != (types.Position{X: 2, Y: 2}) {
		t.Errorf("вторым ожидалась пуля, позиция %v", bullet.Position)
	}
}

// Кирпич и сталь бросают тень, вода и лёд дают блик, лес затеняет частично
func TestLightingUseCases_GetMaterial(t *testing.T) {
	lighting := newLightingTestEnv().lighting

	cases := []struct {
		block      types.BlockType
		opaque     bool
		reflective bool
	}{
		{types.Brick, true, false},
		{types.Steel, true, true},
		{types.Forest, false, false},
		{types.Water, false, true},
		{types.Ice, false, true},
	}
	for _, c := range cases {
		material := lighting.GetMaterial(c.block)
		if (material.Opacity == 1) != c.opaque {
			t.Errorf("%s: непрозрачность %v", c.block, material.Opacity)
		}
		if (material.Reflectivity > 0) != c.reflective {
			t.Errorf("%s: отражение %v", c.block, material.Reflectivity)
		}
	}

	// Сталь — металл: блестит сильнее матового кирпича
	if lighting.GetMaterial(types.Steel).Reflectivity <=
		lighting.GetMaterial(types.Brick).Reflectivity {
		t.Error("сталь должна отражать свет сильнее кирпича")
	}
	if forest := lighting.GetMaterial(types.Forest); forest.Opacity <= 0 {
		t.Error("лес должен частично затенять")
	}
	if unknown := lighting.GetMaterial("unknown"); unknown != (types.SurfaceMaterial{}) {
		t.Errorf("неизвестный блок: материал %v", unknown)
	}
}

// Фара и аура игрока не вытесняются даже взрывами
func TestLightingUseCases_GetLights_PlayerFirst(t *testing.T) {
	env := newLightingTestEnv()
	for range types.MaxLights {
		env.tankCommon.tanks = append(
			env.tankCommon.tanks,
			newEnemyTankInState(types.TankStateExploding),
		)
	}
	env.tankCommon.tanks = append(
		env.tankCommon.tanks,
		newPlayerTank(types.TankRolePlayer1),
	)

	lights := env.lighting.GetLights()

	if len(lights) != types.MaxLights {
		t.Fatalf("источников %d, ожидалось %d", len(lights), types.MaxLights)
	}
	if !lights[0].IsCone() || lights[1].IsCone() {
		t.Error("первыми должны идти фара и аура игрока")
	}
}

// Враг светит фарой короче и тусклее игрока; фара врага уступает пулям
func TestLightingUseCases_GetLights_EnemyHeadlight(t *testing.T) {
	env := newLightingTestEnv()
	env.tankCommon.tanks = []*types.TankEntity{
		newEnemyTankInState(types.TankStateMoving),
		newPlayerTank(types.TankRolePlayer1),
	}
	env.bullets.bullets = []*types.BulletEntity{newBullet(0, 0)}

	lights := env.lighting.GetLights()

	if len(lights) != 5 {
		t.Fatalf("источников %d, ожидалось 5", len(lights))
	}
	playerHeadlight, bullet, enemyHeadlight := lights[0], lights[2], lights[3]
	if !enemyHeadlight.IsCone() {
		t.Fatal("после пули ожидалась фара врага")
	}
	if bullet.IsCone() {
		t.Error("пуля должна идти раньше фары врага")
	}
	if enemyHeadlight.Radius >= playerHeadlight.Radius ||
		enemyHeadlight.Intensity >= playerHeadlight.Intensity {
		t.Error("фара врага должна быть короче и тусклее фары игрока")
	}
}

// Фара врага доворачивается за стволом так же, как у игрока
func TestLightingUseCases_UpdateHeadlights_TurnsEnemy(t *testing.T) {
	env := newLightingTestEnv()
	enemy := newEnemyTankInState(types.TankStateMoving)
	env.tankCommon.tanks = []*types.TankEntity{enemy}
	env.lighting.UpdateHeadlights()
	start := enemy.GetHeadlightAngle()

	enemy.Direction = types.DirectionRight
	env.lighting.UpdateHeadlights()

	if enemy.GetHeadlightAngle() == start {
		t.Error("фара врага не повернулась")
	}
}

// Свечение щита мерцает: соседние фазы отличаются яркостью
func TestLightingUseCases_GetLights_ShieldFlickers(t *testing.T) {
	env := newLightingTestEnv()
	tank := newPlayerTank(types.TankRolePlayer1)
	env.tankCommon.tanks = []*types.TankEntity{tank}

	shieldIntensity := func(ticks uint) float64 {
		tank.ActivateShield(ticks)
		for _, light := range env.lighting.GetLights() {
			if !light.IsCone() {
				return light.Intensity
			}
		}
		return 0
	}

	if shieldIntensity(10) == shieldIntensity(12) {
		t.Error("свечение щита не мерцает")
	}
}

// Усиленная пуля светится иначе, чем обычная
func TestLightingUseCases_GetLights_ReinforcedBullet(t *testing.T) {
	env := newLightingTestEnv()
	regular := newBullet(10, 20)
	reinforced := types.NewBulletEntity(
		types.Position{X: 100, Y: 20},
		types.Size{Width: 4, Height: 4},
		types.GROUND,
		nil,
		types.DirectionUp,
		types.NewSpecsEntity(2, 32, true, 150, 1),
		nil,
	)
	env.bullets.bullets = []*types.BulletEntity{regular, reinforced}

	lights := env.lighting.GetLights()

	if len(lights) != 2 {
		t.Fatalf("lights %d, want 2", len(lights))
	}
	if lights[0].Color == lights[1].Color {
		t.Error("reinforced bullet light matches the regular one")
	}
}

// Враг с бонусом вспыхивает красным только в видимой фазе мигания
func TestLightingUseCases_GetLights_BonusEnemyFlashes(t *testing.T) {
	env := newLightingTestEnv()
	enemy := newEnemyTankInState(types.TankStateMoving)
	enemy.SetWithBonus(true)
	env.tankCommon.tanks = []*types.TankEntity{enemy}

	glow := func() types.LightEntity {
		for _, light := range env.lighting.GetLights() {
			if !light.IsCone() {
				return light
			}
		}
		t.Fatal("нет собственного света танка")
		return types.LightEntity{}
	}

	if glow().Intensity > 1 {
		t.Error("в скрытой фазе танк не должен вспыхивать")
	}
	for !enemy.GetBlinkFlag() {
		enemy.UpdateBlink()
	}
	if light := glow(); light.Intensity <= 1 || light.Color.R <= light.Color.G {
		t.Error("в видимой фазе ожидалась яркая красная вспышка")
	}
}
