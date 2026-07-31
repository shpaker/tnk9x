# tnk9x

A modern remake and tribute to the classic arcade game Battle City (NES, 1985), built with Go and Ebiten. This project combines a passion for the original game with learning Go, Clean Architecture principles, and game development.

**GitHub:** [https://github.com/shpaker/tnk9x](https://github.com/shpaker/tnk9x)

**Play online:** [https://shpaker.github.io/tnk9x/](https://shpaker.github.io/tnk9x/)

## Development Status

**Playable now:**

- Full game loop with HQ and victory/defeat overlays
- Two-player keyboard controls
- Touch controls for mobile browsers: auto-detected virtual D-pad, fire and pause in the letterbox area, tappable menu
- Tank movement with braking and grid snap
- Bullets and destructible terrain with incremental brick chipping: each hit shaves a half-tile slab, reinforced bullets break tiles whole
- All five surface types (brick, steel, forest, water, ice) with ice sliding and water blocking
- Lua-scripted enemies of four types with probability-based levels
- Player lives, levels and damage
- All six bonuses: grenade, tank, star, helmet shield, enemy-freezing timer, HQ-fortifying shovel
- Level selection with a desktop quit item
- Pause menu on Esc/P/touch (continue, exit to level select)
- Sound effects and music
- NES-style sidebar HUD (enemy reserve, player lives, stage flag) on an authentic 256x224 screen
- Runs natively and [in the browser](https://shpaker.github.io/tnk9x/) (WebAssembly, deployed to GitHub Pages on release tags)

**Under the hood:**

- Clean Architecture with depguard-enforced layer boundaries
- Constructor-only DI from a composition root, repository pattern
- Scripting behind a domain-typed engine interface
- App-lifetime GPU sprite cache with startup preload, fail-fast sprite/animation validation on startup
- Unit tests with a >=70% use-cases coverage gate
- CI/CD (fmt, lint, test, build, release)

### Roadmap
- Enemy AI difficulty scaling
- HQ: defeat screen, protection mechanics
- UI: score, main menu, game over screen, settings
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
