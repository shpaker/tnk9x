-- AI для вражеских танков (стиль NES Battle City)
-- Получает данные врага и контекст игры, возвращает решение о движении

-- Вспомогательные функции
function randomDirection()
    -- Возвращает случайное направление (0=up, 1=down, 2=left, 3=right)
    return math.random(0, 3)
end

-- Основная функция AI
function updateEnemyAI(enemy, context)
    local shouldMove = false
    local newDirection = enemy.direction

    -- Если танк остановился (speed == 0) - это значит он столкнулся с препятствием
    if enemy.speed == 0 then
        shouldMove = true

        -- Выбираем новое случайное направление
        -- Это создает эффект "блуждающего" танка как в оригинальной игре
        newDirection = randomDirection()
    else
        -- Танк движется - продолжаем двигаться в том же направлении
        -- Логика: не меняем направление пока танк движется
        -- Направление изменится только когда танк остановится (столкнется)
        newDirection = enemy.direction
        shouldMove = true
    end

    return shouldMove, newDirection
end
