# Удобная реализация AI для Lua файлов

Документ описывает, как должен быть устроен AI в Lua-скриптах для максимального удобства разработки и использования.

## 🎯 Принципы проектирования

### 1. Простота и читаемость
Lua-скрипты должны быть простыми для понимания и модификации, даже для разработчиков без глубокого знания Lua.

### 2. Минимальный API
Предоставлять только необходимые данные и функции, избегать избыточной сложности.

### 3. Естественный интерфейс
API должен интуитивно понятным и следовать идиомам Lua.

---

## 📦 Текущая реализация

### Структура данных

#### Входные данные (enemy, context)

```lua
-- enemy (танк врага)
enemy = {
    x = 100.0,           -- Позиция X
    y = 200.0,           -- Позиция Y
    direction = 0,       -- Направление (0=up, 1=down, 2=left, 3=right)
    speed = 2.0          -- Скорость движения
}

-- context (контекст игры)
context = {
    player = {           -- Игрок (или nil)
        x = 150.0,
        y = 300.0,
        direction = 2,
        speed = 0.0
    },
    enemies = {},       -- Массив врагов (пока не используется)
    bullets = {},       -- Массив пуль (пока не используется)
    blocks = {}         -- Массив блоков (пока не используется)
}
```

#### Выходные данные

```lua
-- Возвращаемые значения
shouldMove = true       -- Двигаться ли танку
newDirection = 2        -- Новое направление (0-3)
```

### Пример использования

```lua
function updateEnemyAI(enemy, context)
    local shouldMove = false
    local newDirection = enemy.direction

    -- Если танк остановился - выбираем новое направление
    if enemy.speed == 0 then
        shouldMove = true
        newDirection = math.random(0, 3)
    else
        -- Продолжаем движение в том же направлении
        shouldMove = true
        newDirection = enemy.direction
    end

    return shouldMove, newDirection
end
```

---

## 🚀 Рекомендуемые улучшения

### 1. Расширенный контекст

Добавить больше полезных данных в контекст:

```lua
context = {
    player = {...},
    enemies = {...},
    bullets = {...},
    blocks = {...},
    
    -- Дополнительные данные
    mapWidth = 104,      -- Ширина карты в пикселях
    mapHeight = 104,     -- Высота карты в пикселях
    basePosition = {     -- Позиция базы
        x = 52,
        y = 100
    },
    currentTime = 1234.5 -- Время игры в секундах
}
```

### 2. Вспомогательные функции в глобальной области

Предоставить готовые функции для типичных задач:

```lua
-- Математические функции
function distance(pos1, pos2)
    local dx = pos1.x - pos2.x
    local dy = pos1.y - pos2.y
    return math.sqrt(dx * dx + dy * dy)
end

function directionToPoint(from, to)
    local dx = to.x - from.x
    local dy = to.y - from.y
    
    if math.abs(dx) > math.abs(dy) then
        return dx > 0 and 3 or 2  -- right or left
    else
        return dy > 0 and 1 or 0  -- down or up
    end
end

-- Работа с направлениями
function isOppositeDirection(dir1, dir2)
    return (dir1 == 0 and dir2 == 1) or
           (dir1 == 1 and dir2 == 0) or
           (dir1 == 2 and dir2 == 3) or
           (dir1 == 3 and dir2 == 2)
end

function rotateDirection(dir, clockwise)
    if clockwise then
        return (dir + 1) % 4
    else
        return (dir + 3) % 4
    end
end

-- Работа с картой
function isPositionValid(x, y, mapWidth, mapHeight)
    return x >= 0 and x < mapWidth and y >= 0 and y < mapHeight
end

function getNearestBlock(pos, blocks)
    local nearest = nil
    local minDist = math.huge
    
    for _, block in ipairs(blocks) do
        local dist = distance(pos, block)
        if dist < minDist then
            minDist = dist
            nearest = block
        end
    end
    
    return nearest, minDist
end
```

### 3. Типизированные данные для блоков и пуль

```lua
-- block структура
block = {
    x = 50.0,
    y = 50.0,
    type = "brick",      -- "brick", "steel", "forest", "water", "ice"
    isDestroyable = true,
    isPassable = false
}

-- bullet структура
bullet = {
    x = 75.0,
    y = 75.0,
    direction = 0,
    owner = "player"    -- "player" или "enemy"
}
```

### 4. Состояние AI между вызовами

Добавить возможность хранить состояние AI между вызовами:

```lua
-- Глобальная таблица состояний (индексируется по ID танка)
AI_STATE = AI_STATE or {}

function updateEnemyAI(enemy, context)
    local state = AI_STATE[enemy.id] or {
        lastDirection = enemy.direction,
        stuckCounter = 0,
        targetDirection = nil
    }
    
    -- Логика AI с использованием state
    
    AI_STATE[enemy.id] = state
    return shouldMove, newDirection
end
```

### 5. Предопределенные стратегии

Создать библиотеку готовых стратегий:

```lua
-- Стратегии поведения
STRATEGIES = {
    -- Случайное блуждание
    randomWander = function(enemy, context)
        if enemy.speed == 0 then
            return true, math.random(0, 3)
        end
        return true, enemy.direction
    end,
    
    -- Преследование игрока
    chasePlayer = function(enemy, context)
        if context.player == nil then
            return STRATEGIES.randomWander(enemy, context)
        end
        
        local targetDir = directionToPoint(
            {x = enemy.x, y = enemy.y},
            {x = context.player.x, y = context.player.y}
        )
        
        return true, targetDir
    end,
    
    -- Защита базы
    defendBase = function(enemy, context)
        if context.basePosition == nil then
            return STRATEGIES.randomWander(enemy, context)
        end
        
        -- Движение к базе
        local targetDir = directionToPoint(
            {x = enemy.x, y = enemy.y},
            context.basePosition
        )
        
        return true, targetDir
    end
}

-- Использование стратегии
function updateEnemyAI(enemy, context)
    return STRATEGIES.chasePlayer(enemy, context)
end
```

### 6. Упрощенный API для получения объектов

```lua
-- Удобные функции для работы с контекстом
function getPlayer(context)
    return context.player
end

function getEnemiesInRange(context, centerX, centerY, range)
    local result = {}
    for _, enemy in ipairs(context.enemies) do
        local dist = distance(
            {x = centerX, y = centerY},
            {x = enemy.x, y = enemy.y}
        )
        if dist <= range then
            table.insert(result, enemy)
        end
    end
    return result
end

function getBulletsInRange(context, centerX, centerY, range)
    local result = {}
    for _, bullet in ipairs(context.bullets) do
        local dist = distance(
            {x = centerX, y = centerY},
            {x = bullet.x, y = bullet.y}
        )
        if dist <= range then
            table.insert(result, bullet)
        end
    end
    return result
end

function getBlocksInRange(context, centerX, centerY, range)
    local result = {}
    for _, block in ipairs(context.blocks) do
        local dist = distance(
            {x = centerX, y = centerY},
            {x = block.x, y = block.y}
        )
        if dist <= range then
            table.insert(result, block)
        end
    end
    return result
end
```

---

## 📋 Итоговая структура AI скрипта

```lua
-- ============================================================================
-- AI для вражеских танков (Battle City стиль)
-- ============================================================================

-- Вспомогательные функции (предоставляются движком)
-- distance(pos1, pos2) - расстояние между точками
-- directionToPoint(from, to) - направление к точке
-- isPositionValid(x, y) - валидна ли позиция на карте

-- Глобальные константы
local DIRECTIONS = {
    UP = 0,
    DOWN = 1,
    LEFT = 2,
    RIGHT = 3
}

-- Состояние AI (опционально)
local aiState = {}

-- Основная функция AI
function updateEnemyAI(enemy, context)
    local shouldMove = false
    local newDirection = enemy.direction
    
    -- Простая логика: если остановился - выбираем случайное направление
    if enemy.speed == 0 then
        shouldMove = true
        newDirection = math.random(0, 3)
    else
        shouldMove = true
        newDirection = enemy.direction
    end
    
    -- TODO: Добавить более сложную логику:
    -- - Преследование игрока
    -- - Избегание препятствий
    -- - Стрельба при обнаружении цели
    
    return shouldMove, newDirection
end
```

---

## ✅ Рекомендации по реализации

1. **Добавить вспомогательные функции** в Lua engine перед загрузкой скрипта
2. **Расширить контекст** дополнительными полезными данными
3. **Документировать API** в отдельном файле или комментариях
4. **Создать библиотеку стратегий** для переиспользования
5. **Добавить отладку** - возможность логировать данные AI в консоль

---

## 🔗 Связанные документы

- [AI Best Practices](./ai_best_practices.md) - лучшие практики реализации AI
- [Architecture](./architecture_deficiencies.md) - архитектура проекта

