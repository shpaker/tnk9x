# Архитектурная диаграмма слоёв

## Структура слоёв приложения gonflict

```mermaid
graph TB
    subgraph "Presentation Layer"
        MAIN[cmd/main.go<br/>Точка входа]
        APP[internal/app.go<br/>App - Ebiten Game Interface]
        RENDERER[internal/adapters/<br/>RendererAdapter<br/>Рендеринг в Ebiten]
        INPUT[internal/adapters/<br/>input_adapters/<br/>KeyboardInputAdapter<br/>AIInputAdapter]
    end

    subgraph "Application/State Layer"
        GAME_STATE[internal/states/<br/>GameState<br/>Состояние игры]
        STATE_FACADE[internal/states/<br/>GameStateUseCasesFacade<br/>Фасад для Use Cases]
    end

    subgraph "Use Cases / Business Logic Layer"
        TANK_UC[internal/use_cases/<br/>TankUseCases<br/>Логика танков]
        BULLET_UC[internal/use_cases/<br/>BulletUseCases<br/>Логика пуль]
        MAP_UC[internal/use_cases/<br/>MapUseCases<br/>Логика карты]
        COLLISION_UC[internal/use_cases/<br/>CollisionUseCases<br/>Логика коллизий]
        TILES_UC[internal/use_cases/<br/>TilesUseCases<br/>Логика тайлов/анимаций]
    end

    subgraph "Services Layer"
        BRAKING_SVC[internal/services/<br/>TankBrakingService<br/>Логика торможения]
    end

    subgraph "Domain / Entities Layer"
        ENTITIES[internal/types/<br/>TankEntity<br/>BulletEntity<br/>BlockEntity<br/>TileAnimationEntity<br/>Сущности домена]
    end

    subgraph "Repository Layer"
        subgraph "Game Repositories"
            GAME_REPO[internal/repositories/game/<br/>GameRepositoriesRegistry<br/>IGameRepositoriesRegistry]
            BLOCKS_REPO[internal/repositories/game/<br/>BlocksRepository<br/>IBlocksRepository]
            BULLETS_REPO[internal/repositories/game/<br/>BulletsRepository<br/>IBulletsRepository]
            TANKS_REPO[internal/repositories/game/<br/>TanksRepository<br/>ITanksRepository]
            ANIM_REPO[internal/repositories/game/<br/>AnimationsRepository<br/>IAnimationsRepository]
        end
        
        subgraph "Processed Repositories"
            PROCESSED_REPO[internal/repositories/processed/<br/>MapsDataRepository<br/>TilesetRepository<br/>ScriptsRepository]
        end
        
        subgraph "Raw Repositories"
            RAW_REPO[internal/repositories/raw/<br/>FileRepository<br/>IFileRepository<br/>Чтение файлов]
        end
    end

    subgraph "Infrastructure Layer"
        CONFIG[config.yml<br/>Конфигурация приложения]
        ASSETS[assets/<br/>levels, tiles, sounds, scripts]
    end

    %% Presentation Layer connections
    MAIN --> APP
    APP --> GAME_STATE
    GAME_STATE --> RENDERER
    GAME_STATE --> INPUT
    GAME_STATE --> STATE_FACADE
    
    %% State Facade connections
    STATE_FACADE --> TANK_UC
    STATE_FACADE --> BULLET_UC
    STATE_FACADE --> MAP_UC
    STATE_FACADE --> COLLISION_UC
    STATE_FACADE --> TILES_UC
    STATE_FACADE --> INPUT
    
    %% Use Cases connections
    TANK_UC --> BRAKING_SVC
    TANK_UC --> BULLET_UC
    TANK_UC --> TILES_UC
    BULLET_UC --> TILES_UC
    COLLISION_UC --> ENTITIES
    
    %% Repository connections
    TANK_UC --> TANKS_REPO
    BULLET_UC --> BULLETS_REPO
    MAP_UC --> BLOCKS_REPO
    TILES_UC --> ANIM_REPO
    GAME_REPO --> BLOCKS_REPO
    GAME_REPO --> BULLETS_REPO
    GAME_REPO --> TANKS_REPO
    GAME_REPO --> ANIM_REPO
    
    %% Processed repositories connections
    STATE_FACADE --> PROCESSED_REPO
    APP --> PROCESSED_REPO
    PROCESSED_REPO --> RAW_REPO
    
    %% Raw repository connections
    RAW_REPO --> ASSETS
    
    %% Entities connections
    TANK_UC --> ENTITIES
    BULLET_UC --> ENTITIES
    MAP_UC --> ENTITIES
    BLOCKS_REPO --> ENTITIES
    BULLETS_REPO --> ENTITIES
    TANKS_REPO --> ENTITIES
    ANIM_REPO --> ENTITIES
    
    %% Infrastructure connections
    APP --> CONFIG
    
    %% Styling
    classDef presentation fill:#e1f5ff,stroke:#01579b,stroke-width:2px
    classDef application fill:#f3e5f5,stroke:#4a148c,stroke-width:2px
    classDef usecases fill:#fff3e0,stroke:#e65100,stroke-width:2px
    classDef services fill:#fff9c4,stroke:#f57f17,stroke-width:2px
    classDef domain fill:#e8f5e9,stroke:#1b5e20,stroke-width:2px
    classDef repository fill:#fce4ec,stroke:#880e4f,stroke-width:2px
    classDef infrastructure fill:#eceff1,stroke:#263238,stroke-width:2px
    
    class MAIN,APP,RENDERER,INPUT presentation
    class GAME_STATE,STATE_FACADE application
    class TANK_UC,BULLET_UC,MAP_UC,COLLISION_UC,TILES_UC usecases
    class BRAKING_SVC services
    class ENTITIES domain
    class GAME_REPO,BLOCKS_REPO,BULLETS_REPO,TANKS_REPO,ANIM_REPO,PROCESSED_REPO,RAW_REPO repository
    class CONFIG,ASSETS infrastructure
```

## Описание слоёв

### 1. Presentation Layer (Слой представления)
**Ответственность:** Взаимодействие с внешним миром (пользователь, графический движок)

- **cmd/main.go** - Точка входа приложения, инициализация и запуск
- **internal/app.go** - Главное приложение, реализует интерфейс Ebiten Game
- **internal/adapters/RendererAdapter** - Адаптер для рендеринга игровых объектов в Ebiten
- **internal/adapters/input_adapters/** - Адаптеры ввода (клавиатура, AI)

**Зависимости:** Зависит только от Application Layer

---

### 2. Application/State Layer (Слой приложения/состояний)
**Ответственность:** Управление состояниями приложения и оркестрация Use Cases

- **internal/states/GameState** - Состояние игры, координирует обновление и отрисовку
- **internal/states/GameStateUseCasesFacade** - Фасад для координации Use Cases

**Зависимости:** Зависит от Use Cases Layer и Presentation Layer (адаптеры)

---

### 3. Use Cases / Business Logic Layer (Слой бизнес-логики)
**Ответственность:** Реализация бизнес-правил и игровой логики

- **internal/use_cases/TankUseCases** - Логика движения, поворотов, стрельбы танков
- **internal/use_cases/BulletUseCases** - Логика движения и взаимодействия пуль
- **internal/use_cases/MapUseCases** - Логика работы с картой и блоками
- **internal/use_cases/CollisionUseCases** - Логика определения и обработки коллизий
- **internal/use_cases/TilesUseCases** - Логика работы с тайлами и анимациями

**Зависимости:** Зависит от Domain Layer и Repository Layer

---

### 4. Services Layer (Слой сервисов)
**Ответственность:** Специализированная бизнес-логика, выделенная из Use Cases

- **internal/services/TankBrakingService** - Специализированная логика торможения танков

**Зависимости:** Зависит от Domain Layer

---

### 5. Domain / Entities Layer (Слой домена/сущностей)
**Ответственность:** Представление доменных сущностей и их поведения

- **internal/types/TankEntity** - Сущность танка
- **internal/types/BulletEntity** - Сущность пули
- **internal/types/BlockEntity** - Сущность блока карты
- **internal/types/TileAnimationEntity** - Сущность анимации тайла
- **internal/types/types.go** - Общие типы (Position, Direction, Size, Altitude и т.д.)

**Зависимости:** Не зависит от других слоёв (чистый домен)

---

### 6. Repository Layer (Слой репозиториев)
**Ответственность:** Управление данными и доступ к хранилищам

#### 6.1. Game Repositories (Игровые репозитории)
- **internal/repositories/game/GameRepositoriesRegistry** - Реестр игровых репозиториев
- **internal/repositories/game/BlocksRepository** - Хранение блоков карты в памяти
- **internal/repositories/game/BulletsRepository** - Хранение пуль в памяти
- **internal/repositories/game/TanksRepository** - Хранение танков в памяти
- **internal/repositories/game/AnimationsRepository** - Хранение анимаций в памяти

#### 6.2. Processed Repositories (Обработанные данные)
- **internal/repositories/processed/MapsDataRepository** - Загрузка и обработка карт уровней
- **internal/repositories/processed/TilesetRepository** - Загрузка и обработка тайлсетов
- **internal/repositories/processed/ScriptsRepository** - Загрузка Lua скриптов

#### 6.3. Raw Repositories (Сырые данные)
- **internal/repositories/raw/FileRepository** - Низкоуровневое чтение файлов из assets

**Зависимости:** Processed зависит от Raw, Game зависит от Processed

---

### 7. Infrastructure Layer (Слой инфраструктуры)
**Ответственность:** Конфигурация и ресурсы приложения

- **config.yml** - Конфигурационный файл приложения
- **assets/** - Ресурсы игры (уровни, тайлсеты, звуки, скрипты)

**Зависимости:** Используется всеми слоями через репозитории

---

## Принципы архитектуры

### Dependency Rule (Правило зависимостей)
Все зависимости направлены **внутрь**:
- Внешние слои зависят от внутренних
- Внутренние слои не знают о внешних
- Domain Layer не зависит ни от чего

### Слои взаимодействия

```
Presentation Layer
      ↓
Application/State Layer
      ↓
Use Cases / Business Logic Layer
      ↓
Services Layer
      ↓
Domain / Entities Layer ← (ядро, никаких зависимостей)
      ↓
Repository Layer
      ↓
Infrastructure Layer
```

### Паттерны проектирования

1. **Facade** - `GameStateUseCasesFacade` скрывает сложность координации Use Cases
2. **Repository** - Абстракция доступа к данным через интерфейсы
3. **Adapter** - `RendererAdapter`, `InputAdapter` адаптируют внешние библиотеки
4. **Factory** - `GameRepositoriesRegistry` создаёт репозитории
5. **Service** - `TankBrakingService` выделяет специализированную логику

---

## Поток данных

### Инициализация приложения:
```
main.go → App.New() → 
  FileRepository → 
  TilesetRepositoryRegistry → 
  ScriptsRepository → 
  MapsDataRepository → 
  GameState → 
  GameStateUseCasesFacade → 
  Use Cases + Repositories
```

### Игровой цикл:
```
Update():
  GameState → 
    InputAdapter.Update() → 
    GameStateUseCasesFacade.Update() → 
    Use Cases → 
    Repositories → 
    Entities
    
Draw():
  GameState → 
    RendererAdapter.DrawAll() → 
    Use Cases (получение данных) → 
    Repositories → 
    Entities → 
    Ebiten
```

