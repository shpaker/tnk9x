# TNK9x: Night Battle

**The tanks game from your childhood — but not a remake.** Childhood tanks, now with seriously upgraded lighting.

**▶ [Play in the browser](https://shpaker.github.io/tnk9x/)** · [Download for Windows, macOS and Linux](https://github.com/shpaker/tnk9x/releases)

[![TNK9x gameplay: headlights, tracers and lit walls on a tube TV](.github/screenshot-desktop.png)](https://shpaker.github.io/tnk9x/)

TNK9x is the tanks game we all remember from childhood. Except in our version it suddenly got proper lighting: headlights cut through the night, tracers light up the corridors, explosions throw shadows off brick walls, and the whole picture glows on a curved tube TV.

It's simple: roll through levels, blow up enemy tanks, defend your HQ, grab power-ups and try not to take a shell to the side.

## What's inside

- **35 levels in 7 sets.** Beat levels, earn stars, unlock new sets. Three stars per level: win, lose no lives, beat the clock.
- **Play solo or with a friend** on the same screen.
- **Brick, steel, water, ice and grass:** each terrain type behaves differently. Brick walls even crumble bit by bit, because why not.
- **Four enemy types, each with its own behavior:** they roam, hunt you down, push for your HQ, dodge bullets and blast through walls. And they can grab power-ups too.
- **Eight power-ups:** grenade, extra life, star, shield, enemy freeze, HQ fortification and weapon upgrades, plus a boat to cross the water.
- **Keyboard, gamepad or on-screen controls**, and a mouse in the menus.
- **And yes, there's that awesome lighting that somehow never existed in our memories:** headlights, glowing tracers, lit explosions with real shadows, bloom and a tube TV filter. Miss the old look? F2 switches to classic graphics.
- **Runs everywhere:** desktop, the browser and mobile browsers; English, Russian and Turkish.

## How to play

The goal is simple: destroy all enemy tanks and don't let them destroy your HQ. Lose your HQ or run out of lives, and that's it: the level is lost.

**Keyboard**

|       | Player 1   | Player 2       |
|-------|------------|----------------|
| Move  | W, A, S, D | I, J, K, L     |
| Fire  | G          | ' (apostrophe) |
| Pause | Esc        | Esc            |

**Gamepad.** D-pad or left stick to move, one button to fire, another to pause. The first gamepad controls player one, the second controls player two.

**Touch.** On phones and tablets the D-pad, fire and pause buttons appear on screen by themselves; in two-player mode each player gets their own side of the device.

**Power-ups.** Pick them up right on the field: they can upgrade your tank, give you an extra life, freeze enemies or stir up a little extra chaos.

And you can remap all of this in the settings, because it's not the 1990s anymore.

[![TNK9x in a mobile browser with on-screen controls](.github/screenshot-mobile.png)](https://shpaker.github.io/tnk9x/)

## For developers

Go + [Ebitengine](https://ebitengine.org/), one codebase for desktop and a single WebAssembly build for the web.

**Under the hood:**

- Clean Architecture with depguard-enforced layer boundaries, constructor-only DI from a composition root
- Enemy behavior in Lua behind a domain-typed engine interface; Dijkstra pathfinding and line-of-fire queries exposed to scripts
- Grid-aligned tank movement from [koleya](https://github.com/shpaker/koleya): tanks roll to a 4px node when you let go, so they slip into tank-wide corridors, and slide farther on ice the faster they go
- Kage shader pipeline: lighting with wall shadows on the 256x224 logical screen; the tube TV on the final screen — afterglow, glow, scanlines, mask, curvature and noise — comes from [kinescope](https://github.com/shpaker/kinescope)
- Data-driven content: section-based map and campaign formats ([assets/levels/README.md](assets/levels/README.md)), locales in `assets/locales`, all validated fail-fast on startup
- Platform specifics behind `IPlatformAdapter`, `IRewardAdapter` and `IPurchaseAdapter`, one adapter per build target; portal bridges in `web/<platform>/` ([web/README.md](web/README.md)): pause on suspend, cloud saves, ads, in-game purchases and rating requests where the platform offers them
- Unit tests with a >=70% use-cases coverage gate; CI builds and lints both desktop and js/wasm (GitHub Pages and Yandex Games)
- Automatic `MAJOR.MINOR` versioning by Conventional Commits: changes pile up on main, and a manual run of the Version workflow tags a commit with a green CI, builds the release (desktop and Yandex Games archives) with generated notes and deploys Pages

**Roadmap:**

- HQ: defeat screen, protection mechanics
- UI: score
- Test coverage >80% total, performance profiling

### Building and running

**Requirements:** Go 1.25+, optionally — [Just](https://github.com/casey/just).

```bash
git clone https://github.com/shpaker/tnk9x.git
cd tnk9x
just deps         # Install dependencies (or go mod download)

# Run the game
just run          # or go run cmd/main.go

# Build binary
just build        # binary will be in ./tnk9x

# Web builds (WebAssembly)
just serve-web            # GitHub Pages build at http://localhost:8000
just package-yandex       # Yandex Games archive in _build/ (also in releases and CI runs), upload it to the console
just serve-yandex         # Yandex build via sdk-dev-proxy (Node.js; YANDEX_APP_ID from .env, dev mode without it)

# Checks
just fmt
just lint
just test
```

### Architecture

The project follows **Clean Architecture** principles with clear separation of concerns. Dependencies point inward: outer layers depend on inner layers through interfaces. The composition root (`internal/app`) is the only place that assembles the object graph — all dependencies are injected via constructors, layer boundaries are enforced by depguard.

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
│  │              │  │ Platform (web/none)  │             │
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
│  │ Sound        │  │ Navigation   │                     │
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
