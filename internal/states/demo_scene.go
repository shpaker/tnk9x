package states

import (
	"github.com/hajimehoshi/ebiten/v2"

	"github.com/shpaker/tnk9x/internal/interfaces"
	"github.com/shpaker/tnk9x/internal/types"
)

// Тайминги демо-сцены в тиках (60 в секунду)
const (
	// demoBlocksPerTick — сколько блоков заголовка встаёт за кадр
	// при сборке
	demoBlocksPerTick = 1
	// demoRepairTicks — как часто чинятся отколотые и снесённые
	// кирпичи заголовка
	demoRepairTicks = 180
	// demoTickSeconds — шаг симуляции: сцена идёт с постоянной
	// скоростью независимо от реального TPS
	demoTickSeconds = 1.0 / 60.0
)

// DemoSceneRenderer — контракт рендера поля демо-сцены,
// определён у потребителя: только поле, без панели HUD
type DemoSceneRenderer interface {
	DrawAll(screen *ebiten.Image)
}

// DemoSceneDependencies — граф демо-сцены; собирается composition
// root'ом из того же графа уровня, что и игра, все поля обязательны
type DemoSceneDependencies struct {
	// Use Cases
	TankCommonUseCases    interfaces.ITankCommonUseCases
	RenderUseCases        interfaces.IRenderUseCases
	TankLifecycleUseCases interfaces.ITankLifecycleUseCases
	TilesUseCases         interfaces.ITilesUseCases
	StageUseCases         interfaces.IStageUseCases
	SoundUseCases         interfaces.ISoundUseCases
	LightingUseCases      interfaces.ILightingUseCases
	VisualEffectsUseCases interfaces.IVisualEffectsUseCases
	MapUseCases           interfaces.IMapUseCases
	BulletUseCases        interfaces.IBulletUseCases

	// Adapters
	EnemyInputAdapter interfaces.IAiInputAdapter
	Renderer          DemoSceneRenderer

	// Entities
	// TitleBlocks — блоки названия игры в порядке сборки
	TitleBlocks types.MapBlocks
}

// DemoScene — живая сцена за главным меню: название игры собирается
// из блоков, затем по полю ездят и стреляют по танку каждого типа
// врагов под управлением настоящего ИИ. Игроков и штаба нет,
// звуков нет; разбитые кирпичи названия со временем встают на место
type DemoScene struct {
	// Use Cases
	tankCommonUseCases    interfaces.ITankCommonUseCases
	renderUseCases        interfaces.IRenderUseCases
	tankLifecycleUseCases interfaces.ITankLifecycleUseCases
	tilesUseCases         interfaces.ITilesUseCases
	stageUseCases         interfaces.IStageUseCases
	soundUseCases         interfaces.ISoundUseCases
	lightingUseCases      interfaces.ILightingUseCases
	visualEffectsUseCases interfaces.IVisualEffectsUseCases
	mapUseCases           interfaces.IMapUseCases
	bulletUseCases        interfaces.IBulletUseCases
	// Adapters
	enemyInputAdapter interfaces.IAiInputAdapter
	renderer          DemoSceneRenderer
	// Entities
	titleBlocks types.MapBlocks

	ticks     uint
	assembled int
}

// NewDemoScene убирает название с поля: оно соберётся заново
// по блоку за кадр
func NewDemoScene(deps DemoSceneDependencies) *DemoScene {
	for _, block := range deps.TitleBlocks {
		_ = deps.MapUseCases.RemoveBlock(block)
	}

	return &DemoScene{
		tankCommonUseCases:    deps.TankCommonUseCases,
		renderUseCases:        deps.RenderUseCases,
		tankLifecycleUseCases: deps.TankLifecycleUseCases,
		tilesUseCases:         deps.TilesUseCases,
		stageUseCases:         deps.StageUseCases,
		soundUseCases:         deps.SoundUseCases,
		lightingUseCases:      deps.LightingUseCases,
		visualEffectsUseCases: deps.VisualEffectsUseCases,
		mapUseCases:           deps.MapUseCases,
		bulletUseCases:        deps.BulletUseCases,
		enemyInputAdapter:     deps.EnemyInputAdapter,
		renderer:              deps.Renderer,
		titleBlocks:           deps.TitleBlocks,
	}
}

// IsAssembled — название собрано, танки выехали
func (s *DemoScene) IsAssembled() bool {
	return s.assembled >= len(s.titleBlocks)
}

// Update продвигает сцену на кадр: сборка названия, затем танки,
// их ИИ, пули, эффекты и починка кирпичей
func (s *DemoScene) Update() {
	s.ticks++
	if !s.IsAssembled() {
		s.assemble()
	} else {
		s.updateTanks()
		if s.ticks%demoRepairTicks == 0 {
			s.repairTitle()
		}
	}

	s.tilesUseCases.UpdateAnimations()
	s.lightingUseCases.UpdateHeadlights()
	s.visualEffectsUseCases.Update()

	// Сцена за меню беззвучна: события звука копятся в очереди
	// уровня, их нужно забирать каждый кадр
	_ = s.soundUseCases.GetEvents()
}

func (s *DemoScene) Draw(screen *ebiten.Image) {
	s.renderer.DrawAll(screen)
}

// assemble ставит очередные блоки названия; с последним блоком
// на поле выезжают танки
func (s *DemoScene) assemble() {
	for range demoBlocksPerTick {
		if s.IsAssembled() {
			break
		}
		s.mapUseCases.RestoreBlock(s.titleBlocks[s.assembled], nil)
		s.assembled++
	}
	if s.IsAssembled() {
		for _, enemy := range s.stageUseCases.SpawnInitialEnemyTanks() {
			if enemy != nil {
				s.addTank(enemy)
			}
		}
	}
}

// updateTanks — жизненный цикл, движение, появление и ИИ танков
func (s *DemoScene) updateTanks() {
	_ = s.tankLifecycleUseCases.UpdateAllTanksLifecycle()
	_ = s.tankCommonUseCases.UpdateAllTanks(demoTickSeconds)
	s.stageUseCases.UpdateGameObjects(demoTickSeconds)

	if spawned := s.stageUseCases.TrySpawnEnemy(); spawned != nil {
		s.addTank(spawned)
	}
	s.enemyInputAdapter.Update(demoTickSeconds)

	var blinking []types.IBlink
	for _, tank := range s.tankCommonUseCases.GetAllTanks() {
		if tank != nil && s.renderUseCases.IsTankBlinking(tank) {
			blinking = append(blinking, tank)
		}
	}
	if len(blinking) > 0 {
		s.renderUseCases.UpdateBlink(blinking)
	}
}

// addTank передаёт танк ИИ. За меню танки не мигают: бонус снимаем
// (по классике его несёт четвёртый враг), а тяжёлому оставляем одно
// попадание — без запаса брони нет и подсветки здоровья. Враги друг
// друга не подбивают, так что на сцене это незаметно
func (s *DemoScene) addTank(tank *types.TankEntity) {
	tank.SetWithBonus(false)
	tank.SetHitPoints(1)
	s.enemyInputAdapter.AddTank(tank)
}

// repairTitle возвращает разбитые кирпичи названия на место, если
// их клетку сейчас не занимают танк или пуля
func (s *DemoScene) repairTitle() {
	var obstacles []types.IEntityCollider
	for _, tank := range s.tankCommonUseCases.GetAllTanks() {
		if tank != nil && tank.IsActive() {
			obstacles = append(obstacles, tank)
		}
	}
	for _, bullet := range s.bulletUseCases.GetBullets() {
		if bullet != nil {
			obstacles = append(obstacles, bullet)
		}
	}

	for _, block := range s.titleBlocks {
		if !s.mapUseCases.IsBlockIntact(block) {
			s.mapUseCases.RestoreBlock(block, obstacles)
		}
	}
}
