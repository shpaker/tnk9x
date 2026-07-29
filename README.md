# tnk9x

A modern remake and tribute to the classic arcade game Battle City (NES, 1985), built with Go and Ebiten. This project combines a passion for the original game with learning Go, Clean Architecture principles, and game development.

**GitHub:** [https://github.com/shpaker/tnk9x](https://github.com/shpaker/tnk9x)

## Development Status

### Core
- [x] Game loop (delta time)
- [x] Collision detection (tank-tank, tank-wall, bullet-wall, bullet-tank, bullet-HQ)
- [x] Boundary collision
- [x] Tank movement/rotation/braking
- [x] Bullet mechanics
- [x] Enemy spawning
- [x] Surface types (brick, steel, forest)
- [ ] Surface types (water, ice)
- [ ] Surface mechanics (ice sliding, water blocking)

### Player
- [x] Controls (keyboard)
- [x] Lives/respawn
- [x] Tank levels (0-3)
- [x] Damage system (level down on hit)
- [x] Bonuses (grenade, tank, star)
- [x] Power-ups (bullet speed, reinforced bullets, bullet limit)
- [ ] Bonuses (helmet, timer, shovel)

### Enemies
- [x] Spawning system
- [x] AI (Lua scripts)
- [x] Level system (probability-based)
- [x] Types (basic, fast, rapid fire, heavy)
- [x] Heavy tank health overlay
- [ ] AI difficulty scaling

### HQ
- [x] Vulnerability/explosion
- [x] Victory/defeat overlays
- [ ] Defeat screen
- [ ] Protection mechanics

### UI
- [x] Level selection
- [x] Pause overlay
- [x] Debug info (FPS, TPS, lives, enemies)
- [ ] HUD (lives, score)
- [ ] Main menu
- [ ] Game over screen
- [ ] Settings

### Audio
- [x] Sound effects
- [x] Background music
- [x] Engine loop
- [x] Volume config

### Infrastructure
- [x] Clean Architecture
- [x] Composition root (internal/app)
- [x] Dependency injection (constructor-only, no setters)
- [x] Layer boundaries enforced by depguard
- [x] Repository pattern
- [x] Specs system
- [x] Scripting behind a domain-typed engine interface
- [x] Unit tests (use cases, services, repositories, app transitions)
- [x] Use cases coverage gate in CI (>=70%)
- [ ] Test coverage (>80% total)
- [x] CI/CD (fmt, lint, test, build, release)
- [ ] Performance profiling

## Installation and Running

**Requirements:** Go 1.24+, optionally — [Just](https://github.com/casey/just).

```bash
git clone https://github.com/shpaker/tnk9x.git
cd tnk9x
just deps         # Install dependencies (or go mod download)

# Run the game
just run          # or go run cmd/main.go

# Build binary
just build        # binary will be in ./tnk9x

# Checks
just fmt
just lint
just test
```

## Architecture

The project follows **Clean Architecture** principles with clear separation of concerns. Dependencies point inward: outer layers depend on inner layers through interfaces. The composition root (`internal/app`) is the only place that assembles the object graph — all dependencies are injected via constructors, layer boundaries are enforced by depguard.

### Dependency Flow

```
┌─────────────────────────────────────────────────────────┐
│  Composition Root (internal/app)                        │
│  game loop · state transitions · object graph assembly  │
└───────────────────┬─────────────────────────────────────┘
                    │ (builds & wires)
┌───────────────────▼─────────────────────────────────────┐
│  Presentation Layer                                     │
│  ┌──────────────┐  ┌──────────────────────┐             │
│  │   States     │  │      Adapters        │             │
│  │              │  │                      │             │
│  │ StageState   │  │ Renderer             │             │
│  │ StageSelect  │  │ Input                │             │
│  │              │  │ Sound                │             │
│  │              │  │ Scripting (Lua)      │             │
│  └──────┬───────┘  └───────┬──────────────┘             │
│         └─────────┬────────┘                            │
│                   │ (depends on)                        │
└───────────────────▼─────────────────────────────────────┘
                    │
┌───────────────────▼─────────────────────────────────────┐
│  Application Layer                                      │
│  ┌──────────────┐  ┌──────────────┐                     │
│  │  Use Cases   │  │   Services   │                     │
│  │              │  │              │                     │
│  │ TankActions  │  │ Collision    │                     │
│  │ Collision    │  │ Animation    │                     │
│  │ Sound        │  │ Braking      │                     │
│  └──────┬───────┘  └──────┬───────┘                     │
│         └─────────┬───────┘                             │
│                   │ (manipulates)                       │
└───────────────────▼─────────────────────────────────────┘
                    │
┌───────────────────▼─────────────────────────────────────┐
│  Domain Layer                                           │
│  ┌───────────────────────────────────────────────────┐  │
│  │  Entities (Tank, Bullet, Block, HQ, etc.)         │  │
│  │  Value Objects (Position, Size, Direction)        │  │
│  │  Session Entities (GameSession, StageSession)     │  │
│  └───────────────────────────────────────────────────┘  │
│                   ▲                                     │
└───────────────────┼─────────────────────────────────────┘
                    │ (reads/writes)
┌───────────────────▼─────────────────────────────────────┐
│  Infrastructure Layer                                   │
│  ┌──────────────┐  ┌──────────────┐                     │
│  │ Repositories │  │   Raw Files  │                     │
│  │              │  │              │                     │
│  │ Game         │  │ FileSystem   │                     │
│  │ Processed    │  │              │                     │
│  └──────────────┘  └──────────────┘                     │
└─────────────────────────────────────────────────────────┘
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
