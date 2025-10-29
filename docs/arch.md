# Нарушения Clean Architecture

## Критичность: Высокая

### 1. Создание конкретных классов в фасаде
**Файл:** `internal/states/game_state_use_cases_facade.go:35`
```go
   gameRepo := game.NewGameRepositoriesRegistry()
```
**Проблема:** Фасад создает конкретную реализацию вместо получения через DI  
**Решение:** Внедрять `game.IGameRepositoriesRegistry` через конструктор

---

### 2. Создание AI адаптера в фасаде
**Файл:** `internal/states/game_state_use_cases_facade.go:48-52`
```go
ai, err := adapters.NewEnemyAILua("assets/scripts/enemies.lua")
if err != nil {
    return nil, err
}
```
**Проблема:** Фасад зависит от конкретной реализации адаптера  
**Решение:** Внедрять `adapters.IEnemyAI` через конструктор

---

### 3. Создание всех tileset репозиториев в App
**Файл:** `internal/app.go:43-81`
```go
tilesetRepo, err := processed.NewTilesetDataRepository(fileRepo, "tiles/blocks")
playerTilesetRepo, err := processed.NewTilesetDataRepository(fileRepo, "tiles/player")
bulletTilesetRepo, err := processed.NewTilesetDataRepository(fileRepo, "tiles/bullet")
spawnerTilesetRepo, err := processed.NewTilesetDataRepository(fileRepo, "tiles/spawner")
explosionTilesetRepo, err := processed.NewTilesetDataRepository(fileRepo, "tiles/explosion")
```
**Проблема:** App создает 5 репозиториев напрямую  
**Решение:** Создать фабрику или внедрять через конструктор

---

## Критичность: Средняя

### 4. Создание адаптеров в GameState
**Файл:** `internal/states/game_state.go:37-51`
```go
rendererAdapter := createRendererAdapter(...)
inputAdapter := createInputAdapter(gameStateServices)
```
**Проблема:** GameState создает адаптеры напрямую  
**Решение:** Внедрять через конструктор

---

### 5. God Object в конструкторе фасада
**Файл:** `internal/states/game_state_use_cases_facade.go:22-31`
```go
func NewGameStateUseCasesFacade(
    mapsRepo processed.IMapsDataRepository,
    levelNumber int,
    mapTilesetRepo processed.ITilesetRepository,
    playerTilesetRepo processed.ITilesetRepository,
    bulletTilesetRepo processed.ITilesetRepository,
    spawnerTilesetRepo processed.ITilesetRepository,
    explosionTilesetRepo processed.ITilesetRepository,
    gameConfig *GameConfig,
)
```
**Проблема:** 8 параметров конструктора (God Object антипаттерн)  
**Решение:** Использовать структуру конфигурации или builder

---

### 6. Зависимость от константы DT
**Файл:** `internal/states/game_state_use_cases_facade.go:126-136`
```go
g.playerUseCases.MoveTank(g.playerUseCases.GetDirection(), use_cases.DT)
for _, enemy := range g.enemyUseCasesList {
    enemy.UpdateAI()
    enemy.MoveTank(use_cases.DT)
}
```
**Проблема:** Глобальные константы усложняют тестирование  
**Решение:** Передавать `dt` извне (например, от GameState)

---

### 7. Хардкод номера уровня
**Файл:** `internal/states/game_state.go:41`
```go
gameStateServices, err := NewGameStateUseCasesFacade(
    mapsRepo,
    13, // Номер уровня - хардкод
    ...
)
```
**Проблема:** Знание о конкретном уровне  
**Решение:** Внедрять через конфигурацию

---

### 8. Создание TilesUseCases внутри GameState
**Файл:** `internal/states/game_state.go:133-137`
```go
mapTilesUseCases := use_cases.NewTilesUseCases(mapTilesetRepo)
playerTilesUseCases := use_cases.NewTilesUseCases(playerTilesetRepo)
bulletTilesUseCases := use_cases.NewTilesUseCases(bulletTilesetRepo)
spawnerTilesUseCases := use_cases.NewTilesUseCases(spawnerTilesetRepo)
explosionTilesUseCases := use_cases.NewTilesUseCases(explosionTilesetRepo)
```
**Проблема:** GameState создает 5 Use Cases напрямую  
**Решение:** Внедрять через конструктор

---

## Критичность: Низкая

### 9. Дублирование GameConfig
**Файлы:** `internal/config.go:16-21` и `internal/states/game_state.go:14-18`
```go
// В internal/config.go
type GameConfig struct {
    EnemySpawners         [][]int `yaml:"enemy_spawners"`
    PlayerSpawners        [][]int `yaml:"players_spawners"`
    AIUpdateIntervalTicks int     `yaml:"ai_update_interval_ticks"`
}

// В internal/states/game_state.go - дублирование
type GameConfig struct {
    EnemySpawners         [][]int `yaml:"enemy_spawners"`
    PlayerSpawners        [][]int `yaml:"players_spawners"`
    AIUpdateIntervalTicks int     `yaml:"ai_update_interval_ticks"`
}
```
**Причина:** Циклический импорт `internal/app.go` → `states` → `internal`  
**Решение:** Перенести в отдельный пакет `internal/config` или использовать интерфейсы

---

### 10. Неполное копирование GameConfig
**Файл:** `internal/app.go:85-88`
```go
gameConfig := &states.GameConfig{
    EnemySpawners:  cfg.EnemySpawners,
    PlayerSpawners: cfg.PlayerSpawners,
    // Отсутствует: AIUpdateIntervalTicks
}
```
**Проблема:** Потеря данных при копировании  
**Решение:** Передавать полную структуру

---

---

## Структура AI компонентов

### Текущая структура

```
internal/
├── use_cases/
│   ├── ai_use_cases.go          // Управление AI логикой
│   └── enemy_ai_interface.go    // Интерфейс IEnemyAI
└── adapters/
    ├── enemy_ai_adapter.go      // Lua реализация (thin wrapper)
    └── lua_adapter.go           // Gopher-lua wrapper + конвертация
```

### Проблемы

1. **Смешанная ответственность в `lua_adapter.go`**:  
   - Содержит и низкоуровневую работу с Lua, и бизнес-логику (конвертация типов)
   - Функции `directionToInt`, `intToDirection`, `convertTankToLua` — это не адаптер

2. **Thin wrapper в `enemy_ai_adapter.go`**:  
   - Просто вызывает `lua_adapter` без добавленной ценности
   - Слишком простая абстракция

3. **Нелогичное размещение**:  
   - AI логика размазана между `use_cases` и `adapters`
   - `lua_adapter.go` делает конвертацию данных (должна быть в use_cases)

### Предлагаемая структура

```
internal/
├── use_cases/
│   ├── ai_use_cases.go              // Оркестрация AI
│   └── ai/                           // 🔥 НОВЫЙ пакет
│       ├── interface.go              // IEnemyAI
│       ├── convertor.go              // Конвертация Go ↔ Lua
│       └── decision.go               // EnemyAIDecision
│
└── adapters/
    ├── ai_lua_adapter.go             // Lua реализация (переименован)
    └── lua/                          // 🔥 НОВЫЙ пакет
        └── engine.go                 // Низкоуровневая работа с Lua VM
```

**Изменения:**
- `ai/convertor.go` — конвертация типов Go/Lua (бизнес-логика)
- `lua/engine.go` — только работа с Lua VM (техническая часть)
- `ai_lua_adapter.go` — соединяет `engine` + `convertor`
- Четкое разделение ответственности

---

## Рефакторинг: Вынос логики торможения в сервис

### Изменения

**Было:** Вся логика торможения находилась в `TankUseCases`

**Стало:** Логика торможения вынесена в отдельный сервис `TankBrakingService`

### Структура

```
internal/
├── use_cases/
│   └── tank_use_cases.go        // Основная логика танков (без торможения)
└── services/
    └── tank_braking_services.go  // Логика торможения (вынесена из Use Cases)
```

### Преимущества

1. **Разделение ответственности:** 
   - `TankUseCases` — базовая логика танков (движение, поворот, стрельба)
   - `TankBrakingService` — специализированная логика торможения

2. **Упрощение тестирования:**
   - Логику торможения можно тестировать изолированно

3. **Улучшение читаемости:**
   - `TankUseCases` стал более компактным и сфокусированным

### Методы сервиса

- `HandleBrakingState(dt)` — обработка движения в состоянии Braking
- `getBrakingMovementContext()` — определение контекста движения
- `checkAndHandleHalfStepBack()` — обработка возврата на 0.5 назад
- `moveTowardsTarget()` — движение к целевому кратному 4
- `moveForwardToTarget()` / `moveBackwardToTarget()` — движение по направлениям
- `completeBraking()` / `finishBraking()` — завершение процесса торможения

### Зависимости

Сервис принимает:
- `*TankEntity` — танк для управления
- `func(bool)` — функция остановки (callback для вызова `Stop` из Use Cases)

Это позволяет сервису вызывать метод остановки без прямой зависимости от `TankUseCases`.

---

## Итого

**Высокая критичность:** 3 нарушения  
**Средняя критичность:** 5 нарушений  
**Низкая критичность:** 2 нарушения  
**Структурные проблемы:** 1 (AI компоненты)

**Всего:** 11 проблем

**Оценка архитектуры: 85/100** — хорошо спроектированная, но требует улучшений в DI и разделении ответственности.
