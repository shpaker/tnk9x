# Code Review: Слоевое разделение и ответственности

## Определения слоев по Clean Architecture (Robert C. Martin)

### 1. Entities Layer (Domain Layer)
**Определение:** Самый внутренний слой. Содержит бизнес-правила высокого уровня, которые не зависят от деталей реализации.

**Что ДОЛЖЕН делать:**
- ✅ Содержать сущности домена (entities)
- ✅ Содержать Value Objects (неизменяемые объекты значений)
- ✅ Содержать доменные интерфейсы (порты без реализаций)
- ✅ Содержать бизнес-правила, которые являются частью сущности
- ✅ Быть независимым от всех внешних слоев (zero dependencies)

**Чего НЕ ДОЛЖЕН делать:**
- ❌ Зависеть от фреймворков, БД, UI
- ❌ Содержать логику сохранения/загрузки данных
- ❌ Содержать логику отрисовки
- ❌ Содержать логику ввода/вывода
- ❌ Содержать зависимости от конкретных технологий

**Примеры правильных сущностей:**
```go
// ✅ ХОРОШО: Чистая сущность без зависимостей
type TankEntity struct {
    Position  Position
    Speed     float64
    Direction Direction
    State     TankState
}
```

---

### 2. Use Cases Layer (Application Layer)
**Определение:** Содержит бизнес-логику приложения, специфичную для конкретных use cases.

**Что ДОЛЖЕН делать:**
- ✅ Оркестрировать выполнение бизнес-логики для конкретных сценариев
- ✅ Координировать работу между сущностями и репозиториями
- ✅ Валидировать входные данные use case
- ✅ Проверять бизнес-правила, выходящие за рамки одной сущности
- ✅ Зависеть только от Domain Layer и интерфейсов репозиториев

**Чего НЕ ДОЛЖЕН делать:**
- ❌ Создавать конкретные реализации сервисов/репозиториев
- ❌ Содержать логику доступа к данным (это репозитории)
- ❌ Содержать логику работы с UI (это адаптеры)
- ❌ Зависеть от фреймворков или внешних библиотек напрямую
- ❌ Содержать детали инфраструктуры (конфигурация, логирование)

**Пример правильного Use Case:**
```go
// ✅ ХОРОШО: Use Case зависит только от интерфейсов
type TankUseCases struct {
    tanksRepo      ITanksRepository      // Интерфейс
    bulletUseCases IBulletUseCases       // Интерфейс
    brakingService ITankBrakingService   // Интерфейс (внедряется)
}

func NewTankUseCases(
    tanksRepo ITanksRepository,
    bulletUseCases IBulletUseCases,
    brakingService ITankBrakingService, // Внедряется, не создается
) *TankUseCases
```

**Пример неправильного Use Case:**
```go
// ❌ ПЛОХО: Use Case создает сервисы напрямую
func NewTankUseCases(...) *TankUseCases {
    uc := &TankUseCases{}
    uc.brakingService = services.NewTankBrakingService(...) // Нарушение!
    return uc
}
```

---

### 3. Interface Adapters Layer (Presentation Layer)
**Определение:** Слой адаптеров, который преобразует данные между Use Cases и внешними системами.

**Что ДОЛЖЕН делать:**
- ✅ Адаптировать данные из Use Cases для представления (UI, API, консоль)
- ✅ Преобразовывать пользовательский ввод в вызовы Use Cases
- ✅ Работать с конкретными фреймворками (Ebiten, HTTP, etc.)
- ✅ Зависеть только от Use Cases Layer (интерфейсы)

**Чего НЕ ДОЛЖЕН делать:**
- ❌ Содержать бизнес-логику (это Use Cases)
- ❌ Прямо обращаться к репозиториям (через Use Cases)
- ❌ Прямо работать с сущностями (через Use Cases)
- ❌ Содержать логику валидации бизнес-правил (это Use Cases)

**Пример правильного адаптера:**
```go
// ✅ ХОРОШО: Адаптер зависит только от интерфейсов Use Cases
type KeyboardInputAdapter struct {
    tankUseCases ITankUseCasesRef // Интерфейс, не конкретная реализация
}

func (a *KeyboardInputAdapter) Update() {
    if ebiten.IsKeyPressed(ebiten.KeyW) {
        a.tankUseCases.Rotate(types.DirectionUp) // Вызов через интерфейс
        a.tankUseCases.Move()
    }
}
```

---

### 4. Frameworks & Drivers Layer (Infrastructure Layer)
**Определение:** Внешний слой, содержащий инструменты и детали реализации.

**Что ДОЛЖЕН делать:**
- ✅ Реализовывать интерфейсы репозиториев
- ✅ Работать с конкретными технологиями (БД, файловая система, сети)
- ✅ Загружать конфигурацию
- ✅ Содержать детали инфраструктуры

**Чего НЕ ДОЛЖЕН делать:**
- ❌ Содержать бизнес-логику
- ❌ Зависеть от Use Cases напрямую (только через интерфейсы)

---

### 5. Services Layer (Application Services / Domain Services)
**Определение:** Специализированные сервисы, которые не относятся к конкретной сущности.

**Что ДОЛЖЕН делать:**
- ✅ Содержать логику, выходящую за рамки одной сущности
- ✅ Предоставлять переиспользуемые бизнес-операции
- ✅ Зависеть только от Domain Layer
- ✅ Иметь интерфейсы для внедрения зависимостей

**Чего НЕ ДОЛЖЕН делать:**
- ❌ Создаваться внутри Use Cases (внедряться через DI)
- ❌ Содержать логику доступа к данным (это репозитории)
- ❌ Зависеть от фреймворков напрямую

**Пример правильного сервиса:**
```go
// ✅ ХОРОШО: Сервис имеет интерфейс и зависит только от домена
type ITankBrakingService interface {
    HandleBrakingState(tank *types.TankEntity, dt float64) error
}

type TankBrakingService struct {
    tank *types.TankEntity // Зависит только от домена
}

func (s *TankBrakingService) HandleBrakingState(dt float64) error {
    // Бизнес-логика торможения
}
```

---

### 6. Repositories Layer (Infrastructure / Persistence)
**Определение:** Абстракция для доступа к данным.

**Что ДОЛЖЕН делать:**
- ✅ Определять интерфейсы доступа к данным (порты)
- ✅ Предоставлять реализацию этих интерфейсов
- ✅ Скрывать детали хранения данных
- ✅ Работать с сущностями домена

**Чего НЕ ДОЛЖЕН делать:**
- ❌ Содержать бизнес-логику
- ❌ Зависеть от Use Cases (только наоборот через интерфейсы)

---

## Анализ текущего разделения слоев в проекте

### ✅ Что сделано правильно

#### 1. Domain Layer (`internal/types/`)
**Статус:** ✅ **ХОРОШО**

- Сущности (`TankEntity`, `BulletEntity`, `BlockEntity`) не зависят от других слоев
- Содержит только данные и методы доступа
- Интерфейсы (`IMapObject`, `IImageIDGetter`) определены на уровне домена

**Пример:**
```go
// ✅ ПРАВИЛЬНО: Чистая сущность
type TankEntity struct {
    Position  Position
    Speed     float64
    Direction Direction
    State     TankState
    Altitude  Altitude
}

// Методы только для доступа, без бизнес-логики
func (t *TankEntity) IsActive() bool {
    return t.State != TankStateExploded && t.State != TankStateSpawning
}
```

#### 2. Interface Adapters (`internal/adapters/`)
**Статус:** ✅ **ХОРОШО**

- Адаптеры работают только с интерфейсами Use Cases
- Не содержат бизнес-логику
- Правильно преобразуют входные данные в вызовы Use Cases

**Пример:**
```go
// ✅ ПРАВИЛЬНО: Адаптер зависит только от интерфейса
type KeyboardInputAdapter struct {
    tankUseCases use_cases.ITankUseCasesRef // Интерфейс
}

func (a *KeyboardInputAdapter) Update() {
    if ebiten.IsKeyPressed(a.upButton) {
        a.tankUseCases.Rotate(types.DirectionUp)
    }
}
```

#### 3. Repositories (`internal/repositories/`)
**Статус:** ✅ **ХОРОШО**

- Четкое разделение на слои: `raw/` → `processed/` → `game/`
- Интерфейсы определены отдельно от реализаций
- Репозитории работают только с сущностями домена

**Пример:**
```go
// ✅ ПРАВИЛЬНО: Интерфейс репозитория
type ITanksRepository interface {
    AddTank(tank *types.TankEntity)
    GetAllTanks() []*types.TankEntity
    RemoveTank(index int)
}
```

---

### ❌ Проблемы в разделении слоев

#### 1. Use Cases создают сервисы напрямую

**Файл:** `internal/use_cases/collision_use_cases.go:37-50`

```go
// ❌ ПРОБЛЕМА: Use Case создает сервисы напрямую
func NewCollisionUseCasesWithEnemies(...) *CollisionUseCases {
    uc := &CollisionUseCases{}
    
    // Нарушение: Use Case не должен создавать конкретные реализации
    uc.boundaryCollisionService = services.NewBoundaryCollisionService(
        MapWidthHeight,
        TankSpriteSize,
    )
    uc.wallCollisionService = services.NewWallCollisionService(...)
    uc.bulletCollisionService = services.NewBulletCollisionService(...)
    
    return uc
}
```

**Проблема:**
- Нарушение Dependency Inversion Principle
- Use Cases зависят от конкретных реализаций сервисов
- Невозможно подменить сервисы моками для тестирования

**Правильное решение:**
```go
// ✅ РЕШЕНИЕ: Внедрение через интерфейсы
type IBoundaryCollisionService interface {
    CheckTankBoundaryCollisions(tank *types.TankEntity, stopAndRound bool) bool
    CheckEnemyBoundaryCollisions(enemy *types.TankEntity)
    CheckBulletBoundaryCollisions(bullets []types.BulletEntity) []int
}

type CollisionUseCases struct {
    boundaryCollisionService IBoundaryCollisionService // Интерфейс
    wallCollisionService     IWallCollisionService
    bulletCollisionService   IBulletCollisionService
}

func NewCollisionUseCasesWithEnemies(
    boundaryService IBoundaryCollisionService, // Внедряется
    wallService IWallCollisionService,
    bulletService IBulletCollisionService,
    // ...
) *CollisionUseCases
```

#### 2. Use Cases создают координатные сервисы напрямую

**Файл:** `internal/use_cases/tank_use_cases.go:53`

```go
// ❌ ПРОБЛЕМА: Use Case создает сервис напрямую
func NewTankUseCases(...) *TankUseCases {
    uc := &TankUseCases{
        coordinateService: services.NewCoordinateService(), // Нарушение!
        // ...
    }
    return uc
}
```

**Правильное решение:**
```go
// ✅ РЕШЕНИЕ: Внедрение через интерфейс
type ICoordinateService interface {
    RoundToNearestMultipleOf4(value float64) float64
}

func NewTankUseCases(
    coordinateService ICoordinateService, // Внедряется
    // ...
) *TankUseCases
```

#### 3. Фасад создает репозитории напрямую

**Файл:** `internal/states/game_state_use_cases_facade.go:43`

```go
// ❌ ПРОБЛЕМА: Фасад создает репозитории напрямую
func NewGameStateUseCasesFacade(...) (*GameStateUseCasesFacade, error) {
    // Нарушение: Фасад не должен создавать конкретные реализации
    gameRepo := game.NewGameRepositoriesRegistry()
    // ...
}
```

**Проблема:**
- Нарушение Dependency Inversion Principle
- Фасад зависит от конкретной реализации репозиториев
- Невозможно подменить репозитории для тестирования

**Правильное решение:**
```go
// ✅ РЕШЕНИЕ: Внедрение через интерфейс
func NewGameStateUseCasesFacade(
    gameRepo game.IGameRepositoriesRegistry, // Интерфейс, внедряется
    // ...
) (*GameStateUseCasesFacade, error)
```

#### 4. Services Layer не имеет интерфейсов

**Проблема:**
- Сервисы (`TankBrakingService`, `CoordinateService`, `BoundaryCollisionService`) не имеют интерфейсов
- Use Cases не могут зависеть от абстракций
- Нарушение Dependency Inversion Principle

**Пример:**
```go
// ❌ ПРОБЛЕМА: Нет интерфейса
type TankBrakingService struct {
    tank *types.TankEntity
}

// ✅ РЕШЕНИЕ: Создать интерфейс
type ITankBrakingService interface {
    HandleBrakingState(dt float64) error
}

// Реализация
type TankBrakingService struct {
    tank *types.TankEntity
}

// Проверка реализации на этапе компиляции
var _ ITankBrakingService = (*TankBrakingService)(nil)
```

#### 5. Смешение ответственности: States и Use Cases

**Проблема:**
- `GameStateUseCasesFacade` находится в слое `states/`, но выполняет оркестрацию Use Cases
- `GameState` создает адаптеры, что является ответственностью инфраструктуры

**Анализ:**
- `states/` - это часть Presentation Layer (управление состояниями UI/приложения)
- Фасад для Use Cases должен быть в Application Layer (рядом с Use Cases)

**Рекомендация:**
```
internal/
├── adapters/           # Presentation Layer
├── application/        # 🔥 НОВЫЙ: Application Layer
│   ├── use_cases/     # Use Cases
│   └── facade/         # Фасад для оркестрации Use Cases
├── services/           # Application Services
├── types/              # Domain Layer
└── repositories/       # Infrastructure Layer
```

#### 6. Константы в Use Cases

**Файл:** `internal/use_cases/constants.go`

```go
// ❌ ПРОБЛЕМА: Константы в Use Cases
const (
    TileMinSize    = 8
    TankSpriteSize = 16
    DT             = 1.0 / 60.0
)
```

**Проблема:**
- Использование глобальных констант нарушает Dependency Rule
- Усложняет тестирование (невозможно подменить значения)
- `DT` должен передаваться как параметр

**Правильное решение:**
```go
// ✅ РЕШЕНИЕ: Константы в конфигурации или передавать как параметры
func (uc *TankUseCases) Update(dt float64) error { // dt передается, не константа
    // ...
}

// Константы в конфигурации
type GameConfig struct {
    TileMinSize    int
    TankSpriteSize int
    DeltaTime      float64
}
```

#### 7. Stateless-сервисы без необходимости инстансов

**Файлы:** `coordinate_services.go`, `image_services.go`, `animation_services.go`

```go
// ❌ ПРОБЛЕМА: Stateless-сервис создается как структура
type CoordinateService struct{}

func NewCoordinateService() *CoordinateService {
    return &CoordinateService{}
}

func (s *CoordinateService) RoundToNearestMultipleOf4(value float64) float64 {
    // ...
}
```

**Проблема:**
- Сервис не хранит состояние, но создается как структура
- Избыточное усложнение

**Правильное решение:**
```go
// ✅ РЕШЕНИЕ: Package-level функции
package services

func RoundToNearestMultipleOf4(value float64) float64 {
    // ...
}

// Или если нужен интерфейс для тестирования:
type Rounder interface {
    RoundToNearestMultipleOf4(value float64) float64
}

func NewRounder() Rounder {
    return &coordinateRounder{}
}

type coordinateRounder struct{}

func (r *coordinateRounder) RoundToNearestMultipleOf4(value float64) float64 {
    // ...
}
```

#### 8. Логирование напрямую в сервисах

**Файл:** `internal/services/tank_braking_services.go`

```go
// ❌ ПРОБЛЕМА: Сервис использует log напрямую
import "log"

func (s *TankBrakingService) HandleBrakingState(dt float64) error {
    log.Printf("DEBUG: Tank braking position...")
    // ...
}
```

**Проблема:**
- Зависимость от конкретной реализации логирования
- Нарушение Dependency Inversion Principle
- Невозможно отключить логирование в тестах

**Правильное решение:**
```go
// ✅ РЕШЕНИЕ: Внедрение логгера через интерфейс
type ILogger interface {
    Debugf(format string, v ...interface{})
    Errorf(format string, v ...interface{})
}

type TankBrakingService struct {
    tank   *types.TankEntity
    logger ILogger // Внедряется
}

func NewTankBrakingService(tank *types.TankEntity, logger ILogger) *TankBrakingService {
    return &TankBrakingService{tank: tank, logger: logger}
}
```

---

## Правильное направление зависимостей

### Текущая структура (с нарушениями)

```
┌───────────────────────────────────────────┐
│ Presentation Layer (adapters/)            │
│  ✅ Правильно зависит от Use Cases        │
└──────────────┬────────────────────────────┘
               │
┌──────────────▼────────────────────────────┐
│ Application/State Layer (states/)         │
│  ❌ Создает репозитории напрямую          │
│  ❌ Создает адаптеры напрямую            │
└──────────────┬────────────────────────────┘
               │
┌──────────────▼────────────────────────────┐
│ Use Cases Layer (use_cases/)              │
│  ❌ Создает сервисы напрямую              │
│  ❌ Использует глобальные константы       │
└──────────────┬────────────────────────────┘
               │
┌──────────────▼────────────────────────────┐
│ Services Layer (services/)                │
│  ❌ Нет интерфейсов                       │
│  ❌ Использует log напрямую              │
└──────────────┬────────────────────────────┘
               │
┌──────────────▼────────────────────────────┐
│ Domain Layer (types/)                     │
│  ✅ Правильно: нет зависимостей          │
└──────────────┬────────────────────────────┘
               │
┌──────────────▼────────────────────────────┐
│ Infrastructure Layer (repositories/)     │
│  ✅ Правильно реализует интерфейсы       │
└───────────────────────────────────────────┘
```

### Правильная структура (Clean Architecture)

```
┌───────────────────────────────────────────┐
│ Presentation Layer (adapters/)           │
│  Зависит от: Use Cases (интерфейсы)      │
└──────────────┬────────────────────────────┘
               │
┌──────────────▼────────────────────────────┐
│ Application Layer (application/)          │
│  ├── use_cases/                          │
│  │   Зависит от: Services (интерфейсы),  │
│  │                Domain, Repositories   │
│  └── facade/                             │
│      Зависит от: Use Cases (интерфейсы)  │
└──────────────┬────────────────────────────┘
               │
┌──────────────▼────────────────────────────┐
│ Services Layer (services/)                │
│  Зависит от: Domain                      │
│  Имеет: Интерфейсы для DI                 │
└──────────────┬────────────────────────────┘
               │
┌──────────────▼────────────────────────────┐
│ Domain Layer (types/)                     │
│  Зависит от: НИЧЕГО                       │
└──────────────┬────────────────────────────┘
               │
┌──────────────▼────────────────────────────┐
│ Infrastructure Layer (repositories/)     │
│  Зависит от: Domain                      │
│  Реализует: Интерфейсы репозиториев      │
└───────────────────────────────────────────┘
```

---

## Рекомендации по рефакторингу

### Приоритет 1: Критические нарушения

1. **Создать интерфейсы для всех сервисов**
   ```go
   // internal/services/interfaces.go
   type ITankBrakingService interface {
       HandleBrakingState(dt float64) error
   }
   
   type ICoordinateService interface {
       RoundToNearestMultipleOf4(value float64) float64
   }
   
   type IBoundaryCollisionService interface { /* ... */ }
   type IWallCollisionService interface { /* ... */ }
   type IBulletCollisionService interface { /* ... */ }
   ```

2. **Убрать создание сервисов из Use Cases**
   - Внедрять сервисы через конструкторы Use Cases
   - Использовать интерфейсы, а не конкретные реализации

3. **Убрать создание репозиториев из фасадов**
   - Внедрять репозитории через конструктор фасада
   - Использовать интерфейсы

### Приоритет 2: Средние нарушения

4. **Внедрить интерфейс логгера**
   - Создать `ILogger` интерфейс
   - Внедрять через конструкторы сервисов

5. **Убрать глобальные константы из Use Cases**
   - Передавать `dt` как параметр
   - Вынести константы в конфигурацию

6. **Преобразовать stateless-сервисы**
   - В package-level функции или
   - С интерфейсами для тестирования

### Приоритет 3: Низкие нарушения

7. **Реорганизация структуры**
   - Создать `internal/application/` для фасада
   - Четче разделить States и Application Layer

---

## Выводы

### Сильные стороны
- ✅ Правильное разделение на слои
- ✅ Domain Layer не зависит от других слоев
- ✅ Адаптеры правильно зависят от Use Cases
- ✅ Репозитории правильно реализуют интерфейсы

### Критические проблемы
- ❌ Use Cases создают сервисы напрямую (нарушение DIP)
- ❌ Фасады создают репозитории напрямую (нарушение DIP)
- ❌ Отсутствие интерфейсов для сервисов
- ❌ Использование глобальных констант

### После исправлений
После устранения нарушений проект будет полностью соответствовать принципам Clean Architecture:
- ✅ Легкое тестирование (все зависимости через интерфейсы)
- ✅ Гибкость (легко менять реализации)
- ✅ Независимость от деталей реализации
- ✅ Масштабируемость

---

**Дата обзора:** 2024  
**Версия:** 1.0

