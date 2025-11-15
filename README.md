# [tnk25](https://github.com/shpaker/tnk25)

A modern remake and tribute to the classic arcade game Battle City (NES, 1985), built with Go and Ebiten. This project combines a passion for the original game with learning Go, Clean Architecture principles, and game development.

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
- [x] Star bonus (tank level up)
- [x] Tank level system (0-3 levels)
- [x] Tank level affects visual appearance
- [x] Player damage system (level down on hit, explode at level 0)
- [ ] Full bonus implementation (helmet, timer, shovel)
- [x] Player power-up system (bullet speed, reinforced bullets, bullet limit)

### Enemies & AI
- [x] Enemy spawning system
- [x] Enemy movement AI
- [x] Lua-based AI scripting
- [x] Enemies with bonus indicators
- [x] Enemy tracking and lifecycle management
- [x] Dynamic enemy level system based on remaining enemies count
- [x] Enemy level progression (levels 0-3 with probability distribution)
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
- [x] Debug information display (FPS, TPS, player lives, enemy count)
- [x] Debug mode toggle (F1 key)
- [x] Debug commands (key 0 - level up player tanks)
- [ ] Loading screen on game start
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
- [x] Named interfaces (no anonymous interfaces)
- [x] Tank specifications system (SpecsEntity)
- [x] Tank specs use cases (SpecsUseCases)
- [x] Unit tests for collision services
- [x] Unit tests for audio system
- [ ] Extended test coverage (>80%)
- [ ] CI/CD pipeline automation
- [ ] Performance profiling and optimization

## Installation and Running

**Requirements:** Go 1.24+, optionally — [Just](https://github.com/casey/just).

```bash
git clone https://github.com/shpaker/tnk25.git
cd tnk25
just deps         # Install dependencies (or go mod download)

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
