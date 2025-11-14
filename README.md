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

The project follows **Clean Architecture** principles with clear separation of concerns across four layers. Dependencies point inward: outer layers depend on inner layers through interfaces.

### Layer Structure

```text
internal/
├── types/                    # Domain Layer - Pure business entities
│   ├── image_providers/      # Image provider implementations
│   ├── session_entities/     # Game and stage session management
│   └── *.go                  # Core entities (Tank, Bullet, Block, HQ, etc.)
│
├── interfaces/                # Contracts between layers
│   ├── adapters.go           # IInputAdapter, ISoundPlayerAdapter, etc.
│   ├── domain.go             # Domain interfaces (IBlink, IImageProvider)
│   ├── repositories.go       # IGameRepositoriesRegistry, ITanksRepository, etc.
│   ├── services.go           # ISoundPlayerAdapter, collision services
│   └── use_cases.go          # ITankCommonUseCases, IStageUseCases, etc.
│
├── use_cases/                # Application Layer - Stateless business logic
│   ├── tank_use_cases/       # Tank-specific use cases
│   │   ├── tank_actions_use_cases.go
│   │   ├── tank_common_use_cases.go
│   │   └── tank_lifecycle_use_cases.go
│   ├── state_use_cases/      # Stage state management
│   └── *.go                  # Other use cases (collision, bonus, sound, etc.)
│
├── services/                 # Application Layer - Specialized services
│   ├── collision_services/   # Collision detection services
│   ├── animation_services.go
│   ├── coordinate_services.go
│   ├── tank_braking_services.go
│   └── tile_services.go
│
├── adapters/                 # Presentation Layer - External system adapters
│   ├── stage/
│   │   ├── input_adapters/   # Keyboard and AI input adapters
│   │   ├── sound_adapter.go  # Audio playback adapter
│   │   └── stage_renderer_adapter.go
│   └── stage_select/         # Level selection screen adapters
│
├── states/                   # Presentation Layer - Application states
│   ├── stage_state.go        # Main game state
│   ├── stage_state_builder.go
│   └── stage_select_state.go
│
└── repositories/             # Infrastructure Layer - Data access
    ├── raw/                  # Direct file system access
    ├── processed/            # Processed resources (tilesets, maps, sounds)
    └── game/                 # In-memory game state (tanks, bullets, bonuses)
```

### Dependency Flow

```
┌─────────────────────────────────────────────────────────┐
│  Presentation Layer                                     │
│  ┌──────────────┐  ┌──────────────┐                    │
│  │   States     │  │   Adapters   │                    │
│  │              │  │              │                    │
│  │ StageState   │  │ Renderer     │                    │
│  │ StageSelect  │  │ Input        │                    │
│  │              │  │ Sound        │                    │
│  └──────┬───────┘  └──────┬───────┘                    │
│         │                  │                             │
│         └──────────┬───────┘                             │
│                   │ (depends on)                        │
│                   ▼                                     │
└─────────────────────────────────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────────────┐
│  Application Layer                                      │
│  ┌──────────────┐  ┌──────────────┐                    │
│  │  Use Cases   │  │   Services   │                    │
│  │              │  │              │                    │
│  │ TankActions  │  │ Collision    │                    │
│  │ Collision    │  │ Animation    │                    │
│  │ Sound        │  │ Coordinate   │                    │
│  └──────┬───────┘  └──────┬───────┘                    │
│         │                  │                             │
│         └──────────┬───────┘                             │
│                   │ (manipulates)                       │
│                   ▼                                     │
└─────────────────────────────────────────────────────────┘
                   │
┌──────────────────▼──────────────────────────────────────┐
│  Domain Layer                                           │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Entities (Tank, Bullet, Block, HQ, etc.)        │  │
│  │  Value Objects (Position, Size, Direction)       │  │
│  │  Session Entities (GameSession, StageSession)     │  │
│  └──────────────────────────────────────────────────┘  │
│                   ▲                                     │
└───────────────────┼─────────────────────────────────────┘
                   │ (reads/writes)
┌──────────────────▼──────────────────────────────────────┐
│  Infrastructure Layer                                  │
│  ┌──────────────┐  ┌──────────────┐                    │
│  │ Repositories │  │   Raw Files  │                    │
│  │              │  │              │                    │
│  │ Game         │  │ FileSystem   │                    │
│  │ Processed    │  │              │                    │
│  └──────────────┘  └──────────────┘                    │
└─────────────────────────────────────────────────────────┘
```

### Key Principles

- **Dependency Inversion**: All dependencies point inward through interfaces
- **Stateless Use Cases**: Use cases don't store entity state; entities are passed as parameters
- **Single Responsibility**: Each use case handles one domain area
- **Interface Segregation**: Small, focused interfaces for each responsibility
- **Repository Pattern**: Data access abstracted through repository interfaces
