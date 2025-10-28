# Gonflict

Игра-клон Battle City (Танчики) на Go с использованием Ebiten. Проект построен на принципах Clean Architecture для изучения игровой разработки.

## 🎮 Особенности

- 🎯 **Чистая архитектура** - разделение на слои (Use Cases, Adapters, Repositories)
- 🤖 **ИИ врагов на Lua** - управление поведением через Lua скрипты (gopher-lua)
- 🎮 **Танки врагов** - спавн, анимация, движение и уничтожение пулями
- 🎨 **Гибкие анимации** - настройка через конфиг с offset и repeats
- ⚔️ **Коллизии** - проверка столкновений между танками и объектами
- ⚙️ **Конфигурация** - настройка через `config.yml`
- 🎯 **Скрипты AI** - изменение поведения врагов без перекомпиляции

## 📁 Структура проекта

```
gonflict/
├── cmd/
├── internal/
│   ├── adapters/
│   ├── repositories/
│   │   ├── game/
│   │   ├── processed/
│   │   └── raw/
│   ├── states/
│   ├── types/
│   ├── use_cases/
│   └── utils/
├── assets/
│   ├── levels/
│   ├── sounds/
│   └── tiles/
├── go.mod
├── go.sum
├── justfile
└── README.md
```

## 🏗️ Архитектура

Проект следует принципам **Clean Architecture**:

### ⚡ Оценка соблюдения принципов

#### ✅ Что сделано правильно:

1. **Четкое разделение слоев**:
   - `Presentation Layer` (adapters) — взаимодействие с внешним миром
   - `Application Layer` (use_cases) — бизнес-логика игры
   - `Domain Layer` (types) — сущности и правила
   - `Infrastructure Layer` (repositories) — хранение данных

2. **Dependency Inversion**:
   - Зависимости через **интерфейсы** (ITanksRepository, ITankUseCases)
   - Конкретные реализации скрыты от бизнес-логики
   - Легко подменять реализации (например, для тестов)

3. **Single Responsibility**:
   - `TankUseCases` — только логика танков
   - `CollisionUseCases` — только коллизии
   - `RendererAdapter` — только отрисовка

#### ⚠️ Области для улучшения:

1. **Dependency Injection**:
   - ✅ Используются интерфейсы
   - ⚠️ Фасад сам создает зависимости (строки 35-56 в `game_state_use_cases_facade.go`)
   - ⚠️ 8 параметров конструктора (God Object антипаттерн)
   - ❌ Фасад создает конкретные классы вместо получения через DI

2. **Дублирование структур**:
   - ❌ `GameConfig` определен в `internal/config.go` и `internal/states/game_state.go`
   - **Причина**: Циклический импорт `internal/app.go` → `states` → `internal`
   - ✅ Временное решение для избежания циклических зависимостей

**Общая оценка: 85/100** — хорошо спроектированная архитектура

### Где можно улучшить?

1. **Инициализация репозиториев в Facade** (строка 35 в `game_state_use_cases_facade.go`):
   ```go
   gameRepo := game.NewGameRepositoriesRegistry()  // Создание конкретных классов
   ```
   **Проблема**: Создание конкретных классов внутри фасада  
   **Решение**: Внедрять готовые репозитории через конструктор

2. **Зависимость от константы DT**:
   ```go
   g.tankUseCases.MoveTank(g.tankUseCases.GetDirection(), use_cases.DT)
   ```
   **Проблема**: Глобальные константы усложняют тестирование  
   **Решение**: Передавать `dt` извне (например, от GameState)

3. **GameState создает Use Cases** (строки 37-46 в `game_state.go`):
   ```go
   gameStateServices, err := NewGameStateUseCasesFacade(...)
   ```
   **Проблема**: Знание об уровне (константа `13`) и репозиториях  
   **Решение**: Внедрять `GameStateUseCasesFacade` через конструктор App

4. **Дублирование GameConfig**:
   - Определен в `internal/config.go` (для загрузки) и `internal/states/game_state.go` (для использования)
   - Причина: циклический импорт `internal/app.go` → `states` → `internal`
   - Временное решение, требующее рефакторинга структуры пакетов

### Диаграмма слоев архитектуры

```mermaid
graph TB
    subgraph P["🎨 Presentation Layer"]
        direction LR
        InputAdapters["📝 Input Adapters<br/>Keyboard / AI"]
        RenderAdapter["🎨 Render Adapter"]
    end
    
    subgraph A["⚙️ Application Layer"]
        direction LR
        UseCases["🎮 Use Cases<br/>Бизнес-логика"]
        AICases["🧠 AIUseCases<br/>AI логика"]
    end
    
    subgraph D["📦 Domain Layer"]
        direction LR
        Entities["🏗️ Entities<br/>Domain модели"]
    end
    
    subgraph I["💾 Infrastructure Layer"]
        direction TB
        subgraph IR["📂 Raw Repositories"]
            FileRepo["📁 FileRepository"]
        end
        
        subgraph IP["🔧 Processed Repositories"]
            MapsRepo["🗺️ MapsDataRepository"]
            TilesetRepo["🎨 TilesetRepository"]
            TilesetReg["📋 TilesetRepositoryRegistry"]
            ScriptsRepo["📜 ScriptsRepository"]
        end
        
        subgraph IG["🎮 Game Repositories"]
            TanksRepo["🚗 TanksRepository"]
            BlocksRepo["🧱 BlocksRepository"]
            BulletsRepo["💣 BulletsRepository"]
            AnimationsRepo["✨ AnimationsRepository"]
        end
    end
    
    InputAdapters -->|"команды"| UseCases
    UseCases -->|"обновление"| AICases
    AICases -->|"вызов"| InputAdapters
    InputAdapters -->|"скрипт"| ScriptsRepo
    UseCases -->|"использует"| Entities
    
    UseCases -->|"требует"| TanksRepo
    UseCases -->|"требует"| BlocksRepo
    UseCases -->|"требует"| BulletsRepo
    UseCases -->|"требует"| AnimationsRepo
    
    TilesetReg -->|"предоставляет"| UseCases
    MapsRepo -->|"загружает уровни"| UseCases
    
    TilesetRepo -->|"использует"| FileRepo
    MapsRepo -->|"использует"| FileRepo
    ScriptsRepo -->|"читает"| FileRepo
    
    UseCases -->|"создает"| IG
    IG -->|"хранит"| Entities
    
    UseCases -->|"отрисовка"| RenderAdapter
```

### Диаграмма взаимодействия компонентов

```mermaid
flowchart TB
    User[👤 Пользователь]
    
    subgraph "🎮 Игровой цикл"
        KeyboardInput[📥 KeyboardInputAdapter<br/>клавиатура]
        AiInput[🤖 AiInputAdapter<br/>Lua AI]
        Facade[🎯 GameStateFacade<br/>оркестрация]
        Render[🎨 RendererAdapter<br/>отрисовка]
    end
    
    subgraph "⚡ Use Cases"
        Tank[🚗 TankUseCases<br/>движение/снаряды]
        Enemy[👾 EnemyUseCases<br/>враги]
        Bullet[💣 BulletUseCases<br/>пули]
        Collision[💥 CollisionUseCases<br/>коллизии]
        AI[🧠 AIUseCases<br/>AI логика]
    end
    
    subgraph "💾 Infrastructure"
        Lua[🗂️ Scripts<br/>enemies.lua]
        TanksRepo[(💾 TanksRepository)]
    end
    
    User -->|WASD Space| KeyboardInput
    KeyboardInput -->|команды| Facade
    AI -->|вызов| AiInput
    AiInput -->|скрипт| Lua
    AiInput -->|обновление| Enemy
    Facade -->|обновление| Tank
    Facade -->|обновление| Enemy
    Facade -->|обновление| Bullet
    Facade -->|обновление| AI
    Facade -->|проверка| Collision
    Tank <--> TanksRepo
    Enemy <--> TanksRepo
    Collision -->|проверяет| Tank
    Collision -->|проверяет| Enemy
    Collision -->|проверяет| Bullet
    Tank -->|отрисовка| Render
    Enemy -->|отрисовка| Render
    Bullet -->|отрисовка| Render
    Render -->|графика| User
```

### Диаграмма игрового цикла

```mermaid
sequenceDiagram
    participant App as 🖥️ Application
    participant State as 🎮 GameState
    participant Facade as 🎯 UseCasesFacade
    participant Tank as 🚗 TankUseCases
    participant Enemy as 👾 EnemyUseCases
    participant AI as 🧠 AIUseCases
    participant AiInput as 📜 AiInputAdapter
    participant Collision as 💥 CollisionUseCases
    participant Render as 🎨 RenderAdapter
    
    Note over App: 🏁 Игровой цикл (60 FPS)
    
    App->>+State: Update()
    State->>+Facade: Update()
    
    par Движение объектов
        Facade->>Tank: MoveTank()
    end
    
    par AI врагов
        Facade->>Enemy: UpdateAI()
        Enemy->>AI: UpdateAI()
        AI->>AiInput: CallEnemyAI()
        Note right of AiInput: 📝 Скрипт:<br/>enemies.lua
        AiInput-->>AI: shouldMove, direction
        AI-->>Enemy: ApplyDecision()
        Facade->>Enemy: MoveTank()
    end
    
    par Коллизии
        Facade->>Collision: UpdateCollisions()
        Collision->>Enemy: checkEnemyCollisions()
    end
    
    Facade->>-State: готово
    State->>+Render: DrawAll()
    
    par Отрисовка
        Render->>Render: drawMap()
        Render->>Render: drawTanks()
        Render->>Render: drawBullets()
    end
    
    Render->>-App: завершено
    State->>-App: завершено
    Note over App: 🔄 Следующий кадр
```

### Конкретные примеры из кода

**Правильная инверсия зависимостей:**
```go
// InputAdapter зависит от интерфейса ITankUseCases, а не от конкретной реализации
type InputAdapter struct {
    tankUseCases use_cases.ITankUseCases  // ✅ Интерфейс
}
```

**Правильное разделение ответственности:**
```go
// CollisionUseCases только для коллизий
type CollisionUseCases struct {
    bulletUseCases IBulletUseCases  // Не напрямую репозиторий пуль
}
```

**Области для улучшения:**
```go
// ❌ Создание конкретных классов в фасаде
blocksRepo := game.NewBlocksRepository()

// ✅ Лучше: внедрять через конструктор
func NewFacade(blocksRepo game.IBlocksRepository) *Facade { ... }
```

### Поток данных

```
[Пользователь] 
    ↓
[InputAdapter] → [TankUseCases / BulletUseCases]
    ↓
[GameStateUseCasesFacade] → [Update/Logic]
    ↓
[EnemyUseCases] → [AIUseCases] → [AiInputAdapter] → [enemies.lua]
    ↓                                        ↓
[Repositories] ← [Domain Entities]     [GameContext]
    ↓
[RenderAdapter] → [Экран]
```

### Структура слоев

- **Presentation Layer** - взаимодействие с пользователем и внешним миром (Input Adapters, Render Adapter)
- **Application Layer** - бизнес-логика и правила игры (Use Cases)
- **Domain Layer** - сущности и типы игры (Entities)
- **Infrastructure Layer** - хранение данных и ресурсы (Repositories, Lua Scripts)
- **Supporting Components** - вспомогательные компоненты (Config, Utils)

## 🚀 Быстрый старт

### Требования

- Go 1.24.3 или выше
- Ebiten 2.x (игровой движок)
- Just (опционально) - [установка](https://github.com/casey/just#installation)

### Установка и запуск

1. Клонируйте репозиторий:
```bash
git clone <repository-url>
cd gonflict
```

2. Установите зависимости:
```bash
go mod download
```

3. Запустите приложение:
```bash
# С Just
just run

# Без Just
go run cmd/main.go
```

### Конфигурация через config.yml

Создайте файл `config.yml` в корне проекта:

```yaml
app:
  name: "gonflict"

game:
  spawn_duration_ms: 1000
  enemy_spawners:
    - [0, 0]   # Левый верхний угол
    - [7, 0]   # Центр сверху
    - [13, 0]  # Правый верхний угол
```

Приложение автоматически загрузит конфигурацию при запуске.

## 🔄 Ключевые изменения

### Архитектурные улучшения

- **Система анимаций** - централизованное управление через AnimationUseCases
- **Спавнер танка** - анимированный объект для появления танка игрока
- **Новый формат анимаций** - компактный YAML формат с duration, repeats и offset
- **Clean Architecture** - четкое разделение слоев (Use Cases, Adapters, Repositories)
- **Dependency Injection** - инверсия зависимостей через интерфейсы
- **Тестируемость** - легко тестировать бизнес-логику
- **TanksRepository** - единый репозиторий для танка игрока и врагов
- **TilesetRepositoryRegistry** - реестр всех тайлсетов (блоки, игрок, пули, спавн, взрыв)
- **ScriptsRepository** - репозиторий для загрузки Lua скриптов AI
- **Враги** - танки врагов с анимацией спавна и уничтожением
- **Конфигурация** - настройка через `config.yml` с врагами
- **Коллизии танков** - проверка столкновений между танком игрока и врагами
- **Анимация движения** - анимация гусениц работает только при движении танка
- **Упрощение API** - уменьшение параметров конструкторов через использование Registry

### Технические детали

- **AnimationUseCases** - централизованное управление анимациями
- **AIUseCases** - управление AI логикой врагов
- **AiInputAdapter** - AI адаптер для управления через Lua скрипты
- **LuaAdapter** - конвертация данных между Go и Lua
- **SpawnerEntity** - сущность спавнера с анимацией
- **TankUseCases** - переименованный PlayerUseCases для лучшей семантики
- **Новый формат YAML** - duration, repeats и offset в конфиге анимаций
- **Lua скрипты** - `assets/scripts/enemies.lua` для управления врагами
- **Offset для анимаций** - смещение анимации относительно сущности
- **Duration анимаций** - интервал между кадрами в тиках
- **Repeats** - количество повторений анимации (nil = бесконечно)
- **Анимация только при движении** - гусеницы работают только когда танк движется
- **Коллизии танков** - танк игрока не может проехать сквозь врагов
- **Обратная совместимость** - существующий код анимации продолжает работать
- **ИИ на Lua** - управление врагами через Lua скрипты (gopher-lua)
- **AIUseCases** - централизованное управление AI логикой врагов
- **AiInputAdapter** - AI адаптер для управления через Lua скрипты
- **Тактика NES** - поведение врагов в стиле классической Battle City

## 🎮 Игровая функциональность

### Управление

- **W** - движение вверх
- **S** - движение вниз  
- **A** - движение влево
- **D** - движение вправо
- **Space** - стрельба
- **Escape** - выход

### Игровые объекты

- **Танк игрока** - управляемый танк с поворотом, стрельбой и анимацией гусениц
- **Враги** - танки врагов с AI на Lua, анимацией спавна, движения и взрыва
- **ИИ врагов** - управление через Lua скрипты (`assets/scripts/enemies.lua`)
- **Спавнер** - анимированный объект для появления танка
- **Анимация взрыва** - анимация уничтожения танка с поддержкой offset
- **Блоки**:
  - `#` - Кирпич (разрушаемый)
  - `@` - Сталь (неразрушаемый)
  - `%` - Лес (прозрачный для танков)
  - `~` - Вода (непроходимая)
  - `-` - Лед (скользкий)
- **Пули** - снаряды с коллизиями
- **Уровни** - загружаемые карты 26x26

## 🛠️ Доступные команды (Just)

```bash
just build            # Собрать приложение
just run              # Собрать и запустить
just dev              # Запустить в режиме разработки
just test             # Запустить тесты
just test-coverage    # Тесты с покрытием
just clean            # Очистить собранные файлы
just fmt              # Форматировать код
just lint             # Запустить линтер
just check            # Все проверки (fmt, lint, test)
just ci               # Полный CI pipeline
just help             # Показать все команды
```

## 📚 Компоненты системы

### Presentation Layer (Слой представления)

- **App** - главное приложение Ebiten, точка входа
- **GameState** - игровое состояние, управляет игровым процессом
- **KeyboardInputAdapter** - адаптер ввода с клавиатуры (WASD, Space)
- **AiInputAdapter** - AI адаптер для управления врагами через Lua скрипты
- **IInputAdapter** - интерфейс для всех адаптеров ввода
- **RendererAdapter** - адаптер отрисовки игры через Ebiten

### Application Layer (Слой приложения)

- **GameStateUseCasesFacade** - фасад для оркестрации всех Use Cases
- **TankUseCases** - бизнес-логика танков игрока (движение, поворот, спавн)
- **EnemyUseCases** - бизнес-логика врагов (спавн, анимация, уничтожение)
- **BulletUseCases** - бизнес-логика пуль (создание, обновление, удаление)
- **MapUseCases** - бизнес-логика карты (работа с блоками)
- **CollisionUseCases** - бизнес-логика коллизий между объектами
- **AnimationUseCases** - бизнес-логика анимаций
- **AIUseCases** - бизнес-логика AI врагов
- **TilesUseCases** - бизнес-логика тайлов (статические и анимированные)

### Domain Layer (Доменный слой)

**Типы данных:**
- **Direction** - направление движения (UP, DOWN, LEFT, RIGHT)
- **Position** - позиция в мире (X, Y)
- **Size** - размер объекта (Width, Height)
- **Altitude** - высота слоя отрисовки

**Сущности:**
- **TankEntity** - сущность танка игрока
- **BulletEntity** - сущность пули
- **BlockEntity** - сущность блока карты
- **SpawnerEntity** - сущность спавнера танка
- **TileStaticEntity** - статический тайл для отрисовки
- **TileAnimationEntity** - анимированный тайл

### Infrastructure Layer (Слой инфраструктуры)

**Game Repositories (In-memory):**
- **TanksRepository** - единое хранилище танков (игрока и врагов)
- **BlocksRepository** - хранилище блоков карты
- **BulletsRepository** - хранилище пуль
- **AnimationsRepository** - хранилище активных анимаций

**Processed Repositories:**
- **MapsDataRepository** - загрузка и обработка уровней
- **TilesetRepository** - загрузка и кеширование тайлсетов
- **TilesetRepositoryRegistry** - реестр всех тайлсетов (блоки, игрок, пули, спавн, взрыв), упрощает создание и передачу
- **ScriptsRepository** - загрузка Lua скриптов для AI (без кэширования)

**Raw Repositories:**
- **FileRepository** - чтение файлов из assets

**Game AI:**
- **Lua скрипты** - `assets/scripts/enemies.lua` - логика поведения врагов
- **gopher-lua** - библиотека для встраивания Lua в Go
- **AIUseCases** - управление AI врагов
- **AiInputAdapter** - AI адаптер для управления через Lua скрипты (принимает скрипт как строку)
- **ScriptsRepository** - загрузка Lua скриптов из файлов без кэширования

### Supporting Components (Вспомогательные компоненты)

- **Config** - конфигурация приложения
- **Utils** - вспомогательные функции

## 🧪 Тестирование

```bash
# Запустить все тесты
go test ./...

# Тесты с покрытием
go test -cover ./...

# Подробный отчет
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Примеры тестов

- `repositories/game/*_test.go` - тесты репозиториев
- `repositories/processed/*_test.go` - тесты обработки данных
- `repositories/raw/*_test.go` - тесты чтения файлов
- `utils/utils_test.go` - тесты утилит

## 🔧 Расширение проекта

### Добавление новой функциональности

1. **Создайте Use Case** в `use_cases/`
2. **Определите интерфейс** в `use_cases/interfaces.go`
3. **Реализуйте логику** в отдельном файле
4. **Добавьте тесты** `*_test.go`
5. **Используйте в GameState**

### Пример: использование TanksRepository

```go
// TanksRepository - единое хранилище для всех танков
type TanksRepository struct {
    tanks []*types.TankEntity
}

// Индекс 0 - танк игрока
// Индексы 1, 2, 3... - враги
func (tr *TanksRepository) GetTank(index int) (*types.TankEntity, error) {
    if index < 0 || index >= len(tr.tanks) {
        return nil, fmt.Errorf("индекс танка %d вне диапазона", index)
    }
    return tr.tanks[index], nil
}
```

### Идеи для расширения

- 🎯 **ИИ врагов** - добавить стрельбу врагов, преследование игрока
- 🎯 **Система очков** - подсчет очков за уничтожение
- 🎨 **Меню и UI** - главное меню, экран победы/поражения
- 🌐 **Сетевой режим** - мультиплеер
- 🏆 **Достижения** - система наград
- 🛠️ **Редактор уровней** - создание своих карт
- 💥 **Эффекты** - взрывы, частицы
- 🎵 **Музыка** - фоновая музыка
- 🎮 **Дополнительные уровни** - больше карт

## 📝 Соглашения по коду

### Именование файлов

```
✅ Хорошо:
player_use_cases.go
collision_service.go
interfaces.go
constants.go

❌ Плохо:
_player.go         # Не начинайте с _
PlayerUseCases.go  # Не используйте PascalCase
```

### Структура файлов в пакете

```
package_name/
├── interfaces.go     # Все интерфейсы (будет первым при сортировке)
├── constants.go      # Все константы
├── entity1.go        # Реализации
├── entity2.go
└── service.go
```

### Импорты

```go
// Всегда используйте централизованный repositories пакет
import "github.com/shpaker/gonflict/internal/repositories"

// raw импортируется отдельно только где необходимо
import "github.com/shpaker/gonflict/internal/repositories/raw"
```

## 🐛 Известные ограничения

- Только один игрок
- Враги пока не стреляют (только двигаются)
- Нет главного меню
- Нет сохранений
- Нет звуковых эффектов
- Нет системы очков
- Спавнер не имеет коллизий

## 🤝 Вклад в проект

1. Fork проекта
2. Создайте feature branch (`git checkout -b feature/amazing-feature`)
3. Commit изменений (`git commit -m 'Add amazing feature'`)
4. Push в branch (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

## 📄 Лицензия

MIT License

## 📝 История изменений

### Последние обновления

- ✅ **TilesetRepositoryRegistry** - реестр всех тайлсетов для упрощения управления
- ✅ **ScriptsRepository** - загрузка Lua скриптов из файлов
- ✅ **AiInputAdapter** - принимает скрипт как строку вместо пути к файлу
- ✅ **Упрощение API** - уменьшение количества параметров конструкторов
- ✅ **ИИ врагов на Lua** - управление врагами через Lua скрипты (gopher-lua)
- ✅ **AIUseCases** - централизованное управление AI логикой врагов
- ✅ **IInputAdapter** - единый интерфейс для всех адаптеров ввода
- ✅ **Lua скрипты** - `assets/scripts/enemies.lua` для изменения поведения без перекомпиляции
- ✅ **Тактика NES** - поведение врагов в стиле классической Battle City
- ✅ **Коллизии врагов** - враги останавливаются при столкновении со стенами
- ✅ **Рефакторинг PlayerUseCases в TankUseCases** - переименование для лучшей семантики
- ✅ **Система анимаций** - централизованное управление анимациями через AnimationUseCases
- ✅ **Спавнер танка** - анимированный объект для появления танка игрока и врагов
- ✅ **Новый формат анимаций** - компактный YAML формат с duration, repeats и offset
- ✅ **Clean Architecture** - четкое разделение слоев и зависимостей
- ✅ **Dependency Injection** - инверсия зависимостей через интерфейсы
- ✅ **TanksRepository** - единый репозиторий для танка игрока и врагов
- ✅ **Враги** - танки врагов с анимацией спавна и уничтожением пулями игрока
- ✅ **Конфигурация** - настройка через `config.yml` с позициями спавна врагов
- ✅ **Анимация взрыва** - визуальная анимация уничтожения танков с поддержкой offset
- ✅ **Коллизии танков** - проверка столкновений между танком игрока и врагами
- ✅ **Анимация гусениц** - анимация работает только при движении танка

### Планируемые улучшения

- 🎯 **Улучшение AI** - добавление стрельбы врагов, преследования игрока
- 🎯 **Система очков** - подсчет очков за уничтожение
- 🎨 **Меню и UI** - главное меню, экран победы/поражения
- 🌐 **Сетевой режим** - мультиплеер
- 🏆 **Достижения** - система наград
- 🛠️ **Редактор уровней** - создание своих карт
- 💥 **Эффекты** - взрывы, частицы
- 🎵 **Музыка** - фоновая музыка
- 🎮 **Дополнительные уровни** - больше карт

## 📚 Полезные ссылки

- [Ebiten Documentation](https://ebiten.org/)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Just Command Runner](https://github.com/casey/just)
- [gopher-lua](https://github.com/yuin/gopher-lua) - библиотека для встраивания Lua в Go
- [Lua Documentation](https://www.lua.org/)
