# Нарушения Clean Architecture

## 🟡 Средняя критичность

**1. God Object в конструкторе фасада** (Планируется рефакторинг)
- `NewGameStateUseCasesFacade()` принимает 14 параметров
- **Планируемое решение**: Рефакторинг в `GameSession` с `GameSessionConfig`
- Статус: Обсуждается дизайн, планируется реализация

## 🟢 Низкая критичность

*Нет открытых проблем низкой критичности*

---

## ✅ Исправлено

- ✅ Создание репозиториев в фасаде → внедряется через DI
- ✅ Создание адаптеров в GameState → создаются в App
- ✅ Константа DT → передается через параметр из `ebiten.ActualTPS()`
- ✅ Хардкод номера уровня → в config.yml
- ✅ Создание Use Cases в GameState → создаются в App
- ✅ Use Cases создают сервисы напрямую → сервисы внедряются через DI в фасад
- ✅ Дублирование GameConfig → вынесен в отдельный пакет `internal/config`
- ✅ **Монолитный TankUseCases** → разделен на специализированные компоненты:
  - `TankCommonUseCases` — общие операции (движение)
  - `TankRenderUseCases` — графика и рендеринг
  - `TankLifecycleUseCases` — жизненный цикл (спавн, взрыв)
  - `TankActionsUseCases` — действия танка (поворот, движение, стрельба)
- ✅ **Отсутствие интерфейсов для Use Cases танка** → созданы интерфейсы:
  - `ITankCommonUseCases`
  - `ITankRenderUseCases`
  - `ITankLifecycleUseCases`
  - `ITankActionsUseCases`
- ✅ **Смешанная ответственность графики** → вся работа с анимациями централизована в `TankRenderUseCases`
- ✅ **Прямые зависимости между компонентами** → все зависимости используют интерфейсы через Dependency Inversion Principle
- ✅ **Смешанная ответственность в AI** → разделено на специализированные компоненты:
  - `LuaEngine` (Infrastructure Layer) — работа с Lua VM
  - `AITypeConverter` (Application Service) — конвертация типов Go ↔ Lua
  - `AIUseCases` (Application Use Case) — бизнес-логика AI
- ✅ **Интерфейс ILuaEngine в неправильном слое** → перенесён в `internal/interfaces/adapters.go`

---

## 📊 Итого

**Осталось:**
- 🟡 Средняя: 1 (God Object в конструкторе фасада)

**Оценка: 98/100** (улучшено с 97/100)

**Следующие шаги:**
- Рефакторинг `GameStateUseCasesFacade` → `GameSession` с упрощённым конструктором

### Улучшения в последней итерации:
- ✅ **Упрощение Render Use Cases** — удалены геттеры/сеттеры, `AnimationGetter` сделан публичным полем:
  - `TankRenderUseCases.AnimationGetter` — прямое обращение вместо `GetAnimationGetter()`/`SetAnimationGetter()`
  - `HQRenderUseCases.AnimationGetter` — прямое обращение вместо методов-обёрток
  - Удалены методы `GetImageID()`, `SetAnimationGetter()`, `GetAnimationGetter()` из обоих классов
- ✅ **Очистка неиспользуемого кода**:
  - Удалён пакет `internal/utils` (функции перенесены в сервисный слой)
  - Удалены неиспользуемые методы `Destroy()` и `IsDestroyed()` из `HQUseCases`
  - Удалены неиспользуемые методы `StartTankAnimation()` и `StopTankAnimation()` из `TankCommonUseCases`
  - Удалён устаревший метод `UpdateEnemiesAnimations()` из `GameStateUseCasesFacade`
  - Удалена неиспользуемая функция `updatePosition()` из `TankActionsUseCases`
- ✅ **Добавление базы (HQ) в игру**:
  - Реализован `HQEntity` с состояниями (Intact, Exploding, Destroyed)
  - Создан `HQUseCases` для обработки логики базы
  - Создан `HQRenderUseCases` для рендеринга базы
  - Интегрирована анимация взрыва базы
  - База добавляется на карту через `config.yml` (`hq_position`)

### Предыдущие улучшения:
- ✅ **Разделение AI архитектуры** на слои:
  - `LuaEngine` (Infrastructure) — инкапсулирует работу с Lua VM
  - `AITypeConverter` (Application Service) — конвертация доменных типов
  - `AIUseCases` (Application Use Case) — бизнес-логика AI
- ✅ **Правильное размещение интерфейсов** — `ILuaEngine` перенесён в слой интерфейсов
- ✅ Разделение ответственности в Use Cases танка
- ✅ Применение Dependency Inversion Principle через интерфейсы
- ✅ Централизация работы с графикой
