# 🎮 Gonflict

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)
[![Ebiten](https://img.shields.io/badge/Ebiten-v2.9.1-orange.svg?style=flat-square)](https://ebiten.org/)

Классическая аркадная игра-клон Battle City на Go с использованием Ebiten. Разработана с применением принципов Clean Architecture и Domain-Driven Design.

## 📖 Описание

Gonflict — это ремейк культовой игры Battle City (Tank 1990) для NES. Игрок управляет танком, сражается с врагами, разрушает блоки и защищает базу на сетке 26x26.

### Основные особенности

- ✅ Полнофункциональный игровой процесс с физикой движения
- ✅ Продвинутая система торможения с доезжанием до сетки
- ✅ AI врагов на Lua-скриптах
- ✅ Система коллизий и разрушаемых блоков
- ✅ Анимации спавна, движения и взрывов
- ✅ 35 уровней с различной конфигурацией
- ✅ Звуковое сопровождение
- ✅ Clean Architecture с четким разделением слоев
- ✅ Разделение Use Cases на специализированные компоненты
- ✅ Интерфейсы для всех Use Cases (Dependency Inversion Principle)

## 🎯 Текущий статус

### ✅ Реализовано

- [x] **Основной игровой цикл**
  - [x] Система состояний игры (Spawning, Moving, Stopped, Braking, Exploding)
  - [x] Обновление и отрисовка игровых объектов
  - [x] Обработка пользовательского ввода

- [x] **Танк игрока**
  - [x] Движение с пошаговой системой (до кратного 4)
  - [x] Продвинутое торможение с доезжанием до сетки
  - [x] Поворот во время движения (с доезжанием до кратного 4)
  - [x] Стрельба пулями
  - [x] Анимации движения (гусеницы)
  - [x] Анимации спавна и взрыва

- [x] **Система торможения**
  - [x] Автоматическое доезжание до координаты кратной 4
  - [x] Обработка смены направления во время торможения
  - [x] Возврат на 0.5 назад при переезде целевой точки
  - [x] Вынесено в отдельный сервис `TankBrakingService`

- [x] **Враги и AI**
  - [x] Система вражеских танков
  - [x] AI на Lua-скриптах (gopher-lua)
  - [x] Анимации спавна и движения врагов
  - [x] Система спавнеров

- [x] **Пули и коллизии**
  - [x] Создание и обновление пуль
  - [x] Коллизии пуль с блоками и танками
  - [x] Разрушение блоков
  - [x] Анимации взрыва

- [x] **Карта и блоки**
  - [x] Загрузка уровней (35 уровней)
  - [x] Типы блоков:
    - Кирпич (разрушаемый)
    - Сталь (неразрушаемый)
    - Лес (прозрачный)
    - Вода (непроходимая)
    - Лед (скользкий)
  - [x] Коллизии танка с блоками

- [x] **Графика и анимации**
  - [x] Система тайлсетов (YAML конфигурация)
  - [x] Статические тайлы
  - [x] Анимационные тайлы с offset
  - [x] Поворот изображений по направлению

- [x] **База и защита**
  - [x] База (HQ) на карте (фиксированная позиция)
  - [x] База разрушается от пуль врагов
  - [x] Визуализация базы (статический спрайт)
  - [x] Анимация взрыва базы

- [x] **Архитектура**
  - [x] Clean Architecture (Presentation, Application, Domain, Infrastructure)
  - [x] Разделение на Use Cases (10 специализированных компонентов)
  - [x] Репозитории для данных (Game, Processed, Raw)
  - [x] Сервисы (9 специализированных сервисов)
  - [x] Адаптеры (Input, Renderer, AI)
  - [x] Разделение AI на слои (LuaEngine, AITypeConverter, AIUseCases)
  - [x] Интерфейсы для всех компонентов (Dependency Inversion)

- [x] **Звуки**
  - [x] Загрузка звуковых файлов
  - [x] Воспроизведение звуков (стрельба, взрывы, фон)

### 🚧 В процессе

- [ ] Система очков и статистики
- [ ] Меню и экраны (главное меню, game over, победа)
- [ ] Бонусы и power-ups
- [ ] Система жизней игрока
- [ ] Game Over при уничтожении базы (требуется интеграция)

### 📋 TODO / Планируется

- [ ] **Игровая механика**
  - [ ] Разные типы врагов с различным поведением
  - [ ] Бонусы при уничтожении врагов
  - [ ] Система уровней сложности
  - [ ] Временные лимиты на уровни
  - [ ] Защита базы от врагов

- [ ] **UI/UX**
  - [ ] Главное меню
  - [ ] Экран выбора уровня
  - [ ] HUD с очками, жизнями, уровнем
  - [ ] Экран Game Over
  - [ ] Экран победы
  - [ ] Пауза (ESC)
  - [ ] Настройки (громкость, управление)

- [ ] **Геймплей**
  - [ ] Разные типы танков для игрока
  - [ ] Улучшения танка (броня, скорость, мощность снарядов)
  - [ ] Боссы на уровнях
  - [ ] Система прогресса и сохранений

- [ ] **Технические улучшения**
  - [ ] Рефакторинг `GameStateUseCasesFacade` → `GameSession` (упрощение конструктора)
  - [ ] Оптимизация производительности
  - [ ] Расширение тестового покрытия
  - [ ] Документация API
  - [ ] CI/CD pipeline

- [ ] **Портирование**
  - [ ] Поддержка других платформ (Windows, Linux)
  - [ ] Мобильная версия (Android/iOS через gomobile)

## 🚀 Установка и запуск

### Требования

- Go 1.24 или выше
- [Just](https://github.com/casey/just) (опционально, для команд сборки)

### Установка

```bash
# Клонируйте репозиторий
git clone https://github.com/shpaker/gonflict.git
cd gonflict

# Установите зависимости
go mod download
```

### Запуск

```bash
# С Just (рекомендуется)
just run

# Или напрямую
go run cmd/main.go

# Сборка релизной версии
just build
./dist/gonflict
```

### Доступные команды (Just)

```bash
just build            # Собрать приложение
just run              # Собрать и запустить
just dev              # Запустить в режиме разработки
just test             # Запустить тесты
just test-coverage    # Тесты с покрытием
just fmt              # Форматировать код
just fmt-check        # Проверить форматирование
just lint             # Запустить линтер
just lint-fix         # Исправить проблемы линтера
just clean            # Очистить собранные файлы
just check            # Все проверки (fmt-check, lint, test)
```

## 🎮 Управление

| Клавиша | Действие |
|---------|----------|
| `W` | Движение вверх |
| `S` | Движение вниз |
| `A` | Движение влево |
| `D` | Движение вправо |
| `Space` | Стрельба |
| `Escape` | Выход (планируется: пауза) |

### Особенности управления

- **Пошаговое движение**: Танк автоматически доезжает до координаты кратной 4 при отпускании клавиши
- **Смена направления**: При нажатии новой клавиши направления во время движения, танк сначала доезжает до кратного 4, затем поворачивается
- **Торможение**: Плавное торможение с выравниванием по сетке

## 🏗️ Архитектура

Проект следует принципам **Clean Architecture** с четким разделением на слои:

```
internal/
├── adapters/          # Presentation Layer
│   ├── input_adapters/    # Адаптеры ввода
│   └── renderer_adapter.go # Адаптер отрисовки
├── states/            # Presentation Layer
│   └── game_state.go      # Управление состоянием игры
├── use_cases/         # Application Layer
│   ├── tank_common_use_cases.go
│   ├── tank_actions_use_cases.go
│   ├── tank_lifecycle_use_cases.go
│   ├── tank_render_use_cases.go
│   ├── bullet_use_cases.go
│   ├── collision_use_cases.go
│   ├── hq_use_cases.go
│   ├── ai_use_cases.go
│   ├── map_use_cases.go
│   └── tile_use_cases.go
├── services/          # Application Layer
│   ├── tank_braking_services.go
│   ├── coordinate_services.go
│   ├── boundary_collision_service.go
│   ├── wall_collision_service.go
│   ├── bullet_collision_service.go
│   ├── tile_services.go
│   ├── animation_services.go
│   ├── image_services.go
│   └── ai_type_converter.go
├── types/             # Domain Layer
│   ├── tank_entity.go
│   ├── bullet_entity.go
│   ├── block_entity.go
│   ├── hq_entity.go
│   ├── tile_animation_entity.go
│   ├── session_entity.go
│   └── battle_entity.go
└── repositories/      # Infrastructure Layer
    ├── game/              # In-memory репозитории
    ├── processed/         # Обработанные данные
    └── raw/               # Загрузка файлов
```

### Архитектурная диаграмма слоёв

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

### Описание слоёв

#### 1. Presentation Layer (Слой представления)
**Ответственность:** Взаимодействие с внешним миром (пользователь, графический движок)
- **cmd/main.go** — точка входа приложения
- **internal/app.go** — главное приложение, реализует интерфейс Ebiten Game
- **internal/adapters/RendererAdapter** — адаптер для рендеринга игровых объектов в Ebiten
- **internal/adapters/input_adapters/** — адаптеры ввода (клавиатура, AI)

#### 2. Application/State Layer (Слой приложения/состояний)
**Ответственность:** Управление состояниями приложения и оркестрация Use Cases
- **internal/states/GameState** — состояние игры, координирует обновление и отрисовку
- **internal/states/GameStateUseCasesFacade** — фасад для координации Use Cases

#### 3. Use Cases / Business Logic Layer (Слой бизнес-логики)
**Ответственность:** Реализация бизнес-правил и игровой логики
- **TankCommonUseCases** — логика движения танков
- **TankActionsUseCases** — логика действий (поворот, движение, стрельба)
- **TankRenderUseCases** — логика графики и анимаций танков
- **TankLifecycleUseCases** — логика жизненного цикла (спавн, взрыв)
- **BulletUseCases** — логика движения и взаимодействия пуль
- **MapUseCases** — логика работы с картой и блоками
- **CollisionUseCases** — логика определения и обработки коллизий
- **TilesUseCases** — логика работы с тайлами и анимациями
- **AIUseCases** — логика AI врагов
- **HQUseCases/HQRenderUseCases** — логика работы с базой (объединены в один файл)

#### 4. Services Layer (Слой сервисов)
**Ответственность:** Специализированная бизнес-логика, выделенная из Use Cases
- **TankBrakingService** — логика торможения танков
- **CoordinateService** — работа с координатами (округление)
- **BoundaryCollisionService** — коллизии с границами карты
- **WallCollisionService** — коллизии со стенами
- **BulletCollisionService** — коллизии пуль с объектами
- **TileService** — работа с тайлами и анимациями
- **AnimationService** — обновление анимаций
- **ImageService** — работа с изображениями (поворот)
- **AITypeConverter** — конвертация типов Go ↔ Lua (Application Service)
- **LuaEngine** — работа с Lua VM (Infrastructure)

#### 5. Domain / Entities Layer (Слой домена/сущностей)
**Ответственность:** Представление доменных сущностей и их поведения
- **TankEntity** — танк (игрок и враги)
- **BulletEntity** — пуля
- **BlockEntity** — блок карты
- **TileAnimationEntity** — анимация тайлов
- **HQEntity** — база
- **SessionEntity** — игровая сессия
- **GameAiContext** — контекст для AI
- Чистый домен без зависимостей от других слоёв

#### 6. Repository Layer (Слой репозиториев)
**Ответственность:** Управление данными и доступ к хранилищам
- **Game Repositories** — хранение игровых объектов в памяти (танки, пули, блоки, анимации)
- **Processed Repositories** — загрузка и обработка карт, тайлсетов, Lua-скриптов
- **Raw Repositories** — низкоуровневое чтение файлов из assets

#### 7. Infrastructure Layer (Слой инфраструктуры)
**Ответственность:** Конфигурация и ресурсы приложения
- **config.yml** — конфигурационный файл
- **assets/** — ресурсы игры (уровни, тайлсеты, звуки, скрипты)

### Принцип зависимостей

Все зависимости направлены **внутрь**:
- Внешние слои зависят от внутренних
- Внутренние слои не знают о внешних
- Domain Layer не зависит ни от чего (чистое ядро)

### Основные компоненты

**Use Cases:**
- **TankCommonUseCases** — общая логика движения танков
- **TankActionsUseCases** — логика действий танков (поворот, движение, стрельба)
- **TankRenderUseCases** — логика графики танков
- **TankLifecycleUseCases** — логика жизненного цикла танков (спавн, взрыв)
- **BulletUseCases** — создание и обновление пуль
- **CollisionUseCases** — обработка коллизий
- **HQUseCases/HQRenderUseCases** — логика базы (объединены)
- **AIUseCases** — логика AI врагов
- **TilesUseCases** — управление тайлами и анимациями

**Services:**
- **TankBrakingService** — логика торможения
- **CollisionServices** — специализированная логика коллизий
- **TileService** — работа с тайлами
- **AITypeConverter** — конвертация типов Go ↔ Lua

**Repositories:**
- **GameRepositoriesRegistry** — реестр игровых репозиториев
- **ITilesetRepositoryRegistry** — реестр тайлсетов
- **IMapsDataRepository** — карты уровней
- **IFileRepository** — чтение файлов

## 🛠️ Технологии

- **[Go 1.24+](https://golang.org/)** — основной язык
- **[Ebiten v2](https://ebiten.org/)** — игровой движок
- **[gopher-lua](https://github.com/yuin/gopher-lua)** — интеграция Lua для AI
- **[YAML](https://yaml.org/)** — конфигурация тайлсетов и уровней

## 📦 Структура проекта

```
gonflict/
├── assets/           # Игровые ресурсы
│   ├── levels/       # 35 уровней
│   ├── scripts/     # Lua скрипты для AI
│   ├── sounds/      # Звуковые файлы
│   └── tiles/       # Тайлсеты и конфигурация
├── cmd/             # Точка входа
├── dist/            # Собранные файлы
├── docs/            # Документация
├── internal/        # Внутренний код приложения
│   ├── adapters/    # Адаптеры
│   ├── repositories/# Репозитории
│   ├── services/    # Сервисы
│   ├── states/      # Состояния игры
│   ├── types/       # Доменные типы
│   └── use_cases/   # Use Cases
├── config.yml       # Конфигурация игры
└── justfile         # Команды сборки
```

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

## 🤝 Вклад

Вклады приветствуются! Если вы хотите внести свой вклад:

1. Fork проекта
2. Создайте ветку для новой функции (`git checkout -b feature/amazing-feature`)
3. Commit изменения (`git commit -m 'Add amazing feature'`)
4. Push в ветку (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

Перед отправкой PR убедитесь:
- Код следует стилю проекта
- Добавлены тесты для новой функциональности
- Документация обновлена

## 📄 Лицензия

Этот проект лицензирован под MIT License — см. файл [LICENSE](LICENSE) для деталей.

## 🙏 Благодарности

- Вдохновлено классической игрой Battle City / Tank 1990 для NES
- Спрайты и ресурсы основаны на оригинальных NES материалах
- Спасибо сообществу Go и разработчикам Ebiten

## 📞 Контакты

- Issues: [GitHub Issues](https://github.com/shpaker/gonflict/issues)
- Discussions: [GitHub Discussions](https://github.com/shpaker/gonflict/discussions)

---

**Приятной игры! 🎮**

