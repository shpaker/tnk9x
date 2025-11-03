# Лучшие практики реализации AI в простых играх

Документ собирает лучшие практики и рекомендации для реализации искусственного интеллекта в простых играх, особенно при использовании скриптовых языков (Lua, JavaScript и т.д.).

---

## 📚 Содержание

1. [Архитектура AI](#архитектура-ai)
2. [Паттерны проектирования](#паттерны-проектирования)
3. [Оптимизация производительности](#оптимизация-производительности)
4. [Работа с состоянием](#работа-с-состоянием)
5. [Отладка и тестирование](#отладка-и-тестирование)
6. [Ресурсы и ссылки](#ресурсы-и-ссылки)

---

## 🏗️ Архитектура AI

### Принцип разделения ответственности

**Практика:** Разделяйте AI на слои:
- **Infrastructure Layer** — работа с движком (Lua VM, выполнение скриптов)
- **Application Service** — конвертация типов между Go и Lua
- **Use Case** — бизнес-логика AI (вызов Lua функций, обработка результатов)

**Преимущества:**
- Легче тестировать каждый слой отдельно
- Проще заменять реализацию (например, другой скриптовый движок)
- Соответствует принципам Clean Architecture

**Реализация в Gonflict:**
```
LuaEngine (Infrastructure) → AITypeConverter (Service) → AIUseCases (Use Case)
```

### Минимальный API

**Практика:** Предоставляйте только необходимые данные в скрипт.

**Пример:**
```lua
-- ❌ Плохо: слишком много данных
enemy = {
    x, y, direction, speed,
    health, maxHealth, armor,
    lastPosition, velocity,
    internalState, metadata, ...
}

-- ✅ Хорошо: только необходимые данные
enemy = {
    x, y, direction, speed
}
```

### Идемпотентность

**Практика:** Функция AI должна быть идемпотентной — одинаковые входные данные должны давать одинаковый результат (или предсказуемое поведение при использовании random).

**Пример:**
```lua
-- ❌ Плохо: состояние влияет на результат
local counter = 0
function updateEnemyAI(enemy, context)
    counter = counter + 1
    if counter % 10 == 0 then
        return true, randomDirection()
    end
    return true, enemy.direction
end

-- ✅ Хорошо: результат зависит только от входных данных
function updateEnemyAI(enemy, context)
    if enemy.speed == 0 then
        return true, randomDirection()
    end
    return true, enemy.direction
end
```

---

## 🎯 Паттерны проектирования

### 1. Конечные автоматы (Finite State Machines)

**Описание:** Управление состояниями AI через явные состояния и переходы.

**Когда использовать:** Когда поведение AI имеет четкие состояния (Idle, Chase, Attack, Flee).

**Пример:**
```lua
local STATE = {
    IDLE = "idle",
    CHASE = "chase",
    ATTACK = "attack",
    FLEE = "flee"
}

function updateEnemyAI(enemy, context)
    local state = getState(enemy.id) or STATE.IDLE
    
    if state == STATE.IDLE then
        if isPlayerNearby(enemy, context) then
            setState(enemy.id, STATE.CHASE)
            return true, directionToPlayer(enemy, context)
        end
        return randomWander(enemy)
    elseif state == STATE.CHASE then
        if distanceToPlayer(enemy, context) < 50 then
            setState(enemy.id, STATE.ATTACK)
            return true, enemy.direction -- остановиться и стрелять
        end
        return true, directionToPlayer(enemy, context)
    end
    
    return true, enemy.direction
end
```

**Ресурсы:**
- [Game Programming Patterns - State Pattern](https://gameprogrammingpatterns.com/state.html) (EN) - Паттерн состояния для управления поведением AI
- [AI for Games - Finite State Machines](https://www.ai-junkie.com/architecture/state_driven/tut_state1.html) (EN) - Введение в конечные автоматы для игрового AI

### 2. Структурированное поведение (Behavior Composition)

**Описание:** Комбинирование простых поведений для создания сложного.

**Пример:**
```lua
local behaviors = {
    wander = function(enemy, context)
        return true, randomDirection()
    end,
    
    chasePlayer = function(enemy, context)
        if context.player then
            return true, directionToPoint(enemy, context.player)
        end
        return false, enemy.direction
    end,
    
    avoidBullets = function(enemy, context)
        local nearestBullet = findNearestBullet(enemy, context)
        if nearestBullet and distance(enemy, nearestBullet) < 30 then
            return true, oppositeDirection(enemy.direction)
        end
        return false, enemy.direction
    end
}

function updateEnemyAI(enemy, context)
    -- Приоритет: избегание > преследование > блуждание
    local shouldMove, direction = behaviors.avoidBullets(enemy, context)
    if not shouldMove then
        shouldMove, direction = behaviors.chasePlayer(enemy, context)
    end
    if not shouldMove then
        shouldMove, direction = behaviors.wander(enemy, context)
    end
    return shouldMove, direction
end
```

### 3. Векторы желаний (Steering Behaviors)

**Описание:** Комбинирование векторов желаний для получения финального направления.

**Когда использовать:** Для плавного движения к цели с учетом препятствий.

**Пример:**
```lua
function updateEnemyAI(enemy, context)
    local desired = {x = 0, y = 0}
    
    -- Желание: двигаться к игроку
    if context.player then
        local toPlayer = normalize({
            x = context.player.x - enemy.x,
            y = context.player.y - enemy.y
        })
        desired.x = desired.x + toPlayer.x * 0.8
        desired.y = desired.y + toPlayer.y * 0.8
    end
    
    -- Желание: избегать препятствий
    for _, block in ipairs(context.blocks) do
        local dist = distance(enemy, block)
        if dist < 20 then
            local away = normalize({
                x = enemy.x - block.x,
                y = enemy.y - block.y
            })
            desired.x = desired.x + away.x * 0.5
            desired.y = desired.y + away.y * 0.5
        end
    end
    
    -- Преобразуем вектор в направление
    local direction = vectorToDirection(desired)
    return true, direction
end
```

**Ресурсы:**
- [Steering Behaviors](http://www.red3d.com/cwr/steer/) (EN) - Классическая статья о steering behaviors
- [Programming Game AI by Example - Steering Behaviors](https://www.amazon.com/Programming-Game-Example-Mat-Buckland/dp/1556220782) (EN) - Книга по программированию игрового AI с примерами steering behaviors

---

## ⚡ Оптимизация производительности

### 1. Кэширование вычислений

**Практика:** Кэшируйте дорогие вычисления между вызовами.

**Пример:**
```lua
-- Кэш для расстояний
local distanceCache = {}

function getCachedDistance(pos1, pos2)
    local key = pos1.x .. "," .. pos1.y .. ":" .. pos2.x .. "," .. pos2.y
    if not distanceCache[key] then
        distanceCache[key] = distance(pos1, pos2)
    end
    return distanceCache[key]
end
```

### 2. Пространственное разбиение

**Практика:** Используйте простые структуры данных для быстрого поиска ближайших объектов.

**Пример:**
```lua
-- Вместо перебора всех объектов
function findNearestEnemy(enemy, context)
    local nearest = nil
    local minDist = math.huge
    
    for _, other in ipairs(context.enemies) do
        if other ~= enemy then
            local dist = distance(enemy, other)
            if dist < minDist then
                minDist = dist
                nearest = other
            end
        end
    end
    return nearest
end

-- Оптимизация: передавать только близкие объекты из Go
-- context.nearbyEnemies уже содержит отфильтрованные объекты
```

### 3. Ленивые вычисления

**Практика:** Вычисляйте только то, что действительно нужно.

**Пример:**
```lua
function updateEnemyAI(enemy, context)
    -- Не вычисляем направление к игроку, если игрока нет
    if context.player == nil then
        return randomWander(enemy)
    end
    
    -- Не вычисляем расстояние, если просто проверяем наличие
    local direction = directionToPoint(enemy, context.player)
    return true, direction
end
```

---

## 💾 Работа с состоянием

### 1. Локальное состояние vs Глобальное состояние

**Практика:** Используйте локальное состояние для индивидуального поведения, глобальное — для общих паттернов.

**Пример:**
```lua
-- Глобальное состояние (общее для всех врагов)
local GLOBAL_STATE = {
    playerLastSeen = nil,
    threatLevel = 0
}

-- Локальное состояние (индивидуальное)
local ENEMY_STATE = {}

function updateEnemyAI(enemy, context)
    local state = ENEMY_STATE[enemy.id] or {
        stuckCounter = 0,
        lastDirection = enemy.direction
    }
    
    -- Обновляем локальное состояние
    if enemy.speed == 0 then
        state.stuckCounter = state.stuckCounter + 1
    else
        state.stuckCounter = 0
    end
    
    -- Обновляем глобальное состояние
    if context.player then
        GLOBAL_STATE.playerLastSeen = context.player
    end
    
    ENEMY_STATE[enemy.id] = state
    
    -- Используем состояние для принятия решения
    if state.stuckCounter > 5 then
        return true, randomDirection()
    end
    
    return true, enemy.direction
end
```

### 2. Очистка состояния

**Практика:** Очищайте неиспользуемое состояние для предотвращения утечек памяти.

```lua
function cleanupState(enemyId)
    ENEMY_STATE[enemyId] = nil
end
```

---

## 🐛 Отладка и тестирование

### 1. Логирование

**Практика:** Добавьте возможность логировать решения AI для отладки.

**Пример:**
```lua
function debugLog(enemy, decision, reason)
    -- В Go: логирование через logger
    -- Здесь только формируем данные
    return {
        enemyId = enemy.id,
        decision = decision,
        reason = reason,
        timestamp = os.time()
    }
end

function updateEnemyAI(enemy, context)
    local shouldMove, direction
    
    if enemy.speed == 0 then
        direction = randomDirection()
        shouldMove = true
        debugLog(enemy, direction, "stuck, random direction")
    else
        direction = enemy.direction
        shouldMove = true
        debugLog(enemy, direction, "continue moving")
    end
    
    return shouldMove, direction
end
```

### 2. Визуализация решений

**Практика:** Визуализируйте решения AI в игровом мире (стрелки, линии, маркеры).

**Идеи:**
- Рисовать стрелку направления движения
- Показывать зону обнаружения
- Отображать текущее состояние FSM

### 3. Unit-тестирование

**Практика:** Тестируйте функции AI с известными входными данными.

**Пример (в Go):**
```go
func TestAIDecisionMaking(t *testing.T) {
    // Подготовка
    enemy := createTestEnemy(100, 100, DirectionUp, 0)
    context := createTestContext()
    
    // Выполнение
    decision, err := aiUseCases.ExecuteAI(enemy, context)
    
    // Проверка
    assert.NoError(t, err)
    assert.True(t, decision.Direction >= 0 && decision.Direction <= 3)
}
```

---

## 📖 Ресурсы и ссылки

### Книги

1. **Programming Game AI by Example** - Mat Buckland
   - [Amazon](https://www.amazon.com/Programming-Game-Example-Mat-Buckland/dp/1556220782) (EN) - Классическая книга по игровому AI с практическими примерами

2. **Artificial Intelligence for Games** - Ian Millington
   - [Amazon](https://www.amazon.com/Artificial-Intelligence-Games-Ian-Millington/dp/0123747317) (EN) - Подробное руководство по AI в играх

3. **Game Programming Patterns** - Robert Nystrom
   - [Сайт книги](https://gameprogrammingpatterns.com/) (EN) - Паттерны проектирования для игр, доступно онлайн бесплатно

### Статьи и туториалы

1. **AI-Junkie**
   - [Сайт](https://www.ai-junkie.com/) (EN) - Туториалы по AI для игр, включая FSM, State Machines

2. **Red Blob Games - Pathfinding**
   - [Сайт](https://www.redblobgames.com/pathfinding/a-star/introduction.html) (EN) - Интерактивные туториалы по поиску пути (A*)

3. **Gamasutra - AI Articles**
   - [Сайт](https://www.gamasutra.com/ai/) (EN) - Статьи по игровому AI от разработчиков индустрии

4. **Steering Behaviors**
   - [Статья](http://www.red3d.com/cwr/steer/) (EN) - Классическая статья о steering behaviors от Craig Reynolds

### Онлайн курсы

1. **Coursera - Game AI**
   - Курсы по игровому AI от различных университетов (EN)

2. **Udemy - Game Development**
   - Курсы по разработке игр с AI компонентами (EN/RU)

### Инструменты и библиотеки

1. **Behavior Designer (Unity)**
   - [Asset Store](https://assetstore.unity.com/packages/tools/visual-scripting/behavior-designer-behavior-trees-for-everyone-15277) (EN) - Визуальный редактор для создания поведенческих деревьев

2. **Gopher-Lua**
   - [GitHub](https://github.com/yuin/gopher-lua) (EN) - Библиотека для встраивания Lua в Go

### Паттерны и практики

1. **State Pattern** (Game Programming Patterns)
   - [Статья](https://gameprogrammingpatterns.com/state.html) (EN) - Паттерн состояния для управления поведением

2. **Command Pattern** (Game Programming Patterns)
   - [Статья](https://gameprogrammingpatterns.com/command.html) (EN) - Для записи и воспроизведения действий AI

3. **Strategy Pattern**
   - Для переключения между различными алгоритмами AI

### Русскоязычные ресурсы

1. **Хабр - Игровой AI**
   - [Хабр](https://habr.com/ru/hub/gamedev/) (RU) - Статьи по разработке игр, включая AI

2. **GameDev.ru**
   - [Форум](https://www.gamedev.ru/) (RU) - Форум и статьи по разработке игр

3. **GameDev.Media**
   - [Сайт](https://gamedev.media/) (RU) - Ресурсы по разработке игр на русском языке

4. **YouTube каналы**
   - [Борис Дворников - Game Development](https://www.youtube.com/c/BorisDvornikov) (RU) - Туториалы по разработке игр и AI
   - [Геймдев - Индустрия игр](https://www.youtube.com/@gamedevindustry) (RU) - Образовательный контент по геймдеву

---

## 🎯 Рекомендации для Gonflict

### Текущая реализация

- ✅ **Правильная архитектура:** Разделение на слои (LuaEngine, AITypeConverter, AIUseCases)
- ✅ **Минимальный API:** Передаются только необходимые данные
- ✅ **Чистый код:** Простые и понятные Lua скрипты

### Рекомендуемые улучшения

1. **Добавить вспомогательные функции** в Lua engine:
   - `distance()`, `directionToPoint()`, `normalize()`
   
2. **Расширить контекст** дополнительными данными:
   - Позиция базы, размеры карты, метаданные уровня

3. **Реализовать FSM** для сложных врагов:
   - Различные состояния (Idle, Chase, Attack, Flee)

4. **Добавить визуализацию** для отладки:
   - Стрелки направления, зоны обнаружения

5. **Оптимизировать передачу данных**:
   - Передавать только близкие объекты вместо всех

---

## ✅ Чеклист при разработке AI

- [ ] AI функция идемпотентна (или предсказуема)
- [ ] Минимальный API — только необходимые данные
- [ ] Кэширование дорогих вычислений
- [ ] Логирование решений для отладки
- [ ] Тестирование с различными входными данными
- [ ] Документация поведения AI
- [ ] Очистка неиспользуемого состояния

---

**Примечание:** Эти практики применимы для простых и средних игр. Для сложных AAA-игр могут потребоваться более продвинутые техники (машинное обучение, нейронные сети, планирование поведения).

