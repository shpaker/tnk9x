-- AI для вражеских танков (стиль NES Battle City)
-- Получает данные врага и контекст игры, возвращает решение о движении

-- Константы направлений
DIRECTION_UP = 0
DIRECTION_DOWN = 1
DIRECTION_LEFT = 2
DIRECTION_RIGHT = 3

-- Константы состояний танка
TANK_STATE_SPAWNING = 0
TANK_STATE_MOVING = 1
TANK_STATE_STOPPED = 2
TANK_STATE_BRAKING = 3
TANK_STATE_EXPLODING = 4
TANK_STATE_EXPLODED = 5

-- Вспомогательные функции
function randomDirection()
    -- Возвращает случайное направление используя константы
    local directions = {DIRECTION_UP, DIRECTION_DOWN, DIRECTION_LEFT, DIRECTION_RIGHT}
    return directions[math.random(1, 4)]
end

-- Основная функция AI
-- Параметры: x, y, direction, state, context
function updateEnemyAI(
    x,        -- Позиция X танка
    y,        -- Позиция Y танка
    direction, -- Направление танка (DIRECTION_UP, DIRECTION_DOWN, DIRECTION_LEFT, DIRECTION_RIGHT)
    state,    -- Состояние танка (TANK_STATE_MOVING, TANK_STATE_STOPPED и т.д.)
    context   -- Контекст игры (игрок, враги, пули, блоки)
)
    local shouldMove = false
    local newDirection = direction

    -- Если танк остановился - это значит он столкнулся с препятствием
    if state == TANK_STATE_STOPPED then
        shouldMove = true

        -- Выбираем новое случайное направление
        -- Это создает эффект "блуждающего" танка как в оригинальной игре
        newDirection = randomDirection()
    elseif state == TANK_STATE_MOVING then
        -- Танк движется - продолжаем двигаться в том же направлении
        -- Логика: не меняем направление пока танк движется
        -- Направление изменится только когда танк остановится (столкнется)
        newDirection = direction
        shouldMove = true
    end

    return shouldMove, newDirection
end
