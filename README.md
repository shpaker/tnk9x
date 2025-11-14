# tnk25

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg?style=flat-square)](LICENSE)
[![Ebiten](https://img.shields.io/badge/Ebiten-v2.9.1-orange.svg?style=flat-square)](https://ebiten.org/)

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
