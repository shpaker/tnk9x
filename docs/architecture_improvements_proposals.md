# Предлагаемые решения архитектурных проблем

## ✅ Выполненные улучшения

### Рефакторинг Use Cases танка

#### Проблема
Изначально вся логика работы с танками была сосредоточена в одном большом `TankUseCases`, что создавало проблемы с разделением ответственности и тестированием.

#### Реализованное решение

**1. Разделение Use Cases на специализированные компоненты:**

- **`TankCommonUseCases`** — общие операции с танком (движение)
  - `Update(tank *types.TankEntity, dt float64) error`
  
- **`TankRenderUseCases`** — графика и рендеринг танка
  - `GetImageID() (string, error)`
  - `GetAnimationGetter() types.IImageIDGetter`
  - `SetAnimationGetter(animationGetter types.IImageIDGetter)`
  - `IsSpawnAnimationFinished() bool`
  - `IsExplosionAnimationFinished() bool`

- **`TankLifecycleUseCases`** — жизненный цикл танка (спавн, взрыв, анимации)
  - `Spawn(tank *types.TankEntity) error`
  - `Explode(tank *types.TankEntity) error`
  - `IsSpawnFinished(tank *types.TankEntity, currentTime float64)`
  - `IsExplosionFinished(tank *types.TankEntity)`

- **`TankActionsUseCases`** — действия танка (движение и боевые действия)
  - `Update(tank *types.TankEntity, dt float64) error`
  - `Rotate(tank *types.TankEntity, direction types.Direction) error`
  - `Move(tank *types.TankEntity) error`
  - `Stop(tank *types.TankEntity, byCollision bool)`
  - `IsStopped(tank *types.TankEntity) bool`
  - `Shoot(tank *types.TankEntity) error`
  - `ApplyDecision(tank *types.TankEntity, decision types.EnemyAIDecision)`

**2. Создание интерфейсов для всех Use Cases:**

Созданы интерфейсы в `internal/interfaces/use_cases.go`:
- `ITankCommonUseCases`
- `ITankRenderUseCases`
- `ITankLifecycleUseCases`
- `ITankActionsUseCases`

**3. Централизация работы с анимациями:**

Вся работа с `animationGetter` и графикой теперь осуществляется через `TankRenderUseCases`:
- `TankLifecycleUseCases` использует `ITankRenderUseCases` для установки и проверки анимаций
- Удалена прямая зависимость между компонентами
- Инкапсуляция работы с графикой в одном месте

**4. Обновление зависимостей:**

Все зависимости обновлены для использования интерфейсов:
- `GameStateUseCasesFacade` — использует интерфейсы вместо конкретных типов
- `RendererAdapter` — принимает `ITankRenderUseCases`
- `CollisionUseCases` — использует `ITankActionsUseCases` и `ITankCommonUseCases`
- `InputAdapters` — используют `ITankActionsUseCases`

#### Преимущества реализации

- ✅ **Чёткое разделение ответственности**: Каждый Use Case отвечает за свою область
- ✅ **Улучшенная тестируемость**: Можно легко создать mock-объекты для интерфейсов
- ✅ **Гибкость**: Легко заменить реализацию без изменения зависимостей
- ✅ **Соответствие принципам SOLID**: Single Responsibility, Dependency Inversion
- ✅ **Централизация графики**: Вся работа с рендерингом в одном месте

#### Файлы, затронутые изменениями

- `internal/use_cases/tank_common_use_cases.go`
- `internal/use_cases/tank_render_use_cases.go`
- `internal/use_cases/tank_lifecycle_use_cases.go`
- `internal/use_cases/tank_actions_use_cases.go`
- `internal/interfaces/use_cases.go`
- `internal/states/game_state_use_cases_facade.go`
- `internal/adapters/renderer_adapter.go`
- `internal/use_cases/collision_use_cases.go`
- `internal/adapters/input_adapters/keyboard_input_adapter.go`
- `internal/adapters/input_adapters/ai_input_adapter.go`
- `internal/use_cases/interfaces_check.go`

---

## Предлагаемые решения архитектурных проблем

## 🟡 Проблема 1: God Object в конструкторе фасада

### Текущая ситуация

`NewGameStateUseCasesFacade()` принимает **13 параметров**, что делает его сложным для использования и тестирования:

```go
func NewGameStateUseCasesFacade(
	mapsRepo interfaces.IMapsDataRepository,
	scriptsRepo interfaces.IScriptsRepository,
	levelNumber int,
	mapTilesetRepo interfaces.ITilesetRepository,
	playerTilesetRepo interfaces.ITilesetRepository,
	bulletTilesetRepo interfaces.ITilesetRepository,
	spawnerTilesetRepo interfaces.ITilesetRepository,
	explosionTilesetRepo interfaces.ITilesetRepository,
	gameConfig *config.GameConfig,
	gameRepo interfaces.IGameRepositoriesRegistry,
	boundaryCollisionService interfaces.IBoundaryCollisionService,
	wallCollisionService interfaces.IWallCollisionService,
	coordinateService interfaces.ICoordinateService,
	tankBrakingService interfaces.ITankBrakingService,
) (*GameStateUseCasesFacade, error)
```

### Проблемы

1. **Сложность использования**: Легко перепутать порядок параметров
2. **Сложность тестирования**: Придётся передавать множество mock-объектов
3. **Плохая расширяемость**: Добавление нового параметра требует изменения всех вызовов
4. **Нарушение Single Responsibility**: Конструктор берёт на себя слишком много ответственности

### Решение: GameSession с конфигурацией (Планируется)

Создать структуру `GameSession`, которая представляет одну игровую сессию (раунд/катку), и структуру `GameSessionConfig` для конфигурации:

```go
// internal/states/game_session.go
package states

import (
	"github.com/shpaker/gonflict/internal/config"
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
	"github.com/shpaker/gonflict/internal/use_cases"
)

// GameSession представляет одну игровую сессию (раунд/катку)
type GameSession struct {
	// Игрок
	Player struct {
		Tank       *types.TankEntity
		Common     interfaces.ITankCommonUseCases
		Render     interfaces.ITankRenderUseCases
		Lifecycle  interfaces.ITankLifecycleUseCases
		Actions    interfaces.ITankActionsUseCases
	}

	// Враги
	Enemies []struct {
		Tank         *types.TankEntity
		Common       interfaces.ITankCommonUseCases
		Render       interfaces.ITankRenderUseCases
		Lifecycle    interfaces.ITankLifecycleUseCases
		Actions      interfaces.ITankActionsUseCases
		InputAdapter interfaces.IInputAdapter // AI адаптер
	}

	// Игровые системы
	Bullets         *use_cases.BulletUseCases
	Map             *use_cases.MapUseCases
	Collisions      *use_cases.CollisionUseCases
	Animations      *use_cases.TilesUseCases
	AIContext       *types.GameAiContext
}

// GameSessionConfig содержит всю конфигурацию для создания игровой сессии
type GameSessionConfig struct {
	// Репозитории данных
	MapsRepo    interfaces.IMapsDataRepository
	ScriptsRepo interfaces.IScriptsRepository
	GameRepo    interfaces.IGameRepositoriesRegistry
	
	// Конфигурация уровня
	LevelNumber int
	GameConfig  *config.GameConfig
	
	// Репозитории тайлсетов
	Tilesets struct {
		Map       interfaces.ITilesetRepository
		Player    interfaces.ITilesetRepository
		Bullet    interfaces.ITilesetRepository
		Spawner   interfaces.ITilesetRepository
		Explosion interfaces.ITilesetRepository
	}
	
	// Сервисы
	Services struct {
		BoundaryCollision interfaces.IBoundaryCollisionService
		WallCollision     interfaces.IWallCollisionService
		Coordinate        interfaces.ICoordinateService
		TankBraking       interfaces.ITankBrakingService
	}
}

// NewGameSession создает новую игровую сессию из конфигурации
func NewGameSession(cfg GameSessionConfig) (*GameSession, error) {
	// Логика создания из текущего NewGameStateUseCasesFacade
	// ...
}
```

**Преимущества:**
- ✅ **Семантическая ясность**: `GameSession` явно представляет игровую сессию
- ✅ **Логическая группировка**: Все компоненты игрока и врагов сгруппированы
- ✅ **Упрощённый конструктор**: Один параметр `GameSessionConfig` вместо 13
- ✅ **Лучшая инкапсуляция**: Вся логика сессии в одном месте
- ✅ **Улучшенная читаемость**: Понятно, что относится к какой части игры

---

### Решение 2: Builder паттерн

Альтернативный подход с использованием Builder паттерна для пошагового конструирования:

```go
// GameStateUseCasesFacadeBuilder для пошагового создания фасада
type GameStateUseCasesFacadeBuilder struct {
	config *GameStateUseCasesFacadeConfig
}

func NewGameStateUseCasesFacadeBuilder() *GameStateUseCasesFacadeBuilder {
	return &GameStateUseCasesFacadeBuilder{
		config: &GameStateUseCasesFacadeConfig{},
	}
}

func (b *GameStateUseCasesFacadeBuilder) WithRepositories(
	mapsRepo interfaces.IMapsDataRepository,
	scriptsRepo interfaces.IScriptsRepository,
	gameRepo interfaces.IGameRepositoriesRegistry,
) *GameStateUseCasesFacadeBuilder {
	b.config.MapsRepo = mapsRepo
	b.config.ScriptsRepo = scriptsRepo
	b.config.GameRepo = gameRepo
	return b
}

func (b *GameStateUseCasesFacadeBuilder) WithLevel(
	levelNumber int,
	gameConfig *config.GameConfig,
) *GameStateUseCasesFacadeBuilder {
	b.config.LevelNumber = levelNumber
	b.config.GameConfig = gameConfig
	return b
}

func (b *GameStateUseCasesFacadeBuilder) WithTilesets(
	mapRepo, playerRepo, bulletRepo, spawnerRepo, explosionRepo interfaces.ITilesetRepository,
) *GameStateUseCasesFacadeBuilder {
	b.config.Tilesets = TilesetsConfig{
		Map:      mapRepo,
		Player:   playerRepo,
		Bullet:   bulletRepo,
		Spawner:  spawnerRepo,
		Explosion: explosionRepo,
	}
	return b
}

func (b *GameStateUseCasesFacadeBuilder) WithServices(
	boundaryCollision, wallCollision interfaces.ICollisionService,
	coordinate interfaces.ICoordinateService,
	tankBraking interfaces.ITankBrakingService,
) *GameStateUseCasesFacadeBuilder {
	b.config.Services = ServicesConfig{
		BoundaryCollision: boundaryCollision,
		WallCollision:     wallCollision,
		Coordinate:        coordinate,
		TankBraking:       tankBraking,
	}
	return b
}

func (b *GameStateUseCasesFacadeBuilder) Build() (*GameStateUseCasesFacade, error) {
	// Валидация
	if err := b.validate(); err != nil {
		return nil, err
	}
	
	return NewGameStateUseCasesFacade(b.config)
}
```

**Использование:**

```go
facade, err := states.NewGameStateUseCasesFacadeBuilder().
	WithRepositories(mapsRepo, scriptsRepo, gameRepo).
	WithLevel(gameConfig.LevelNumber, gameConfig).
	WithTilesets(
		tilesetRegistry.Blocks(),
		tilesetRegistry.Player(),
		tilesetRegistry.Bullet(),
		tilesetRegistry.Spawner(),
		tilesetRegistry.Explosion(),
	).
	WithServices(
		boundaryCollisionService,
		wallCollisionService,
		coordinateService,
		tankBrakingService,
	).
	Build()
```

**Преимущества:**
- ✅ Гибкость — можно пропускать опциональные параметры
- ✅ Читаемость — понятна последовательность настройки
- ✅ Fluent interface — удобный синтаксис

**Недостатки:**
- ⚠️ Больше кода для поддержки
- ⚠️ Возможность создать невалидную конфигурацию до вызова `Build()`

---

### Текущий статус

**Планируется реализация `GameSession`** — структуры, представляющей игровую сессию. Это решение лучше отражает доменную модель (игровая сессия/раунд) и упрощает использование.

---

## 🟢 Проблема 2: Смешанная ответственность в AI адаптере

### Текущая ситуация

`ai_input_adapter.go` содержит две разные ответственности:

1. **Техническая работа с Lua** (Infrastructure Layer):
   - Создание и управление Lua VM
   - Загрузка и выполнение скриптов
   - Вызов Lua функций

2. **Бизнес-логика конвертации** (Application Layer):
   - Конвертация Go типов (`TankEntity`, `GameAiContext`) в products Lua
   - Преобразование результатов Lua в Go типы
   - Интерпретация бизнес-логики решений AI

### Проблемы

1. **Нарушение Single Responsibility Principle**: Класс делает две разные вещи
2. **Сложность тестирования**: Невозможно тестировать конвертацию без Lua VM
3. **Плохая переиспользуемость**: Логику конвертации нельзя использовать отдельно
4. **Нарушение Dependency Rule**: Application Layer логика смешана с Infrastructure Layer

### Решение: Разделение на Engine и Converter

Разделить на два компонента:

1. **LuaEngine** (Infrastructure Layer) — низкоуровневая работа с Lua
2. **AITypeConverter** (Application Layer) — конвертация типов и бизнес-логика

---

### Структура решения

```
internal/
├── adapters/
│   └── input_adapters/
│       ├── ai_input_adapter.go        # Основной адаптер (использует engine + converter)
│       └── ai/
│           ├── lua_engine.go          # Работа с Lua VM (Infrastructure)
│           └── type_converter.go      # Конвертация типов (Application)
```

---

### 1. LuaEngine (Infrastructure Layer)

Ответственность: Управление Lua VM, выполнение скриптов, вызов функций.

```go
// internal/adapters/input_adapters/ai/lua_engine.go
package ai

import (
	"errors"
	lua "github.com/yuin/gopher-lua"
)

// LuaEngine инкапсулирует работу с Lua VM
type LuaEngine interface {
	// Execute выполняет Lua скрипт из строки
	Execute(script string) error
	
	// CallFunction вызывает Lua функцию с параметрами и возвращает результаты
	CallFunction(functionName string, args ...lua.LValue) ([]lua.LValue, error)
	
	// Close освобождает ресурсы Lua VM
	Close()
}

// luaEngineImpl реализация LuaEngine
type luaEngineImpl struct {
	L *lua.LState
}

// NewLuaEngine создает новый Lua engine
func NewLuaEngine() LuaEngine {
	L := lua.NewState()
	// Инициализируем генератор случайных чисел
	L.DoString("math.randomseed(os.time())")
	
	return &luaEngineImpl{L: L}
}

func (e *luaEngineImpl) Execute(script string) error {
	return e.L.DoString(script)
}

func (e *luaEngineImpl) CallFunction(functionName string, args ...lua.LValue) ([]lua.LValue, error) {
	fn := e.L.GetGlobal(functionName)
	if fn == lua.LNil {
		return nil, errors.New("function not found: " + functionName)
	}
	
	err := e.L.CallByParam(lua.P{
		Fn:      fn,
		NRet:    2,
		Protect: true,
	}, args...)
	
	if err != nil {
		return nil, err
	}
	
	// Собираем результаты
	results := make([]lua.LValue, 2)
	results[0] = e.L.Get(-2)
	results[1] = e.L.Get(-1)
	e.L.Pop(2)
	
	return results, nil
}

func (e *luaEngineImpl) Close() {
	if e.L != nil {
		e.L.Close()
	}
}
```

---

### 2. AITypeConverter (Application Layer)

Ответственность: Конвертация доменных типов Go ↔ Lua, интерпретация бизнес-логики.

```go
// internal/adapters/input_adapters/ai/type_converter.go
package ai

import (
	lua "github.com/yuin/gopher-lua"
	"github.com/shpaker/gonflict/internal/types"
)

// AITypeConverter конвертирует типы между Go и Lua
type AITypeConverter interface {
	// TankToLua конвертирует TankEntity в Lua таблицу
	TankToLua(tank *types.TankEntity) (*lua.LTable, error)
	
	// ContextToLua конвертирует GameAiContext в Lua таблицу
	ContextToLua(context *types.GameAiContext) (*lua.LTable, error)
	
	// LuaToDecision конвертирует результаты Lua функции в EnemyAIDecision
	LuaToDecision(results []lua.LValue) (types.EnemyAIDecision, error)
}

// aiTypeConverterImpl реализация AITypeConverter
type aiTypeConverterImpl struct {
	L *lua.LState
}

// NewAITypeConverter создает новый конвертер типов
func NewAITypeConverter(luaEngine LuaEngine) AITypeConverter {
	// Получаем доступ к LState для создания таблиц
	// В идеале LuaEngine должен предоставлять метод для создания таблиц
	// Для простоты используем type assertion
	if impl, ok := luaEngine.(*luaEngineImpl); ok {
		return &aiTypeConverterImpl{L: impl.L}
	}
	// Если не удалось получить LState, можно создать новый или использовать другой подход
	return nil
}

func (c *aiTypeConverterImpl) TankToLua(tank *types.TankEntity) (*lua.LTable, error) {
	if tank == nil {
		return nil, errors.New("tank is nil")
	}
	
	t := c.L.NewTable()
	t.RawSetString("x", lua.LNumber(tank.Position.X))
	t.RawSetString("y", lua.LNumber(tank.Position.Y))
	t.RawSetString("direction", lua.LNumber(int(tank.Direction)))
	t.RawSetString("speed", lua.LNumber(tank.Speed))
	
	return t, nil
}

func (c *aiTypeConverterImpl) ContextToLua(context *types.GameAiContext) (*lua.LTable, error) {
	if context == nil {
		return nil, errors.New("context is nil")
	}
	
	ctx := c.L.NewTable()
	
	// Добавляем игрока если есть
	if context.Player != nil {
		playerTable, err := c.TankToLua(context.Player)
		if err != nil {
			return nil, err
		}
		ctx.RawSetString("player", playerTable)
	}
	
	// Можно добавить другие поля контекста
	
	return ctx, nil
}

func (c *aiTypeConverterImpl) LuaToDecision(results []lua.LValue) (types.EnemyAIDecision, error) {
	if len(results) < 2 {
		return types.EnemyAIDecision{}, errors.New("insufficient results from Lua")
	}
	
	shouldMove := c.L.ToBool(results[0])
	if !shouldMove {
		return types.EnemyAIDecision{}, nil
	}
	
	directionInt := int(c.L.ToNumber(results[1]))
	
	return types.EnemyAIDecision{
		Direction: types.Direction(directionInt),
	}, nil
}
```

**Альтернативный вариант** — LuaEngine предоставляет метод для создания таблиц:

```go
// В LuaEngine добавляем метод
func (e *luaEngineImpl) NewTable() *lua.LTable {
	return e.L.NewTable()
}

// Тогда конвертер не нуждается в прямом доступе к LState
type aiTypeConverterImpl struct {
	luaEngine LuaEngine
}

func (c *aiTypeConverterImpl) TankToLua(tank *types.TankEntity) (*lua.LTable, error) {
	t := c.luaEngine.NewTable()
	// ...
}
```

---

### 3. Обновлённый AiInputAdapter

Теперь адаптер использует оба компонента:

```go
// internal/adapters/input_adapters/ai_input_adapter.go
package input_adapters

import (
	"github.com/shpaker/gonflict/internal/adapters/input_adapters/ai"
	"github.com/shpaker/gonflict/internal/interfaces"
	"github.com/shpaker/gonflict/internal/types"
)

type AiInputAdapter struct {
	tankUseCases   interfaces.ITankUseCasesRef
	aiContext      *types.GameAiContext
	updateInterval int
	tickCounter    int
	
	// Используем компоненты вместо прямого доступа к Lua
	luaEngine      ai.LuaEngine
	typeConverter  ai.AITypeConverter
}

func NewAiInputAdapter(
	tankUseCases interfaces.ITankUseCasesRef,
	aiContext *types.GameAiContext,
	updateInterval int,
	script string,
) (*AiInputAdapter, error) {
	// Создаем Lua engine
	luaEngine := ai.NewLuaEngine()
	
	// Загружаем скрипт
	if err := luaEngine.Execute(script); err != nil {
		luaEngine.Close()
		return nil, err
	}
	
	// Создаем конвертер типов
	typeConverter := ai.NewAITypeConverter(luaEngine)
	
	return &AiInputAdapter{
		tankUseCases:   tankUseCases,
		aiContext:      aiContext,
		updateInterval: updateInterval,
		tickCounter:    0,
		luaEngine:      luaEngine,
		typeConverter:  typeConverter,
	}, nil
}

func (a *AiInputAdapter) Update(dt float64) {
	if !a.tankUseCases.IsActive() {
		return
	}
	
	if a.tickCounter == 0 {
		decision, err := a.callEnemyAI()
		if err == nil && decision.ShouldMove() {
			a.applyDecision(decision)
		}
	}
	
	a.tickCounter++
	if a.tickCounter >= a.updateInterval {
		a.tickCounter = 0
	}
	
	if err := a.tankUseCases.Update(dt); err != nil {
		log.Printf("ERROR: Failed to update AI tank: %v", err)
	}
}

func (a *AiInputAdapter) callEnemyAI() (types.EnemyAIDecision, error) {
	// Получаем танк
	tank := a.tankUseCases.GetTank()
	if tank == nil {
		return types.EnemyAIDecision{}, errors.New("tank is nil")
	}
	
	// Конвертируем в Lua типы
	tankTable, err := a.typeConverter.TankToLua(tank)
	if err != nil {
		return types.EnemyAIDecision{}, err
	}
	
	contextTable, err := a.typeConverter.ContextToLua(a.aiContext)
	if err != nil {
		return types.EnemyAIDecision{}, err
	}
	
	// Вызываем Lua функцию через engine
	results, err := a.luaEngine.CallFunction("updateEnemyAI", tankTable, contextTable)
	if err != nil {
		return types.EnemyAIDecision{}, err
	}
	
	// Конвертируем результаты обратно в Go типы
	return a.typeConverter.LuaToDecision(results)
}

func (a *AiInputAdapter) applyDecision(decision types.EnemyAIDecision) {
	if a.tankUseCases.IsStopped() {
		a.tankUseCases.Rotate(decision.Direction)
		a.tankUseCases.Move()
	}
}

// Close освобождает ресурсы
func (a *AiInputAdapter) Close() {
	if a.luaEngine != nil {
		a.luaEngine.Close()
	}
}
```

---

### Преимущества решения

1. **Разделение ответственности**:
   - `LuaEngine` — только работа с Lua VM
   - `AITypeConverter` — только конвертация типов и бизнес-логика

2. **Улучшенная тестируемость**:
   - Можно тестировать конвертер с mock Lua engine
   - Можно тестировать адаптер с mock engine и converter

3. **Переиспользуемость**:
   - `LuaEngine` можно использовать для других Lua-скриптов
   - `AITypeConverter` можно переиспользовать в других местах

4. **Соответствие Clean Architecture**:
   - Infrastructure Layer (`LuaEngine`) не зависит от Application Layer
   - Application Layer (`AITypeConverter`) зависит только от Domain Layer

5. **Расширяемость**:
   - Легко добавить новые типы конвертации
   - Можно заменить Lua на другой скриптовый движок, изменив только `LuaEngine`

---

### План внедрения

1. **Этап 1**: Создать интерфейсы `LuaEngine` и `AITypeConverter`
2. **Этап 2**: Реализовать `luaEngineImpl` с минимальным API
3. **Этап 3**: Реализовать `aiTypeConverterImpl` с конвертацией типов
4. **Этап 4**: Рефакторинг `AiInputAdapter` для использования новых компонентов
5. **Этап 5**: Написать тесты для каждого компонента отдельно
6. **Этап 6**: Удалить старый код

---

## Итоговые рекомендации

### Приоритет 1: God Object в конструкторе фасада
- **Критичность**: 🟡 Средняя
- **Рекомендуемое решение**: Структура конфигурации
- **Оценка усилий**: Средние (1-2 дня)
- **Влияние**: Улучшит читаемость и поддерживаемость кода

### Приоритет 2: Смешанная ответственность в AI
- **Критичность**: 🟢 Низкая
- **Рекомендуемое решение**: Разделение на Engine и Converter
- **Оценка усилий**: Средние (2-3 дня)
- **Влияние**: Улучшит тестируемость и переиспользуемость

---

## Ожидаемый результат

После внедрения обоих решений:

- ✅ **Улучшенная архитектура**: Чёткое разделение ответственности
- ✅ **Лучшая тестируемость**: Компоненты можно тестировать изолированно
- ✅ **Упрощённое расширение**: Легче добавлять новые функции
- ✅ **Повышенная оценка**: Clean Architecture оценка может достичь **98/100**

