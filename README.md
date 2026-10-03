# tnk9x

Tanks from the 90s — played not the way they really were, but the way we remember them. Plus lighting no cartridge ever had: real-time 2D lights and shadows, bloom and a tube TV look. Built with Go and Ebitengine.

**Play online:** [https://shpaker.github.io/tnk9x/](https://shpaker.github.io/tnk9x/)

[![tnk9x gameplay with normal graphics](.github/screenshot-desktop.png)](https://shpaker.github.io/tnk9x/)

*Stage 1 with normal graphics: headlights, bullet tracers, lit bricks and steel, tube TV filter.*

[![tnk9x running in a mobile browser](.github/screenshot-mobile.png)](https://shpaker.github.io/tnk9x/)

*WebAssembly build in a mobile browser — touch controls are auto-detected.*

## Development Status

**Playable now:**

- Campaign of 35 stages in 7 packs: stage select with a minimap, 3 stars per stage (win, no lives lost, time), results screen, continue with carried lives, a briefing before stage 1
- One or two players: rebindable keyboard, gamepads, touch controls in mobile browsers; mouse and taps in menus
- Battlefield: brick chipped by half-tiles, steel, forest, water and ice; eight bonuses including boat and pistol; enemies race for bonuses too
- Four enemy types scripted in Lua, waves set per map ([assets/levels/README.md](assets/levels/README.md)): roam, hunt the player, push to the HQ, aimed fire, pathfinding, wall demolition, bullet dodging
- Lighting and effects: headlights, glowing tracers, lit explosions and bonuses with shadows from walls, metallic steel and matte brick, bloom, particles, screen shake, tube TV filter on every screen; F2 switches to classic graphics
- Interface: splash with a loading bar, live title scene, pause menu, settings (graphics, fullscreen, volume, language, controls); English, Russian and Turkish with auto-detection
- Desktop and one WebAssembly build for GitHub Pages and Yandex Games: pause on suspend, cloud saves, interstitial ads at most once per 5 minutes of gameplay, rewarded ads
- In-game purchases where the platform sells them: SHOP in the main menu with no ads, stage packs and REVIVE/BOOST tokens that replace watching an ad

**Under the hood:**

- Clean Architecture with depguard-enforced layer boundaries, constructor-only DI from a composition root
- Enemy behavior in Lua behind a domain-typed engine interface; Dijkstra pathfinding and line-of-fire queries exposed to scripts
- Kage shader pipeline: lighting on the logical screen, bloom, phosphor afterglow and CRT on the final screen (see [Rendering Pipeline](#rendering-pipeline))
- Data-driven content: section-based map and campaign formats, locales in `assets/locales`, all validated fail-fast on startup
- Platform specifics behind `IPlatformAdapter`, `IRewardAdapter` and `IPurchaseAdapter`, one adapter per build target; portal bridges in `web/<platform>/` ([web/README.md](web/README.md))
- Unit tests with a >=70% use-cases coverage gate; CI builds and lints both desktop and js/wasm (GitHub Pages and Yandex Games)
- Automatic `MAJOR.MINOR` versioning by Conventional Commits: after a green CI on main CI tags the commit, builds the release (desktop and Yandex Games archives) with generated notes and deploys Pages

### Roadmap
- HQ: defeat screen, protection mechanics
- UI: score
- Test coverage >80% total, performance profiling

## Installation and Running

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
just package-yandex       # Yandex Games archive in _build/ (also in releases and CI runs on main), upload it to the console
just serve-yandex         # Yandex build via sdk-dev-proxy (Node.js; YANDEX_APP_ID from .env, dev mode without it)

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

### Rendering Pipeline

Each frame goes through two passes: the game draws a 256x224 logical screen, then Ebitengine's final-screen hook scales it to the window. Effects (lighting, bloom, CRT) are switched by the graphics setting; with classic graphics both passes skip the shaders.

```
 UPDATE (every tick)                                      internal/states
 ┌──────────────────────────────────────────────────────────────────────┐
 │ gameplay use cases ──RequestEffect──▶ visual events queue            │
 │ VisualEffectsUseCases.Update ──▶ particles · flashes · screen shake  │
 │ LightingUseCases.UpdateHeadlights ──▶ headlight cones follow tanks   │
 └──────────────────────────────────┬───────────────────────────────────┘
                                    │ per-stage state (repositories/game)
 DRAW → logical screen 256x224      ▼          App.Draw → State.Draw
 ┌──────────────────────────────────────────────────────────────────────┐
 │ StageState / DemoScene → StageRendererAdapter.DrawAll                │
 │                                                                      │
 │   drawField → scene buffer (EffectsRendererAdapter.BeginScene)       │
 │     background · ground and surface blocks · HQ · tanks · bullets    │
 │     · bonuses · explosions · particles · forest on top               │
 │   buildSurfaces → material mask                                      │
 │     R opacity · G reflectivity · B sparkle · A dimming               │
 │   buildLights → LightingUseCases.GetLights                           │
 │     headlights · tracers · flashes (up to MaxLights)                 │
 │                                                                      │
 │   lighting.kage (scene + mask + lights, shadows cast by walls)       │
 │     + screen shake ──▶ logical screen                                │
 │                                                                      │
 │ then on top: sidebar HUD, pause / results / settings / briefing      │
 │ menus and overlays draw straight to the logical screen               │
 └──────────────────────────────────┬───────────────────────────────────┘
                                    │ offscreen image
 FINAL SCREEN → window              ▼          App.DrawFinalScreen
 ┌──────────────────────────────────────────────────────────────────────┐
 │ GameRect (touch controls adapter): integer scale, centered           │
 │                                                                      │
 │   bloom.kage     bright pass + horizontal blur → vertical blur       │
 │   phosphor.kage  frame + previous frame (ping-pong) → afterglow      │
 │   crt.kage       afterglow + bloom: curvature, scanlines, aperture   │
 │                  grille, convergence, grain, flicker, line tearing   │
 │                                                    ──▶ window        │
 │ touch controls drawn in the letterbox                                │
 └──────────────────────────────────────────────────────────────────────┘
```

- Shaders are loaded through `IShadersRepository` and compiled fail-fast at startup.
- All offscreen buffers live in `EffectsRendererAdapter` (`internal/adapters/effects`) for the whole app lifetime and are recreated only when the screen size changes.
- Classic graphics: the field is drawn straight to the logical screen and upscaled with nearest-neighbor filtering.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
