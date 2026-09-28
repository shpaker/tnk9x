-- AI вражеских танков
--
-- Вся логика поведения живёт здесь; движок только передаёт снимок мира,
-- отвечает на запросы таблицы ai и исполняет решение. Скрипт загружается
-- заново на каждом уровне, поэтому память танков сбрасывается.
--
-- Движок вызывает updateEnemyAI(ctx), когда танк стоит: после каждой
-- клетки пути, после столкновения и каждый тик, пока танк стоит на месте.
--
-- ctx:
--   tick     — номер тика AI
--   stage    — номер уровня
--   self     — {id, x, y, size, dir, level, hp, bonus, reinforced}
--              level: 0 Basic, 1 Fast, 2 Power, 3 Armor
--   players  — активные игроки: {x, y, size, dir, ...}
--   enemies  — союзники без самого танка
--   bullets  — {x, y, w, h, dir, enemy}
--   hq       — {x, y, size, intact}
--
-- ai (запросы к движку, решений не принимают):
--   ai.findPath(fromX, fromY, toX, toY [, {brickCost=, steelPassable=}])
--     -> направление первого шага и длина пути в шагах, либо nil
--   ai.castRay(x, y, dir) -> "edge"|"brick"|"steel"|"player"|"enemy"|"hq", расстояние
--   ai.tileAt(x, y) -> "brick"|"steel"|"water"|"forest"|"ice" или nil
--
-- Глобальные параметры карты: MAP_WIDTH_PX, MAP_HEIGHT_PX, TANK_SIZE_PX.
--
-- Ответ: {direction=, move=, shoot=}. Без direction танк сохраняет
-- направление, без move стоит, без shoot не стреляет.

UP, DOWN, LEFT, RIGHT = 0, 1, 2, 3
BASIC, FAST, POWER, ARMOR = 0, 1, 2, 3

local OPPOSITE = { [UP] = DOWN, [DOWN] = UP, [LEFT] = RIGHT, [RIGHT] = LEFT }
local SIDES = {
    [UP] = { LEFT, RIGHT },
    [DOWN] = { LEFT, RIGHT },
    [LEFT] = { UP, DOWN },
    [RIGHT] = { UP, DOWN },
}

-- Застревание: столько решений подряд без сдвига при попытке ехать
local STUCK_LIMIT = 12
-- Максимум тиков на пробивание одной стены
local BREACH_LIMIT = 150
-- Дистанция, с которой быстрый танк замечает летящую в него пулю
local DODGE_RANGE = 72
-- Минимальный интервал фоновой стрельбы в тиках: наугад и по стенам
local AMBIENT_COOLDOWN = 30
local DEMOLITION_COOLDOWN = 20
-- Дальность, с которой танк обстреливает стену по курсу
local DEMOLITION_RANGE = 80

-- Память танков по id; сбрасывается при загрузке скрипта
memory = {}

-- Характеры по типу танка. targetRule — чередование целей:
--   nes     — бродит, потом охотится на игрока, потом идёт к штабу
--   player  — бродит, потом охотится на игрока
--   nearest — бродит, потом давит на ближайшую цель: игрока или штаб
-- Разрушение окружения:
--   brickCost  — доплата за кирпич в поиске пути (меньше — чаще напролом)
--   roamBreach — шанс пробить кирпич перед собой, когда бродит
--   brickFire  — шанс на ходу обстрелять стену по курсу
--   sideBreach — шанс развернуться и прорубить проход в боковой стене
PROFILES = {
    [BASIC] = {
        sightMul = 0.9,
        targetRule = "nes",
        roamTicks = 480,
        huntTicks = 600,
        brickCost = 2,
        reactionMul = 1.4,
        aimMul = 0.6,
        holdToAim = false,
        wanderMul = 1.5,
        ambientFire = 0.18,
        roamBreach = 0.85,
        brickFire = 0.5,
        sideBreach = 0.15,
        hqOpportunist = false,
        dodge = false,
    },
    [FAST] = {
        sightMul = 1.1,
        targetRule = "player",
        roamTicks = 150,
        huntTicks = 0,
        brickCost = 3,
        reactionMul = 1.0,
        aimMul = 1.0,
        holdToAim = false,
        wanderMul = 0.8,
        ambientFire = 0.08,
        roamBreach = 0.6,
        brickFire = 0.3,
        sideBreach = 0.05,
        hqOpportunist = false,
        dodge = true,
    },
    [POWER] = {
        sightMul = 1.0,
        targetRule = "nes",
        roamTicks = 180,
        huntTicks = 360,
        brickCost = 0,
        reactionMul = 0.9,
        aimMul = 0.8,
        holdToAim = false,
        wanderMul = 0.7,
        ambientFire = 0.15,
        roamBreach = 1.0,
        brickFire = 0.75,
        sideBreach = 0.25,
        hqOpportunist = true,
        dodge = false,
    },
    [ARMOR] = {
        sightMul = 1.2,
        targetRule = "nearest",
        roamTicks = 240,
        huntTicks = 0,
        brickCost = 1,
        reactionMul = 0.8,
        aimMul = 1.2,
        holdToAim = true,
        wanderMul = 0.5,
        ambientFire = 0.08,
        roamBreach = 0.8,
        brickFire = 0.45,
        sideBreach = 0.15,
        hqOpportunist = true,
        dodge = false,
    },
}

-- Общие утилиты

local function clamp(value, low, high)
    return math.max(low, math.min(high, value))
end

local function lerp(from, to, t)
    return from + (to - from) * t
end

local function center(tank)
    local half = tank.size / 2
    return tank.x + half, tank.y + half
end

local function manhattan(ax, ay, bx, by)
    return math.abs(ax - bx) + math.abs(ay - by)
end

local function decide(direction, move, shoot)
    return { direction = direction, move = move, shoot = shoot }
end

-- Сложность растёт с номером уровня и выходит на максимум к 16-му
function difficulty(stage)
    local t = clamp(((stage or 1) - 1) / 15, 0, 1)
    local dodgeChance = 0
    if (stage or 1) >= 5 then
        dodgeChance = lerp(0.35, 0.75, t)
    end
    return {
        reaction = lerp(60, 15, t),      -- тики от обнаружения цели до выстрела
        aimCooldown = lerp(120, 35, t),  -- тики между прицельными выстрелами
        sightRange = lerp(96, 224, t),   -- дальность, на которой замечает игрока
        wander = lerp(0.20, 0.08, t),    -- шанс свернуть с пути
        phaseMul = lerp(1.5, 0.6, t),    -- множитель длительности фаз целей
        aimChance = lerp(0.25, 0.9, t),  -- шанс довернуть ствол к цели
        dodgeChance = dodgeChance,
        ambientMul = lerp(0.8, 1.2, t),
    }
end

-- Окружение танка

-- Что стоит вплотную перед танком в направлении dir:
-- "edge", "steel", "water", "brick", "tank" или nil, если проезд свободен
local function aheadBlock(ctx, tank, dir)
    local size = tank.size
    local points = {}
    for offset = 2, size - 2, 4 do
        if dir == UP then
            points[#points + 1] = { tank.x + offset, tank.y - 2 }
        elseif dir == DOWN then
            points[#points + 1] = { tank.x + offset, tank.y + size + 2 }
        elseif dir == LEFT then
            points[#points + 1] = { tank.x - 2, tank.y + offset }
        else
            points[#points + 1] = { tank.x + size + 2, tank.y + offset }
        end
    end

    local found = nil
    for _, point in ipairs(points) do
        local x, y = point[1], point[2]
        if x < 0 or y < 0 or x >= MAP_WIDTH_PX or y >= MAP_HEIGHT_PX then
            return "edge"
        end
        local tile = ai.tileAt(x, y)
        if tile == "steel" or tile == "water" then
            return tile
        end
        if tile == "brick" then
            found = "brick"
        end
    end
    if found then
        return found
    end

    -- Другие танки перед носом
    local left, top, right, bottom = points[1][1], points[1][2], points[#points][1], points[#points][2]
    local function overlaps(other)
        return left < other.x + other.size and right >= other.x
            and top < other.y + other.size and bottom >= other.y
    end
    for _, other in ipairs(ctx.enemies) do
        if overlaps(other) then
            return "tank"
        end
    end
    for _, other in ipairs(ctx.players) do
        if overlaps(other) then
            return "tank"
        end
    end
    return nil
end

-- Препятствие перед танком лежит на линии выстрела и пробивается
local function canBreach(tank, dir)
    local cx, cy = center(tank)
    local kind, distance = ai.castRay(cx, cy, dir)
    local breakable = kind == "brick" or (kind == "steel" and tank.reinforced)
    return breakable and distance <= tank.size / 2 + 4
end

-- Свободное направление, кроме current: боковые вдвое вероятнее обратного
local function randomTurn(ctx, tank, current)
    local candidates = {}
    for _, side in ipairs(SIDES[current]) do
        if aheadBlock(ctx, tank, side) == nil then
            candidates[#candidates + 1] = side
            candidates[#candidates + 1] = side
        end
    end
    local opposite = OPPOSITE[current]
    if aheadBlock(ctx, tank, opposite) == nil then
        candidates[#candidates + 1] = opposite
    end
    if #candidates == 0 then
        return nil
    end
    return candidates[math.random(1, #candidates)]
end

-- Память и цели

local function getMemory(ctx, profile, diff)
    local tank = ctx.self
    local m = memory[tank.id]
    if m then
        return m
    end

    -- Разброс фаз, чтобы танки одного типа не ходили строем
    local function jitter()
        return 0.7 + math.random() * 0.6
    end
    local roamUntil = ctx.tick + profile.roamTicks * diff.phaseMul * jitter()
    m = {
        roamUntil = roamUntil,
        huntUntil = roamUntil + profile.huntTicks * diff.phaseMul * jitter(),
        stuck = 0,
        detourUntil = 0,
        lastShot = -1000,
        lastAimedShot = -1000,
        nextDodgeCheck = 0,
        dodgeUntil = 0,
    }
    memory[tank.id] = m
    return m
end

local function nearestPlayer(ctx, tank)
    local best, bestDistance = nil, math.huge
    for _, player in ipairs(ctx.players) do
        local distance = manhattan(tank.x, tank.y, player.x, player.y)
        if distance < bestDistance then
            best, bestDistance = player, distance
        end
    end
    return best, bestDistance
end

-- Текущая фаза танка: "roam", "player" или "hq"
local function currentMode(ctx, m, profile)
    if ctx.tick < m.roamUntil then
        return "roam"
    end

    local hasPlayer = #ctx.players > 0
    local hqAlive = ctx.hq.intact
    local mode = "hq"
    if profile.targetRule == "player" then
        mode = "player"
    elseif profile.targetRule == "nes" then
        if ctx.tick < m.huntUntil then
            mode = "player"
        end
    elseif profile.targetRule == "nearest" then
        local _, playerDistance = nearestPlayer(ctx, ctx.self)
        local hqDistance = manhattan(ctx.self.x, ctx.self.y, ctx.hq.x, ctx.hq.y)
        if playerDistance < hqDistance then
            mode = "player"
        end
    end

    if mode == "player" and not hasPlayer then
        mode = "hq"
    end
    if mode == "hq" and not hqAlive then
        mode = hasPlayer and "player" or "roam"
    end
    return mode
end

-- Поведения. Каждое возвращает решение или nil, если не применимо

local function startBreach(m, dir, tick)
    m.breachDir = dir
    m.breachStart = tick
    return decide(dir, false, true)
end

local function startDetour(m, tick)
    m.detourUntil = tick + math.random(40, 90)
    m.forceTurn = true
    m.breachDir = nil
end

-- Бродит как в оригинале: едет прямо, изредка сворачивает,
-- иногда пробивает кирпич перед собой
local function roam(ctx, tank, m, profile)
    local forward = aheadBlock(ctx, tank, tank.dir)
    if not m.forceTurn then
        -- Прорубает проход в стене сбоку
        if forward == nil and math.random() < profile.sideBreach then
            local side = SIDES[tank.dir][math.random(1, 2)]
            if canBreach(tank, side) then
                return startBreach(m, side, ctx.tick)
            end
        end
        if forward == nil and math.random(1, 8) ~= 1 then
            return decide(tank.dir, true, false)
        end
        if forward == "brick" and canBreach(tank, tank.dir)
            and math.random() < profile.roamBreach then
            return startBreach(m, tank.dir, ctx.tick)
        end
    end
    m.forceTurn = false

    local turn = randomTurn(ctx, tank, tank.dir)
    if turn then
        return decide(turn, true, false)
    end
    if forward == nil then
        return decide(tank.dir, true, false)
    end
    return decide(OPPOSITE[tank.dir], true, false)
end

-- Продолжает пробивать стену, пока она на линии выстрела
local function continueBreach(ctx, tank, m)
    if not m.breachDir then
        return nil
    end
    if ctx.tick - m.breachStart > BREACH_LIMIT or not canBreach(tank, m.breachDir) then
        m.breachDir = nil
        return nil
    end
    return decide(m.breachDir, false, true)
end

-- Направление на цель, если танк стоит с ней на одной линии
local function alignment(tank, target)
    local cx, cy = center(tank)
    local half = target.size / 2
    local tx, ty = target.x + half, target.y + half
    if math.abs(tx - cx) <= half then
        return ty < cy and UP or DOWN
    end
    if math.abs(ty - cy) <= half then
        return tx < cx and LEFT or RIGHT
    end
    return nil
end

-- Ищет цель на линии огня: игрока в пределах дальности обзора или штаб
local function findShot(ctx, tank, mode, profile, diff)
    local cx, cy = center(tank)
    local sight = diff.sightRange * profile.sightMul
    for _, player in ipairs(ctx.players) do
        local dir = alignment(tank, player)
        if dir then
            local kind, distance = ai.castRay(cx, cy, dir)
            if kind == "player" and distance <= sight then
                return { dir = dir }
            end
        end
    end

    local hq = ctx.hq
    if hq.intact and (mode == "hq" or profile.hqOpportunist) then
        local dir = alignment(tank, hq)
        if dir then
            local kind = ai.castRay(cx, cy, dir)
            if kind == "hq" then
                return { dir = dir }
            end
            -- Идя на штаб, пробивает стену между собой и штабом
            local breakable = kind == "brick" or (kind == "steel" and tank.reinforced)
            if mode == "hq" and breakable then
                return { dir = dir, wall = true }
            end
        end
    end
    return nil
end

-- Прицеливание: доворачивает ствол к цели на линии и стреляет
-- после задержки реакции, выдерживая паузу между прицельными выстрелами
local function aim(ctx, tank, m, profile, diff, mode)
    local shot = findShot(ctx, tank, mode, profile, diff)
    if not shot then
        m.seenSince = nil
        return nil
    end

    if not m.seenSince then
        m.seenSince = ctx.tick
        m.willAim = math.random() < diff.aimChance * profile.aimMul
    end

    if shot.dir ~= tank.dir then
        if not m.willAim then
            return nil
        end
        return decide(shot.dir, not profile.holdToAim, false)
    end

    local move = not profile.holdToAim and aheadBlock(ctx, tank, shot.dir) == nil
    if shot.wall then
        return decide(shot.dir, move, true)
    end

    local reacted = ctx.tick - m.seenSince >= diff.reaction * profile.reactionMul
    local rested = ctx.tick - m.lastAimedShot >= diff.aimCooldown
    local decision = decide(shot.dir, move, reacted and rested)
    decision.aimed = decision.shoot
    return decision
end

-- Пуля игрока, летящая в танк
local function incomingBullet(ctx, tank)
    local cx, cy = center(tank)
    local half = tank.size / 2
    for _, bullet in ipairs(ctx.bullets) do
        if not bullet.enemy then
            local bx, by = bullet.x + bullet.w / 2, bullet.y + bullet.h / 2
            local onColumn = math.abs(bx - cx) < half + bullet.w / 2
            local onRow = math.abs(by - cy) < half + bullet.h / 2
            if (bullet.dir == DOWN and onColumn and by < cy and cy - by < DODGE_RANGE)
                or (bullet.dir == UP and onColumn and by > cy and by - cy < DODGE_RANGE)
                or (bullet.dir == RIGHT and onRow and bx < cx and cx - bx < DODGE_RANGE)
                or (bullet.dir == LEFT and onRow and bx > cx and bx - cx < DODGE_RANGE) then
                return bullet
            end
        end
    end
    return nil
end

-- Уклонение: встречный выстрел, если ствол смотрит на пулю, иначе шаг в сторону
local function dodge(ctx, tank, m, profile, diff)
    if not profile.dodge or diff.dodgeChance <= 0 then
        return nil
    end

    if ctx.tick < m.dodgeUntil and m.dodgeDir then
        if aheadBlock(ctx, tank, m.dodgeDir) == nil then
            return decide(m.dodgeDir, true, false)
        end
        m.dodgeDir = nil
    end

    if ctx.tick < m.nextDodgeCheck then
        return nil
    end
    local bullet = incomingBullet(ctx, tank)
    if not bullet then
        return nil
    end
    m.nextDodgeCheck = ctx.tick + 20
    if math.random() >= diff.dodgeChance then
        return nil
    end

    if tank.dir == OPPOSITE[bullet.dir] then
        return decide(tank.dir, false, true)
    end

    local sides = SIDES[bullet.dir]
    local first = math.random(1, 2)
    for i = 0, 1 do
        local side = sides[(first + i - 1) % 2 + 1]
        if aheadBlock(ctx, tank, side) == nil then
            m.dodgeDir = side
            m.dodgeUntil = ctx.tick + 16
            return decide(side, true, false)
        end
    end
    return nil
end

-- Движение к цели по пути с долей случайности
local function navigate(ctx, tank, m, profile, diff, mode)
    if mode == "roam" or ctx.tick < m.detourUntil then
        return roam(ctx, tank, m, profile)
    end

    local target = ctx.hq
    if mode == "player" then
        target = nearestPlayer(ctx, tank)
    end

    local dir, length = ai.findPath(tank.x, tank.y, target.x, target.y, {
        brickCost = profile.brickCost,
        steelPassable = tank.reinforced,
    })
    if not dir or length == 0 then
        return roam(ctx, tank, m, profile)
    end

    if math.random() < diff.wander * profile.wanderMul then
        local turn = randomTurn(ctx, tank, dir)
        if turn and turn ~= OPPOSITE[dir] then
            return decide(turn, true, false)
        end
    end

    local block = aheadBlock(ctx, tank, dir)
    if block == nil then
        return decide(dir, true, false)
    end
    if (block == "brick" or block == "steel") and canBreach(tank, dir) then
        return startBreach(m, dir, ctx.tick)
    end

    -- Путь перекрыт танком или обломком вне линии выстрела — объезд
    startDetour(m, ctx.tick)
    return roam(ctx, tank, m, profile)
end

-- Фоновая стрельба на ходу: по стенам по курсу и наугад, как в оригинале
local function ambientFire(ctx, tank, m, profile, diff, decision)
    if not decision.move then
        return false
    end
    local sinceShot = ctx.tick - m.lastShot

    -- Стену по курсу обстреливает охотнее и чаще, чем пустоту
    local cx, cy = center(tank)
    local kind, distance = ai.castRay(cx, cy, decision.direction)
    local breakable = kind == "brick" or (kind == "steel" and tank.reinforced)
    if breakable and distance <= DEMOLITION_RANGE then
        return sinceShot >= DEMOLITION_COOLDOWN
            and math.random() < profile.brickFire * diff.ambientMul
    end
    return sinceShot >= AMBIENT_COOLDOWN
        and math.random() < profile.ambientFire * diff.ambientMul
end

-- Не тратит выстрел на союзника и непробиваемый бетон
local function worthShooting(tank, dir)
    local cx, cy = center(tank)
    local kind = ai.castRay(cx, cy, dir)
    if kind == "enemy" then
        return false
    end
    if kind == "steel" and not tank.reinforced then
        return false
    end
    return true
end

local function updateStuck(ctx, tank, m)
    if m.lastMove and m.lastX == tank.x and m.lastY == tank.y then
        m.stuck = m.stuck + 1
    else
        m.stuck = 0
    end
    if m.stuck >= STUCK_LIMIT then
        m.stuck = 0
        startDetour(m, ctx.tick)
    end
end

-- Точка входа

function updateEnemyAI(ctx)
    local tank = ctx.self
    local profile = PROFILES[tank.level] or PROFILES[BASIC]
    local diff = difficulty(ctx.stage)
    local m = getMemory(ctx, profile, diff)
    local mode = currentMode(ctx, m, profile)

    updateStuck(ctx, tank, m)

    local decision = dodge(ctx, tank, m, profile, diff)
        or aim(ctx, tank, m, profile, diff, mode)
        or continueBreach(ctx, tank, m)
        or navigate(ctx, tank, m, profile, diff, mode)

    if not decision.shoot then
        decision.shoot = ambientFire(ctx, tank, m, profile, diff, decision)
    end
    if decision.shoot and not worthShooting(tank, decision.direction) then
        decision.shoot = false
    end

    m.lastX, m.lastY, m.lastMove = tank.x, tank.y, decision.move
    m.mode = mode
    if decision.shoot then
        m.lastShot = ctx.tick
        if decision.aimed then
            m.lastAimedShot = ctx.tick
        end
    end
    return { direction = decision.direction, move = decision.move, shoot = decision.shoot }
end
