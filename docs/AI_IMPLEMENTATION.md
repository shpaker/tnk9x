# AI для вражеских танков - Архитектура и Реализация

## Проблема

Необходимо реализовать AI поведение для вражеских танков в стиле классической игры Battle City.

### Требования:
- AI должен получать полный контекст игры каждый тик
- Все пули, все блоки, все танки
- Решения принимаются скриптовым языком (Lua/Python) для гибкости
- Без необходимости перекомпиляции при изменении логики

## Архитектура решения

### Концепция

```
Game State → Update() → AI Update() → Lua/Python Script → Decision → Apply
```

Каждый тик игры:
1. Собирается контекст (все танки, блоки, пули)
2. Передается в AI модуль
3. Скрипт принимает решение
4. Решение применяется к вражескому танку

### Данные для AI

**EnemyGameContext:**
```go
type EnemyGameContext struct {
    AllTanks  []*types.TankEntity    // Все танки на карте
    AllBlocks []types.BlockEntity    // Все блоки карты
    AllBullets []types.BulletEntity  // Все активные пули
}
```

**Решение AI:**
```go
type EnemyAIDecision struct {
    ShouldMove      bool              // Должен ли танк двигаться
    ShouldShoot     bool              // Должен ли танк стрелять
    NewDirection    types.Direction   // Новое направление движения
}
```

## Сравнение вариантов реализации

### Вариант 1: Нативный Go код

**Плюсы:**
- ⚡ Максимальная производительность
- 🎯 Полный контроль компилятора
- 📦 Нет дополнительных зависимостей
- 🔒 Типобезопасность на этапе компиляции

**Минусы:**
- 🔨 Требует перекомпиляции при изменении логики
- 🧪 Сложно тестировать разные стратегии
- 📝 Изменения требуют знания Go
- 🚀 Медленный цикл разработки

**Оценка:** 3/5 для игрового AI (подходит для простых паттернов, но не гибко)

---

### Вариант 2: Lua скрипты ⭐ (РЕКОМЕНДУЕТСЯ)

**Плюсы:**
- ⚡ Быстрый и легковесный (~5-10MB размер)
- 🎮 Идеально подходит для игровой логики
- 🔧 Легко встраивается в Go (`github.com/yuin/gopher-lua`)
- 📝 Простой и понятный синтаксис
- 🔄 Изменение без перекомпиляции
- 📚 Широко используется в играх (World of Warcraft, Civilization)
- 🎯 Нет проблем с GIL (как в Python)
- 🛡️ Изолированное выполнение
- 🔍 Хорошие отладочные возможности

**Минусы:**
- 📦 Нужна дополнительная зависимость
- 📖 Язык менее популярен чем Python/JS
- 🧠 Ограниченная стандартная библиотека

**Производительность:**
- Скорость выполнения: ~95% от нативного Go кода
- Объем памяти: ~5-15MB на AI модуль
- Время инициализации: <10ms
- **Размер бинарника:** +5-10 MB (Lua VM + скрипты)

**Зависимость:**
```bash
go get github.com/yuin/gopher-lua
```

**Пример использования:**
```go
ai, _ := NewEnemyAILua("assets/scripts/enemy_ai.lua")
decision := ai.Update(enemy, ctx.AllTanks, ctx.AllBlocks, ctx.AllBullets)
```

**Пример Lua кода:**
```lua
function updateEnemyAI(enemy, tanks, blocks, bullets)
    -- Простая и понятная логика
    if enemy.speed == 0 then
        return true, false, math.random(0, 3)  -- Двигаться в случайном направлении
    end
    return false, true, enemy.direction  -- Стрелять
end
```

**Оценка:** 5/5 - Оптимальный выбор для игрового AI

---

### Вариант 3: Python скрипты

**Плюсы:**
- 🔥 Мощный язык для сложной логики
- 📚 Огромная экосистема библиотек (numpy, scipy для AI)
- 👨‍💼 Язык знаком большинству разработчиков
- 🔬 Отлично для машинного обучения и сложных алгоритмов
- 📖 Богатая документация
- 🛠️ Отличные инструменты для отладки

**Минусы:**
- 📏 Тяжелый (~50MB+ размер)
- ⏱️ Сложная интеграция через cgo
- 🐌 Медленнее Lua (в 2-3 раза)
- 🔒 Проблемы с GIL (Global Interpreter Lock)
- 📦 Требует установленный Python интерпретатор
- 💾 Высокое потребление памяти
- 🔄 Медленная инициализация

**Производительность:**
- Скорость выполнения: ~30-40% от нативного Go кода
- Объем памяти: ~50-100MB на AI модуль
- Время инициализации: 50-200ms

**Зависимость:**
```bash
go get github.com/DataDog/go-python3
# Требует Python 3.x установленный в системе
```

**Пример использования:**
```go
import "github.com/DataDog/go-python3"

// Инициализация Python интерпретатора
python3.Py_Initialize()

// Загрузка модуля
module := python3.PyImport_ImportModule("enemy_ai")
if module == nil {
    log.Fatal("Failed to import enemy_ai module")
}

// Вызов функции
fn := module.GetAttrString("updateEnemyAI")
result := fn.Call(convertedArgs)

// Конвертация результатов обратно в Go типы
```

**Детальный пример интеграции см. в разделе "Практическая интеграция Python" ниже.**

**Оценка:** 3/5 - Перегружен для простого игрового AI, но хорош для сложных ML стратегий

---

### Вариант 4: Конфигурационный подход (YAML/JSON + паттерны)

**Плюсы:**
- ✅ Очень легковесный (нет скриптовых движков)
- 🔧 Нет дополнительных зависимостей
- 📖 Человекочитаемый формат
- 🛡️ Полная типобезопасность

**Минусы:**
- 🎯 Очень ограниченная гибкость
- 📝 Нельзя выразить сложную логику
- 🔨 Все равно требует кода-обработчика в Go

**Пример:**
```yaml
# enemy_behaviors.yaml
aggressive:
  move_probability: 0.8
  shoot_probability: 0.7
  avoid_blocks: false
  
defensive:
  move_probability: 0.3
  shoot_probability: 0.2
  avoid_blocks: true
```

**Оценка:** 2/5 - Слишком ограничен для полноценного AI

---

### Вариант 5: Модульная система на Go (composable AI)

**Плюсы:**
- ⚡ Лучшая производительность
- 🧩 Модульная архитектура
- 🔧 Легко тестировать компоненты
- 📚 Без внешних зависимостей

**Минусы:**
- 🔨 Требует перекомпиляции
- 🏗️ Нужно продумать архитектуру заранее

**Пример:**
```go
type Behavior interface {
    Update(tank *Tank, ctx *GameContext) Decision
}

type RandomBehavior struct {}
type ChaseBehavior struct { target *Tank }
type PatrolBehavior struct { points []Point }

// Композиция:
tank.Behavior = NewBehaviorTree(
    NewSelector(
        NewChaseBehavior(),
        NewRandomBehavior(),
    ),
)
```

**Оценка:** 4/5 - Отличный выбор для serious game development

---

### Вариант 6: Встроенные DSL для Go (expr, Starlark, YAL)

**Что это?** Специально созданные маленькие языки, которые компилируются в Go код или выполняются очень быстро.

#### 6.1. **expr** (Expression Language) ⭐

**Плюсы:**
- ⚡ Компилируется в нативный Go код (100% производительность!)
- 📦 Очень легковесный (~2MB)
- 🔧 Логические операторы из коробки
- 🛡️ Полная типобезопасность
- 📝 Синтаксис похож на Go/JavaScript
- 🔄 Можно вызывать Go функции из скрипта

**Минусы:**
- 🎯 Ограничен функциональностью (выражения, не полноценный язык)
- 📖 Нужно изучать синтаксис expr
- 🔬 Не подходит для очень сложной логики (нет циклов, ограниченные условия)

**Производительность:**
- Скорость выполнения: 100% от нативного Go кода (компилируется!)
- Объем памяти: ~2-5MB
- Время инициализации: <5ms

**Зависимость:**
```bash
go get github.com/antonmedv/expr
```

**Пример:**
```go
// AI скрипт
program, _ := expr.Compile(`
    enemy.speed == 0 && distance(enemy.position, player.position) < 100 ?
        {shouldMove: true, shouldShoot: true, direction: getDirectionTo(enemy, player)} :
        {shouldMove: false, shouldShoot: false, direction: enemy.direction}
`, expr.Env(env))

// Выполнение
result, _ := expr.Run(program, env)
decision := result.(EnemyAIDecision)
```

**Оценка:** 4/5 - Отличный выбор для простых и средних игровых правил

#### 6.2. **Starlark** (Python-подобный)

**Плюсы:**
- 📝 Синтаксис похож на Python (знаком многим)
- ⚡ Написан на Go, быстрый
- 🛡️ Безопасное выполнение (изоляция)
- 📦 Средний размер (~10MB)
- 🔧 Подходит для кода с функциями и циклами

**Минусы:**
- 📚 Меньше популярен чем Lua
- 🎮 Меньше игрового опыта чем Lua

**Производительность:**
- Скорость выполнения: ~70-80% от нативного Go кода
- Объем памяти: ~10-15MB

**Зависимость:**
```bash
go get go.starlark.net/starlark
```

**Пример:**
```python
# enemy_ai.star
def updateEnemyAI(enemy, tanks, blocks, bullets):
    player = findNearestPlayer(enemy, tanks)
    
    if player and distance(enemy, player) < 100:
        return {"shouldMove": True, "shouldShoot": True}
    
    return {"shouldMove": enemy.speed == 0, "shouldShoot": False}
```

**Оценка:** 3.5/5 - Хорошая альтернатива Lua для тех кто знает Python

#### 6.3. **YAL** (Yet Another Language)

**Плюсы:**
- ⚡ Очень легковесный (~500KB)
- 🎯 Специально для Go
- 📝 Простой синтаксис

**Минусы:**
- 📚 Мало документации
- 🔧 Ограниченная функциональность
- 🛠️ На ранней стадии разработки

**Оценка:** 2.5/5 - Интересно, но не готово для продакшена

---

## Сравнительная таблица

| Критерий | Нативный Go | Lua ⭐ | Python | **expr** | Starlark | YAML/JSON |
|----------|-------------|--------|--------|----------|----------|-----------|
| **Описание** | Логика в Go коде | Скриптовый язык для игр | Язык для ML/AI | Язык выражений | Python-подобный язык | Конфигурация |
| **Репозиторий** | — | [yuin/gopher-lua](https://github.com/yuin/gopher-lua) | [DataDog/go-python3](https://github.com/DataDog/go-python3) | [antonmedv/expr](https://github.com/antonmedv/expr) | [bazelbuild/starlark](https://github.com/bazelbuild/starlark) | [go-yaml](https://github.com/go-yaml/yaml) |
| **Производительность** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Гибкость** | ⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐ |
| **Увеличение бинарника** | 0 MB | +5-10 MB | +50-100 MB | +2-5 MB | +10-15 MB | 0 MB |
| **Внешние зависимости** | ❌ Нет | ❌ Нет | ✅ Python | ❌ Нет | ❌ Нет | ❌ Нет |
| **Легковесность** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Простота интеграции** | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Без перекомпиляции** | ⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Отладка** | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |
| **Знакомость языка** | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Игровой опыт** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐ | ⭐⭐ |

### Детальная таблица размеров

**Размер базового исполняемого файла (без AI):** ~5-10 MB

| Вариант | Описание (RU) | Пакет | Увеличение размера | Итоговый размер | Процент увеличения |
|---------|---------------|--------|-------------------|-----------------|-------------------|
| **Нативный Go** | Логика в Go коде | — | 0 MB | ~5-10 MB | 0% ✅ |
| **Lua** | Скриптовый язык для игр | [gopher-lua](https://github.com/yuin/gopher-lua) | +5-10 MB | ~10-20 MB | +50-100% |
| **Python** | Язык для ML/AI | [go-python3](https://github.com/DataDog/go-python3) | +50-100 MB | ~55-110 MB | +500-1000% ❌ |
| **expr** | Язык выражений | [expr](https://github.com/antonmedv/expr) | +2-5 MB | ~7-15 MB | +20-50% |
| **Starlark** | Python-подобный язык | [starlark](https://github.com/bazelbuild/starlark) | +10-15 MB | ~15-25 MB | +100-150% |
| **YAML/JSON** | Конфигурация | [go-yaml](https://github.com/go-yaml/yaml) | 0 MB | ~5-10 MB | 0% ✅ |

**Примечания:**
- Python требует установленный Python 3.x на целевой системе (+ еще 50-100 MB)
- Lua и expr - статическая компиляция, один бинарный файл
- YAML/JSON - только если используется существующий Go пакет yaml.v3
- Starlark также статически компилируется

**Итоговая оценка (для игрового AI):**
1. 🥇 **Lua** - 5/5 (лучший баланс для игр) ⭐
2. 🥈 **expr** - 4.5/5 (идеален для простых правил, компилируется в Go!)
3. 🥉 **Go поведенческая система** - 4.5/5 (для больших проектов)
4. **Starlark** - 3.5/5 (если команда знает Python синтаксис)
5. **Python** - 3/5 (для сложных ML стратегий)
6. **YAML/JSON** - 2/5 (слишком ограничен)
7. **Нативный Go** - 2/5 (нет гибкости)

---

## Рекомендация

### Вариант 1: Для гибкого игрового AI → **Lua** 🎯 (РЕКОМЕНДУЕТСЯ)

**Для Battle City клона: Lua** — лучший выбор для полноценного игрового AI.

Почему:
- ⚡ Отличная производительность для игрового AI
- 📦 Минимальный overhead
- 🎮 Проверен десятилетиями использования в играх
- 📝 Простой синтаксис для игровой логики
- 🔄 Быстрый цикл разработки и тестирования
- 💡 Идеальный баланс между гибкостью и производительностью
- 🎮 Поддержка сложной логики (циклы, функции, условия)

### Вариант 2: Для простых правил → **expr** 🚀

Если ваши AI правила можно описать простыми выражениями (if/then, математика, сравнения):

**Примеры подходящих правил:**
- ✅ `enemy.speed == 0 && distance(enemy, player) < 100 → chase`
- ✅ `enemy.health < 30% → retreat`
- ✅ `path_clear(direction) → move`
- ✅ `random() > 0.7 → shoot`

**Почему expr:**
- ⚡ Компилируется в нативный Go код (100% производительность!)
- 📦 Очень легковесный (~2MB)
- 🛡️ Полная типобезопасность
- 🎯 Отлично подходит для простых игровых правил
- 🔄 Нет необходимости в полноценном скриптовом языке

**Когда не использовать expr:**
- ❌ Нужны циклы (for/while)
- ❌ Нужны сложные функции с множеством локальных переменных
- ❌ Нужна рекурсия

### Вариант 3: Если команда знает Python → **Starlark** 🐍

Если ваша команда более знакома с Python синтаксисом, чем с Lua:

**Плюсы:**
- 📝 Синтаксис как у Python
- ⚡ Написан на Go, быстрый
- 🛡️ Изолированное выполнение
- 📚 Лучше подходит для сложной логики чем expr

**Оценка:** 3.5/5 — хорошая альтернатива Lua

### Золотая середина: Комбинированный подход 💡

```go
// Простые правила → expr
simpleBehavior := expr.Compile(`
    enemy.speed == 0 && distance(enemy, player) < 50 ?
        {shouldMove: true, shouldShoot: true, direction: getDirectionTo(enemy, player)} :
        {shouldMove: false, shouldShoot: false, direction: enemy.direction}
`)

// Сложная логика → Lua
complexBehavior := NewLuaAI("assets/ai/advanced_behaviors.lua")

// Применение в зависимости от типа врага
func GetDecision(enemy *TankEntity) {
    if enemy.BehaviorType == "simple" {
        return simpleBehavior.Evaluate(enemy)
    } else {
        return complexBehavior.Update(enemy)
    }
}
```

**Преимущества:**
- ⚡ Максимальная производительность для простых случаев
- 🎯 Гибкость для сложных AI
- 📦 Оптимальное использование ресурсов

## Текущая структура кода

### Уже реализовано:

1. **EnemyUseCases** - управление одним вражеским танком
   - Спавн с анимацией
   - Взрыв с анимацией
   - Анимации движения

2. **GameStateUseCasesFacade** - оркестрация
   - Массив до 3 врагов
   - Обновление спавна врагов
   - Обновление анимаций

3. **CollisionUseCases** - коллизии
   - Проверка танк-блок
   - Проверка танк-танк
   - Проверка пуля-объекты

### Не реализовано:

1. **Логика движения врагов** - танки просто стоят
2. **Логика стрельбы врагов** - не стреляют
3. **AI поведение** - нет принятия решений

## Реализация через Lua

### Структура файлов:

```
internal/
  use_cases/
    enemy_ai_lua.go      # Go wrapper для Lua
    enemy_use_cases.go    # Интеграция AI
  states/
    game_state_use_cases_facade.go  # Вызов UpdateAI()
assets/
  scripts/
    enemy_ai.lua          # Lua скрипт с логикой
```

### Интерфейс для AI:

```go
// internal/use_cases/enemy_ai_interface.go

type IEnemyAI interface {
    Update(
        enemy *types.TankEntity,
        allTanks []*types.TankEntity,
        allBlocks []types.BlockEntity,
        allBullets []types.BulletEntity,
    ) EnemyAIDecision
}

type EnemyAIDecision struct {
    ShouldMove      bool
    ShouldShoot     bool
    NewDirection    types.Direction
}
```

### Реализация EnemyAILua:

```go
// internal/use_cases/enemy_ai_lua.go

package use_cases

import (
    lua "github.com/yuin/gopher-lua"
    "github.com/shpaker/gonflict/internal/types"
)

type EnemyAILua struct {
    L          *lua.LState
    scriptPath string
}

func NewEnemyAILua(scriptPath string) (*EnemyAILua, error) {
    ai := &EnemyAILua{
        L:          lua.NewState(),
        scriptPath: scriptPath,
    }
    
    // Загружаем скрипт
    if err := ai.L.DoFile(scriptPath); err != nil {
        return nil, err
    }
    
    return ai, nil
}

func (ai *EnemyAILua) Update(enemy, allTanks, allBlocks, allBullets) EnemyAIDecision {
    // Вызываем Lua функцию
    err := ai.L.CallByParam(lua.P{
        Fn:      ai.L.GetGlobal("updateEnemyAI"),
        NRet:    3,
        Protect: true,
    }, 
    ai.convertToLuaTable(enemy),
    ai.convertToLuaArray(allTanks),
    ai.convertToLuaArray(allBlocks),
    ai.convertToLuaArray(allBullets))
    
    if err != nil {
        return EnemyAIDecision{} // Вернуть безопасное решение
    }
    
    // Получаем результаты
    ret := ai.L.Get(-3).(lua.LBool)
    shouldMove := bool(ret)
    ret = ai.L.Get(-2).(lua.LBool)
    shouldShoot := bool(ret)
    direction := ai.L.Get(-1).(lua.LNumber)
    
    ai.L.Pop(3)
    
    return EnemyAIDecision{
        ShouldMove:   shouldMove,
        ShouldShoot:  shouldShoot,
        NewDirection: intToDirection(int(direction)),
    }
}

// Вспомогательные функции конвертации...
```

### Пример Lua скрипта:

```lua
-- assets/scripts/enemy_ai.lua

-- Таймеры
local moveTimer = 0
local shootTimer = 0

-- Основная функция AI
function updateEnemyAI(enemy, allTanks, allBlocks, allBullets)
    local shouldMove = false
    local shouldShoot = false
    local newDirection = enemy.direction
    
    -- Находим игрока (первый заспавненный танк)
    local player = nil
    for i = 1, #allTanks do
        local tank = allTanks[i]
        if tank.isSpawned and not tank.isExploding then
            player = tank
            break
        end
    end
    
    -- Если танк стоит, выбираем направление
    if enemy.speed == 0 then
        shouldMove = true
        
        -- Простая логика: случайное направление
        -- TODO: добавить проверку препятствий
        -- TODO: добавить поиск игрока
        newDirection = math.random(0, 3)
    end
    
    -- Периодическая стрельба
    shootTimer = shootTimer + (1/60)
    if shootTimer > 2.0 then
        shouldShoot = true
        shootTimer = 0
    end
    
    return shouldMove, shouldShoot, newDirection
end

-- Вспомогательные функции
function distance(x1, y1, x2, y2)
    local dx = x2 - x1
    local dy = y2 - y1
    return math.sqrt(dx*dx + dy*dy)
end
```

### Интеграция в EnemyUseCases:

```go
// internal/use_cases/enemy_use_cases.go

type EnemyUseCases struct {
    // ... existing fields ...
    ai         IEnemyAI
}

func NewEnemyUseCases(..., ai IEnemyAI) *EnemyUseCases {
    return &EnemyUseCases{
        // ... existing initialization ...
        ai: ai,
    }
}

func (uc *EnemyUseCases) UpdateAI(
    allTanks []*types.TankEntity,
    allBlocks []types.BlockEntity,
    allBullets []types.BulletEntity,
) {
    if uc.enemyTank == nil || !uc.enemyTank.IsSpawned || uc.enemyTank.IsExploding {
        return
    }
    
    // Получаем решение от AI
    decision := uc.ai.Update(uc.enemyTank, allTanks, allBlocks, allBullets)
    
    // Применяем движение
    if decision.ShouldMove && uc.enemyTank.Speed == 0 {
        uc.enemyTank.Direction = decision.NewDirection
        uc.enemyTank.Speed = 32.0
    }
    
    // Применяем стрельбу
    if decision.ShouldShoot {
        // Используем IBulletUseCases для стрельбы
        // Нужно передать ссылку на bulletUseCases
    }
}

// Методы для движения вражеского танка
func (uc *EnemyUseCases) MoveTank(dt float64) {
    if uc.enemyTank == nil || !uc.enemyTank.IsSpawned {
        return
    }
    
    delta := uc.enemyTank.Speed * dt
    
    switch uc.enemyTank.Direction {
    case types.DirectionUp:
        uc.enemyTank.WorldPosition.Y -= delta
    case types.DirectionDown:
        uc.enemyTank.WorldPosition.Y += delta
    case types.DirectionLeft:
        uc.enemyTank.WorldPosition.X -= delta
    case types.DirectionRight:
        uc.enemyTank.WorldPosition.X += delta
    }
}
```

### Вызов из Facade:

```go
// internal/states/game_state_use_cases_facade.go

func (g *GameStateUseCasesFacade) Update() {
    // ... existing code ...
    
    // Обновляем AI врагов
    allTanks := g.getAllTanks()  // Получить всех танков
    allBlocks := g.mapUseCases.GetBlocks()
    allBullets := g.bulletUseCases.GetBullets()
    
    for _, enemyUseCases := range g.enemyUseCasesList {
        enemyUseCases.UpdateAI(allTanks, allBlocks, allBullets)
        enemyUseCases.MoveTank(use_cases.DT)
    }
    
    // ... existing collision check ...
}
```

## Альтернативная реализация через Python

Если вы решили использовать Python вместо Lua (не рекомендуется для большинства случаев, но может быть полезно для ML стратегий):

### Установка и настройка Python

1. **Установить Python 3.x** (обязательно на систему):
```bash
# macOS
brew install python3

# Linux
sudo apt-get install python3-dev

# Windows
# Скачать с python.org
```

2. **Добавить зависимость в Go:**
```bash
go get github.com/DataDog/go-python3
```

3. **Установить необходимые флаги компиляции:**
```bash
# В macOS/Linux может потребоваться
export CGO_CFLAGS="-I/usr/include/python3.11"
export CGO_LDFLAGS="-L/usr/lib/python3.11"

# В Windows используйте TDM-GCC или MinGW
```

### Создание Python AI скрипта

**Файл `assets/scripts/enemy_ai.py`:**

```python
# -*- coding: utf-8 -*-

import random
import math

def distance(x1, y1, x2, y2):
    """Вычисляет расстояние между двумя точками"""
    dx = x2 - x1
    dy = y2 - y1
    return math.sqrt(dx * dx + dy * dy)

def find_nearest_player(enemy, all_tanks):
    """Находит ближайшего игрока"""
    nearest = None
    min_dist = float('inf')
    
    for tank in all_tanks:
        if tank['id'] == enemy['id']:
            continue
        if not tank.get('is_spawned', False):
            continue
        if tank.get('is_exploding', False):
            continue
        
        dist = distance(
            enemy['x'], enemy['y'],
            tank['x'], tank['y']
        )
        
        if dist < min_dist:
            min_dist = dist
            nearest = tank
    
    return nearest

def get_direction_to(target_x, target_y, current_x, current_y):
    """Определяет направление к цели"""
    dx = target_x - current_x
    dy = target_y - current_y
    
    if abs(dx) > abs(dy):
        return 3 if dx > 0 else 2  # right or left
    else:
        return 1 if dy > 0 else 0  # down or up

def update_enemy_ai(enemy, all_tanks, all_blocks, all_bullets):
    """Основная функция AI - принимает решение для врага"""
    should_move = False
    should_shoot = False
    new_direction = enemy['direction']
    
    # Находим ближайшего игрока
    player = find_nearest_player(enemy, all_tanks)
    
    # Если танк стоит, решаем куда двигаться
    if enemy['speed'] == 0:
        should_move = True
        
        if player and distance(enemy['x'], enemy['y'], player['x'], player['y']) < 100:
            # Преследуем игрока
            new_direction = get_direction_to(
                player['x'], player['y'],
                enemy['x'], enemy['y']
            )
        else:
            # Случайное блуждание
            new_direction = random.randint(0, 3)
    
    # Периодическая стрельба (раз в 2 секунды при 60 FPS)
    shoot_timer = enemy.get('shoot_timer', 0)
    shoot_timer += 1/60
    
    if shoot_timer > 2.0:
        should_shoot = True
        enemy['shoot_timer'] = 0
    else:
        enemy['shoot_timer'] = shoot_timer
    
    return should_move, should_shoot, new_direction
```

### Реализация EnemyAIPython

**Создайте файл `internal/use_cases/enemy_ai_python.go`:**

```go
package use_cases

import (
    "log"
    python3 "github.com/DataDog/go-python3"
    "github.com/shpaker/gonflict/internal/types"
)

type EnemyAIPython struct {
    module *python3.PyObject
    fn     *python3.PyObject
}

func NewEnemyAIPython(scriptPath string) (*EnemyAIPython, error) {
    // Инициализация Python интерпретатора
    python3.Py_Initialize()
    if !python3.Py_IsInitialized() {
        return nil, errors.New("Failed to initialize Python")
    }
    
    ai := &EnemyAIPython{}
    
    // Добавляем путь к скриптам
    sysModule := python3.PyImport_ImportModule("sys")
    path := sysModule.GetAttrString("path")
    python3.PyList_Insert(path, 0, python3.PyUnicode_FromString("assets/scripts"))
    
    // Загружаем модуль
    module := python3.PyImport_ImportModule("enemy_ai")
    if module == nil {
        python3.PyErr_Print()
        return nil, errors.New("Failed to import enemy_ai module")
    }
    
    ai.module = module
    
    // Получаем функцию
    fn := module.GetAttrString("update_enemy_ai")
    if fn == nil {
        python3.PyErr_Print()
        return nil, errors.New("Failed to get update_enemy_ai function")
    }
    
    ai.fn = fn
    
    log.Println("Python AI module loaded successfully")
    return ai, nil
}

func (ai *EnemyAIPython) Update(
    enemy *types.TankEntity,
    allTanks []*types.TankEntity,
    allBlocks []types.BlockEntity,
    allBullets []types.BulletEntity,
) (shouldMove bool, shouldShoot bool, newDirection types.Direction) {
    
    // Конвертируем данные в Python объекты
    enemyDict := ai.convertTankToDict(enemy)
    tanksList := ai.convertTanksToList(allTanks)
    blocksList := ai.convertBlocksToList(allBlocks)
    bulletsList := ai.convertBulletsToList(allBullets)
    
    // Вызываем функцию
    result := ai.fn.Call(
        python3.PyTuple_Pack(4, enemyDict, tanksList, blocksList, bulletsList),
    )
    
    if result == nil {
        python3.PyErr_Print()
        return false, false, enemy.Direction
    }
    
    // Извлекаем результаты
    shouldMove = python3.PyBool_Check(result.GetItem(0)) && bool(result.GetItem(0).Bool())
    shouldShoot = python3.PyBool_Check(result.GetItem(1)) && bool(result.GetItem(1).Bool())
    directionInt := int(result.GetItem(2).PyLong_AsLong())
    
    result.DecRef()
    
    return shouldMove, shouldShoot, intToDirection(directionInt)
}

func (ai *EnemyAIPython) convertTankToDict(tank *types.TankEntity) *python3.PyObject {
    dict := python3.PyDict_New()
    
    python3.PyDict_SetItemString(dict, "id", python3.PyLong_FromLong(0))
    python3.PyDict_SetItemString(dict, "x", python3.PyFloat_FromDouble(tank.WorldPosition.X))
    python3.PyDict_SetItemString(dict, "y", python3.PyFloat_FromDouble(tank.WorldPosition.Y))
    python3.PyDict_SetItemString(dict, "speed", python3.PyFloat_FromDouble(tank.Speed))
    python3.PyDict_SetItemString(dict, "direction", python3.PyLong_FromLong(int(directionToInt(tank.Direction))))
    python3.PyDict_SetItemString(dict, "is_spawned", python3.PyBool_FromLong(1))
    
    if tank.IsExploding {
        python3.PyDict_SetItemString(dict, "is_exploding", python3.PyBool_FromLong(1))
    } else {
        python3.PyDict_SetItemString(dict, "is_exploding", python3.PyBool_FromLong(0))
    }
    
    return dict
}

func (ai *EnemyAIPython) convertTanksToList(tanks []*types.TankEntity) *python3.PyObject {
    list := python3.PyList_New(len(tanks))
    
    for i, tank := range tanks {
        dict := ai.convertTankToDict(tank)
        python3.PyList_SetItem(list, i, dict)
    }
    
    return list
}

// Аналогичные функции для blocks и bullets...
func (ai *EnemyAIPython) convertBlocksToList(blocks []types.BlockEntity) *python3.PyObject {
    // Реализация конвертации блоков
    return python3.PyList_New(0)
}

func (ai *EnemyAIPython) convertBulletsToList(bullets []types.BulletEntity) *python3.PyObject {
    // Реализация конвертации пуль
    return python3.PyList_New(0)
}

// Cleanup при завершении
func (ai *EnemyAIPython) Close() {
    if ai.fn != nil {
        ai.fn.DecRef()
    }
    if ai.module != nil {
        ai.module.DecRef()
    }
    python3.Py_Finalize()
}
```

### Использование Python AI

```go
// В EnemyUseCases
ai, err := NewEnemyAIPython("assets/scripts/enemy_ai.py")
if err != nil {
    log.Fatalf("Failed to load Python AI: %v", err)
}

defer ai.Close() // Важно вызвать в defer для cleanup
```

### Проблемы Python интеграции

⚠️ **Важно знать:**

1. **Требует Python в системе** - пользователь должен иметь Python 3.x установленный
2. **CGO необходимо** - усложняет кросс-компиляцию и сборку
3. **Медленная инициализация** - ~50-200ms при старте
4. **Высокое потребление памяти** - ~50-100MB
5. **GIL проблемы** - Global Interpreter Lock может замедлять многопоточность

**Когда использовать Python:**
- ✅ Нужны ML/нейронные сети (tensorflow, pytorch)
- ✅ Сложные математические алгоритмы
- ✅ Команда знает только Python
- ✅ Нужны библиотеки (numpy, scipy, pandas)

**Когда НЕ использовать Python:**
- ❌ Простая игровая логика (Lua лучше)
- ❌ Нужна максимальная производительность
- ❌ Нужна простая дистрибуция (один бинарник)
- ❌ Мобильная платформа (iOS/Android)

---

## Стратегии AI для Battle City

### Базовые правила:

1. **Случайное блуждание**
   - Выбирать случайное направление каждые N секунд
   - Избегать блоков
   - Стрелять периодически

2. **Преследование игрока**
   - Вычислять направление к игроку
   - Двигаться в направлении игрока
   - Стрелять при приближении

3. **Защита базы**
   - Движение вокруг базы
   - Атаковать приближающихся врагов

4. **Агрессивное поведение**
   - Быстро двигаться к игроку
   - Часто стрелять
   - Не избегать опасностей

### Примеры Lua функций:

```lua
-- Поиск ближайшего игрока
function findNearestPlayer(enemy, allTanks)
    local minDist = math.huge
    local nearest = nil
    
    for i = 1, #allTanks do
        local tank = allTanks[i]
        if tank ~= enemy and tank.isSpawned and not tank.isExploding then
            local dist = distance(
                enemy.x, enemy.y,
                tank.x, tank.y
            )
            if dist < minDist then
                minDist = dist
                nearest = tank
            end
        end
    end
    
    return nearest
end

-- Проверка свободен ли путь
function isPathClear(enemy, direction, allBlocks)
    local checkX = enemy.x
    local checkY = enemy.y
    
    if direction == 0 then -- up
        checkY = checkY - 16
    elseif direction == 1 then -- down
        checkY = checkY + 16
    elseif direction == 2 then -- left
        checkX = checkX - 16
    else -- right
        checkX = checkX + 16
    end
    
    for i = 1, #allBlocks do
        local block = allBlocks[i]
        if block.x == checkX and block.y == checkY then
            return false
        end
    end
    
    return true
end

-- Выбор направления к цели
function getDirectionTo(targetX, targetY, currentX, currentY)
    local dx = targetX - currentX
    local dy = targetY - currentY
    
    if math.abs(dx) > math.abs(dy) then
        return dx > 0 and 3 or 2 -- right or left
    else
        return dy > 0 and 1 or 0 -- down or up
    end
end
```

## План реализации

### Шаг 1: Добавить зависимость
```bash
go get github.com/yuin/gopher-lua
```

### Шаг 2: Создать интерфейс AI
- `internal/use_cases/enemy_ai_interface.go`

### Шаг 3: Реализовать Lua AI
- `internal/use_cases/enemy_ai_lua.go`

### Шаг 4: Создать Lua скрипт
- `assets/scripts/enemy_ai.lua`

### Шаг 5: Интегрировать в EnemyUseCases
- Добавить поле `ai IEnemyAI`
- Реализовать метод `UpdateAI()`
- Реализовать метод `MoveTank()`

### Шаг 6: Обновить Facade
- Собрать контекст (все танки, блоки, пули)
- Вызвать `UpdateAI()` для каждого врага
- Вызвать `MoveTank()` для каждого врага

### Шаг 7: Добавить поддержку стрельбы
- Передать `IBulletUseCases` в `EnemyUseCases`
- Реализовать стрельбу врагов

## Преимущества подхода

1. **Гибкость** - изменение логики без перекомпиляции
2. **Тестируемость** - легко тестировать разные стратегии
3. **Модульность** - AI отделен от игровой логики
4. **Расширяемость** - легко добавить новые типы поведения
5. **Производительность** - Lua быстро выполняется

## Заключение

Реализация AI через Lua предоставляет оптимальный баланс между:
- Гибкостью и легкостью изменения логики
- Производительностью выполнения
- Простотой интеграции

Это позволит создавать разнообразное поведение врагов без изменения Go кода.

