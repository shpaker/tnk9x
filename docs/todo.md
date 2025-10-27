# TODO: AI для вражеских танков (Lua)

## Этап 1: Интеграция Lua (2ч)
- [x] `go get github.com/yuin/gopher-lua`
- [x] Создать `enemy_ai_interface.go`: `IEnemyAI`, `EnemyAIDecision { ShouldMove, NewDirection }`
- [x] Создать `enemy_ai_lua.go`: wrapper для Lua VM, конвертация типов, метод `Update()`
- [x] Создать `assets/scripts/enemies.lua`: функция `updateEnemyAI()` → возвращает 2 значения

## Этап 2: Интеграция в EnemyUseCases (1.5ч)
- [x] Добавить поля `ai IEnemyAI`, `aiContext`
- [x] Обновить конструктор `NewEnemyUseCases()`
- [x] Реализовать `MoveTank(dt)` - обновление позиции по направлению
- [x] Реализовать `UpdateAI()` - вызов `ai.Update()`, применение решений

## Этап 3: Интеграция в Facade (1ч)
- [x] В `NewGameStateUseCasesFacade()` создать AI, передать в `NewEnemyUseCases()`
- [x] В `Update()` вызвать `UpdateAI()` и `MoveTank()` для каждого врага

## Этап 4: Продвинутое поведение (1ч)
- [ ] В Lua: `findNearestPlayer()`, `distance()`, `getDirectionTo()`
- [ ] Преследование игрока, периодическая стрельба

## Этап 5: Тестирование (1ч)
- [ ] Враги двигаются, стреляют, реагируют на игрока

---
**MVP (Этапы 1-3): ~4.5ч | Полная версия: ~7ч**
