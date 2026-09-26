package use_cases_test

import (
	"testing"

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
) error {
	return nil
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

	return &lightingTestEnv{
		tankCommon: tankCommon,
		bullets:    bullets,
		bonuses:    bonuses,
		hq:         hq,
		lighting: use_cases.NewLightingUseCases(
			tankCommon,
			bullets,
			&stubHQUseCases{hq: hq},
			bonuses,
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

// Щит заменяет собственный свет танка
func TestLightingUseCases_GetLights_ShieldReplacesTankLight(t *testing.T) {
	env := newLightingTestEnv()
	plain := newPlayerTank(types.TankRolePlayer1)
	shielded := newPlayerTank(types.TankRolePlayer1)
	shielded.ActivateShield(60)
	env.tankCommon.tanks = []*types.TankEntity{plain, shielded}

	lights := env.lighting.GetLights()

	if len(lights) != 2 {
		t.Fatalf("источников %d, ожидалось 2", len(lights))
	}
	if lights[0].Color == lights[1].Color {
		t.Error("свет щита должен отличаться от света танка")
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
// вытесняют танки
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
	if explosion.Radius <= bullet.Radius {
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

	if forest := lighting.GetMaterial(types.Forest); forest.Opacity <= 0 {
		t.Error("лес должен частично затенять")
	}
	if unknown := lighting.GetMaterial("unknown"); unknown != (types.SurfaceMaterial{}) {
		t.Errorf("неизвестный блок: материал %v", unknown)
	}
}
