# TNK25

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)
[![Ebiten](https://img.shields.io/badge/Ebiten-v2.9.1-orange.svg?style=flat-square)](https://ebiten.org/)

tnk25 — a modern remake of the classic arcade game Battle City (NES, 1985), built with Go and Ebiten.

## Development Status

| Category | Completed | In Progress |
| --- | --- | --- |
| Game Loop | ✅ World updates, collisions, controls, enemy spawning with delay | ➖ Automatic level progression |
| Player | ✅ Braking system, lives, respawn, basic bonuses (grenade, tank) | ➖ Full bonus implementation (helmet, timer, shovel, star) |
| Enemies & AI | ✅ Spawning, movement, Lua scripts, enemies with bonuses | ➖ Different enemy types (only regular) |
| HQ Base | ✅ Vulnerability, explosion, victory/defeat overlays | ➖ Separate defeat screen |
| UI | ✅ Level selection screen, pause/result overlays | ➖ HUD (lives, score), main menu, final screens |
| Technical | ✅ Clean Architecture, interfaces, DI, unit tests for collision services, audio | ➖ Extended test coverage and automation |

## Installation and Running

**Requirements:** Go 1.24+, optionally — [Just](https://github.com/casey/just).

### System Dependencies

System libraries are required for building the project (especially for Linux, where CGO is used for audio):

**Linux (Ubuntu/Debian):**
```bash
sudo apt-get update
sudo apt-get install -y \
  libglfw3 \
  libglfw3-dev \
  libglu1-mesa-dev \
  mesa-common-dev \
  libx11-dev \
  libxrandr-dev \
  libxinerama-dev \
  libxi-dev \
  libxcursor-dev \
  libxxf86vm-dev \
  libasound2-dev \
  pkg-config
```

**macOS:**
Dependencies are usually installed automatically via Homebrew when installing Go and Ebiten.

**Windows:**
No additional dependencies required.

### Build and Run

```bash
git clone https://github.com/shpaker/tnk25.git
cd tnk25
go mod download
```

```bash
# Run the game
just run          # or go run cmd/main.go

# Build binary
just build        # binary will be in dist/tnk25

# Checks
just fmt
just lint
just test
```

## Architecture

The project follows Clean Architecture principles: the core (`Domain`) describes pure entities, the `Application` layer manages business logic through interfaces, and external layers (`Presentation`, `Infrastructure`) depend only on public contracts of internal layers.

```text
internal/
├── types/            # Domain: entities (tanks, bullets, HQ, maps)
│   └── session_entities/ # Game and stage session entities
├── interfaces/       # Contracts between layers
├── use_cases/        # Application: stateless business logic
├── services/         # Application: specialized services (collisions, animations, coordinates)
├── adapters/         # Presentation: input/output implementation
│   ├── stage/        # StageRendererAdapter, StageKeyboardInputAdapter, AI adapters
│   └── stage_select/ # Level selection screen: render and input
├── states/           # Presentation: StageSelectState, StageState, etc.
└── repositories/     # Infrastructure: raw, processed, game storage
```

Layer interaction (dependencies point inward):

```
[ Presentation ]  adapters + states
        │   (implementation of IState, IInputAdapter, IRendererAdapter)
        ▼
[ Application ]   use_cases + services (via interfaces.*)
        │   (manipulate Domain, access infrastructure through interfaces)
        ▼
[ Domain ]        types (entity, value objects, session entities)
        ▲
        │   (repositories read/write data, implementing Application interfaces)
[ Infrastructure ] repositories/raw/processed/game
```
