# TNK25

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)
[![Ebiten](https://img.shields.io/badge/Ebiten-v2.9.1-orange.svg?style=flat-square)](https://ebiten.org/)

tnk25 — ремейк Battle City на Go.

## Статус разработки

| Категория | Готово | В работе |
| --- | --- | --- |
| Игровой цикл | ✅ Обновление мира, столкновения, управление, спавн врагов с задержкой | ➖ Автоматическая прогрессия уровней |
| Игрок | ✅ Система торможения, жизни, респавн, базовые бонусы (граната, танк) | ➖ Полная реализация бонусов (helmet, timer, shovel, star) |
| Враги и AI | ✅ Спавн, движение, Lua-сценарии, враги с бонусами | ➖ Разные типы врагов (только regular) |
| База HQ | ✅ Уязвимость, взрыв, оверлеи победы/поражения | ➖ Отдельный экран поражения |
| UI | ✅ Экран выбора уровня, оверлеи паузы/результата | ➖ HUD (жизни, очки), главное меню, итоговые экраны |
| Техническое | ✅ Clean Architecture, интерфейсы, DI, unit-тесты сервисов коллизий, аудио | ➖ Расширение покрытия тестами и автоматизация |

## Установка и запуск

**Требования:** Go 1.24+, опционально — [Just](https://github.com/casey/just).

```bash
git clone https://github.com/shpaker/tnk25.git
cd tnk25
go mod download
```

```bash
# Запустить игру
just run          # или go run cmd/main.go

# Собрать бинарник
just build        # бинарник появится в dist/tnk25

# Проверки
just fmt
just lint
just test
```

## Архитектура

Проект следует принципам Clean Architecture: ядро (`Domain`) описывает чистые сущности, слой `Application` управляет бизнес-логикой через интерфейсы, а внешние слои (`Presentation`, `Infrastructure`) зависят только от публичных контрактов внутренних слоев.

```text
internal/
├── types/            # Domain: сущности (танки, пули, HQ, карты)
│   └── session_entities/ # Сущности игровых и стадийных сессий
├── interfaces/       # Контракты между слоями
├── use_cases/        # Application: stateless бизнес-логика
├── services/         # Application: специализированные сервисы (коллизии, анимации, координаты)
├── adapters/         # Presentation: реализация ввода/вывода
│   ├── stage/        # StageRendererAdapter, StageKeyboardInputAdapter, AI adapters
│   └── stage_select/ # Экран выбора уровня: рендер и ввод
├── states/           # Presentation: StageSelectState, StageState и др.
└── repositories/     # Infrastructure: raw, processed, game-хранилища
```

Взаимодействие слоев (зависимости направлены внутрь):

```
[ Presentation ]  adapters + states
        │   (реализация IState, IInputAdapter, IRendererAdapter)
        ▼
[ Application ]   use_cases + services (через interfaces.*)
        │   (манипулируют Domain, обращаются к инфраструктуре через интерфейсы)
        ▼
[ Domain ]        types (entity, value objects, session entities)
        ▲
        │   (репозитории читают/пишут данные, реализуя интерфейсы Application)
[ Infrastructure ] repositories/raw/processed/game
```
