# tnk9x

A modern remake and tribute to the classic arcade game Battle City (NES, 1985), built with Go and Ebiten. This project combines a passion for the original game with learning Go, Clean Architecture principles, and game development.

**GitHub:** [https://github.com/shpaker/tnk9x](https://github.com/shpaker/tnk9x)

## Development Status

**Playable now:** NES-faithful campaign flow (title screen, STAGE N curtain with stage select, 35 stages with auto-advance and loop, per-stage score tally, GAME OVER screen), lives/score/star levels carried across stages, scoring with 100-400 points per enemy tier, 500 per power-up, extra life at 20000 and session HI-SCORE, fixed NES enemy wave tables with cyclic spawners and stage-scaled spawn delay, all six bonuses (star, grenade, tank, helmet, timer, shovel with steel-ring fortification), canonical combat rules (death on hit, spawn shield, friendly-fire freeze, flashing bonus tanks #4/#11/#18 dropping power-ups on first hit), full-cell-wide brick strip destruction with bullet impact explosions, phase-based enemy AI (wander, hunt player, siege HQ), two-player keyboard controls, all five surface types with ice sliding, pause, sound, NES-style sidebar HUD on an authentic 256x224 screen.

**Under the hood:** Clean Architecture with depguard-enforced layer boundaries, constructor-only DI from a composition root, repository pattern, scripting behind a domain-typed engine interface, app-lifetime GPU sprite cache with startup preload, fail-fast sprite/animation validation on startup, unit tests with a >=70% use-cases coverage gate, CI/CD (fmt, lint, test, build, release).

### Roadmap
- Persistent HI-SCORE and settings
- Two-player tally bonuses, red palette flash for bonus tanks
- Test coverage >80% total, performance profiling

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
│  │ Title        │  │ Renderer             │             │
│  │ StageCurtain │  │ Input                │             │
│  │ StageState   │  │ Sound                │             │
│  │ Score        │  │ Scripting (Lua)      │             │
│  │ GameOver     │  │                      │             │
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
