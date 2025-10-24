# Gonflict

Игра-клон Battle City (Танчики) на Go с использованием Ebiten. Проект построен на принципах Clean Architecture для изучения игровой разработки.

## 🎮 Особенности

- ✅ **Clean Architecture** - четкое разделение слоев (Use Cases, Adapters, Repositories)
- ✅ **Dependency Injection** - инверсия зависимостей через интерфейсы
- ✅ **Тестируемость** - легко тестировать бизнес-логику
- ✅ **Расширяемость** - простое добавление новых функций
- ✅ **Конфигурируемость** - настройка через переменные окружения

## 📁 Структура проекта

```
gonflict/
├── cmd/
│   └── main.go                    # Точка входа приложения
├── internal/
│   ├── app.go                     # Главное приложение и Ebiten интеграция
│   ├── config.go                  # Конфигурация приложения
│   │
│   ├── use_cases/                 # 🎯 Бизнес-логика (Use Cases Layer)
│   │   ├── interfaces.go          # Интерфейсы use cases
│   │   ├── constants.go           # Константы игровой логики
│   │   ├── player_use_cases.go    # Логика игрока
│   │   ├── bullet_use_cases.go    # Логика пуль
│   │   ├── map_use_cases.go       # Логика карты
│   │   └── collision_use_cases.go # Логика коллизий
│   │
│   ├── adapters/                  # 🔌 Адаптеры (Infrastructure Layer)
│   │   ├── interfaces.go          # Интерфейсы адаптеров
│   │   ├── constants.go           # Константы адаптеров
│   │   ├── input_adapter.go       # Адаптер ввода (клавиатура)
│   │   └── renderer_adapter.go    # Адаптер рендеринга (Ebiten)
│   │
│   ├── repositories/              # 💾 Репозитории (Data Layer)
│   │   ├── repositories.go        # Централизованные реимпорты
│   │   ├── game/                  # Игровые репозитории (in-memory)
│   │   │   ├── interfaces.go
│   │   │   ├── blocks_repository.go
│   │   │   └── bullets_repository.go
│   │   ├── processed/             # Обработанные данные
│   │   │   ├── interfaces.go
│   │   │   ├── constants.go
│   │   │   ├── maps_data_repository.go
│   │   │   └── sprites_repository.go
│   │   └── raw/                   # Сырые данные (файлы)
│   │       ├── interfaces.go
│   │       └── file_repository.go
│   │
│   ├── states/                    # 🎭 Состояния игры
│   │   ├── interfaces.go          # Интерфейс State
│   │   └── game_state.go          # Игровое состояние
│   │
│   ├── types/                     # 📦 Типы данных
│   │   └── types.go               # Базовые структуры (Tank, Bullet, Block)
│   │
│   └── utils/                     # 🛠️ Утилиты
│       ├── utils.go               # Вспомогательные функции
│       └── utils_test.go          # Тесты утилит
│
├── assets/                        # Игровые ресурсы
│   ├── levels/                    # Уровни игры
│   ├── sounds/                    # Звуковые эффекты
│   ├── new/                       # Новые ресурсы
│   │   ├── blocks.png            # Изображения блоков
│   │   ├── blocks.yml            # Конфигурация блоков
│   │   ├── player.png            # Изображения игрока
│   │   └── player.yml            # Конфигурация игрока
│   └── blocks.yml                 # Конфигурация блоков
│
├── go.mod                         # Go модуль
├── go.sum                         # Зависимости
├── justfile                       # Команды для разработки
└── README.md                      # Документация
```

## 🏗️ Архитектура

Проект следует принципам **Clean Architecture**:

### Слои архитектуры (High-Level)

```mermaid
graph TB
    Entry[cmd/main.go<br/>Entry Point]
    App[internal/app.go<br/>Application Layer<br/>Ebiten Integration]
    States[states/<br/>Game States<br/>Orchestration]
    
    UseCases[use_cases/<br/>Business Logic<br/>• Player<br/>• Bullet<br/>• Map<br/>• Collision]
    Adapters[adapters/<br/>Infrastructure<br/>• Input<br/>• Renderer]
    
    Repos[repositories/<br/>Data Layer<br/>• Game<br/>• Processed<br/>• Raw]
    
    Entry --> App
    App --> States
    States --> UseCases
    States --> Adapters
    Adapters -.-> UseCases
    UseCases --> Repos
```

### Принципы

1. **Dependency Rule** - зависимости направлены внутрь (к бизнес-логике)
2. **Interface Segregation** - интерфейсы определяются там, где используются
3. **Dependency Inversion** - зависимость от абстракций, а не реализаций
4. **Single Responsibility** - каждый компонент отвечает за одну задачу

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

### Конфигурация через переменные окружения

```bash
# Настройка приложения
export APP_NAME="My Battle City"
export SCREEN_WIDTH=240
export SCREEN_HEIGHT=240
export GAME_SPEED=0.016667
export TILE_SIZE=8

# Запуск
go run cmd/main.go
```

## 🎮 Игровая функциональность

### Управление

- **W** - движение вверх
- **S** - движение вниз  
- **A** - движение влево
- **D** - движение вправо
- **Space** - стрельба
- **Escape** - выход

### Игровые объекты

- **Танк игрока** - управляемый танк с поворотом и стрельбой
- **Блоки**:
  - `#` - Кирпич (разрушаемый)
  - `@` - Сталь (неразрушаемый)
  - `%` - Лес (прозрачный для танков)
  - `~` - Вода (непроходимая)
  - `=` - Лед (скользкий)
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

### Use Cases (Бизнес-логика)

- **PlayerUseCases** - движение, поворот, управление игроком
- **BulletUseCases** - создание, обновление, удаление пуль
- **MapUseCases** - работа с блоками карты
- **CollisionUseCases** - проверка коллизий между объектами

### Adapters (Инфраструктура)

- **InputAdapter** - обработка ввода с клавиатуры
- **RendererAdapter** - отрисовка игры через Ebiten

### Repositories (Данные)

- **BlocksRepository** - хранение блоков карты (in-memory)
- **BulletsRepository** - хранение пуль (in-memory)
- **MapsDataRepository** - загрузка уровней из файлов
- **TilesetRepository** - загрузка и кеширование изображений из тайлсетов
- **FileRepository** - чтение файлов из assets

### States (Состояния)

- **GameState** - основное игровое состояние
- Легко добавить: MenuState, PauseState, GameOverState

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

### Пример: добавление врагов

```go
// use_cases/enemy_use_cases.go
type EnemyUseCases struct {
    enemiesRepo repositories.IEnemiesRepository
    tilesetRepo repositories.ITilesetRepository
}

func (uc *EnemyUseCases) SpawnEnemy() error {
    // Логика создания врага
}

func (uc *EnemyUseCases) UpdateEnemies(dt float64) error {
    // Логика обновления врагов
}
```

### Идеи для расширения

- 🤖 **ИИ врагов** - добавить вражеские танки
- 🎯 **Система очков** - подсчет очков за уничтожение
- 🎨 **Меню и UI** - главное меню, экран победы/поражения
- 🌐 **Сетевой режим** - мультиплеер
- 🏆 **Достижения** - система наград
- 🛠️ **Редактор уровней** - создание своих карт
- 💥 **Эффекты** - взрывы, частицы
- 🎵 **Музыка** - фоновая музыка

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
- Нет ИИ врагов
- Нет системы уровней (только уровень 1)
- Нет главного меню
- Нет сохранений

## 🤝 Вклад в проект

1. Fork проекта
2. Создайте feature branch (`git checkout -b feature/amazing-feature`)
3. Commit изменений (`git commit -m 'Add amazing feature'`)
4. Push в branch (`git push origin feature/amazing-feature`)
5. Откройте Pull Request

## 📄 Лицензия

MIT License

## 📚 Полезные ссылки

- [Ebiten Documentation](https://ebiten.org/)
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Just Command Runner](https://github.com/casey/just)
