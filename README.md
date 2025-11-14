# TNK25

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)
[![Ebiten](https://img.shields.io/badge/Ebiten-v2.9.1-orange.svg?style=flat-square)](https://ebiten.org/)

tnk25 — a modern remake of the classic arcade game Battle City (NES, 1985), built with Go and Ebiten.

## Development Status

### Core Gameplay
- [x] Game loop with delta time
- [x] Collision detection system (tank-tank, tank-wall, bullet-wall, bullet-tank, bullet-HQ)
- [x] Boundary collision handling
- [x] Tank movement and rotation
- [x] Tank braking system
- [x] Bullet shooting mechanics
- [x] Enemy spawning with delay system
- [x] Basic surface types (brick, steel)
- [x] Forest surface type with hiding mechanics
- [ ] Advanced surface types (water, ice)
- [ ] Surface-specific mechanics (ice sliding, water blocking)

### Player Features
- [x] Player tank controls (keyboard input)
- [x] Player lives system
- [x] Player respawn mechanism
- [x] Basic bonuses (grenade, tank)
- [ ] Full bonus implementation (helmet, timer, shovel, star)
- [ ] Player power-up system (bullet speed, armor)

### Enemies & AI
- [x] Enemy spawning system
- [x] Enemy movement AI
- [x] Lua-based AI scripting
- [x] Enemies with bonus indicators
- [x] Enemy tracking and lifecycle management
- [ ] Different enemy types (fast, armored, heavy)
- [ ] Enemy AI difficulty scaling

### HQ Base
- [x] HQ vulnerability system
- [x] HQ explosion animation
- [x] Victory/defeat overlays
- [ ] Separate defeat screen
- [ ] HQ protection mechanics (shovel bonus)

### User Interface
- [x] Level selection screen
- [x] Pause overlay
- [x] Result overlays (victory/defeat)
- [ ] HUD (lives counter, score display)
- [ ] Main menu screen
- [ ] Game over screen
- [ ] Settings screen

### Audio System
- [x] Sound effects (fire, explosion, brick, steel, bonus, score)
- [x] Background music (game start, game over)
- [x] Engine sound with loop
- [x] Sound adapter implementation
- [x] Volume configuration in config.yml

### Technical Infrastructure
- [x] Clean Architecture implementation
- [x] Interface-based dependency injection
- [x] Repository pattern for data access
- [x] Unit tests for collision services
- [x] Unit tests for audio system
- [ ] Extended test coverage (>80%)
- [ ] CI/CD pipeline automation
- [ ] Performance profiling and optimization

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
