# 🎮 Gonflict

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)
[![Ebiten](https://img.shields.io/badge/Ebiten-v2.9.1-orange.svg?style=flat-square)](https://ebiten.org/)

Gonflict — ремейк Battle City на Go с использованием Ebiten. Проект строится вокруг принципов Clean Architecture и обеспечивает четкое разделение ответственности по слоям.

## Кратко

- Полноценный игровой цикл: спавн, движение, стрельба, уничтожение
- 35 уровней, загружаемых из конфигураций
- Lua-скрипты для поведения врагов
- Отрисовка, анимации, звук и экран выбора уровня с выбором числа игроков и подсказками
- Все ресурсы и состояния обслуживаются через репозитории и use cases
- Сессии игры и уровня (`GameSessionEntity`, `StageSessionEntity`) единообразно управляют прогрессом, а `StageUseCases` проверяет победу/поражение и формирует оверлей результата

## Статус разработки

| Категория | Готово | В работе |
| --- | --- | --- |
| Игровой цикл | ✅ Обновление мира, столкновения, управление | ➖ Волны врагов, прогрессия уровней |
| Игрок | ✅ Система торможения, жизни, респавн | ➖ Улучшения и бонусы |
| Враги и AI | ✅ Спавн, движение, Lua-сценарии | ➖ Разные типы врагов, бонусы |
| База HQ | ✅ Уязвимость, взрыв, оверлей поражения | ➖ Переход к отдельному экрану поражения |
| UI | ✅ Экран выбора уровня, оверлеи паузы/результата | ➖ HUD, меню, итоговые экраны |
| Техническое | ✅ Слои, интерфейсы, DI, unit-тесты сервисов коллизий | ➖ Расширение покрытия и автоматизация |

Полный список открытых задач с пометками «готово» и «в работе» см. в `docs/completion_checklist.md`.

## Установка и запуск

**Требования:** Go 1.24+, опционально — [Just](https://github.com/casey/just).

```bash
git clone https://github.com/shpaker/gonflict.git
cd gonflict
go mod download
```

```bash
# Запустить игру
just run          # или go run cmd/main.go

# Собрать бинарник
just build        # бинарник появится в dist/gonflict

# Проверки
just fmt
just lint
just test
```

## Управление

| Контекст | Клавиши |
| --- | --- |
| Игра | `WASD` — движение, `Space` — стрельба, `F1` — переключить отладочный оверлей, `Escape` — выход |
| Выбор уровня | `Up/Down` или `W/S` — переключение между пунктами меню, `Left/Right` или `A/D` — изменить значение активного пункта, `Enter` — запустить игру, `Escape` — выход |

На экране выбора уровня активный пункт подсвечивается белым. Дополнительно выводится подсказка `PRESS ENTER TO START` с тем же размером шрифта, что и прочие субтитры.

## Конфигурация

Все пользовательские настройки хранятся в `config.yml`:

- `app.title_font_size`, `app.subtitle_font_size`, `app.regular_font_size` — размеры шрифтов для заголовков, подсказок и элементов меню.
- `app.game_title` — текст заголовка, отображаемого на экране выбора уровня.
- Прочие параметры (`level_number`, `screen_px`, `game.*`) описывают стартовый уровень, размер экрана, карту и поведение врагов.

Режим отладки можно переключать по клавише `F1`. Флаг `internal.Debug` по умолчанию `false` и устанавливается в `true` при сборке Debug-версий (`just build-macos`, `just build-windows`, GitHub Actions для debug-артефактов).

Дополнительно: танк автоматически выравнивается по сетке 4×4, тормозит плавно и завершает маневр перед сменой направления.

## Архитектура

Проект придерживается Clean Architecture: Presentation → Application → Domain → Infrastructure. Все зависимости направлены внутрь, а работа с ресурсами строго идет через интерфейсы репозиториев.

```text
internal/
├── adapters/       # Presentation (реализация вывода/ввода через адаптеры)
│   ├── game/       # Уровень: StageRendererAdapter, StageKeyboardInputAdapter, AI adapters
│   └── stage_select/ # Экран выбора уровня: отрисовка и ввод меню
├── states/         # Presentation (StageSelectState, StageState и т.д.)
├── use_cases/      # Application (stateless бизнес-логика)
├── services/       # Application services (коллизии, анимации, координаты)
├── types/          # Domain types (танки, пули, HQ, карта)
│   └── session_entities/ # Сессии игры и уровней
├── repositories/   # Infrastructure (in-memory, processed, raw)
└── interfaces/     # Контракты между слоями
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

Основные правила:

- Use cases stateless, принимают сущности как параметры
- Отрисовка и ввод осуществляются только через адаптеры (`*renderer_adapter`, `*input_adapter`)
- Не использовать устаревшие (deprecated) API движка и стандартной библиотеки
- Сервисы коллизий только проверяют, use cases вносят изменения
- Сырые ресурсы доступны только через `IFileRepository`; обработанные — через специализированные репозитории
- Все зависимости внедряются через конструкторы (`New*`)

Детали и список оставшихся архитектурных задач описаны в `docs/architecture_deficiencies.md`.

## Тестирование

```bash
go test ./...
go test -cover ./...
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Вклад

1. Форкните репозиторий
2. Создайте ветку: `git checkout -b feature/your-feature`
3. Внесите изменения, запустите `just check`
4. Отправьте PR с описанием изменений

Перед PR убедитесь, что код отформатирован, покрыт тестами и сопровождается обновленной документацией.

## Дополнительные материалы

- `docs/architecture_deficiencies.md` — текущее состояние архитектуры и план улучшений
- `docs/completion_checklist.md` — чеклист геймдизайн-задач
- `docs/resource_management.md` — правила работы с ресурсами (репозитории, ассеты)

## Лицензия

MIT — см. [LICENSE](LICENSE).

---

Приятной игры и удачных пул-реквестов! 🎮
