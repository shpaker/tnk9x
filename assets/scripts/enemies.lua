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

-- Основная функция AI
-- Параметры: x, y, direction, state
function updateEnemyAI(
    x,        -- Позиция X танка
    y,        -- Позиция Y танка
    direction, -- Направление танка
    state     -- Состояние танка
)
    -- В одном из 8 случаев выбираем случайное направление из разрешенных
    -- Обратное направление в 2 раза реже, чем боковое
    if math.random(1, 8) == 1 then
        local newDirection = randomAllowedDirection(x, y, direction)
        return true, newDirection
    end

    -- Иначе возвращаем текущее направление (если оно разрешено)
    -- Если текущее направление ведет к краю, выбираем другое разрешенное
    if isNearEdge(x, y, direction) then
        local newDirection = randomAllowedDirection(x, y, direction)
        return true, newDirection
    end

    return true, direction
end
