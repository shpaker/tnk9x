-- AI для вражеских танков
-- В одном из 4 случаев запускает рандомный выбор направления
-- Не позволяет развернуться в сторону края карты

-- Константы направлений
DIRECTION_UP = 0
DIRECTION_DOWN = 1
DIRECTION_LEFT = 2
DIRECTION_RIGHT = 3

-- Проверяет, близок ли танк к краю карты в указанном направлении
function isNearEdge(x, y, direction)
    -- Размеры карты в пикселях инжектируются движком
    local mapWidthPx = MAP_WIDTH_PX
    local mapHeightPx = MAP_HEIGHT_PX

    if direction == DIRECTION_UP then
        -- Проверяем верхний край: если Y близок к 0 (меньше размера танка)
        return y <= 0
    elseif direction == DIRECTION_DOWN then
        -- Проверяем нижний край: если Y + размер танка >= высота карты
        return y + TANK_SIZE_PX >= mapHeightPx
    elseif direction == DIRECTION_LEFT then
        -- Проверяем левый край: если X близок к 0 (меньше размера танка)
        return x <= 0
    elseif direction == DIRECTION_RIGHT then
        -- Проверяем правый край: если X + размер танка >= ширина карты
        return x + TANK_SIZE_PX >= mapWidthPx
    end
    return false
end

-- Возвращает обратное направление для заданного
function getOppositeDirection(direction)
    if direction == DIRECTION_UP then
        return DIRECTION_DOWN
    elseif direction == DIRECTION_DOWN then
        return DIRECTION_UP
    elseif direction == DIRECTION_LEFT then
        return DIRECTION_RIGHT
    elseif direction == DIRECTION_RIGHT then
        return DIRECTION_LEFT
    end
    return direction
end

-- Возвращает боковые направления для заданного (перпендикулярные)
function getSideDirections(direction)
    if direction == DIRECTION_UP or direction == DIRECTION_DOWN then
        return {DIRECTION_LEFT, DIRECTION_RIGHT}
    elseif direction == DIRECTION_LEFT or direction == DIRECTION_RIGHT then
        return {DIRECTION_UP, DIRECTION_DOWN}
    end
    return {}
end

-- Возвращает список разрешенных направлений (исключая направления к краю)
function getAllowedDirections(x, y)
    local allowed = {}

    if not isNearEdge(x, y, DIRECTION_UP) then
        table.insert(allowed, DIRECTION_UP)
    end
    if not isNearEdge(x, y, DIRECTION_DOWN) then
        table.insert(allowed, DIRECTION_DOWN)
    end
    if not isNearEdge(x, y, DIRECTION_LEFT) then
        table.insert(allowed, DIRECTION_LEFT)
    end
    if not isNearEdge(x, y, DIRECTION_RIGHT) then
        table.insert(allowed, DIRECTION_RIGHT)
    end

    -- Если все направления запрещены (маловероятно, но на всякий случай)
    if #allowed == 0 then
        return {DIRECTION_UP, DIRECTION_DOWN, DIRECTION_LEFT, DIRECTION_RIGHT}
    end

    return allowed
end

-- Проверяет, есть ли направление в списке разрешенных
function isDirectionAllowed(direction, allowed)
    for i = 1, #allowed do
        if allowed[i] == direction then
            return true
        end
    end
    return false
end

-- Выбирает случайное направление с учетом вероятностей:
-- обратное направление в 2 раза реже, чем боковое
function randomAllowedDirection(x, y, currentDirection)
    local allowed = getAllowedDirections(x, y)

    -- Получаем обратное и боковые направления
    local opposite = getOppositeDirection(currentDirection)
    local sides = getSideDirections(currentDirection)

    -- Фильтруем боковые направления, оставляя только разрешенные
    local allowedSides = {}
    for i = 1, #sides do
        if isDirectionAllowed(sides[i], allowed) then
            table.insert(allowedSides, sides[i])
        end
    end

    -- Проверяем, разрешено ли обратное направление
    local oppositeAllowed = isDirectionAllowed(opposite, allowed)

    -- Если есть боковые направления, выбираем с вероятностью 2/3 боковое, 1/3 обратное
    if #allowedSides > 0 and oppositeAllowed then
        -- 1 из 3 случаев = обратное, 2 из 3 = боковое
        if math.random(1, 3) == 1 then
            return opposite
        else
            return allowedSides[math.random(1, #allowedSides)]
        end
    elseif #allowedSides > 0 then
        -- Если обратное не разрешено, выбираем только из боковых
        return allowedSides[math.random(1, #allowedSides)]
    elseif oppositeAllowed then
        -- Если боковых нет, но обратное разрешено
        return opposite
    else
        -- Если ничего не подходит, выбираем из всех разрешенных
        return allowed[math.random(1, #allowed)]
    end
end

-- Возвращает направления, ведущие к цели
function directionsToward(x, y, targetX, targetY)
    local toward = {}
    if targetY < y then
        table.insert(toward, DIRECTION_UP)
    end
    if targetY > y then
        table.insert(toward, DIRECTION_DOWN)
    end
    if targetX < x then
        table.insert(toward, DIRECTION_LEFT)
    end
    if targetX > x then
        table.insert(toward, DIRECTION_RIGHT)
    end
    return toward
end

-- Выбор направления со смещением к цели: в половине случаев танк
-- поворачивает к цели, иначе — обычный случайный выбор
function biasedDirection(x, y, currentDirection, targetX, targetY)
    if math.random(1, 2) == 1 then
        local allowed = getAllowedDirections(x, y)
        local toward = directionsToward(x, y, targetX, targetY)

        local candidates = {}
        for i = 1, #toward do
            if isDirectionAllowed(toward[i], allowed) then
                table.insert(candidates, toward[i])
            end
        end

        if #candidates > 0 then
            return candidates[math.random(1, #candidates)]
        end
    end

    return randomAllowedDirection(x, y, currentDirection)
end

-- Основная функция AI
-- Фазы этапа, как в оригинале: 1 — свободный обход со смещением вниз,
-- 2 — охота на игрока, 3 — атака штаба (цель передаёт движок)
function updateEnemyAI(
    x,         -- Позиция X танка
    y,         -- Позиция Y танка
    direction, -- Направление танка
    state,     -- Состояние танка
    phase,     -- Фаза AI (1-3)
    targetX,   -- Цель X (игрок или штаб)
    targetY,   -- Цель Y
    hasTarget  -- Цель задана движком
)
    -- Без цели смещаемся к нижней части карты (к базе игроков)
    local goalX, goalY = targetX, targetY
    if not hasTarget then
        goalX, goalY = x, MAP_HEIGHT_PX
    end

    -- В одном из 8 случаев выбираем новое направление со смещением к цели
    if math.random(1, 8) == 1 then
        return true, biasedDirection(x, y, direction, goalX, goalY)
    end

    -- Если текущее направление ведет к краю, выбираем другое
    if isNearEdge(x, y, direction) then
        return true, biasedDirection(x, y, direction, goalX, goalY)
    end

    return true, direction
end
