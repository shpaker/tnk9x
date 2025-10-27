# Lua vs Luau: Выбор скриптинга для AI в gonflict

## Введение

При интеграции AI в проект gonflict (Battle City на Go) нужно выбрать между классическим Lua и современным Luau. Этот документ сравнивает две доступные библиотеки для встраивания скриптов в Go.

## Сравнение библиотек

### Lua + gopher-lua

**gopher-lua** (`github.com/yuin/gopher-lua`) - зрелая и проверенная библиотека для встраивания Lua в Go.

#### ✅ Преимущества:
- **Стабильность**: Зрелая библиотека, используемая в продакшене
- **Простота интеграции**: Чистая Go-реализация, без внешних зависимостей
- **Богатая экосистема**: Множество примеров и документации
- **Совместимость**: Полная поддержка Lua 5.1
- **Отличная документация**: Множество примеров использования
- **Большое сообщество**: Тысячи пользователей GitHub

#### ❌ Недостатки:
- **Медленнее нативного Lua**: Реализация на чистом Go
- **Нет Luau возможностей**: Типизация, современный синтаксис
- **Ограниченная обработка ошибок**: Базовый `pcall` поддержка

#### Пример использования:
```go
import "github.com/yuin/gopher-lua"

L := lua.NewState()
defer L.Close()

// Простая регистрация функций
L.SetGlobal("getPlayer", L.NewFunction(func(L *lua.LState) int {
    gameRepo := GetGameRepository()
    context := gameRepo.GetGameContext()
    
    table := L.NewTable()
    if context.Player != nil {
        table.RawSetString("x", lua.LNumber(context.Player.WorldPosition.X))
        table.RawSetString("y", lua.LNumber(context.Player.WorldPosition.Y))
    }
    L.Push(table)
    return 1
}))

if err := L.DoFile("ai/enemy_ai.lua"); err != nil {
    panic(err)
}
```

### Luau + gluau

**gluau** ([github.com/gluau/gluau](https://github.com/gluau/gluau)) - современная библиотека для встраивания Luau в Go, использует Rust через mlua crate.

#### ✅ Преимущества:
- **Современный Luau**: Типизация, улучшенная производительность
- **Лучшая производительность**: Rust-бэкенд через mlua
- **Статическая типизация**: Опциональная типизация как в Luau
- **Exception handling**: Полная поддержка `pcall` и `xpcall` для Go
- **Interrupt-система**: Контроль времени выполнения скриптов
- **Luau Buffers**: Эффективная работа с данными

#### ❌ Недостатки:
- **Heavy WIP**: Проект в активной разработке (версия ещё не стабильная)
- **Сложная сборка**: Требует Rust и CGO
- **Нестабильный API**: API может кардинально измениться
- **Мало примеров**: 18 ⭐ на GitHub против тысяч для gopher-lua
- **Ограниченная документация**: Работа в процессе

#### Что реализовано в gluau:
- ✅ VM инициализация и выключение
- ✅ Базовый API для Lua значений (через Go интерфейсы)
- ✅ Lua Strings с API
- ✅ Lua Tables с API
- ✅ Lua Functions (базовый функционал)
- ✅ Lua Userdata
- ✅ Lua Threads (resume и yield)
- ✅ Luau Buffers

#### Что в разработке:
- 🚧 Luau require by string
- 🚧 Другие Luau-специфичные возможности

#### Пример использования:
```go
import "github.com/gluau/gluau/vm/vmlib"

// Создание Luau VM с Rust бэкендом
vm, err := vmlib.CreateLuaVm()
if err != nil {
    panic(err)
}
defer vm.Close()

// Создание глобальной таблицы
globTab, err := vm.CreateTableWithCapacity(16)

// Interrupt для контроля времени выполнения
vm.SetInterrupt(func(funcVm *vmlib.CallbackLua) (vmlib.VmState, error) {
    if time.Since(startTime) > 10*time.Millisecond {
        return vmlib.VmStateYield, nil // Остановить выполнение
    }
    return vmlib.VmStateContinue, nil
})

// Регистрация Go функции
err = globTab.Set(vmlib.GoString("getPlayer"), vm.CreateGoFunction(func(args []vmlib.Value) ([]vmlib.Value, error) {
    gameRepo := GetGameRepository()
    context := gameRepo.GetGameContext()
    
    if context.Player == nil {
        return []vmlib.Value{}, nil
    }
    
    // Создаем Lua таблицу
    playerTab, err := vm.CreateTable()
    playerTab.Set(vmlib.GoString("x"), vm.CreateNumber(context.Player.WorldPosition.X))
    playerTab.Set(vmlib.GoString("y"), vm.CreateNumber(context.Player.WorldPosition.Y))
    
    return []vmlib.Value{playerTab}, nil
}))
```

### Ключевые отличия синтаксиса

#### Lua (через gopher-lua):
```lua
-- Классический Lua
function calculateDistance(x1, y1, x2, y2)
    return math.sqrt((x2 - x1)^2 + (y2 - y1)^2)
end

-- Динамическая типизация
player = {
    x = 100,
    y = 100,
    health = 100
}
```

#### Luau (через gluau):
```lua
-- Luau с опциональной типизацией
type Enemy = {
    x: number,
    y: number,
    health: number
}

function calculateDistance(a: Enemy, b: Enemy): number
    return math.sqrt((b.x - a.x)^2 + (b.y - a.y)^2)
end

-- Статическая проверка типов
local enemy: Enemy = {
    x = 100,
    y = 100,
    health = 100
}
```

## Детальное сравнение

| Критерий | Lua + gopher-lua | Luau + gluau |
|----------|------------------|--------------|
| **Стабильность** | ✅ Зрелая | ⚠️ Heavy WIP |
| **API стабильность** | ✅ Стабильная | ⚠️ Может меняться |
| **Сборка** | ✅ Простая (чистый Go) | ⚠️ Сложная (Rust/CGO) |
| **Документация** | ✅ Обильная | ⚠️ Ограниченная |
| **Сообщество** | ✅ Тысячи звезд | ⚠️ 18 звезд |
| **Производительность** | ✅ Хорошая | ✅ Лучшая (Rust) |
| **Типизация** | ❌ Нет | ✅ Есть (Luau) |
| **Exception handling** | ⚠️ Ограниченная | ✅ Полная |
| **Interrupt-система** | ❌ Нет | ✅ Есть |
| **Luau Buffers** | ❌ Нет | ✅ Есть |
| **Отладка** | ✅ Простая | ✅ Luau типы |
| **Windows поддержка** | ✅ Есть | ⚠️ Требует MinGW |

## Рекомендация для gonflict

### Для проекта gonflict рекомендуется использовать **Lua + gopher-lua** потому что:

1. **Стабильность и надежность**
   - ✅ gopher-lua - проверенная временем библиотека
   - ✅ gluau находится в активной разработке (Heavy WIP)
   - ✅ API может кардинально измениться в любой момент

2. **Простота интеграции**
   - ✅ Не требует Rust и CGO
   - ✅ Легкая сборка проекта
   - ✅ Простая работа с `go get`

3. **Богатая экосистема**
   - ✅ Множество готовых библиотек для Lua
   - ✅ Большое сообщество разработчиков
   - ✅ Много примеров использования

4. **Достаточная производительность**
   - ✅ Для AI скриптов Lua достаточно быстр
   - ✅ gopher-lua хорошо оптимизирован для Go

5. **Гибкость**
   - ✅ Легко менять логику без перекомпиляции
   - ✅ Идеально для быстрого прототипирования AI

### Когда имеет смысл рассматривать gluau:

**В будущем, когда gluau станет стабильным (версия 1.0+), если вам потребуется:**
- Сложная AI логика с большим количеством кода
- Статическая типизация для безопасности типов
- Максимальная производительность скриптов
- Точный контроль времени выполнения (для сетевых игр)
- Exception handling в AI логике

## Пример интеграции с gonflict

### Используя gopher-lua:

```go
package main

import (
    "github.com/yuin/gopher-lua"
    "github.com/shpaker/gonflict/internal/repositories/game"
)

func SetupAI(gameRepo game.IGameRepository) *lua.LState {
    L := lua.NewState()
    
    // Регистрируем функцию для получения игрока
    L.SetGlobal("getPlayer", L.NewFunction(func(L *lua.LState) int {
        context := gameRepo.GetGameContext()
        table := L.NewTable()
        
        if context.Player != nil {
            table.RawSetString("x", lua.LNumber(context.Player.WorldPosition.X))
            table.RawSetString("y", lua.LNumber(context.Player.WorldPosition.Y))
        }
        
        L.Push(table)
        return 1
    }))
    
    // Регистрируем функцию для получения врагов
    L.SetGlobal("getEnemies", L.NewFunction(func(L *lua.LState) int {
        context := gameRepo.GetGameContext()
        table := L.NewTable()
        
        for i, enemy := range context.Enemies {
            enemyTable := L.NewTable()
            enemyTable.RawSetString("x", lua.LNumber(enemy.WorldPosition.X))
            enemyTable.RawSetString("y", lua.LNumber(enemy.WorldPosition.Y))
            table.RawSetInt(i+1, enemyTable)
        }
        
        L.Push(table)
        return 1
    }))
    
    // Регистрируем функцию для получения пуль
    L.SetGlobal("getBullets", L.NewFunction(func(L *lua.LState) int {
        context := gameRepo.GetGameContext()
        table := L.NewTable()
        
        for i, bullet := range context.Bullets {
            bulletTable := L.NewTable()
            bulletTable.RawSetString("x", lua.LNumber(bullet.WorldPosition.X))
            bulletTable.RawSetString("y", lua.LNumber(bullet.WorldPosition.Y))
            table.RawSetInt(i+1, bulletTable)
        }
        
        L.Push(table)
        return 1
    }))
    
    return L
}

// Использование
func RunAI(gameRepo game.IGameRepository) {
    L := SetupAI(gameRepo)
    defer L.Close()
    
    if err := L.DoFile("ai/enemy_ai.lua"); err != nil {
        panic(err)
    }
}
```

## Вывод

Для проекта **gonflict** используйте **Lua + gopher-lua**. Это обеспечит:
- ✅ Простую и быструю интеграцию
- ✅ Достаточную производительность
- ✅ Богатую экосистему
- ✅ Гибкость разработки AI
- ✅ Стабильность и надежность

**gluau** имеет смысл рассматривать для будущих версий проекта, когда:
1. Библиотека станет стабильной (версия 1.0+)
2. Возникнет потребность в статической типизации
3. Потребуется максимальная производительность
4. Нужен контроль времени выполнения скриптов

## Дополнительные ресурсы

- [Lua официальный сайт](https://www.lua.org/)
- [Luau документация](https://luau-lang.org/)
- [gopher-lua GitHub](https://github.com/yuin/gopher-lua) - Рекомендуется
- [gluau GitHub](https://github.com/gluau/gluau) - В разработке
- [Lua в 5 минут](https://learnxinyminutes.com/docs/lua/)
