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
- [x] Dependency injection
- [x] Repository pattern
- [x] Specs system
- [x] Unit tests (collision, audio)
- [ ] Test coverage (>80%)
- [ ] CI/CD
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
just build        # binary will be in dist/tnk9x

# Checks
just fmt
just lint
just test
```

## Architecture

The project follows **Clean Architecture** principles with clear separation of concerns across four layers. Dependencies point inward: outer layers depend on inner layers through interfaces.

### Dependency Flow

```
┌────────────────────────────────────────────────────────┐
│  Presentation Layer                                    │
│  ┌──────────────┐  ┌──────────────┐                    │
│  │   States     │  │   Adapters   │                    │
│  │              │  │              │                    │
│  │ StageState   │  │ Renderer     │                    │
│  │ StageSelect  │  │ Input        │                    │
│  │              │  │ Sound        │                    │
│  └──────┬───────┘  └───────┬──────┘                    │
│         │                  │                           │
│         └─────────┬────────┘                           │
│                   │ (depends on)                       │
│                   ▼                                    │
└────────────────────────────────────────────────────────┘
                    │
┌───────────────────▼─────────────────────────────────────┐
│  Application Layer                                      │
│  ┌──────────────┐  ┌──────────────┐                     │
│  │  Use Cases   │  │   Services   │                     │
│  │              │  │              │                     │
│  │ TankActions  │  │ Collision    │                     │
│  │ Collision    │  │ Animation    │                     │
│  │ Sound        │  │ Coordinate   │                     │
│  └──────┬───────┘  └──────┬───────┘                     │
│         │                 │                             │
│         └─────────┬───────┘                             │
│                   │ (manipulates)                       │
│                   ▼                                     │
└─────────────────────────────────────────────────────────┘
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
