package use_cases_test

import (
	"errors"
	"testing"

	"github.com/shpaker/tnk9x/internal/types"
	"github.com/shpaker/tnk9x/internal/types/session_entities"
	"github.com/shpaker/tnk9x/internal/use_cases"
)

type fakeScriptEngine struct {
	decision types.EnemyAIDecision
	err      error

	gotContext              types.EnemyAIContext
	updateEnemyAICallsCount int
}

func (e *fakeScriptEngine) LoadScript(source string) error { return nil }

func (e *fakeScriptEngine) SetGlobalNumber(name string, value float64) {}

func (e *fakeScriptEngine) UpdateEnemyAI(
	context types.EnemyAIContext,
) (types.EnemyAIDecision, error) {
	e.updateEnemyAICallsCount++
	e.gotContext = context
	return e.decision, e.err
}

func (e *fakeScriptEngine) Close() {}

type fakeAITanksRepository struct {
	players []*types.TankEntity
	enemies []*types.TankEntity
}

func (r *fakeAITanksRepository) SetPlayer(
	types.PlayerTankNum,
	*types.TankEntity,
) {
}

func (r *fakeAITanksRepository) GetPlayer(
	types.PlayerTankNum,
) *types.TankEntity {
	return nil
}

func (r *fakeAITanksRepository) HasPlayer(
	types.PlayerTankNum,
) bool {
	return false
}

func (r *fakeAITanksRepository) GetAllPlayers() []*types.TankEntity { return r.players }
func (r *fakeAITanksRepository) GetActivePlayerTanks() []*types.TankEntity {
	return r.players
}

func (r *fakeAITanksRepository) AddEnemy(enemy *types.TankEntity) {
	r.enemies = append(r.enemies, enemy)
}

func (r *fakeAITanksRepository) GetAllEnemies() []*types.TankEntity { return r.enemies }
func (r *fakeAITanksRepository) GetAllTanks() []*types.TankEntity {
	return append(append([]*types.TankEntity{}, r.players...), r.enemies...)
}

type fakeAIBulletsRepository struct {
	bullets []*types.BulletEntity
}

func (r *fakeAIBulletsRepository) AddBullet(bullet *types.BulletEntity) error {
	r.bullets = append(r.bullets, bullet)
	return nil
}

func (r *fakeAIBulletsRepository) GetAllBullets() []*types.BulletEntity { return r.bullets }
func (r *fakeAIBulletsRepository) RemoveBullet(*types.BulletEntity) error {
	return nil
}

func newTestTank(role types.TankRole, x, y float64) *types.TankEntity {
	tank := types.NewDefaultTankEntity(role, types.DirectionUp)
	tank.Position = types.Position{X: x, Y: y}
	tank.State = types.TankStateStopped
	return &tank
}

func newAIUseCases(
	engine *fakeScriptEngine,
	tanks *fakeAITanksRepository,
	bullets *fakeAIBulletsRepository,
) *use_cases.AIUseCases {
	stageSession := session_entities.NewStageSessionEntity()
	stageSession.SetStageNumber(7)
	mapEntity := types.NewMapEntity(
		types.Size{Width: 64, Height: 64},
		types.MapBlocks{
			types.NewBlockEntity(string(types.Steel), 8, 8, 8, nil),
		},
		nil,
	)
	hq := &types.HQEntity{
		Position: types.Position{X: 24, Y: 48},
		State:    types.HQStateIntact,
	}
	return use_cases.NewAIUseCases(
		stageSession,
		mapEntity,
		hq,
		tanks,
		bullets,
		engine,
	)
}

func TestAIUseCasesExecuteAINilTank(t *testing.T) {
	engine := &fakeScriptEngine{}
	uc := newAIUseCases(
		engine,
		&fakeAITanksRepository{},
		&fakeAIBulletsRepository{},
	)

	if _, err := uc.ExecuteAI(nil, 0); err == nil {
		t.Fatal("expected error for nil tank")
	}
	if engine.updateEnemyAICallsCount != 0 {
		t.Fatal("engine must not be called for nil tank")
	}
}

func TestAIUseCasesExecuteAIBuildsContext(t *testing.T) {
	engine := &fakeScriptEngine{
		decision: types.EnemyAIDecision{
			Direction: types.DirectionLeft,
			Move:      true,
		},
	}
	tanks := &fakeAITanksRepository{}
	self := newTestTank(types.TankRoleEnemy, 48, 0)
	self.SetID(1)
	self.SetSpecs(types.NewSpecsEntity(2, 1, true, 2, 1))
	ally := newTestTank(types.TankRoleEnemy, 0, 0)
	ally.SetID(2)
	exploding := newTestTank(types.TankRoleEnemy, 16, 0)
	exploding.State = types.TankStateExploding
	tanks.AddEnemy(self)
	tanks.AddEnemy(ally)
	tanks.AddEnemy(exploding)
	player := newTestTank(types.TankRolePlayer1, 16, 48)
	tanks.players = []*types.TankEntity{player}

	bullets := &fakeAIBulletsRepository{}
	_ = bullets.AddBullet(types.NewBulletEntity(
		types.Position{X: 20, Y: 30}, types.Size{Width: 4, Height: 4},
		types.SURFACE, nil, types.DirectionUp, nil, player,
	))

	uc := newAIUseCases(engine, tanks, bullets)
	decision, err := uc.ExecuteAI(self, 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decision != engine.decision {
		t.Fatalf("decision not passed through: %+v", decision)
	}

	context := engine.gotContext
	if context.Tick != 42 || context.StageNumber != 7 {
		t.Fatalf("tick/stage not passed: %+v", context)
	}
	if context.Self.ID != 1 || context.Self.Level != 2 ||
		!context.Self.Reinforced {
		t.Fatalf("self not described: %+v", context.Self)
	}
	if len(context.Enemies) != 1 || context.Enemies[0].ID != 2 {
		t.Fatalf("only active allies expected: %+v", context.Enemies)
	}
	if len(context.Players) != 1 ||
		context.Players[0].Position != player.Position {
		t.Fatalf("players not passed: %+v", context.Players)
	}
	if len(context.Bullets) != 1 || context.Bullets[0].FromEnemy {
		t.Fatalf("bullets not passed: %+v", context.Bullets)
	}
	if !context.HQ.Intact || context.HQ.Position.Y != 48 {
		t.Fatalf("hq not passed: %+v", context.HQ)
	}
	if context.Grid.GetBlockAt(10, 10) != types.Steel {
		t.Fatal("grid must be built from map blocks")
	}
}

func TestAIUseCasesExecuteAIPropagatesError(t *testing.T) {
	engine := &fakeScriptEngine{err: errors.New("script failed")}
	uc := newAIUseCases(
		engine,
		&fakeAITanksRepository{},
		&fakeAIBulletsRepository{},
	)

	if _, err := uc.ExecuteAI(newTestTank(types.TankRoleEnemy, 0, 0), 0); err == nil {
		t.Fatal("expected engine error to propagate")
	}
}
